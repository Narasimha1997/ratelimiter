package ratelimiter

import (
	"time"
)

// Window represents the structure of timing-window at given point of time.
type Window struct {
	count     uint64
	startTime time.Time
}

func (w *Window) updateCount(n uint64) {
	w.count += n
}

func (w *Window) getStartTime() time.Time {
	return w.startTime
}

func (w *Window) setStateFrom(other *Window) {
	w.count = other.count
	w.startTime = other.startTime
}

func (w *Window) resetToTime(startTime time.Time) {
	w.count = 0
	w.startTime = startTime
}

func (w *Window) setToState(startTime time.Time, count uint64) {
	w.startTime = startTime
	w.count = count
}

// Creates and returns a pointer to the new Window instance.
//
// Parameters:
//
// 1. count: The initial count of the window.
//
// 2. startTime: The initial starting time of the window.
func NewWindow(count uint64, startTime time.Time) *Window {

	return &Window{
		count:     count,
		startTime: startTime,
	}
}
