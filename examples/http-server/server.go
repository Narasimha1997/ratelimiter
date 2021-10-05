package main

import (
	"log"
	"net/http"
	"time"

	"github.com/Narasimha1997/ratelimiter"
)

func main() {
	perIntervalRecv := 0
	perIntervalAllowed := 0
	nIterations := 0

	duration := time.Second * 5
	requestsAllowed := uint64(100)

	reporter := func() {
		for {
			time.Sleep(duration)
			log.Printf(
				"Iteration: %d, Requests received: %d, Allowed: %d",
				nIterations+1, perIntervalRecv, perIntervalAllowed,
			)

			perIntervalRecv = 0
			perIntervalAllowed = 0
			nIterations++
		}
	}

	// add a middleware:
	limiter := ratelimiter.NewSyncLimiter(requestsAllowed, duration)

	rateLimiterHandler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			allowed, err := limiter.ShouldAllow(1)
			if err != nil {
				log.Fatalln(err)
			}

			perIntervalRecv++

			if allowed {
				perIntervalAllowed++
				next.ServeHTTP(w, r)
			}
		})
	}

	ponger := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Pong!!"))
	}

	// attach the ratelimiter middleware:
	muxServer := http.NewServeMux()
	muxServer.Handle("/", rateLimiterHandler(
		http.HandlerFunc(ponger),
	))

	// start reporter routine:
	go reporter()
	err := http.ListenAndServe(":6000", muxServer)
	if err != nil {
		log.Fatalln(err)
	}
}
