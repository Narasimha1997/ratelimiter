package core

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

func NewWindow(count uint64, startTime time.Time) *Window {
	nsTime := startTime.UnixNano()
	return &Window{
		count:     count,
		startTime: int64(nsTime),
	}
}
