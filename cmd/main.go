package main

import (
	"log"
	"net/http"

	"github.com/BabouZ17/rate-limiter/pkg/handler"
	"github.com/BabouZ17/rate-limiter/pkg/limiter"
	"github.com/gorilla/mux"
)

func main() {
	requestors := map[string]limiter.Requestor{
		"dim":  *limiter.NewRequestor(),
		"liam": *limiter.NewRequestor(),
	}
	limiter := limiter.NewLimiterMiddleware(requestors)

	r := mux.NewRouter()
	r.HandleFunc("/", handler.HomeHandler)
	r.Use(limiter.Middleware)
	log.Fatal(http.ListenAndServe(":8080", r))
}
