package ratelimiter

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestInvalidLimiterConfiguration(t *testing.T) {
	limiter := NewDefaultLimiter(10, time.Microsecond*800)
	if _, err := limiter.ShouldAllow(3); err == nil {
		t.Fatalf("ShouldAllow() failed, did not throw error when window size <= 1 millisecond")
	}

	limiter1 := NewSyncLimiter(0, 10*time.Second)
	if _, err := limiter1.ShouldAllow(10); err == nil {
		t.Fatalf("ShouldAllow() failed, did not throw error when limit == 0")
	}
}

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
	var allowanceRange uint64 = 20

	// will be set to true once the go routine completes all `nRuns`

	limiter := NewDefaultLimiter(limit, duration)
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
	sharedLimiter := NewDefaultLimiter(limit, duration)
	defer sharedLimiter.Kill()

	// launch N go-routines:
	nRoutines := 4

	// test with accuracy +/- 3, modify this variable to
	// test accuracy for various error offsets, 0 is the most
	// ideal case.
	var allowanceRange uint64 = 20

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

func TestConcurrentSyncLimiter(t *testing.T) {
	nRuns := 10
	duration := time.Second * 1

	// 100 tasks must be allowed to execute
	// for every `duration` interval.
	var limit uint64 = 100

	// create a limiter, that is shared across go routines:
	sharedLimiter := NewSyncLimiter(limit, duration)
	defer sharedLimiter.Kill()

	// launch N go-routines:
	nRoutines := 4

	// dry run, this will allow rate-limiter to stabilize:
	isDry := true

	// test with accuracy +/- 3, modify this variable to
	// test accuracy for various error offsets, 0 is the most
	// ideal case.
	var allowanceRange uint64 = 20

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
		if !isDry {
			if (limit-allowanceRange) <= count && count <= (limit+allowanceRange) {
				fmt.Printf(
					"Iteration %d, Allowed tasks: %d, passed rate limiting accuracy test.\n",
					i, count,
				)
			} else {
				t.Fatalf(
					"Accuracy test failed, expected results to be in +/- %d error range, but got %d",
					allowanceRange, count,
				)
			}
		}

		isDry = false
	}
}

func TestLimiterCleanup(t *testing.T) {
	var limit uint64 = 10
	var size time.Duration = 5 * time.Second

	limiter := NewDefaultLimiter(limit, size)

	// call allow check on limiter:
	_, err := limiter.ShouldAllow(1)
	if err != nil {
		t.Fatalf("Error when calling ShouldAllow() on active limiter, Error: %v", err)
	}

	// kill the limiter:
	if err = limiter.Kill(); err != nil {
		t.Fatalf("Failed to kill an active limiter, Error: %v", err)
	}

	// try to call kill again on already killed limiter:
	if err = limiter.Kill(); err == nil {
		t.Fatalf("Failed to throw error when Kill() was called on the same limiter twice.")
	}

	// call ShouldAllow() on inactive limiter, this should throw an error
	_, err = limiter.ShouldAllow(4)
	if err == nil {
		t.Fatalf("Calling ShouldAllow() on inactive limiter did not throw any errors.")
	}
}

func TestSyncLimiterCleanup(t *testing.T) {
	var limit uint64 = 10
	var size time.Duration = 5 * time.Second

	limiter := NewSyncLimiter(limit, size)

	// call allow check on limiter:
	_, err := limiter.ShouldAllow(1)
	if err != nil {
		t.Fatalf("Error when calling ShouldAllow() on active limiter, Error: %v", err)
	}

	// kill the limiter:
	if err = limiter.Kill(); err != nil {
		t.Fatalf("Failed to kill an active limiter, Error: %v", err)
	}

	// try to call kill again on already killed limiter:
	if err = limiter.Kill(); err == nil {
		t.Fatalf("Failed to throw error when Kill() was called on the same limiter twice.")
	}

	// call ShouldAllow() on inactive limiter, this should throw an error
	_, err = limiter.ShouldAllow(4)
	if err == nil {
		t.Fatalf("Calling ShouldAllow() on inactive limiter did not throw any errors.")
	}
}

func BenchmarkDefaultLimiter(b *testing.B) {
	limiter := NewDefaultLimiter(100, 1*time.Second)

	for i := 0; i < b.N; i++ {
		_, err := limiter.ShouldAllow(1)
		if err != nil {
			b.Fatalf("Error when calling ShouldAllow() on active limiter, Error: %v", err)
		}
	}
}

func BenchmarkSyncLimiter(b *testing.B) {
	limiter := NewSyncLimiter(100, 1*time.Second)

	for i := 0; i < b.N; i++ {
		_, err := limiter.ShouldAllow(1)
		if err != nil {
			b.Fatalf("Error when calling ShouldAllow() on active limiter, Error: %v", err)
		}
	}
}

func BenchmarkConcurrentDefaultLimiter(b *testing.B) {
	limiter := NewDefaultLimiter(100, 1*time.Second)

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			_, err := limiter.ShouldAllow(1)
			if err != nil {
				b.Fatalf("Error when calling ShouldAllow() on active limiter, Error: %v", err)
			}
		}
	})
}

func BenchmarkConcurrentSyncLimiter(b *testing.B) {
	limiter := NewSyncLimiter(100, 1*time.Second)

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			_, err := limiter.ShouldAllow(1)
			if err != nil {
				b.Fatalf("Error when calling ShouldAllow() on active limiter, Error: %v", err)
			}
		}
	})
}
