package limiter

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"strconv"
	"time"

	"github.com/BabouZ17/rate-limiter/pkg/config"
	"github.com/redis/go-redis/v9"
)

const BUCKETS_SET_KEY = "buckets"

var ctx = context.Background()
var ErrBucketNotFound = errors.New("Bucket not found")

type ErrRedisRepository struct {
	Msg string
	Err error
}

func NewErrRedisRepository(msg string, err error) *ErrRedisRepository {
	return &ErrRedisRepository{Msg: msg, Err: err}
}

func (e *ErrRedisRepository) Error() string {
	log.Printf("Redis operation failed, reason: %s", e.Err)
	return e.Err.Error()
}

type RedisRepository struct {
	redis *redis.Client
}

func NewRedisRepository(c config.Config) *RedisRepository {
	return &RedisRepository{
		redis: redis.NewClient(
			&redis.Options{
				Addr: c.RedisConfig.Address,
				DB:   c.RedisConfig.Database,
			},
		),
	}
}

func (rr *RedisRepository) AddBucket(owner string, capacity int32, expiration int32) error {
	// Store new key in the Sorted Set with expiration time time.Now() + expiration

	expiration_delta := float64(expiration)
	expiration_time := expiration_delta + float64(time.Now().Unix())

	_, err := rr.redis.ZAdd(ctx, BUCKETS_SET_KEY, redis.Z{Score: expiration_time, Member: owner}).Result()
	if err != nil {
		return NewErrRedisRepository(fmt.Sprintf("failed to save the key %s in set", owner), err)
	}

	// Store the bucket data into a hash
	if _, err := rr.redis.Pipelined(ctx, func(redis redis.Pipeliner) error {
		redis.HSet(ctx, owner, "owner", owner)
		redis.HSet(ctx, owner, "capacity", capacity)
		redis.HSet(ctx, owner, "count", capacity)
		return nil
	}); err != nil {
		return NewErrRedisRepository(fmt.Sprintf("could not add the new bucket for %s", owner), err)
	}
	return nil
}

func (rr *RedisRepository) RemoveToken(owner string) error {
	bucket, err := rr.GetBucket(owner)
	if err != nil {
		return err
	}

	if bucket.Count == 0 {
		return NewErrRedisRepository(fmt.Sprintf("bucket %s has no more tokens", owner), ErrBucketEmpty)
	} else {
		bucket.Count--
		_, err = rr.redis.HSet(ctx, owner, "count", bucket.Count).Result()
		if err != nil {
			return NewErrRedisRepository(fmt.Sprintf("failed to update count of bucket %s", owner), err)
		}
		return nil
	}
}

func (rr *RedisRepository) GetBucket(owner string) (*Bucket, error) {
	var bucket Bucket
	if err := rr.redis.HGetAll(ctx, owner).Scan(&bucket); err != nil {
		return nil, NewErrRedisRepository(fmt.Sprintf("could not load the bucket %s", owner), err)
	}

	if bucket.Owner == "" {
		return nil, NewErrRedisRepository(fmt.Sprintf("could not find the bucket %s", owner), ErrBucketNotFound)
	}
	return &bucket, nil
}

func (rr *RedisRepository) RefillBucket(owner string) error {
	bucket, err := rr.GetBucket(owner)
	if err != nil {
		return err
	}
	_, err = rr.redis.HSet(ctx, owner, "count", bucket.Capacity).Result()
	if err != nil {
		return NewErrRedisRepository(fmt.Sprintf("could not refill bucket %s", owner), err)
	}
	return nil
}

func (rr *RedisRepository) RefillBuckets() error {
	members, err := rr.redis.ZRange(ctx, BUCKETS_SET_KEY, 0, math.MaxInt64).Result()
	if err != nil {
		return NewErrRedisRepository("failed to retrieve buckets", err)
	}
	for _, member := range members {
		if err := rr.RefillBucket(member); err != nil {
			return err
		}
	}
	return nil
}

func (rr *RedisRepository) DeleteBucket(owner string) error {
	if _, err := rr.redis.Del(ctx, owner).Result(); err != nil {
		return NewErrRedisRepository(fmt.Sprintf("failed to delete bucket %s hash", owner), err)
	}
	if _, err := rr.redis.ZRem(ctx, BUCKETS_SET_KEY, owner).Result(); err != nil {
		return NewErrRedisRepository(fmt.Sprintf("failed to remove key %s from set", owner), err)
	}
	return nil
}

func (rr *RedisRepository) DeleteBuckets() error {
	min := "0"
	max := strconv.Itoa(int(time.Now().Unix()))
	members, err := rr.redis.ZRangeByScore(ctx, BUCKETS_SET_KEY, &redis.ZRangeBy{Min: min, Max: max}).Result()
	if err != nil {
		return NewErrRedisRepository("failed to retrieve buckets keys", err)
	}
	for _, member := range members {
		if err := rr.DeleteBucket(member); err != nil {
			return err
		}
	}
	return nil
}
