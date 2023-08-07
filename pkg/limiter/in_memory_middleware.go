package limiter

import (
	"net/http"
)

type InMemoryLimiterMiddleware struct {
	capacity int32
	buckets  map[string]*Bucket
}

func NewInMemoryLimiterMiddleware(capacity int32) *InMemoryLimiterMiddleware {
	return &InMemoryLimiterMiddleware{capacity: capacity, buckets: make(map[string]*Bucket)}
}

func (lm *InMemoryLimiterMiddleware) AddBucket(key string, bucket *Bucket) *Bucket {
	lm.buckets[key] = bucket
	return bucket
}

func (lm *InMemoryLimiterMiddleware) RefillBuckets() {
	for _, bucket := range lm.buckets {
		bucket.RefillTokens()
	}
}

func (lm *InMemoryLimiterMiddleware) DeleteBuckets() {
	lm.buckets = make(map[string]*Bucket)
}

func (lm *InMemoryLimiterMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestor_key := r.Header.Get("X-Requestor")

		if bucket, found := lm.buckets[requestor_key]; found {
			err := bucket.RemoveToken()
			if err != nil {
				http.Error(w, "Too many requests sent :(", http.StatusTooManyRequests)
			} else {
				next.ServeHTTP(w, r)
			}
		} else {
			bucket := lm.AddBucket(requestor_key, NewBucket(requestor_key, lm.capacity))
			bucket.RemoveToken()
			next.ServeHTTP(w, r)
		}
	})
}
