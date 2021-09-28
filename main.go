package main

import (
	"time"

	"github.com/Narasimha1997/ratelimit/core"
)

func main() {
	limiter := core.NewLimiter(10, time.Duration(10*time.Second))
	for {
		println(limiter.ShouldAllow(4))
		time.Sleep(3 * time.Second)
	}
}
