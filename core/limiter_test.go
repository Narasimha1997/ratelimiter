package core

import (
	"fmt"
	"testing"
	"time"
)

func TestLimiterAccuracy(t *testing.T) {

	nRuns := 20
	var count uint64 = 0

	// Time duration of the window.
	duration := time.Second * 1

	// 100 tasks must be allowed to execute
	// for every `duration` interval.
	var limit uint64 = 100

	// test with accuracy +/- 3, modify this variable to
	// test accuracy for various error offsets, 0 is the most
	// ideal case.
	var allowanceRange uint64 = 3

	// will be set to true once the go routine completes all `nRuns`

	limiter := NewLimiter(limit, duration)

	for i := 0; i < nRuns; i++ {
		count = 0
		nTicks := 0
		for range time.Tick(time.Millisecond * 2) {

			if limiter.ShouldAllow(1) {
				count++
			}

			nTicks++

			if nTicks%500 == 0 {
				break
			}
		}

		if (limit-allowanceRange) <= count && count <= (limit+allowanceRange) {
			fmt.Printf(
				"Iteration %d, Allowed tasks: %d, passed rate limiting accuracy test.\n",
				i+1, count,
			)
			count = 0
		} else {
			t.Fatalf(
				"Accuracy test failed, expected results to be in +/- %d error range, but got %d",
				allowanceRange, count,
			)
		}
	}

	fmt.Printf("Running accuracy measurement check for %d trials", nRuns)
}
