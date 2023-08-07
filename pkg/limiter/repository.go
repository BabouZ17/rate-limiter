package limiter

import (
	"fmt"
	"log"

	"github.com/BabouZ17/rate-limiter/pkg/config"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
)

const BUCKETS_SET_KEY = "buckets"

type RedisRepository struct {
	redis *redis.Client
}

func (rr *RedisRepository) AddBucket(owner string, capacity int32) {
	// Store new key in the Set
	id := uuid.New().String()
	_, err := rr.redis.SAdd(BUCKETS_SET_KEY, id).Result()
	if err != nil {
		fmt.Println(err)
		log.Println("failed to save the key in set")
	}

	// Store new hash
	//bucket := NewBucket(owner, capacity)

}

func (rr *RedisRepository) RemoveBuckets() {
	// to be done
}

func NewRedisRepository(c config.Config) *RedisRepository {
	return &RedisRepository{
		redis: redis.NewClient(
			&redis.Options{
				Addr:     c.RedisConfig.Address,
				Password: c.GetEnvValue(c.RedisConfig.Password),
				DB:       c.RedisConfig.Database,
			},
		),
	}
}
