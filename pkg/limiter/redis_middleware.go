package limiter

import (
	"errors"
	"fmt"
	"log"
	"net/http"
)

type RedisLimiterMiddleware struct {
	capacity   int32
	expiration int32
	rr         *RedisRepository
}

func NewRedisLimiterMiddleware(capacity, expiration int32, redis *RedisRepository) *RedisLimiterMiddleware {
	return &RedisLimiterMiddleware{capacity: capacity, expiration: expiration, rr: redis}
}

func (lm *RedisLimiterMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestor := r.Header.Get("X-Requestor")

		err := lm.rr.RemoveToken(requestor)
		if errors.Is(err, ErrBucketNotFound) {
			if err := lm.rr.AddBucket(requestor, lm.capacity, lm.expiration); err != nil {
				log.Fatal(err)
			}
			if err := lm.rr.RemoveToken(requestor); err != nil {
				log.Fatal(err)
			}
			next.ServeHTTP(w, r)
		} else if errors.Is(err, ErrBucketEmpty) {
			http.Error(w, fmt.Sprintf("Too many requests sent :(, sorry %s", requestor), http.StatusTooManyRequests)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
