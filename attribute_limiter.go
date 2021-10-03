package ratelimiter

import (
	"fmt"
	"sync"
	"time"
)

type AttributeMap map[string]Limiter

type AttributeBasedLimiter struct {
	attributeMap AttributeMap
	m            sync.Mutex
	syncMode     bool
}

func (a *AttributeBasedLimiter) HasKey(key string) bool {
	a.m.Lock()
	_, ok := a.attributeMap[key]
	a.m.Unlock()
	return ok
}

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

func (a *AttributeBasedLimiter) ShouldAllow(key string, n uint64) (bool, error) {
	a.m.Lock()
	defer a.m.Unlock()

	limiter, ok := a.attributeMap[key]
	if ok {
		return limiter.ShouldAllow(n)
	}

	return false, fmt.Errorf("key %s not found", key)
}

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

func NewAttributeBasedLimiter(backgroundSliding bool) *AttributeBasedLimiter {
	return &AttributeBasedLimiter{
		attributeMap: make(AttributeMap),
		syncMode:     !backgroundSliding,
	}
}
