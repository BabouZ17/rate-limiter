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
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const BUCKETS_SET_KEY = "buckets"

var ctx = context.Background()
var ErrBucketNotFound = errors.New("Bucket not found")

type RedisRepository struct {
	redis *redis.Client
}

func (rr *RedisRepository) AddBucket(owner string, capacity int32, expiration int32) error {
	// Store new key in the Sorted Set with expiration time time.Now() + expiration
	id := uuid.New().String()

	expiration_delta := float64(expiration)
	expiration_time := expiration_delta + float64(time.Now().Unix())

	_, err := rr.redis.ZAdd(ctx, BUCKETS_SET_KEY, redis.Z{Score: expiration_time, Member: owner}).Result()
	if err != nil {
		return errors.New("failed to save the key in set")
	}

	// Store the bucket data into a hash
	if _, err := rr.redis.Pipelined(ctx, func(redis redis.Pipeliner) error {
		redis.HSet(ctx, owner, "id", id)
		redis.HSet(ctx, owner, "owner", owner)
		redis.HSet(ctx, owner, "capacity", capacity)
		redis.HSet(ctx, owner, "count", capacity)
		return nil
	}); err != nil {
		return errors.New("could not save the bucket")
	}
	return nil
}

func (rr *RedisRepository) RemoveToken(owner string) error {
	bucket, err := rr.GetBucket(owner)
	if err != nil {
		return err
	}

	if bucket.Count == 0 {
		msg, _ := fmt.Printf("Bucket %s belonging to %s has no more tokens", bucket.Id, bucket.Owner)
		log.Println(msg)
		return ErrBucketEmpty
	} else {
		bucket.Count--
		_, err = rr.redis.HSet(ctx, owner, "count", bucket.Count).Result()
		if err != nil {
			return errors.New("failed to update count of bucket")
		}
		return nil
	}
}

func (rr *RedisRepository) GetBucket(owner string) (*Bucket, error) {
	var bucket Bucket
	if err := rr.redis.HGetAll(ctx, owner).Scan(&bucket); err != nil {
		return nil, err
	}

	if bucket.Id == "" {
		return nil, ErrBucketNotFound
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
		return errors.New("could not refill bucket")
	}
	return nil
}

func (rr *RedisRepository) RefillBuckets() error {
	members, err := rr.redis.ZRange(ctx, BUCKETS_SET_KEY, 0, math.MaxInt64).Result()
	if err != nil {
		return errors.New("failed to retrieve buckets")
	}
	for _, member := range members {
		rr.RefillBucket(member)
	}
	return nil
}

func (rr *RedisRepository) DeleteBucket(owner string) error {
	if _, err := rr.redis.Del(ctx, owner).Result(); err != nil {
		return errors.New("failed to delete bucket hash")
	}
	if _, err := rr.redis.ZRem(ctx, BUCKETS_SET_KEY, owner).Result(); err != nil {
		return errors.New("failed to remove key from set")
	}
	return nil
}

func (rr *RedisRepository) DeleteBuckets() error {
	min := "0"
	max := strconv.Itoa(int(time.Now().Unix()))
	members, err := rr.redis.ZRangeByScore(ctx, BUCKETS_SET_KEY, &redis.ZRangeBy{Min: min, Max: max}).Result()
	if err != nil {
		return errors.New("failed to delete buckets")
	}
	for _, member := range members {
		rr.DeleteBucket(member)
	}
	return nil
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
