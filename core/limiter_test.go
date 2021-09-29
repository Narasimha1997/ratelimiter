package core

import (
	"fmt"
	"sync"
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
}

func TestConcurrentLimiterAccuracy(t *testing.T) {
	nRuns := 20
	duration := time.Second * 1

	// 100 tasks must be allowed to execute
	// for every `duration` interval.
	var limit uint64 = 100

	// create a limiter, that is shared across go routines:
	sharedLimiter := NewLimiter(limit, duration)

	// launch N go-routines:
	nRoutines := 4

	// test with accuracy +/- 3, modify this variable to
	// test accuracy for various error offsets, 0 is the most
	// ideal case.
	var allowanceRange uint64 = 3

	counterSlice := make([]uint64, nRoutines)

	routine := func(idx int, wg *sync.WaitGroup) {

		defer wg.Done()

		// no need of mutex locking the counterSlice
		// because each goroutine has access to only a
		// unique index `idx` of the slice.
		counterSlice[idx] = 0
		j := 0

		// Use of time.Tick in production is discouraged.
		// time.Tick cannot be stopped, we are using it because
		// this is a test code.
		for range time.Tick(2 * time.Millisecond) {
			if sharedLimiter.ShouldAllow(1) {
				counterSlice[idx]++
			}

			j++
			if j%500 == 0 {
				break
			}
		}
	}

	for i := 0; i < nRuns; i++ {
		// create a wait group and
		wg := sync.WaitGroup{}
		for j := 0; j < nRoutines; j++ {
			wg.Add(1)
			go routine(j, &wg)
		}

		wg.Wait()

		// sum over the counterSlice and check accuracy:
		var count uint64 = 0
		for _, partialCount := range counterSlice {
			count += partialCount
		}

		// check accuracy of counter
		if (limit-allowanceRange) <= count && count <= (limit+allowanceRange) {
			fmt.Printf(
				"Iteration %d, Allowed tasks: %d, passed rate limiting accuracy test.\n",
				i+1, count,
			)
		} else {
			t.Fatalf(
				"Accuracy test failed, expected results to be in +/- %d error range, but got %d",
				allowanceRange, count,
			)
		}
	}
}
