package distributedlock

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

type DistributedLock struct {
	client *redis.Client
}

func New(client *redis.Client) *DistributedLock {
	return &DistributedLock{client: client}
}

func (dl *DistributedLock) AcquireLock(key string, expiration time.Duration) (bool, error) {
	ctx := context.Background()
	// 尝试获取锁
	success, err := dl.client.SetNX(ctx, key, "locked", expiration).Result()
	if err != nil {
		return false, err
	}

	return success, nil
}

func (dl *DistributedLock) ReleaseLock(key string) error {
	ctx := context.Background()
	// 释放锁
	_, err := dl.client.Del(ctx, key).Result()
	return err
}
