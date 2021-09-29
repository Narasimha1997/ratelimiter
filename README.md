# ratelimiter
A generic concurrent rate limiter library for Golang based on Sliding-window rate limitng algorithm.

The implementation of rate-limiter algorithm is based on Scalable Distributed Rate Limiter algorithm  used in Kong API gateway. Read [this blog](https://konghq.com/blog/how-to-design-a-scalable-rate-limiting-algorithm/) for more details.

This library can be used in your code base to rate-limit literally anything. For example, you can integrate this library to provide rate-limiting for your REST/gRPC APIs or you can use this library to 
rate-limit the number of go-routines spawned or number of tasks submitted to a function/module per given time interval. This library provides generic rate check APIs that can be used anywhere. The library is built with concurrency in mind from the groud up, the rate-limiter can be used across go-routines without having to worry about synchronization issues. This library also provides capability to create and manage multiple rate-limiters with different configurations assiociated with unique keys.

### Installation:
The package can be installed as a Go module.

```
go get github.com/Narasimha1997/ratelimiter
```

### Using the library:
There are two types of rate-limiters used.

#### Generic rate-limiter
The generic rate-limiter instance can be created if you want to have a single rate-limiter with single configuration for everything. The generic rate-limiter can be created by calling `NewLimiter()` function and by passing the `limit` and `size` as parameters. Example:

```go
import (
	"fmt"
	"log"
	"time"

	"github.com/Narasimha1997/ratelimiter"
)

func GenericRateLimiter() {
	/* create an instance of Limiter.
	format: NewLimiter(limit uint64, size time.Duration),
	where:
		limit: The number of tasks/items that should be allowed.
		size: The window size, i.e the time interval during which the limit
				should be imposed.
		To summarize, if limit = 100 and duration = 5s, then allow 100 items per 5 seconds
	*/

	limiter := ratelimiter.NewLimiter(
		100, time.Second*5,
	)

	/*
		the limiter provides ShouldAllow(N uint64) function which
		returns true/false if N items/tasks can be allowed during current
		time interval
	*/

	// ShouldAllow(N uint64) -> returns true/false

	// should return true
	fmt.Println(limiter.ShouldAllow(60))
	// should return false, because (60 + 50 = 110) > 100 during this window
	fmt.Println(limiter.ShouldAllow(50))
	// sleep for some time
	time.Sleep(5 * time.Second)
	// should return true, because the previous window has been slided over
	fmt.Println(limiter.ShouldAllow(20))
}
```

#### Attribute based rate-limiter:
Attribute based rate-limiter can hold multiple rate-limiters with different configurations in a map
of <string, Limiter> type. Each limiter is uniquely identified by a key. Calling  `NewAttributeBasedLimiter()` will create an empty rate limiter with no entries.

```go
func AttributeRateLimiter() {
	/*
		Attribute based rate-limiter can hold multiple
		rate-limiters with different configurations in a map
		of <string, Limiter> type. Each limiter is uniquely identified
		by a key. Calling NewAttributeBasedLimiter() will create an empty
		rate limiter with no entries.
	*/
	limiter := ratelimiter.NewAttributeBasedLimiter()

	/*
		Now we are adding a new entry to the limiter, we pass:
			key: A string that is used to uniquely identify the rate-limiter.
			limit: The number of tasks/items that should be allowed.
			size: The window size, i.e the time interval during which the limit
				should be imposed.

		returns error if the key already exists in the map.
	*/
	// we have two articles here (for example)
	article_ids := []string{"article_id=10", "article_id=11"}

	// for article_id=10, allow 10 tasks/items per every second
	err := limiter.CreateNewKey(&article_ids[0], 10, 5*time.Second)
	if err != nil {
		log.Fatalln(err)
	}
	// for article_id=11, allow 100 tasks/items per every 6 minutes
	err = limiter.CreateNewKey(&article_ids[1], 100, 6*time.Minute)
	if err != nil {
		log.Fatalln(err)
	}
	// rates can be checked by passing key and N as parameters
	// Can I make 8 requests to article_id=10 during this time window?

	// ShouldAllow(key *string, N uint64) returns (bool, error)
	// the bool is true/false, true if it can be allowed
	// false if it cant be allowed.
	// error if key is not found.

	fmt.Println(limiter.ShouldAllow(&article_ids[0], 8))
	// Can I make 104 requests to article_id=11 during this time window?
	fmt.Println(limiter.ShouldAllow(&article_ids[0], 104))

	/*
		Other functions:
			1. HasKey: to check if the attribute already has given key
			   call: HasKey(key *string) function.
			   Example: limiter.HasKey(&article_id[0])
			   Returns a bool, true if exists, false otherwise

			2. DeleteKey: to remove the key from attribute map
			   call: DeleteKey(key *string) function.
			   Example: limiter.DeleteKey(&article_id[1])
			   Returns an error, if key was not in the map.
	*/
}
```

### Testing
Tests are written in `attribute_limiter_test.go` and `limiter_test.go` files. To execute the tests, 
simply run:
```
go test ./ -v
```

These are some of the results from tests:
1. **Single goroutine, Generic limiter**: This test configures the rate-limiter to allow 100 requests/sec and fires 500 requests/sec with a time gap of 2ms each, allowed requests are counted and is tested with difference +/- 3. The same test is run for 10 samples. Here are the results:

```
=== RUN   TestLimiterAccuracy
Iteration 1, Allowed tasks: 100, passed rate limiting accuracy test.
Iteration 2, Allowed tasks: 101, passed rate limiting accuracy test.
Iteration 3, Allowed tasks: 100, passed rate limiting accuracy test.
Iteration 4, Allowed tasks: 100, passed rate limiting accuracy test.
Iteration 5, Allowed tasks: 100, passed rate limiting accuracy test.
Iteration 6, Allowed tasks: 100, passed rate limiting accuracy test.
Iteration 7, Allowed tasks: 101, passed rate limiting accuracy test.
Iteration 8, Allowed tasks: 100, passed rate limiting accuracy test.
Iteration 9, Allowed tasks: 100, passed rate limiting accuracy test.
Iteration 10, Allowed tasks: 100, passed rate limiting accuracy test.
--- PASS: TestLimiterAccuracy (10.01s)
```

2. **4 goroutines, Generic Limiter**: This test configures the limiter to allow 100 requests/sec and spins up 4 goroutines, the same limiter is shared across all the routines. Each goroutine generates 500 requests/sec with 2ms time gap between 2 requests. Allowed requests are counted per each goroutine, the result sum of all counts should be almost equal to 100. The accuracy is measured considering +/- 3 as error offset. The same test is conducted 10 times. Here are the results:

```
=== RUN   TestConcurrentLimiterAccuracy
Iteration 1, Allowed tasks: 101, passed rate limiting accuracy test.
Iteration 2, Allowed tasks: 100, passed rate limiting accuracy test.
Iteration 3, Allowed tasks: 100, passed rate limiting accuracy test.
Iteration 4, Allowed tasks: 100, passed rate limiting accuracy test.
Iteration 5, Allowed tasks: 100, passed rate limiting accuracy test.
Iteration 6, Allowed tasks: 100, passed rate limiting accuracy test.
Iteration 7, Allowed tasks: 100, passed rate limiting accuracy test.
Iteration 8, Allowed tasks: 100, passed rate limiting accuracy test.
Iteration 9, Allowed tasks: 100, passed rate limiting accuracy test.
Iteration 10, Allowed tasks: 100, passed rate limiting accuracy test.
--- PASS: TestConcurrentLimiterAccuracy (10.01s)
```

3. **2 goroutines, 2 attribute keys, Attribute based limiter**: An attribute based limiter is created with 2 keys, these keys are configured to allow 100 requests/sec and 123 requests/sec respectively. Two goroutines are created and same attribute based limiter is shared across. Each goroutine produces 500 requests/sec per key. The overall count is then verified for each goroutine with error offset of +/- 3. Here are the results:

```
=== RUN   TestAttributeBasedLimiterAccuracy
Iteration 1, Allowed tasks: 100, passed rate limiting accuracy test.
Iteration 1, Allowed tasks: 123, passed rate limiting accuracy test.
Iteration 2, Allowed tasks: 101, passed rate limiting accuracy test.
Iteration 2, Allowed tasks: 124, passed rate limiting accuracy test.
Iteration 3, Allowed tasks: 100, passed rate limiting accuracy test.
Iteration 3, Allowed tasks: 123, passed rate limiting accuracy test.
Iteration 4, Allowed tasks: 100, passed rate limiting accuracy test.
Iteration 4, Allowed tasks: 123, passed rate limiting accuracy test.
Iteration 5, Allowed tasks: 100, passed rate limiting accuracy test.
Iteration 5, Allowed tasks: 123, passed rate limiting accuracy test.
--- PASS: TestAttributeBasedLimiterAccuracy (5.00s)
```

#### Notes on test:
The testing code produces 500 requests/sec with `2ms` precision time gap between each request. The accuracy of this `2ms` time tick generation can differ from platform to platform, even a small difference of 500 micorseconds can add up together and give more time for test to run in the end because of clock drift, as a result the error offset +/- 3 might not always work. On Windows for example, the `2ms` precision time ticks can be inconsistent because the windows scheduler wakes up every `15ms` causing a drift in the clock time, however Linux based distros have precise timers that allow us to obtain precise `2ms` time tikcs.

### Contributing
Feel free to raise issues, make pull requests or suggest new features.
