# ratelimiter

![Tests](https://github.com/Narasimha1997/ratelimiter/actions/workflows/test.yml/badge.svg)
[![Go Reference](https://pkg.go.dev/badge/github.com/Narasimha1997/ratelimiter.svg)](https://pkg.go.dev/github.com/Narasimha1997/ratelimiter)

A generic concurrent rate limiter library for Golang based on Sliding-window rate limitng algorithm.

The implementation of rate-limiter algorithm is based on Scalable Distributed Rate Limiter algorithm  used in Kong API gateway. Read [this blog](https://konghq.com/blog/how-to-design-a-scalable-rate-limiting-algorithm/) for more details.

This library can be used in your codebase to rate-limit literally anything. For example, you can integrate this library to provide rate-limiting for your REST/gRPC APIs or you can use this library to 
rate-limit the number of go-routines spawned or number of tasks submitted to a function/module per given time interval. This library provides generic rate check APIs that can be used anywhere. The library is built with concurrency in mind from the groud up, the rate-limiter can be used across go-routines without having to worry about synchronization issues. This library also provides capability to create and manage multiple rate-limiters with different configurations assiociated with unique keys.

### Installation:
The package can be installed as a Go module.

```
go get github.com/Narasimha1997/ratelimiter
```

### Using the library:
There are two types of rate-limiters used.

#### All APIs:
1. **Generic rate-limiter**:
```go 
	/* creates an instance of DefaultLimiter and returns it's pointer.
	   Parameters:
	 		limit: The number of tasks to be allowd
			size: duration
	*/
	func NewDefaultLimiter(limit uint64, size time.Duration) *DefaultLimiter

	/*
		Kill the limiter, returns error if the limiter has been killed already.
	*/
	func (s *DefaultLimiter) Kill() error

	/*
		Makes decison whether n tasks can be allowed or not.
		Parameters:
			n: number of tasks to be processed, set this as 1 for a single task. 
				(Example: An HTTP request)
		Returns (bool, error),
			if limiter is inactive (or it is killed), returns an error
			the boolean flag is either true - i.e n tasks can be allowed or false otherwise.
	*/
	func (s *DefaultLimiter) ShouldAllow(n uint64) (bool, error)

	/*
		Kill the limiter, returns error if the limiter has been killed already.
	*/
	func (s *DefaultLimiter) Kill() error	
```

2. **On-demand rate-limiter**
```go
	/*  creates an instance of SyncLimiter and returns it's pointer.
	 	Parameters:
	 		limit: The number of tasks to be allowd
			size: duration
	*/
	func NewSyncLimiter(limit uint64, size time.Duration) *SyncLimiter

	/*
		Kill the limiter, returns error if the limiter has been killed already.
	*/
	func (s *SyncLimiter) Kill() error

	/*
		Makes decison whether n tasks can be allowed or not.
		Parameters:
			n: number of tasks to be processed, set this as 1 for a single task. 
				(Example: An HTTP request)
		Returns (bool, error),
			if limiter is inactive (or it is killed), returns an error
			the boolean flag is either true - i.e n tasks can be allowed or false otherwise.
	*/
	func (s *SyncLimiter) ShouldAllow(n uint64) (bool, error)

	/*
		Kill the limiter, returns error if the limiter has been killed already.
	*/
	func (s *SyncLimiter) Kill() error
```

3. **Attribute based Rate Limiter**
```go
	/*
		Creates an instance of AttributeBasedLimiter and returns it's pointer.
		Parameters:
			backgroundSliding: if set to true, DefaultLimiter will be used as an underlying limiter.
							   else, SyncLimiter will be used.
	*/
	func NewAttributeBasedLimiter(backgroundSliding bool) *AttributeBasedLimiter

	/*
		Check if AttributeBasedLimiter has a limiter for the key.
		Parameters:
			key: a unique key string, example: IP address, token, uuid etc
		Returns a boolean flag, if true, the key is already present, false otherwise.
	*/
	func (a *AttributeBasedLimiter) HasKey(key string) bool

	/*
		Create a new key-limiter assiociation.
		Parameters:
			key: a unique key string, example: IP address, token, uuid etc
			limit: The number of tasks to be allowd
			size: duration
		Returns error if the key already exist.
	*/

	func (a *AttributeBasedLimiter) CreateNewKey(
		key string, limit uint64, 
		size time.Duration,
	) error

	/* 
	   check if AttributeBasedLimiter has a limiter for the key.
	   Create a new key-limiter assiociation if the key not exists.
	   Parameters:
	    key: a unique key string, example: IP address, token, uuid etc.
		limit: The number of tasks to be allowd
		size: duration
		Return true if the key exists or is created successfully.
	*/
	func (a *AttributeBasedLimiter) HasOrCreateKey(key string, limit uint64, size time.Duration);

	/*
		Makes decison whether n tasks can be allowed or not.
		Parameters:
			key: a unique key string, example: IP address, token, uuid etc
			n: number of tasks to be processed, set this as 1 for a single task. 
				(Example: An HTTP request)
		Returns (bool, error),
			if limiter is inactive (or it is killed) or key is not present, returns an error
			the boolean flag is either true - i.e n tasks can be allowed or false otherwise.
	*/
	func (a *AttributeBasedLimiter) ShouldAllow(key string, n uint64) (bool, error)

	/* 
		MustShouldAllow makes decison whether n tasks can be allowed or not.
		Creates a new key if it does not exist.
		Parameters:
			key: a unique key string, example: IP address, token, uuid etc
			n: number of tasks to be processed, set this as 1 for a single task.
			(Example: An HTTP request)
			limit: The number of tasks to be allowd
			size: duration

		Returns bool.
			(false) when limiter is inactive (or it is killed) or n tasks can be not allowed.
			(true) when n tasks can be allowed or new key-limiter.
	*/
	func (a *AttributeBasedLimiter) MustShouldAllow(key string, n uint64, limit uint64, size time.Duration) bool

	/*
		Remove the key and kill its underlying limiter.
		Parameters:
			key: a unique key string, example: IP address, token, uuid etc
		Returns an error if the key is not present.
	*/
	func (a *AttributeBasedLimiter) DeleteKey(key string) error
```

### Examples and Explanation of each type of rate-limiter:
#### Generic rate-limiter
The generic rate-limiter instance can be created if you want to have a single rate-limiter with single configuration for everything. The generic rate-limiter can be created by calling `NewDefaultLimiter()` function and by passing the `limit` and `size` as parameters. Example:

```go
func GenericRateLimiter() {
	/* create an instance of Limiter.
	format: NewLimiter(limit uint64, size time.Duration),
	where:
		limit: The number of tasks/items that should be allowed.
		size: The window size, i.e the time interval during which the limit
				should be imposed.
		To summarize, if limit = 100 and duration = 5s, then allow 100 items per 5 seconds
	*/

	limiter := ratelimiter.NewDefaultLimiter(
		100, time.Second*5,
	)

	/*
		Cleaning up the limiter: Once the limiter is no longer required,
		the underlying goroutines and resources used by the limiter can be cleaned up.
		This can be done using:
			limiter.Kill(),
		Returns an error if the limiter is already being killed.
	*/

	defer limiter.Kill()

	/*
		the limiter provides ShouldAllow(N uint64) function which
		returns true/false if N items/tasks can be allowed during current
		time interval.

		An error is returned if the limiter is already killed.
	*/

	// ShouldAllow(N uint64) -> returns bool, error

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

#### On demand window sliding:
The previous method i.e the Generic Rate limiter spins up a background goroutine that takes care of sliding the rate-limiting window whenever it's size expires, because of this, rate-limiting check function `ShouldAllow` has fewer steps and takes very less time to make decision. But if your application manages a large number of Limiters, for example a web-server that performs rate-limiting across hundreds of different IPs, then your `AttributeBasedRateLimiter` spins up a goroutine for each unique IP and thus lot of such routines needs to be manitanied, this might induce scheduling pressure.

An alternative solution is to use a rate-limiter does not require a background routine, instead the window is sliding is taken care by `ShouldAllow` function itself, this method can be used to maintain large number of rate limiters without any scheduling pressure. This limiter is called `SyncLimiter` and can be used just like `DefaultLimiter`, because `SyncLimiter` and `DefaultLimiter` are built on top of the same `Limiter` interface. To use this, just replace `NewDefaultLimiter` with `NewSyncLimiter`
```go
......

	limiter := ratelimiter.NewSyncLimiter(
		100, time.Second*5,
	)
......
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
	/*
		Attribute based rate-limiter has a boolean parameter called:
		`backgroundSliding` - if set to true, the attribute based rate-limiter
		uses Limiter instance and each Limiter instance have it's own background goroutine
		to manage sliding window. This might be resource expensive for large number of attributes,
		but is faster than SyncLimiter.

		Disable this, i.e pass `false` if you want to manage large number of attributes
		in less memory and compute, sacrifcing a minimal amount of performance.
	*/
	limiter := ratelimiter.NewAttributeBasedLimiter(true)

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
			   call: HasKey(key string) function.
			   Example: limiter.HasKey(&article_id[0])
			   Returns a bool, true if exists, false otherwise

			2. DeleteKey: to remove the key from attribute map
			   call: DeleteKey(key string) function.
			   Example: limiter.DeleteKey(&article_id[1])
			   Returns an error, if key was not in the map.
	*/
}
```

### Using ratelimiter as a middleware with HTTP web server:
ratelimiter is pluggable and can be used anywhere. This code snippet shows how it can be used with
Go's standard HTTP library when building a web server:

```go
.....
// allow 100 requests every 5 seconds
limiter := ratelimiter.NewSyncLimiter(100, time.Second * 5)

// register the handler
rateLimiterHandler := func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowed, err := limiter.ShouldAllow(1)
		if err != nil {
			log.Fatalln(err)
		}
		if allowed {
			next.ServeHTTP(w, r)
		}
	})
}

