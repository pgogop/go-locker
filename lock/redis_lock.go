package lock

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"time"
)

type RedisLocker struct {
	pool *redis.Pool
}

type lockChan struct {
	val bool
	err error
}

func (l *RedisLocker) SetPool(pool *redis.Pool) {
	l.pool = pool
}

func (l *RedisLocker) lock(conn redis.Conn, key string, reqId string, expireTime uint) (bool, error) {
	script := redis.NewScript(1, l.luaLockScript())
	val, err := redis.String(script.Do(conn, key, reqId, expireTime))
	if err != nil && err != redis.ErrNil {
		return false, err
	}
	return "OK" == val, nil
}

func (l *RedisLocker) Lock(req LockReq) (bool, error) {
	conn := l.pool.Get()
	defer conn.Close()
	return l.lock(conn, req.Key, req.Value, req.ExpireTime)
}

func (l *RedisLocker) MLock(reqs []LockReq) (map[string]bool, error) {
	conn := l.pool.Get()
	defer conn.Close()
	var (
		lockRes = make(map[string]bool)
		lockErr error
	)
	defer func() {
		if lockErr != nil && len(lockRes) > 0 {
			var releaseReqs []LockReq
			for _, req := range reqs {
				if res, ok := lockRes[req.Key]; ok && res {
					releaseReqs = append(releaseReqs, req)
				}
			}
			l.MReleaseLock(releaseReqs)
		}
	}()

	for _, req := range reqs {
		if _, ok := lockRes[req.Key]; ok {
			continue
		}
		res, err := l.lock(conn, req.Key, req.Value, req.ExpireTime)
		if err != nil {
			lockErr = err
			break
		}
		lockRes[req.Key] = res
	}

	if lockErr != nil {
		return nil, lockErr
	}
	return lockRes, nil
}

func (l *RedisLocker) luaLockScript() string {
	return `return redis.call('set',KEYS[1],ARGV[1],'EX',ARGV[2],'NX')`
}

func (l *RedisLocker) ReleaseLock(req LockReq) (bool, error) {
	conn := l.pool.Get()
	defer conn.Close()
	script := redis.NewScript(1, l.luaReleaseLockScript())
	val, err := redis.Int(script.Do(conn, req.Key, req.Value))
	if err != nil {
		return false, err
	}

	return 1 == val, nil
}

func (l *RedisLocker) MReleaseLock(reqs []LockReq) map[string]bool {
	conn := l.pool.Get()
	defer conn.Close()
	var (
		res = make(map[string]bool)
	)
	script := redis.NewScript(1, l.luaReleaseLockScript())
	for _, req := range reqs {
		if _, ok := res[req.Key]; ok {
			continue
		}

		val, err := redis.Int(script.Do(conn, req.Key, req.Value))
		if err != nil {
			res[req.Key] = false
			continue
		}

		res[req.Key] = 1 == val
	}

	return res
}

func (l *RedisLocker) luaReleaseLockScript() string {
	return `if redis.call('get',KEYS[1]) == ARGV[1] then 
return redis.call('del',KEYS[1]) 
else 
return 0 
end`
}

func (l *RedisLocker) TryLock(req LockReq, maxWaitTime time.Duration) (bool, error) {
	conn := l.pool.Get()
	defer conn.Close()
	c1 := make(chan lockChan, 1)
	go func() {
		for true {
			res, err := l.lock(conn, req.Key, req.Value, req.ExpireTime)
			if err != nil {
				c1 <- lockChan{
					val: res,
					err: err,
				}
				break
			}

			if res {
				c1 <- lockChan{
					val: res,
					err: nil,
				}
				break
			}

			time.Sleep(10 * time.Millisecond)
		}
	}()

	select {
	case res := <-c1:
		return res.val, res.err
	case <-time.After(maxWaitTime):
		return false, GetLockTimeout
	}
}

func (l *RedisLocker) MTryLock(reqs []LockReq, maxWaitTime time.Duration) map[string]bool {
	conn := l.pool.Get()
	defer conn.Close()
	var lockRes = make(map[string]bool)
	c1 := make(chan bool, 1)
	go func() {
		over := false
		count := 0
		totalLen := len(reqs)
	out:
		for !over {
			clone := reqs
			reqs = []LockReq{}
			if count == totalLen {
				over = true
				c1 <- true
				continue out
			}

			for _, req := range clone {
				if v, ok := lockRes[req.Key]; ok && v {
					count++
					continue
				}

				res, err := l.lock(conn, req.Key, req.Value, req.ExpireTime)
				if err == nil && res {
					lockRes[req.Key] = true
					count++
				} else {
					reqs = append(reqs, req)
				}
			}

			time.Sleep(10 * time.Millisecond)
		}
	}()

	select {
	case <-c1:
		return l.fillResult(reqs, lockRes)
	case <-time.After(maxWaitTime):
		fmt.Println("get lock timeout")
		return l.fillResult(reqs, lockRes)
	}
}

func (l *RedisLocker) fillResult(reqs []LockReq, res map[string]bool) map[string]bool {
	for _, req := range reqs {
		if _, ok := res[req.Key]; !ok {
			res[req.Key] = false
		}
	}

	return res
}
