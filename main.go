package main

import (
	"time"

	"github.com/Narasimha1997/ratelimit/core"
)

func main() {
	limiter := core.NewLimiter(100, time.Duration(1*time.Second))
	for {
		println(limiter.ShouldAllow(1))
		time.Sleep(2 * time.Millisecond)
	}
}
