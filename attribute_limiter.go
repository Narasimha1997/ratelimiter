package ratelimiter

import (
	"fmt"
	"sync"
	"time"
)

// AttributeMap is a custom map type of string key and Limiter instance as value
type AttributeMap map[string]Limiter

// AttributeBasedLimiter is an instance that can manage multiple rate limiter instances
// with different configutations.
type AttributeBasedLimiter struct {
	attributeMap AttributeMap
	m            sync.Mutex
	syncMode     bool
}

// Check if AttributeBasedLimiter has a limiter for the key.
//
// Parameters:
//
// 1. key: a unique key string, example: IP address, token, uuid etc
//
// Returns a boolean flag, if true, the key is already present, false otherwise.
func (a *AttributeBasedLimiter) HasKey(key string) bool {
	a.m.Lock()
	_, ok := a.attributeMap[key]
	a.m.Unlock()
	return ok
}

// Create a new key-limiter assiociation.
//
// Parameters:
//
// 1. key: a unique key string, example: IP address, token, uuid etc
//
// 2. limit: The number of tasks to be allowd
//
// 3. size: duration
//
// Returns error if the key already exists.
func (a *AttributeBasedLimiter) CreateNewKey(key string, limit uint64, size time.Duration) error {
	a.m.Lock()
	defer a.m.Unlock()

	if _, ok := a.attributeMap[key]; ok {
		return fmt.Errorf(
			"key %s is already defined", key,
		)
	}

	// create a new entry:
	if !a.syncMode {
		a.attributeMap[key] = NewDefaultLimiter(limit, size)
	} else {
		a.attributeMap[key] = NewSyncLimiter(limit, size)
	}
	return nil
}

// Makes decison whether n tasks can be allowed or not.
//
// Parameters:
//
// key: a unique key string, example: IP address, token, uuid etc
//
// n: number of tasks to be processed, set this as 1 for a single task.
// (Example: An HTTP request)
//
// Returns (bool, error).
// (false, error) when limiter is inactive (or it is killed) or key is not present.
// (true/false, nil) if key exists and n tasks can be allowed or not.
func (a *AttributeBasedLimiter) ShouldAllow(key string, n uint64) (bool, error) {
	a.m.Lock()
	defer a.m.Unlock()

	limiter, ok := a.attributeMap[key]
	if ok {
		return limiter.ShouldAllow(n)
	}

	return false, fmt.Errorf("key %s not found", key)
}

// Remove the key and kill its underlying limiter.
//
// Parameters:
//
// 1.key: a unique key string, example: IP address, token, uuid etc
//
// Returns an error if the key is not present.
func (a *AttributeBasedLimiter) DeleteKey(key string) error {

	a.m.Lock()
	defer a.m.Unlock()

	if limiter, ok := a.attributeMap[key]; ok {
		err := limiter.Kill()
		if err != nil {
			return err
		}
		delete(a.attributeMap, key)
		return nil
	}

	return fmt.Errorf("key %s not found", key)
}

// Creates an instance of AttributeBasedLimiter and returns it's pointer.
//
// Parameters:
//
// 1. backgroundSliding: if set to true, DefaultLimiter will be used as an underlying limiter,
// else, SyncLimiter will be used.
func NewAttributeBasedLimiter(backgroundSliding bool) *AttributeBasedLimiter {
	return &AttributeBasedLimiter{
		attributeMap: make(AttributeMap),
		syncMode:     !backgroundSliding,
	}
}
