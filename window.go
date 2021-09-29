package ratelimit

import (
	"time"
)

type Window struct {
	count     uint64
	startTime int64
}

func (w *Window) updateCount(n uint64) {
	w.count += n
}

func (w *Window) getStartTime() time.Time {
	return time.Unix(0, w.startTime)
}

func (w *Window) setStateFrom(other *Window) {
	w.count = other.count
	w.startTime = other.startTime
}

func (w *Window) resetToTime(startTime time.Time) {
	nsTime := startTime.UnixNano()
	w.count = 0
	w.startTime = nsTime
}

func NewWindow(count uint64, startTime time.Time) *Window {
	nsTime := startTime.UnixNano()
	return &Window{
		count:     count,
		startTime: int64(nsTime),
	}
}
