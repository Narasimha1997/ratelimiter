package core

import (
	"fmt"
	"testing"
	"time"
)

func TestLimiterAccuracy(t *testing.T) {

	nRuns := 60
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
	hasFinished := false

	rateLimiterRoutine := func() {
		limiter := NewLimiter(limit, duration)
		for i := 0; i < (nRuns * 500); i++ {
			if limiter.ShouldAllow(1) {
				count++
			}

			time.Sleep(2 * time.Millisecond)
		}

		hasFinished = true
	}

	go rateLimiterRoutine()

	fmt.Printf("Running accuracy measurement check for %d trials", nRuns)

	for !hasFinished {
		time.Sleep(duration)
		if (limit-allowanceRange) <= count && count >= (limit+allowanceRange) {
			count = 0
		} else {
			t.Fatalf(
				"Accuracy test failed, expected results to be in +/- %d error range, but got %d",
				allowanceRange, count,
			)
		}
	}
}
