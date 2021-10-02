package ratelimiter

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Limiter interface {
	Kill() error
	ShouldAllow(n uint64) (bool, error)
}

type DefaultLimiter struct {
	previous      *Window
	current       *Window
	lock          sync.Mutex
	size          time.Duration
	limit         uint64
	killed        bool
	windowContext context.Context
	cancelFn      func()
}

func (l *DefaultLimiter) ShouldAllow(n uint64) (bool, error) {

	l.lock.Lock()
	defer l.lock.Unlock()

	if l.killed {
		return false, fmt.Errorf("function ShouldAllow called on an inactive instance")
	}

	currentTime := time.Now()
	currentWindowBoundary := currentTime.Sub(l.current.getStartTime())

	w := float64(l.size-currentWindowBoundary) / float64(l.size)

	currentSlidingRequests := uint64(w*float64(l.previous.count)) + l.current.count

	if currentSlidingRequests+n > l.limit {
		return false, nil
	}

	// add current request count to window of current count
	l.current.updateCount(n)
	return true, nil
}

func (l *DefaultLimiter) progressiveWindowSlider() {
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

func (l *DefaultLimiter) Kill() error {

	l.lock.Lock()
	defer l.lock.Unlock()

	if l.killed {
		return fmt.Errorf("called Kill on already killed limiter")
	}

	defer l.cancelFn()
	l.killed = true
	return nil
}

func NewDefaultLimiter(limit uint64, size time.Duration) *DefaultLimiter {
	previous := NewWindow(0, time.Unix(0, 0))
	current := NewWindow(0, time.Unix(0, 0))

	childCtx, cancelFn := context.WithCancel(context.Background())

	limiter := &DefaultLimiter{
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

type SyncLimiter struct {
	previous *Window
	current  *Window
	lock     sync.Mutex
	size     time.Duration
	limit    uint64
	killed   bool
}

func (s *SyncLimiter) getNSlidesSince(now time.Time) (time.Duration, time.Time) {
	sizeAlignedTime := now.Truncate(s.size)
	timeSinceStart := sizeAlignedTime.Sub(s.current.getStartTime())

	return timeSinceStart / s.size, sizeAlignedTime
}

func (s *SyncLimiter) ShouldAllow(n uint64) (bool, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.killed {
		return false, fmt.Errorf("function ShouldAllow called on an inactive instance")
	}

	currentTime := time.Now()

	// advance the window on demand, as this doesn't make use of goroutine.
	nSlides, alignedCurrentTime := s.getNSlidesSince(currentTime)

	// window slide shares both current and previous windows.
	if nSlides == 1 {
		s.previous.setToState(
			alignedCurrentTime.Add(-s.size),
			s.current.count,
		)

		s.current.resetToTime(
			alignedCurrentTime,
		)

	} else if nSlides > 1 {
		s.previous.resetToTime(
			alignedCurrentTime.Add(-s.size),
		)
		s.current.resetToTime(
			alignedCurrentTime,
		)
	}

	currentWindowBoundary := currentTime.Sub(s.current.getStartTime())

	w := float64(s.size-currentWindowBoundary) / float64(s.size)

	currentSlidingRequests := uint64(w*float64(s.previous.count)) + s.current.count

	if currentSlidingRequests+n > s.limit {
		return false, nil
	}

	// add current request count to window of current count
	s.current.updateCount(n)
	return true, nil
}

func (s *SyncLimiter) Kill() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.killed {
		return fmt.Errorf("called Kill on already killed limiter")
	}

	// kill is a dummy implementation for SyncLimiter,
	// because there is no need of stopping a go-routine.
	s.killed = true
	return nil
}

func NewSyncLimiter(limit uint64, size time.Duration) *SyncLimiter {

	current := NewWindow(0, time.Unix(0, 0))
	previous := NewWindow(0, time.Unix(0, 0))

	return &SyncLimiter{
		previous: previous,
		current:  current,
		lock:     sync.Mutex{},
		killed:   false,
		size:     size,
		limit:    limit,
	}
}
