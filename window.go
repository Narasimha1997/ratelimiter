package ratelimiter

import (
	"time"
)

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

func NewWindow(count uint64, startTime time.Time) *Window {

	return &Window{
		count:     count,
		startTime: startTime,
	}
}
