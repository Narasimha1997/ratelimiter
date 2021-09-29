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
