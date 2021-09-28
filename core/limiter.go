package core

import (
	"fmt"
	"sync"
	"time"
)

type Limiter struct {
	previous *Window
	current  *Window
	lock     sync.Mutex
	size     time.Duration
	limit    uint64
}

func (l *Limiter) ShouldAllow(n uint64) bool {
	l.lock.Lock()

	currentTime := time.Now()
	currentWindowBoundary := currentTime.Sub(l.current.getStartTime())

	w := float64(l.size-currentWindowBoundary) / float64(l.size)

	currentSlidingRequests := uint64(w*float64(l.previous.count)) + l.current.count

	defer l.lock.Unlock()
	if currentSlidingRequests+n > l.limit {
		return false
	}

	// add current request count to window of current count
	l.current.updateCount(n)
	return true
}

func (l *Limiter) progressiveWindowSlider() {
	for {
		toSleepDuration := l.size - time.Since(l.current.getStartTime())
		time.Sleep(toSleepDuration)
		l.lock.Lock()
		// make current as previous and create a new current window
		l.previous.setStateFrom(l.current)
		l.current.resetToTime(time.Now())
		l.lock.Unlock()
	}
}

func NewLimiter(limit uint64, size time.Duration) *Limiter {
	previous := NewWindow(0, time.Now())
	current := NewWindow(0, time.Now())

	limiter := &Limiter{
		previous: previous,
		current:  current,
		lock:     sync.Mutex{},
		size:     size,
		limit:    limit,
	}

	go limiter.progressiveWindowSlider()
	return limiter
}

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

func NewAttributeBasedLimiter() *AttributeBasedLimiter {
	return &AttributeBasedLimiter{
		attributeMap: make(AttributeMap),
	}
}
