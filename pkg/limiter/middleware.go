package limiter

import (
	"net/http"

	"github.com/google/uuid"
)

type Requestor struct {
	id     string
	bucket *Bucket
}

func NewRequestor() *Requestor {
	return &Requestor{id: uuid.New().String(), bucket: NewBucket(10)}
}

type LimiterMiddleware struct {
	requestors map[string]Requestor
}

func NewLimiterMiddleware(requestors map[string]Requestor) *LimiterMiddleware {
	return &LimiterMiddleware{requestors: requestors}
}

func (lm *LimiterMiddleware) AddRequestor(key string, requestor Requestor) {
	lm.requestors[key] = requestor
}

func (lm *LimiterMiddleware) RemoveRequestor(key string) {
	delete(lm.requestors, key)
}

func (lm *LimiterMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestor_key := r.Header.Get("X-Requestor")

		if requestor, found := lm.requestors[requestor_key]; found {
			err := requestor.bucket.RemoveToken()
			if err != nil {
				http.Error(w, "Too many requests sent :(", http.StatusTooManyRequests)
			} else {
				next.ServeHTTP(w, r)
			}
		} else {
			http.Error(w, "Requestor not found :(", http.StatusBadRequest)
		}
	})
}
