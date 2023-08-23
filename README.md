# rate-limiter
A little rate limiter proof of concept using the [token bucket algorithm](https://www.linkedin.com/advice/0/what-benefits-drawbacks-using-token-bucket-leaky) using golang and [mux framework](https://github.com/gorilla/mux).

## How does it work
In order to keep track of the callers rates, we use a bucket. The bucket is a simple data structure holding tokens. Everytime a call is made to the application endpoint, a token is taken from the bucket. After a certain amount of time, the bucket is refilled to allow new calls to be made. If the total of tokens reaches zero, a 429 status code is returned to the caller, otherwise a normal 200 status code is returned.

For this example:
- A single bucket is created per user
- A user must pass a custom header X-Requestor when doing a call (which allow to bind the bucket to the caller)
- By default, the buckets are refilled every minutes.
- By default, the buckets are deleted after 10 minutes.

```curl
# example of a call
curl localhost:8080 -H X-Requestor:test

```

## Using in memory implementation
The simpliest way to implement the rate limiter is done in memory. The buckets live in memory
at runtime.

```golang
# in the cmd/main.go
# initialize a InMemoryRateLimiter instance

func main() {
	cfg := config.NewConfig()
	limiter := initializeInMemoryRateLimiter(cfg)

	r := mux.NewRouter()
	r.HandleFunc("/", handler.HomeHandler)
	r.Use(limiter.Middleware)
	log.Panic(http.ListenAndServe(":8080", r))
}

go run cmd/main.go
```

## Using redis implementation
Another way to implement the rate limiter is to use redis.
Acting as a simple key store database, it delivers fast read / writes
operations while being single threaded.

```golang
# in the cmd/main.go
# initialize a RedisRateLimiter instance

func main() {
	cfg := config.NewConfig()
	limiter := initializeRedisRateLimiter(cfg)

	r := mux.NewRouter()
	r.HandleFunc("/", handler.HomeHandler)
	r.Use(limiter.Middleware)
	log.Panic(http.ListenAndServe(":8080", r))
}

docker-compose -f scripts/docker-compose.yml up --build
```

If you want you can also have a look on what is going on redis using [redis-commander link](http://localhost:8081/)