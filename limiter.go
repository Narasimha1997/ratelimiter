package ratelimiter

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Limiter struct {
	previous      *Window
	current       *Window
	lock          sync.Mutex
	size          time.Duration
	limit         uint64
	killed        bool
	windowContext context.Context
	cancelFn      func()
}

func (l *Limiter) ShouldAllow(n uint64) (bool, error) {

	if l.killed {
		return false, fmt.Errorf("function ShouldAllow called on an inactive instance")
	}

	l.lock.Lock()

	currentTime := time.Now()
	currentWindowBoundary := currentTime.Sub(l.current.getStartTime())

	w := float64(l.size-currentWindowBoundary) / float64(l.size)

	currentSlidingRequests := uint64(w*float64(l.previous.count)) + l.current.count

	defer l.lock.Unlock()
	if currentSlidingRequests+n > l.limit {
		return false, nil
	}

	// add current request count to window of current count
	l.current.updateCount(n)
	return true, nil
}

func (l *Limiter) progressiveWindowSlider() {
	for {
		select {
		case <-l.windowContext.Done():
			return
		default:
			toSleepDuration := l.size - time.Since(l.current.getStartTime())
			time.Sleep(toSleepDuration)
			l.lock.Lock()
			// make current as previous and create a new current window
			l.previous.setStateFrom(l.current)
			l.current.resetToTime(time.Now())
			l.lock.Unlock()
		}
	}
}

func (l *Limiter) Kill() error {

	l.lock.Lock()
	defer l.lock.Unlock()

	if l.killed {
		return fmt.Errorf("called Kill on already killed limiter")
	}

	defer l.cancelFn()
	l.killed = true
	return nil
}

func NewLimiter(limit uint64, size time.Duration) *Limiter {
	previous := NewWindow(0, time.Now())
	current := NewWindow(0, time.Now())

	childCtx, cancelFn := context.WithCancel(context.Background())

	limiter := &Limiter{
		previous:      previous,
		current:       current,
		lock:          sync.Mutex{},
		size:          size,
		limit:         limit,
		killed:        false,
		windowContext: childCtx,
		cancelFn:      cancelFn,
	}

	go limiter.progressiveWindowSlider()
	return limiter
}
