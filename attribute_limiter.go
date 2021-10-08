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

// HasKey check if AttributeBasedLimiter has a limiter for the key.
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

// CreateNewKey create a new key-limiter assiociation.
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

	return a.createNewKey(key, limit, size)
}

func (a *AttributeBasedLimiter) createNewKey(key string, limit uint64, size time.Duration) error {
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

// HasOrCreateKey check if AttributeBasedLimiter has a limiter for the key.
// Create a new key-limiter assiociation if the key not exists.
//
// Parameters:
//
// 1. key: a unique key string, example: IP address, token, uuid etc
//
// 2. limit: The number of tasks to be allowd
//
// 3. size: duration
//
// Return true if the key exists or is created successfully.
func (a *AttributeBasedLimiter) HasOrCreateKey(key string, limit uint64, size time.Duration) bool {
	a.m.Lock()
	defer a.m.Unlock()

	if _, ok := a.attributeMap[key]; ok {
		return true
	}

	if err := a.createNewKey(key, limit, size); err == nil {
		return true
	}

	return false
}

// ShouldAllow makes decison whether n tasks can be allowed or not.
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

// MustShouldAllow makes decison whether n tasks can be allowed or not.
//
// Parameters:
//
// key: a unique key string, example: IP address, token, uuid etc
//
// n: number of tasks to be processed, set this as 1 for a single task.
// (Example: An HTTP request)
//
// Returns bool.
// (false) when limiter is inactive (or it is killed) or n tasks can be not allowed.
// (true) when n tasks can be allowed or new key-limiter.
func (a *AttributeBasedLimiter) MustShouldAllow(key string, n uint64, limit uint64, size time.Duration) bool {
	a.m.Lock()
	defer a.m.Unlock()

	if limiter, ok := a.attributeMap[key]; ok {
		allowed, err := limiter.ShouldAllow(n)
		return allowed && err == nil
	}

	err := a.createNewKey(key, limit, size)
	return err == nil
}

// DeleteKey remove the key and kill its underlying limiter.
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

// NewAttributeBasedLimiter creates an instance of AttributeBasedLimiter and returns it's pointer.
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
