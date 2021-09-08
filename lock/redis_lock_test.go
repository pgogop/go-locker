package lock

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func getRedisPool() (*redis.Pool, error) {
	db := redis.DialDatabase(0)
	pwd := redis.DialPassword("")

	pool := &redis.Pool{
		MaxIdle:   5,
		MaxActive: 200,
		Dial: func() (redis.Conn, error) {
			rc, err := redis.Dial("tcp", "localhost:16379", db, pwd)
			if err != nil {
				return nil, err
			}
			return rc, nil
		},
		IdleTimeout: 29 * time.Second,
		Wait:        true,
	}

	return pool, nil
}

func TestLock(t *testing.T) {
	pool, _ := getRedisPool()
	locker := RedisLocker{
		pool: pool,
	}
	//ctx := context.Background()
	req := LockReq{
		Key:        "l-key1",
		Value:      "tttt",
		ExpireTime: 20,
	}

	res, _ := locker.Lock(req)
	res2, err := locker.Lock(req)
	releaseRess, _ := locker.ReleaseLock(req)
	fmt.Println(res, res2, err)
	fmt.Println(releaseRess)

	assert.Equal(t, nil, err)
}

func TestMLock(t *testing.T) {
	pool, _ := getRedisPool()
	locker := RedisLocker{
		pool: pool,
	}
	//ctx := context.Background()
	reqs := []LockReq{
		{
			Key:        "l-key1",
			Value:      "tttt",
			ExpireTime: 20,
		},
		{
			Key:        "l-key1",
			Value:      "tttt",
			ExpireTime: 20,
		},
		{
			Key:        "l-key2",
			Value:      "tttt2",
			ExpireTime: 20,
		},
	}

	res, err := locker.MLock(reqs)
	releaseRess := locker.MReleaseLock(reqs)
	fmt.Println(res, err)
	fmt.Println(releaseRess)

	assert.Equal(t, nil, err)
}

func TestMTryLock(t *testing.T) {
	pool, _ := getRedisPool()
	locker := RedisLocker{
		pool: pool,
	}

	reqs := []LockReq{
		{
			Key:        "l-key1",
			Value:      "tttt",
			ExpireTime: 10,
		},
		{
			Key:        "l-key1",
			Value:      "tttt",
			ExpireTime: 10,
		},
		{
			Key:        "l-key1.1",
			Value:      "tttt",
			ExpireTime: 3,
		},
		{
			Key:        "l-key2",
			Value:      "tttt2",
			ExpireTime: 20,
		},
		{
			Key:        "l-key3",
			Value:      "tttt2",
			ExpireTime: 20,
		},
	}

	//ctx := context.Background()
	res := locker.MTryLock(reqs, 2*time.Second)
	//releaseRess2 := locker.MReleaseLock(reqs)
	fmt.Println(res)
	//fmt.Println(releaseRess2)
	//assert.Equal(t, nil,err)
}

func TestTryLock(t *testing.T) {
	pool, _ := getRedisPool()
	locker := RedisLocker{
		pool: pool,
	}

	req := LockReq{
		Key:        "l-key1",
		Value:      "tttt",
		ExpireTime: 1,
	}

	//ctx := context.Background()
	res, err1 := locker.TryLock(req, 2*time.Second)
	//releaseRess, _ := locker.ReleaseLock( "tttt", "l-key1")
	res2, err2 := locker.TryLock(req, 4*time.Second)
	releaseRess2, _ := locker.ReleaseLock(req)
	fmt.Println(res, res2)
	fmt.Println(err1, err2)
	//fmt.Println(releaseRess)
	fmt.Println(releaseRess2)
}
