package ratelimiter

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestLimiterAccuracy(t *testing.T) {

	nRuns := 10
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
	defer limiter.Kill()

	for i := 0; i < nRuns; i++ {
		count = 0
		nTicks := 0
		for range time.Tick(time.Millisecond * 2) {

			canAllow, err := limiter.ShouldAllow(1)
			if err != nil {
				t.Fatalf("%v", err)
			}

			if canAllow {
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
	nRuns := 10
	duration := time.Second * 1

	// 100 tasks must be allowed to execute
	// for every `duration` interval.
	var limit uint64 = 100

	// create a limiter, that is shared across go routines:
	sharedLimiter := NewLimiter(limit, duration)
	defer sharedLimiter.Kill()

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
			canAllow, err := sharedLimiter.ShouldAllow(1)
			if err != nil {
				break
			}

			if canAllow {
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

func TestLimiterCleanup(t *testing.T) {
	var limit uint64 = 10
	var size time.Duration = 5 * time.Second

	limiter := NewLimiter(limit, size)

	// call allow check on limiter:
	_, err := limiter.ShouldAllow(1)
	if err != nil {
		t.Fatalf("Error when calling ShouldAllow() on active limiter, Error: %v", err)
	}

	// kill the limiter:
	err = limiter.Kill()
	if err != nil {
		t.Fatalf("Failed to kill an active limiter, Error: %v", err)
	}

	// call ShouldAllow() on inactive limiter, this should throw an error
	_, err = limiter.ShouldAllow(4)
	if err == nil {
		t.Fatalf("Calling ShouldAllow() on inactive limiter did not throw any errors.")
	}
}
