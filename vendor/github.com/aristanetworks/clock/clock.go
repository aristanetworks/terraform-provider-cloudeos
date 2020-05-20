// Copyright (c) 2014 Ben Johnson

package clock

import (
	"time"
)

// Clock represents an interface to the functions in the standard library time
// package. Two implementations are available in the clock package. The first
// is a real-time clock which simply wraps the time package's functions. The
// second is a mock clock which will only make forward progress when
// programmatically adjusted.
type Clock interface {
	After(d time.Duration) <-chan time.Time
	AfterFunc(d time.Duration, f func()) Timer
	Now() time.Time
	Since(t time.Time) time.Duration
	Sleep(d time.Duration)
	Tick(d time.Duration) <-chan time.Time
	Ticker(d time.Duration) Ticker
	Timer(d time.Duration) Timer
}

// A Ticker delivers `ticks' of a clock at intervals.
type Ticker interface {
	C() <-chan time.Time
	Stop()
}

// The Timer type represents a single event.
// When the Timer expires, the current time will be sent on C(),
// unless the Timer was created by AfterFunc.
// A Timer must be created with clock.Timer or clock.AfterFunc.
type Timer interface {
	C() <-chan time.Time
	Reset(d time.Duration) bool
	Stop() bool
}

// New returns an instance of a real-time clock.
func New() Clock {
	return &clock{}
}

// clock implements a real-time clock by simply wrapping the time package functions.
type clock struct{}

func (c *clock) After(d time.Duration) <-chan time.Time { return time.After(d) }

func (c *clock) AfterFunc(d time.Duration, f func()) Timer {
	return &timer{t: time.AfterFunc(d, f)}
}

func (c *clock) Now() time.Time { return time.Now() }

func (c *clock) Since(t time.Time) time.Duration { return time.Since(t) }

func (c *clock) Sleep(d time.Duration) { time.Sleep(d) }

func (c *clock) Tick(d time.Duration) <-chan time.Time { return time.Tick(d) }

func (c *clock) Ticker(d time.Duration) Ticker {
	return &ticker{t: time.NewTicker(d)}
}

func (c *clock) Timer(d time.Duration) Timer {
	return &timer{t: time.NewTimer(d)}
}

// ticker implements the Ticker interface
type ticker struct {
	t *time.Ticker
}

func (t ticker) C() <-chan time.Time {
	return t.t.C
}

func (t ticker) Stop() {
	t.t.Stop()
}

// timer implements the timer interface
type timer struct {
	t *time.Timer
}

func (t timer) C() <-chan time.Time {
	return t.t.C
}

func (t timer) Reset(d time.Duration) bool {
	return t.t.Reset(d)
}

func (t timer) Stop() bool {
	return t.t.Stop()
}
