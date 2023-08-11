package main

import (
	"log"
	"net/http"

	"github.com/BabouZ17/rate-limiter/pkg/config"
	"github.com/BabouZ17/rate-limiter/pkg/handler"
	"github.com/BabouZ17/rate-limiter/pkg/limiter"
	"github.com/gorilla/mux"
	"github.com/robfig/cron"
)

func initializeInMemoryRateLimiter(cfg config.Config) *limiter.InMemoryLimiterMiddleware {
	limiter := limiter.NewInMemoryLimiterMiddleware(cfg.RateLimiterConfig.Capacity)

	cronScheduler := cron.New()
	cronScheduler.AddFunc(cfg.RateLimiterConfig.TokensRefreshTime, func() { limiter.RefillBuckets() })
	cronScheduler.AddFunc(cfg.RateLimiterConfig.FlushBucketsTime, func() { limiter.DeleteBuckets() })
	cronScheduler.Start()

	return limiter
}

func initializeRedisLimiter(cfg config.Config) *limiter.RedisLimiterMiddleware {
	rr := limiter.NewRedisRepository(cfg)
	rr.DeleteBuckets()
	limiter := limiter.NewRedisLimiterMiddleware(cfg.RateLimiterConfig.Capacity, cfg.RateLimiterConfig.Expiration, rr)

	cronScheduler := cron.New()
	cronScheduler.AddFunc(cfg.RateLimiterConfig.TokensRefreshTime, func() { rr.RefillBuckets() })
	cronScheduler.AddFunc(cfg.RateLimiterConfig.FlushBucketsTime, func() { rr.DeleteBuckets() })
	cronScheduler.Start()

	return limiter
}

func main() {
	cfg := config.NewConfig()
	limiter := initializeRedisLimiter(cfg)

	r := mux.NewRouter()
	r.HandleFunc("/", handler.HomeHandler)
	r.Use(limiter.Middleware)
	log.Fatal(http.ListenAndServe(":8080", r))
}
