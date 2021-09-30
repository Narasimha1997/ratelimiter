package main

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
	err := limiter.CreateNewKey(article_ids[0], 10, 5*time.Second)
	if err != nil {
		log.Fatalln(err)
	}
	// for article_id=11, allow 100 tasks/items per every 6 minutes
	err = limiter.CreateNewKey(article_ids[1], 100, 6*time.Minute)
	if err != nil {
		log.Fatalln(err)
	}
	// rates can be checked by passing key and N as parameters
	// Can I make 8 requests to article_id=10 during this time window?

	// ShouldAllow(key string, N uint64) returns (bool, error)
	// the bool is true/false, true if it can be allowed
	// false if it cant be allowed.
	// error if key is not found.

	fmt.Println(limiter.ShouldAllow(article_ids[0], 8))
	// Can I make 104 requests to article_id=11 during this time window?
	fmt.Println(limiter.ShouldAllow(article_ids[0], 104))

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

func main() {

	fmt.Println("Generic rate limiter:")
	GenericRateLimiter()
	fmt.Println("Attribute based rate limiter:")
	AttributeRateLimiter()
}