// create a test route handler
ponger := func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Pong!!"))
}

// attach the ratelimiter middleware:
muxServer := http.NewServeMux()
muxServer.Handle("/", rateLimiterHandler(
	http.HandlerFunc(ponger),
))

// start the server
err := http.ListenAndServe(":6000", muxServer)
if err != nil {
	log.Fatalln(err)
}
```
The complete example can be found at `examples/http-server/server.go`.
`curl` was used to simulate `X` requests per second and following was the output, as logged.
```
................................
2021/10/05 14:12:38 Iteration: 7, Requests received: 522, Allowed: 99
2021/10/05 14:12:43 Iteration: 8, Requests received: 533, Allowed: 101
2021/10/05 14:12:48 Iteration: 9, Requests received: 515, Allowed: 100
2021/10/05 14:12:53 Iteration: 10, Requests received: 505, Allowed: 100
2021/10/05 14:12:58 Iteration: 11, Requests received: 508, Allowed: 100
2021/10/05 14:13:03 Iteration: 12, Requests received: 474, Allowed: 100
2021/10/05 14:13:08 Iteration: 13, Requests received: 495, Allowed: 100
2021/10/05 14:13:13 Iteration: 14, Requests received: 478, Allowed: 100
..................................
```
The ratelimiter was able to balance the requested limit as specified.
If you have installed the package, you can simply run the webserver as follows:
```
go run examples/http-server/server.go
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

**Code coverage**:
To generate code coverage report, execute:
```
go test -coverprofile=c.out
```

This should print the following after running all the tests.
```
coverage: 99.0% of statements
ok      github.com/Narasimha1997/ratelimiter    25.099s
```

You can also save the results as HTML for more detailed code view of the coverage.
```
go tool cover -html=c.out -o coverage.html
```

This will generate a file called `coverage.html`. The `coverage.html` is provided in the repo which is pre-generated.

#### Notes on test:
The testing code produces 500 requests/sec with `2ms` precision time gap between each request. The accuracy of this `2ms` time tick generation can differ from platform to platform, even a small difference of 500 micorseconds can add up together and give more time for test to run in the end because of clock drift, as a result the error offset +/- 3 might not always work. On Windows for example, the `2ms` precision time ticks can be inconsistent because the windows scheduler wakes up every `15ms` causing a drift in the clock time, however Linux based distros have precise timers that allow us to obtain precise `2ms` time tikcs.

### Contributing
Feel free to raise issues, make pull requests or suggest new features.
