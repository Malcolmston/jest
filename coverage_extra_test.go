package jest

import (
	"reflect"
	"testing"
)

func TestObjectContainingUnexportedField(t *testing.T) {
	type rec struct {
		Name   string
		secret int
	}
	r := rec{Name: "bob", secret: 42}
	// Matching an unexported field exercises the forceInterface path.
	expectPass(t, "unexported field", func(rp TestReporter) {
		Expect[any](rp, r).ToEqual(ObjectContaining(map[string]any{"secret": 42}))
	})
	expectFail(t, "unexported field mismatch", func(rp TestReporter) {
		Expect[any](rp, r).ToEqual(ObjectContaining(map[string]any{"secret": 7}))
	})
}

func TestToBeInstanceOfReflectType(t *testing.T) {
	expectPass(t, "reflect.Type target", func(r TestReporter) {
		Expect[any](r, "hi").ToBeInstanceOf(reflect.TypeOf(""))
	})
	msg := expectFail(t, "nil target with nil actual", func(r TestReporter) {
		Expect[any](r, nil).ToBeInstanceOf(nil)
	})
	if msg == "" {
		t.Error("expected failure message")
	}
}

func TestSpyOnNilPointerArg(t *testing.T) {
	deref := func(p *int) int {
		if p == nil {
			return -1
		}
		return *p
	}
	spy := SpyOn(&deref)
	defer spy.Restore()
	// Passing a nil pointer exercises callReflect's zero-value substitution and
	// paramType.
	Expect(t, deref(nil)).ToBe(-1)
	Expect(t, spy).ToHaveBeenCalledWith((*int)(nil))
	n := 5
	Expect(t, deref(&n)).ToBe(5)
}

func TestClockClearDuringRun(t *testing.T) {
	c := NewClock()
	var id2 int
	c.SetTimeout(10, func() { c.ClearTimer(id2) })
	id2 = c.SetTimeout(20, func() { t.Error("cleared timer should not fire") })
	c.RunOnlyPendingTimers()
}
