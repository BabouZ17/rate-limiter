package limiter

import "net/http"

type RedisLimiterMiddleware struct {
	capacity int32
	buckets  map[string]*Bucket
	rr       *RedisRepository
}

func NewRedisLimiterMiddleware(capacity int32, redis *RedisRepository) *RedisLimiterMiddleware {
	return &RedisLimiterMiddleware{capacity: capacity, buckets: make(map[string]*Bucket), rr: redis}
}

func (lm *RedisLimiterMiddleware) AddBucket(key string, bucket *Bucket) *Bucket {
	return bucket
}

func (lm *RedisLimiterMiddleware) RefillBuckets() {
}

func (lm *RedisLimiterMiddleware) DeleteBuckets() {
}

func (lm *RedisLimiterMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
