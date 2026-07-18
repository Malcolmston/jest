package jest

import "time"

// AdvanceTimersToNextTimer advances virtual time to the deadline of the next
// scheduled timer and fires it, repeating for the given number of steps (default
// 1), mirroring Jest's advanceTimersToNextTimer. Repeating timers are
// rescheduled and may be fired again by a later step. It stops early if no
// timers remain and returns the number of callbacks invoked.
func (c *Clock) AdvanceTimersToNextTimer(steps ...int) int {
	n := 1
	if len(steps) > 0 {
		n = steps[0]
	}
	fired := 0
	for i := 0; i < n; i++ {
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

// ClearAllTimers cancels every pending timer without firing it, mirroring Jest's
// clearAllTimers. It returns the number of timers removed.
func (c *Clock) ClearAllTimers() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	n := len(c.timers)
	c.timers = nil
	return n
}

// GetTimerCount returns the number of timers currently scheduled, mirroring
// Jest's getTimerCount. It is the Jest spelling of [Clock.PendingCount].
func (c *Clock) GetTimerCount() int {
	return c.PendingCount()
}

// SetSystemTime sets the clock's current virtual time to t without firing any
// timers, mirroring Jest's setSystemTime. Timers scheduled relative to the old
// time keep their absolute deadlines, so moving time forward past a deadline and
// then advancing will fire them.
func (c *Clock) SetSystemTime(t time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = t
}
