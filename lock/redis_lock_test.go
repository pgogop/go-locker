package lock

import (
	"fmt"
	"testing"
	"time"
)

func TestLock(t *testing.T) {
	locker := RedisLocker{}
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
}
func TestMLock(t *testing.T) {
	locker := RedisLocker{}
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

	res := locker.MLock(reqs)
	releaseRess := locker.MReleaseLock(reqs)
	fmt.Println(res)
	fmt.Println(releaseRess)
}

func TestMTryLock(t *testing.T) {
	locker, _ := New("redis")

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
}

func TestTryLock(t *testing.T) {
	locker, _ := New("redis")

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
