package lock

import (
	"errors"
	"time"
)

const (
	DriverRedis = "redis"
)

var (
	GetLockTimeout = errors.New("get lock timeout")
)

type Locker interface {
	Lock(req LockReq) (bool, error)
	MLock(reqs []LockReq) (map[string]bool, error)
	TryLock(req LockReq, maxWaitTime time.Duration) (bool, error)
	MTryLock(reqs []LockReq, maxWaitTime time.Duration) map[string]bool
	ReleaseLock(req LockReq) (bool, error)
	MReleaseLock(reqs []LockReq) map[string]bool
}

type LockReq struct {
	Key        string
	Value      string
	ExpireTime uint
}

//func New(driver string) (Locker, error) {
//	switch driver {
//	case DriverRedis:
//		return &RedisLocker{}, nil
//	default:
//		return nil, errors.New("unsupported driver : " + driver)
//	}
//}
