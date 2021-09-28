package core

import (
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
		l.previous = l.current
		l.current = NewWindow(0, time.Now())
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
