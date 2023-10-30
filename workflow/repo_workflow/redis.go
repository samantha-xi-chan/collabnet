package repo_workflow

import (
	"collab-net-v2/util/distributedlock"
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"log"
	"time"
)

//var err error

type RedisManager struct {
	Address string
	Client  *redis.Client
	MaxSize int64
}

var redisManager RedisManager

func GetRedisMgr() (x *RedisManager) {
	return &redisManager
}

func InitRedis(ctx context.Context, dsn string, maxSize int64, slowThresholdMs int) (e error) {
	redisManager = RedisManager{
		Address: dsn,
		Client:  nil,
		MaxSize: maxSize,
	}

	return redisManager.Init(ctx)
}

func (mgr *RedisManager) AcquireEnQueue(ctx context.Context, lockKey string, myfun func(string) int) (x bool, err error) {
	// 创建分布式锁
	lock := distributedlock.New(mgr.Client)

	// 锁的过期时间
	expiration := 5 * time.Second

	// 尝试获取锁
	success, err := lock.AcquireLock(lockKey, expiration)
	if err != nil {
		fmt.Println("Error acquiring lock:", err)
		return
	}

	if success {
		// 成功获取锁，执行临界区代码
		fmt.Println("Lock acquired!")
		// ... 执行临界区代码 ...
		myfun(lockKey)

		// 释放锁
		err := lock.ReleaseLock(lockKey)
		if err != nil {
			fmt.Println("Error releasing lock:", err)
		}
	} else {
		// 未能获取锁
		fmt.Println("Failed to acquire lock.")
	}

	return success, nil
}

func (mgr *RedisManager) Init(ctx context.Context) (e error) {

	mgr.Client = redis.NewClient(&redis.Options{
		Addr:     mgr.Address, // Replace with the address of your Redis server
		Password: "",          // No password if not set
		DB:       0,           // Use default DB
	})

	// Ping the Redis server to check the connection
	pong, err := mgr.Client.Ping(ctx).Result()
	if err != nil {
		fmt.Println("Error pinging Redis server:", err)
		return
	}
	fmt.Println("Redis server responded:", pong)

	return nil
}

func (mgr *RedisManager) NewLog(ctx context.Context, trim bool, key string, val string) (e error) {
	result := mgr.Client.LPush(ctx, key, val)
	if err := result.Err(); err != nil {
		return errors.Wrap(err, fmt.Sprint("mgr.Client.LPush: key=", key))
	} else {
		log.Println("LPush ok, len(val)=", len(val))
	}

	if trim {
		_, e = mgr.Client.LTrim(ctx, key, 0, mgr.MaxSize-1).Result()
		if e != nil {
			log.Println("NewLog LTrim e: ", e)
			return errors.Wrap(e, ".LTrim(ctx, key ：")
		}
	}

	return nil
}

func (mgr *RedisManager) NewLogMulti(ctx context.Context, trim bool, key string, vals []string) (e error) {

	mgr.Client.LPush(ctx, key, vals)

	if trim {
		_, e = mgr.Client.LTrim(ctx, key, 0, mgr.MaxSize-1).Result()
		if e != nil {
			log.Println("NewLogMulti LTrim e: ", e)
		}
	}

	return nil
}

func (mgr *RedisManager) Traversal(ctx context.Context, trim bool, key string, startFromRear bool) (e error) {

	if trim {
		_, e = mgr.Client.LTrim(ctx, key, 0, mgr.MaxSize-1).Result()
		if e != nil {
			log.Println("Traversal LTrim e: ", e)
			return e
		}
	}

	elements, err := mgr.Client.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		log.Println("Error:", err)
		return
	}

	if startFromRear {
		log.Println("List elements:")
		for _, element := range elements {
			log.Println(element)
		}
	} else {
		for i := len(elements) - 1; i >= 0; i-- {
			fmt.Println(elements[i])
		}
	}

	return nil
}

func (mgr *RedisManager) Exists(ctx context.Context, key string) (bool, error) {

	exists, err := mgr.Client.Exists(context.Background(), key).Result()
	if err != nil {
		fmt.Println("Error:", err)
		return false, errors.Wrap(err, ".Exists: ")
	}

	if exists == 1 {
		return true, nil
	} else {
		return false, nil
	}
}

func (mgr *RedisManager) Query(ctx context.Context, trim bool, key string, timeAsc bool, pageId int, pageSize int) (elem []string, total int64, e error) {

	if pageId == 0 || pageSize == 0 {
		return nil, 0, errors.New("TMD")
	}

	total, err := mgr.Client.LLen(ctx, key).Result()
	if err != nil {
		fmt.Println("获取列表长度时出错:", err)
		return
	}

	if pageSize > 1000 {
		// warning, return a certain warning error

		return nil, total, nil
	}

	maxPageId := total/int64(pageSize) + 1

	if int64(pageId) > maxPageId {
		return nil, 0, errors.New("TMD")
	}

	// 时间 降序
	start := int64(pageSize * (pageId - 1))
	stop := int64(pageSize * pageId)

	if timeAsc { // 时间升序 = 从队头开始
		tmpStart := start
		start = total - stop
		stop = total - tmpStart
	}

	elements, err := mgr.Client.LRange(ctx, key, start, stop).Result()
	if err != nil {
		log.Println("Error:", err)
		return
	}

	log.Println("List elements :")
	for _, element := range elements {
		log.Println(element)
	}

	if timeAsc {
		reverseArray(elements)
	}

	log.Println("List elements result:")
	for _, element := range elements {
		log.Println(element)
	}

	return elements, total, nil
}

func reverseArray(arr []string) {
	length := len(arr)
	for i := 0; i < length/2; i++ {
		arr[i], arr[length-i-1] = arr[length-i-1], arr[i]
	}
}
