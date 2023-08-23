package limiter

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
)

type InMemoryLimiterMiddleware struct {
	capacity int32
	mu       sync.Mutex
	buckets  map[string]*Bucket
}

func NewInMemoryLimiterMiddleware(capacity int32) *InMemoryLimiterMiddleware {
	return &InMemoryLimiterMiddleware{capacity: capacity, buckets: make(map[string]*Bucket)}
}

func (lm *InMemoryLimiterMiddleware) AddBucket(owner string, capacity int32) *Bucket {
	lm.mu.Lock()
	lm.buckets[owner] = NewBucket(owner, capacity)
	defer lm.mu.Unlock()
	return lm.buckets[owner]
}

func (lm *InMemoryLimiterMiddleware) RefillBuckets() {
	for _, bucket := range lm.buckets {
		lm.mu.Lock()
		bucket.RefillTokens()
		lm.mu.Unlock()
	}
}

func (lm *InMemoryLimiterMiddleware) DeleteBuckets() {
	lm.mu.Lock()
	lm.buckets = make(map[string]*Bucket)
	lm.mu.Unlock()
	log.Println("deleted all buckets")
}

func (lm *InMemoryLimiterMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestor := r.Header.Get("X-Requestor")

		if bucket, found := lm.buckets[requestor]; found {
			err := bucket.RemoveToken()
			if errors.Is(err, ErrBucketEmpty) {
				http.Error(w, fmt.Sprintf("Too many requests sent :(, sorry %s", requestor), http.StatusTooManyRequests)
			} else {
				next.ServeHTTP(w, r)
			}
		} else {
			bucket := lm.AddBucket(requestor, lm.capacity)
			bucket.RemoveToken()
			next.ServeHTTP(w, r)
		}
	})
}
