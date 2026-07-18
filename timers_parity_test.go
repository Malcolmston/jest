package jest

import (
	"testing"
	"time"
)

func TestAdvanceTimersToNextTimer(t *testing.T) {
	c := NewClock()
	var order []int
	c.SetTimeout(10*time.Millisecond, func() { order = append(order, 1) })
	c.SetTimeout(20*time.Millisecond, func() { order = append(order, 2) })
	c.SetTimeout(30*time.Millisecond, func() { order = append(order, 3) })

	if n := c.AdvanceTimersToNextTimer(); n != 1 {
		t.Errorf("first step fired %d, want 1", n)
	}
	if len(order) != 1 || order[0] != 1 {
		t.Errorf("order=%v, want [1]", order)
	}
	if !c.Now().Equal(time.Unix(0, 0).Add(10 * time.Millisecond)) {
		t.Errorf("time not advanced to first timer: %v", c.Now())
	}
	if n := c.AdvanceTimersToNextTimer(2); n != 2 {
		t.Errorf("two steps fired %d, want 2", n)
	}
	if len(order) != 3 {
		t.Errorf("order=%v, want 3 entries", order)
	}
	// No timers remain.
	if n := c.AdvanceTimersToNextTimer(5); n != 0 {
		t.Errorf("expected 0 fired when empty, got %d", n)
	}
}

func TestAdvanceTimersToNextTimerRepeating(t *testing.T) {
	c := NewClock()
	count := 0
	c.SetInterval(5*time.Millisecond, func() { count++ })
	c.AdvanceTimersToNextTimer(3)
	if count != 3 {
		t.Errorf("repeating fired %d, want 3", count)
	}
}

func TestClearAllTimers(t *testing.T) {
	c := NewClock()
	c.SetTimeout(time.Second, func() {})
	c.SetTimeout(time.Second, func() {})
	if c.GetTimerCount() != 2 {
		t.Fatalf("GetTimerCount=%d, want 2", c.GetTimerCount())
	}
	if n := c.ClearAllTimers(); n != 2 {
		t.Errorf("ClearAllTimers removed %d, want 2", n)
	}
	if c.GetTimerCount() != 0 {
		t.Errorf("timers remain after clear: %d", c.GetTimerCount())
	}
}

func TestSetSystemTime(t *testing.T) {
	c := NewClock()
	target := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	c.SetSystemTime(target)
	if !c.Now().Equal(target) {
		t.Errorf("Now=%v, want %v", c.Now(), target)
	}
	// Setting time does not fire timers.
	fired := false
	c.SetSystemTime(time.Unix(0, 0))
	c.SetTimeout(10*time.Millisecond, func() { fired = true })
	c.SetSystemTime(time.Unix(100, 0))
	if fired {
		t.Error("SetSystemTime should not fire timers")
	}
}

func BenchmarkAdvanceTimersToNextTimer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		c := NewClock()
		c.SetInterval(time.Millisecond, func() {})
		c.AdvanceTimersToNextTimer(10)
	}
}
