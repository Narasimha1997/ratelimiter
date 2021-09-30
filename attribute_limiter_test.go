package ratelimiter

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestAttributeMapGetSetDelete(t *testing.T) {

	duration := 1 * time.Second
	limit := 100

	attributeLimiter := NewAttributeBasedLimiter()

	// create a new key attribute:
	// Example scenario, the rate-limiter
	testKey1 := "/api/getArticle?id=10"
	testKey2 := "/api/getArticle?id=20"

	// check keys:
	if attributeLimiter.HasKey(testKey1) {
		t.Fatalf(
			"AttributeBasedLimiter.HasKey() failed, returned true for non-existing key %s",
			testKey1,
		)
	}

	if attributeLimiter.HasKey(testKey2) {
		t.Fatalf(
			"AttributeBasedLimiter.HasKey() failed, returned true for non-existing key %s",
			testKey2,
		)
	}

	// create key:
	if err := attributeLimiter.CreateNewKey(testKey1, uint64(limit), duration); err != nil {
		t.Fatalf(
			"AttributeBasedLimiter.CreateNewKey() failed, returned error on creating key %s, Error: %v\n",
			testKey1, err,
		)
	}

	if err := attributeLimiter.CreateNewKey(testKey2, uint64(limit), duration); err != nil {
		t.Fatalf(
			"AttributeBasedLimiter.CreateNewKey() failed, returned error on creating key %s, Error: %v\n",
			testKey2, err,
		)
	}

	// create an already existing key:
	if err := attributeLimiter.CreateNewKey(testKey2, uint64(limit), duration); err == nil {
		t.Fatalf(
			"AttributeBasedLimiter.CreateNewKey() failed, did not return error when creating existing key %s\n",
			testKey2,
		)
	}

	// check existing keys:
	if !attributeLimiter.HasKey(testKey1) {
		t.Fatalf(
			"AttributeBasedLimiter.HasKey() failed, returned false for existing key %s",
			testKey1,
		)
	}

	if !attributeLimiter.HasKey(testKey1) {
		t.Fatalf(
			"AttributeBasedLimiter.HasKey() failed, returned false for existing key %s",
			testKey2,
		)
	}

	// remove key
	if err := attributeLimiter.DeleteKey(testKey1); err != nil {
		t.Fatalf(
			"AttributeBasedLimiter.DeleteKey() failed, returned error when removing existing key %s, Error: %v",
			testKey1, err,
		)
	}

	if err := attributeLimiter.DeleteKey(testKey2); err != nil {
		t.Fatalf(
			"AttributeBasedLimiter.DeleteKey() failed, returned error when removing existing key %s, Error: %v",
			testKey2, err,
		)
	}

	// check keys again:
	if attributeLimiter.HasKey(testKey1) {
		t.Fatalf(
			"AttributeBasedLimiter.HasKey() failed, returned true for non-existing key %s",
			testKey1,
		)
	}

	if attributeLimiter.HasKey(testKey2) {
		t.Fatalf(
			"AttributeBasedLimiter.HasKey() failed, returned true for non-existing key %s",
			testKey2,
		)
	}

	// check ShouldAllow on non-existing key:
	if _, err := attributeLimiter.ShouldAllow("noKey", 5); err == nil {
		t.Fatalf(
			"AttributeBasedLimiter.ShouldAllow() failed, did not return error when checking non-existing key.",
		)
	}

	// Remove the non-existing key:
	if err := attributeLimiter.DeleteKey("noKey"); err == nil {
		t.Fatalf(
			"AttributeBasedLimiter.DeleteKey failed, did not return error when deleting non-existing key.",
		)
	}
}

func TestAttributeBasedLimiterAccuracy(t *testing.T) {

	// number of unique keys to be tested
	keys := []string{"/api/getArticle?id=10", "/api/getArticle?id=20"}

	// key1 has limit of 100 hits/sec and key2 has 123 hits/sec allowed.
	limits := []uint64{100, 123}
	counters := make([]uint64, len(keys))

	// per second window
	duration := 1 * time.Second

	// 10 samples will be executed.
	nRuns := 5

	// test with accuracy +/- 3, modify this variable to
	// test accuracy for various error offsets, 0 is the most
	// ideal case.
	var allowanceRange uint64 = 20

	sharedLimiter := NewAttributeBasedLimiter()

	for idx, key := range keys {
		err := sharedLimiter.CreateNewKey(key, limits[idx], duration)
		if err != nil {
			t.Fatalf("%v", err)
		}
	}

	routine := func(key string, idx int, wg *sync.WaitGroup) {
		defer wg.Done()
		j := 0
		counters[idx] = 0
		for range time.Tick(2 * time.Millisecond) {
			allowed, err := sharedLimiter.ShouldAllow(key, 1)
			if err != nil {
				break
			}

			if allowed {
				counters[idx]++
			}

			j++

			if j%500 == 0 {
				break
			}
		}
	}

	// run for nRuns:
	for i := 0; i < nRuns; i++ {
		wg := sync.WaitGroup{}
		for idx, key := range keys {
			wg.Add(1)
			go routine(key, idx, &wg)
		}

		wg.Wait()

		// loop over the keys and check rate-limit:
		for idx, count := range counters {
			limit := limits[idx]

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
}
