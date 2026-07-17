package jest

import (
	"sort"
	"sync"
	"time"
)

// Clock is a deterministic, manually-advanced replacement for real time. Instead
// of relying on the wall clock, code under test schedules work with
// [Clock.SetTimeout] and [Clock.SetInterval] (or waits on [Clock.After]) and the
// test drives time forward with [Clock.AdvanceTimersByTime], [Clock.RunAllTimers]
// or [Clock.RunOnlyPendingTimers]. A Clock is safe for concurrent use.
type Clock struct {
	mu     sync.Mutex
	now    time.Time
	nextID int
	timers []*fakeTimer
}

// fakeTimer is one scheduled callback.
type fakeTimer struct {
	id       int
	at       time.Time
	interval time.Duration // >0 for repeating timers
	fn       func()
}

// NewClock creates a Clock whose current time starts at the Unix epoch.
func NewClock() *Clock {
	return &Clock{now: time.Unix(0, 0)}
}

// NewClockAt creates a Clock whose current time starts at start.
func NewClockAt(start time.Time) *Clock {
	return &Clock{now: start}
}

// Now returns the clock's current (virtual) time.
func (c *Clock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

// SetTimeout schedules fn to run once after delay of virtual time elapses,
// returning an id that can be passed to [Clock.ClearTimer] to cancel it.
func (c *Clock) SetTimeout(delay time.Duration, fn func()) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.nextID++
	c.timers = append(c.timers, &fakeTimer{id: c.nextID, at: c.now.Add(delay), fn: fn})
	return c.nextID
}

// SetInterval schedules fn to run repeatedly every interval of virtual time,
// returning an id that can be passed to [Clock.ClearTimer] to cancel it. A
// non-positive interval is treated as a single [Clock.SetTimeout].
func (c *Clock) SetInterval(interval time.Duration, fn func()) int {
	if interval <= 0 {
		return c.SetTimeout(interval, fn)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.nextID++
	c.timers = append(c.timers, &fakeTimer{id: c.nextID, at: c.now.Add(interval), interval: interval, fn: fn})
	return c.nextID
}

// ClearTimer cancels the timer with the given id (from [Clock.SetTimeout] or
// [Clock.SetInterval]), reporting whether a pending timer was removed.
func (c *Clock) ClearTimer(id int) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, t := range c.timers {
		if t.id == id {
			c.timers = append(c.timers[:i], c.timers[i+1:]...)
			return true
		}
	}
	return false
}

// After returns a channel that receives the clock's virtual time once delay has
// elapsed, mirroring time.After for code driven by a Clock.
func (c *Clock) After(delay time.Duration) <-chan time.Time {
	ch := make(chan time.Time, 1)
	c.SetTimeout(delay, func() { ch <- c.Now() })
	return ch
}

// PendingCount returns the number of timers currently scheduled.
func (c *Clock) PendingCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.timers)
}

// AdvanceTimersByTime advances the clock by d, firing every timer whose deadline
// falls within the interval in chronological order. Repeating timers are
// rescheduled and may fire multiple times. It returns the number of callbacks
// invoked.
func (c *Clock) AdvanceTimersByTime(d time.Duration) int {
	c.mu.Lock()
	target := c.now.Add(d)
	fired := 0
	for {
		t := c.earliestDueLocked(target)
		if t == nil {
			break
		}
		c.now = t.at
		fn := t.fn
		if t.interval > 0 {
			t.at = t.at.Add(t.interval)
		} else {
			c.removeLocked(t.id)
		}
		c.mu.Unlock()
		fn()
		fired++
		c.mu.Lock()
	}
	c.now = target
	c.mu.Unlock()
	return fired
}

// RunAllTimers runs scheduled timers until none remain, advancing virtual time
// to each timer's deadline. To avoid looping forever on repeating timers it
// stops after a large safety bound and returns the number of callbacks invoked.
func (c *Clock) RunAllTimers() int {
	const safety = 100000
	fired := 0
	for fired < safety {
		c.mu.Lock()
		t := c.earliestLocked()
		if t == nil {
			c.mu.Unlock()
			break
		}
		c.now = t.at
		fn := t.fn
		if t.interval > 0 {
			t.at = t.at.Add(t.interval)
		} else {
			c.removeLocked(t.id)
		}
		c.mu.Unlock()
		fn()
		fired++
	}
	return fired
}

// RunOnlyPendingTimers runs exactly the timers scheduled at the moment of the
// call (advancing time to each), without running callbacks that those callbacks
// themselves schedule. It returns the number of callbacks invoked.
func (c *Clock) RunOnlyPendingTimers() int {
	c.mu.Lock()
	due := make([]*fakeTimer, len(c.timers))
	copy(due, c.timers)
	sort.SliceStable(due, func(i, j int) bool { return due[i].at.Before(due[j].at) })
	c.mu.Unlock()

	fired := 0
	for _, t := range due {
		c.mu.Lock()
		if !c.stillScheduledLocked(t.id) {
			c.mu.Unlock()
			continue
		}
		c.now = t.at
		fn := t.fn
		if t.interval > 0 {
			t.at = t.at.Add(t.interval)
		} else {
			c.removeLocked(t.id)
		}
		c.mu.Unlock()
		fn()
		fired++
	}
	return fired
}

// earliestDueLocked returns the earliest timer with a deadline at or before
// limit, or nil. The caller must hold c.mu.
func (c *Clock) earliestDueLocked(limit time.Time) *fakeTimer {
	t := c.earliestLocked()
	if t == nil || t.at.After(limit) {
		return nil
	}
	return t
}

// earliestLocked returns the timer with the earliest deadline, or nil. The
// caller must hold c.mu.
func (c *Clock) earliestLocked() *fakeTimer {
	var best *fakeTimer
	for _, t := range c.timers {
		if best == nil || t.at.Before(best.at) {
			best = t
		}
	}
	return best
}

// removeLocked deletes the timer with the given id. The caller must hold c.mu.
func (c *Clock) removeLocked(id int) {
	for i, t := range c.timers {
		if t.id == id {
			c.timers = append(c.timers[:i], c.timers[i+1:]...)
			return
		}
	}
}

// stillScheduledLocked reports whether a timer id is still pending. The caller
// must hold c.mu.
func (c *Clock) stillScheduledLocked(id int) bool {
	for _, t := range c.timers {
		if t.id == id {
			return true
		}
	}
	return false
}
