package main

import (
	"log"
	"net/http"

	"github.com/BabouZ17/rate-limiter/pkg/config"
	"github.com/BabouZ17/rate-limiter/pkg/handler"
	"github.com/BabouZ17/rate-limiter/pkg/limiter"
	"github.com/gorilla/mux"
)

func main() {
	config := config.NewConfig()

	// Using in Memory Rate Limiter
	// limiter := limiter.NewInMemoryLimiterMiddleware(config.RateLimiterConfig.Capacity)

	// cronScheduler := cron.New()
	// cronScheduler.AddFunc(config.RateLimiterConfig.RefreshTime, func() { limiter.RefillBuckets() })
	// cronScheduler.AddFunc(config.RateLimiterConfig.FlushBucketsTime, func() { limiter.DeleteBuckets() })
	// cronScheduler.Start()

	// Using Redis Rate Limiter
	rr := limiter.NewRedisRepository(config)
	limiter := limiter.NewRedisLimiterMiddleware(config.RateLimiterConfig.Capacity, rr)

	rr.AddBucket("bob", 10)
	rr.RemoveBuckets()

	r := mux.NewRouter()
	r.HandleFunc("/", handler.HomeHandler)
	r.Use(limiter.Middleware)
	log.Fatal(http.ListenAndServe(":8080", r))
}
