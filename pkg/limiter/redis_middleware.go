package limiter

import (
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
		if e, ok := err.(*ErrRedisRepository); ok && e.Err == ErrBucketNotFound {
			if err := lm.rr.AddBucket(requestor, lm.capacity, lm.expiration); err != nil {
				log.Panic(err)
			}
			if err := lm.rr.RemoveToken(requestor); err != nil {
				log.Panic(err)
			}
			next.ServeHTTP(w, r)
		} else if e, ok = err.(*ErrRedisRepository); ok && e.Err == ErrBucketEmpty {
			http.Error(w, fmt.Sprintf("Too many requests sent :(, sorry %s", requestor), http.StatusTooManyRequests)
		} else if err != nil {
			http.Error(w, fmt.Sprintf("Internal error, %s", err), http.StatusInternalServerError)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
