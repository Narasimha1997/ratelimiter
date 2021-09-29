package ratelimit

import (
	"fmt"
	"sync"
	"time"
)

type AttributeMap map[string]Limiter

type AttributeBasedLimiter struct {
	attributeMap AttributeMap
	m            sync.Mutex
}

func (a *AttributeBasedLimiter) HasKey(key *string) bool {
	a.m.Lock()
	_, ok := a.attributeMap[*key]
	a.m.Unlock()
	return ok
}

func (a *AttributeBasedLimiter) CreateNewKey(key *string, limit uint64, size time.Duration) error {
	a.m.Lock()
	defer a.m.Unlock()

	if _, ok := a.attributeMap[*key]; ok {
		return fmt.Errorf(
			"key %s is already defined", *key,
		)
	}

	// create a new entry:
	a.attributeMap[*key] = *NewLimiter(limit, size)
	return nil
}

func (a *AttributeBasedLimiter) ShouldAllow(key *string, n uint64) (bool, error) {
	a.m.Lock()
	defer a.m.Unlock()

	limiter, ok := a.attributeMap[*key]
	if ok {
		return limiter.ShouldAllow(n), nil
	}

	return false, fmt.Errorf("key %s not found", *key)
}

func (a *AttributeBasedLimiter) DeleteKey(key *string) error {

	a.m.Lock()
	defer a.m.Unlock()

	if _, ok := a.attributeMap[*key]; ok {
		delete(a.attributeMap, *key)
		return nil
	}

	return fmt.Errorf("key %s not found", *key)
}

func NewAttributeBasedLimiter() *AttributeBasedLimiter {
	return &AttributeBasedLimiter{
		attributeMap: make(AttributeMap),
	}
}
