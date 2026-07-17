package jest

import (
	"testing"
	"time"
)

func TestClockSetTimeout(t *testing.T) {
	c := NewClock()
	fired := 0
	c.SetTimeout(100*time.Millisecond, func() { fired++ })
	Expect(t, c.PendingCount()).ToBe(1)

	c.AdvanceTimersByTime(50 * time.Millisecond)
	Expect(t, fired).ToBe(0)
	c.AdvanceTimersByTime(60 * time.Millisecond)
	Expect(t, fired).ToBe(1)
	Expect(t, c.PendingCount()).ToBe(0)
	Expect(t, c.Now()).ToEqual(time.Unix(0, 0).Add(110 * time.Millisecond))
}

func TestClockOrdering(t *testing.T) {
	c := NewClock()
	var order []int
	c.SetTimeout(30*time.Millisecond, func() { order = append(order, 3) })
	c.SetTimeout(10*time.Millisecond, func() { order = append(order, 1) })
	c.SetTimeout(20*time.Millisecond, func() { order = append(order, 2) })
	n := c.AdvanceTimersByTime(100 * time.Millisecond)
	Expect(t, n).ToBe(3)
	Expect(t, order).ToEqual([]int{1, 2, 3})
}

func TestClockInterval(t *testing.T) {
	c := NewClock()
	ticks := 0
	id := c.SetInterval(10*time.Millisecond, func() { ticks++ })
	c.AdvanceTimersByTime(35 * time.Millisecond)
	Expect(t, ticks).ToBe(3)
	Expect(t, c.ClearTimer(id)).ToBeTrue()
	c.AdvanceTimersByTime(100 * time.Millisecond)
	Expect(t, ticks).ToBe(3)
	Expect(t, c.ClearTimer(id)).ToBeFalse()
}

func TestClockRunAllTimers(t *testing.T) {
	c := NewClock()
	done := false
	c.SetTimeout(10*time.Millisecond, func() {
		c.SetTimeout(10*time.Millisecond, func() { done = true })
	})
	fired := c.RunAllTimers()
	Expect(t, fired).ToBe(2)
	Expect(t, done).ToBeTrue()
}

func TestClockRunOnlyPendingTimers(t *testing.T) {
	c := NewClock()
	nested := false
	c.SetTimeout(10*time.Millisecond, func() {
		c.SetTimeout(10*time.Millisecond, func() { nested = true })
	})
	fired := c.RunOnlyPendingTimers()
	Expect(t, fired).ToBe(1)
	Expect(t, nested).ToBeFalse() // the callback scheduled by the first timer did not run
	Expect(t, c.PendingCount()).ToBe(1)
}

func TestClockAfter(t *testing.T) {
	c := NewClockAt(time.Unix(1000, 0))
	ch := c.After(5 * time.Second)
	select {
	case <-ch:
		t.Fatal("channel fired before advance")
	default:
	}
	c.AdvanceTimersByTime(5 * time.Second)
	got := <-ch
	Expect(t, got).ToEqual(time.Unix(1005, 0))
}

func TestClockSetIntervalNonPositive(t *testing.T) {
	c := NewClock()
	ran := false
	c.SetInterval(0, func() { ran = true })
	c.RunAllTimers()
	Expect(t, ran).ToBeTrue()
}
