package main

import (
	"fmt"
	"time"

	"github.com/Narasimha1997/ratelimit/core"
)

func main() {

	windowDuration := time.Duration(10 * time.Second)

	limiter := core.NewLimiter(300, windowDuration)
	counter := 0

	printCounter := func() {
		for {
			time.Sleep(windowDuration)
			fmt.Printf("Rate allowed: %d\n", counter)
			counter = 0
		}
	}

	go printCounter()

	for {
		if limiter.ShouldAllow(1) {
			counter++
		}

		time.Sleep(2 * time.Millisecond)
	}
}
