package jest

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// fakeReporter records failures instead of failing a test, so both the passing
// and failing branch of every matcher can be exercised deterministically.
type fakeReporter struct {
	errors []string
	fatals []string
}

func (f *fakeReporter) Errorf(format string, args ...any) {
	f.errors = append(f.errors, fmt.Sprintf(format, args...))
}

func (f *fakeReporter) Fatalf(format string, args ...any) {
	f.fatals = append(f.fatals, fmt.Sprintf(format, args...))
}

func (f *fakeReporter) Helper() {}

func (f *fakeReporter) failed() bool { return len(f.errors)+len(f.fatals) > 0 }

func (f *fakeReporter) message() string {
	return strings.Join(append(append([]string{}, f.errors...), f.fatals...), "\n")
}

// expectPass runs an assertion and fails the surrounding test if the matcher
// reported any failure.
func expectPass(t *testing.T, name string, run func(r TestReporter)) {
	t.Helper()
	r := &fakeReporter{}
	run(r)
	if r.failed() {
		t.Errorf("%s: expected assertion to PASS, but it reported:\n%s", name, r.message())
	}
}

// expectFail runs an assertion and fails the surrounding test unless the matcher
// reported a failure. It returns the recorded message for further inspection.
func expectFail(t *testing.T, name string, run func(r TestReporter)) string {
	t.Helper()
	r := &fakeReporter{}
	run(r)
	if !r.failed() {
		t.Errorf("%s: expected assertion to FAIL, but it passed", name)
	}
	return r.message()
}

func TestToBe(t *testing.T) {
	expectPass(t, "int equal", func(r TestReporter) { Expect(r, 5).ToBe(5) })
	expectPass(t, "string equal", func(r TestReporter) { Expect(r, "hi").ToBe("hi") })
	expectFail(t, "int not equal", func(r TestReporter) { Expect(r, 5).ToBe(6) })
	expectPass(t, "not different", func(r TestReporter) { Expect(r, 5).Not().ToBe(6) })
	expectFail(t, "not same", func(r TestReporter) { Expect(r, 5).Not().ToBe(5) })

	// Pointer identity semantics.
	a, b := 1, 1
	expectPass(t, "same ptr", func(r TestReporter) { Expect(r, &a).ToBe(&a) })
	expectFail(t, "different ptr", func(r TestReporter) { Expect(r, &a).ToBe(&b) })
}

func TestToEqual(t *testing.T) {
	expectPass(t, "slice deep", func(r TestReporter) { Expect(r, []int{1, 2, 3}).ToEqual([]int{1, 2, 3}) })
	expectPass(t, "map deep", func(r TestReporter) {
		Expect(r, map[string]int{"a": 1}).ToEqual(map[string]int{"a": 1})
	})
	msg := expectFail(t, "slice diff", func(r TestReporter) { Expect(r, []int{1, 2, 3}).ToEqual([]int{1, 9, 3}) })
	if !strings.Contains(msg, "diff:") {
		t.Errorf("expected slice diff output, got: %s", msg)
	}

	type point struct {
		X, Y int
	}
	msg = expectFail(t, "struct diff", func(r TestReporter) {
		Expect(r, point{1, 2}).ToEqual(point{1, 3})
	})
	if !strings.Contains(msg, "Y:") {
		t.Errorf("expected struct field diff, got: %s", msg)
	}
	msg = expectFail(t, "map diff", func(r TestReporter) {
		Expect(r, map[string]int{"a": 1, "b": 2}).ToEqual(map[string]int{"a": 1, "c": 2})
	})
	if !strings.Contains(msg, "missing") || !strings.Contains(msg, "unexpected") {
		t.Errorf("expected map key diff, got: %s", msg)
	}
	expectPass(t, "not deep equal", func(r TestReporter) { Expect(r, []int{1}).Not().ToEqual([]int{2}) })
}

func TestToBeNil(t *testing.T) {
	var p *int
	expectPass(t, "nil ptr", func(r TestReporter) { Expect(r, p).ToBeNil() })
	expectPass(t, "nil any", func(r TestReporter) { Expect[any](r, nil).ToBeNil() })
	var s []int
	expectPass(t, "nil slice", func(r TestReporter) { Expect(r, s).ToBeNil() })
	expectFail(t, "non nil", func(r TestReporter) { Expect(r, 5).ToBeNil() })
	x := 3
	expectPass(t, "not nil ptr", func(r TestReporter) { Expect(r, &x).Not().ToBeNil() })
}

func TestToBeTrueFalse(t *testing.T) {
	expectPass(t, "true", func(r TestReporter) { Expect(r, true).ToBeTrue() })
	expectFail(t, "true fail", func(r TestReporter) { Expect(r, false).ToBeTrue() })
	expectPass(t, "false", func(r TestReporter) { Expect(r, false).ToBeFalse() })
	expectFail(t, "false fail", func(r TestReporter) { Expect(r, true).ToBeFalse() })
	expectFail(t, "non-bool true", func(r TestReporter) { Expect(r, 1).ToBeTrue() })
}

func TestNumericComparisons(t *testing.T) {
	expectPass(t, "gt", func(r TestReporter) { Expect(r, 5).ToBeGreaterThan(3) })
	expectFail(t, "gt fail", func(r TestReporter) { Expect(r, 3).ToBeGreaterThan(5) })
	expectPass(t, "gte", func(r TestReporter) { Expect(r, 5).ToBeGreaterThanOrEqual(5) })
	expectFail(t, "gte fail", func(r TestReporter) { Expect(r, 4).ToBeGreaterThanOrEqual(5) })
	expectPass(t, "lt", func(r TestReporter) { Expect(r, 3).ToBeLessThan(5) })
	expectFail(t, "lt fail", func(r TestReporter) { Expect(r, 5).ToBeLessThan(3) })
	expectPass(t, "lte", func(r TestReporter) { Expect(r, 5).ToBeLessThanOrEqual(5) })
	expectFail(t, "lte fail", func(r TestReporter) { Expect(r, 6).ToBeLessThanOrEqual(5) })
	expectPass(t, "float gt", func(r TestReporter) { Expect(r, 2.5).ToBeGreaterThan(2.4) })

	// Non-numeric operands report a distinct error.
	msg := expectFail(t, "non numeric", func(r TestReporter) { Expect(r, "x").ToBeGreaterThan("y") })
	if !strings.Contains(msg, "numeric") {
		t.Errorf("expected numeric error, got %s", msg)
	}
}

func TestToContain(t *testing.T) {
	expectPass(t, "substr", func(r TestReporter) { Expect(r, "hello world").ToContain("world") })
	expectFail(t, "substr fail", func(r TestReporter) { Expect(r, "hello").ToContain("z") })
	expectPass(t, "slice elem", func(r TestReporter) { Expect(r, []int{1, 2, 3}).ToContain(2) })
	expectFail(t, "slice elem fail", func(r TestReporter) { Expect(r, []int{1, 2, 3}).ToContain(9) })
	expectPass(t, "map key", func(r TestReporter) { Expect(r, map[string]int{"a": 1}).ToContain("a") })
	expectFail(t, "map key fail", func(r TestReporter) { Expect(r, map[string]int{"a": 1}).ToContain("b") })
	expectPass(t, "not contain", func(r TestReporter) { Expect(r, "abc").Not().ToContain("z") })
	// Wrong map key type.
	expectFail(t, "map wrong key", func(r TestReporter) { Expect(r, map[string]int{"a": 1}).ToContain(1) })
}

func TestToHaveLen(t *testing.T) {
	expectPass(t, "slice len", func(r TestReporter) { Expect(r, []int{1, 2, 3}).ToHaveLen(3) })
	expectPass(t, "string len", func(r TestReporter) { Expect(r, "abcd").ToHaveLen(4) })
	expectPass(t, "map len", func(r TestReporter) { Expect(r, map[int]int{1: 1, 2: 2}).ToHaveLen(2) })
	expectFail(t, "len fail", func(r TestReporter) { Expect(r, []int{1}).ToHaveLen(3) })
	msg := expectFail(t, "len wrong type", func(r TestReporter) { Expect(r, 5).ToHaveLen(1) })
	if !strings.Contains(msg, "requires") {
		t.Errorf("expected type error, got %s", msg)
	}
}

func TestToMatch(t *testing.T) {
	expectPass(t, "match", func(r TestReporter) { Expect(r, "abc123").ToMatch(`\d+`) })
	expectFail(t, "match fail", func(r TestReporter) { Expect(r, "abc").ToMatch(`\d+`) })
	msg := expectFail(t, "bad pattern", func(r TestReporter) { Expect(r, "abc").ToMatch(`[`) })
	if !strings.Contains(msg, "invalid pattern") {
		t.Errorf("expected invalid pattern error, got %s", msg)
	}
	expectPass(t, "not match", func(r TestReporter) { Expect(r, "abc").Not().ToMatch(`\d`) })
}

func TestToBeCloseTo(t *testing.T) {
	expectPass(t, "close default", func(r TestReporter) { Expect(r, 1.0000000001).ToBeCloseTo(1.0) })
	expectPass(t, "close eps", func(r TestReporter) { Expect(r, 3.14159).ToBeCloseTo(3.14, 0.01) })
	expectFail(t, "not close", func(r TestReporter) { Expect(r, 3.2).ToBeCloseTo(3.14, 0.01) })
	msg := expectFail(t, "non numeric close", func(r TestReporter) { Expect(r, "x").ToBeCloseTo(1.0) })
	if !strings.Contains(msg, "numeric") {
		t.Errorf("expected numeric error, got %s", msg)
	}
}

func TestToPanicAndToThrow(t *testing.T) {
	boom := func() { panic("boom") }
	safe := func() {}
	expectPass(t, "panics", func(r TestReporter) { Expect(r, boom).ToPanic() })
	expectFail(t, "no panic", func(r TestReporter) { Expect(r, safe).ToPanic() })
	expectPass(t, "panic msg", func(r TestReporter) { Expect(r, boom).ToPanic("boom") })
	expectFail(t, "panic wrong msg", func(r TestReporter) { Expect(r, boom).ToPanic("nope") })
	expectPass(t, "throw alias", func(r TestReporter) { Expect(r, boom).ToThrow() })
	expectPass(t, "not panic", func(r TestReporter) { Expect(r, safe).Not().ToPanic() })
	// Non-func value.
	msg := expectFail(t, "not a func", func(r TestReporter) { Expect(r, 5).ToPanic() })
	if !strings.Contains(msg, "func()") {
		t.Errorf("expected func requirement error, got %s", msg)
	}
}

func TestFormatAndHelpers(t *testing.T) {
	if got := format(nil); got != "<nil>" {
		t.Errorf("format(nil) = %q", got)
	}
	if got := format("x"); got != `"x"` {
		t.Errorf("format string = %q", got)
	}
	if got := format(errors.New("boom")); got != `"boom"` {
		t.Errorf("format error = %q", got)
	}
	if !isNil((func())(nil)) {
		t.Error("expected nil func to be nil")
	}
	if _, ok := toFloat(uint8(5)); !ok {
		t.Error("expected uint8 to convert")
	}
}

func TestMockBasics(t *testing.T) {
	m := NewMock("adder")
	m.Return(42)
	if got := m.Call(1, 2); len(got) != 1 || got[0] != 42 {
		t.Errorf("unexpected results %v", got)
	}
	m.Call(3, 4)
	Expect(t, m.Name()).ToBe("adder")
	Expect(t, m.CallCount()).ToBe(2)
	Expect(t, m.Called()).ToBeTrue()
	Expect(t, m.CalledWith(1, 2)).ToBeTrue()
	Expect(t, m.CalledWith(9, 9)).ToBeFalse()

	last, ok := m.LastCall()
	Expect(t, ok).ToBeTrue()
	Expect(t, last.Args).ToEqual([]any{3, 4})

	first, ok := m.NthCall(0)
	Expect(t, ok).ToBeTrue()
	Expect(t, first.Args).ToEqual([]any{1, 2})
	_, ok = m.NthCall(99)
	Expect(t, ok).ToBeFalse()

	Expect(t, len(m.Calls())).ToBe(2)

	m.Reset()
	Expect(t, m.CallCount()).ToBe(0)
	Expect(t, m.Called()).ToBeFalse()
	_, ok = m.LastCall()
	Expect(t, ok).ToBeFalse()
}

func TestMockReturnValues(t *testing.T) {
	m := NewMock("seq")
	m.ReturnValues([]any{"a"}, []any{"b"})
	Expect(t, m.Call()[0]).ToBe("a")
	Expect(t, m.Call()[0]).ToBe("b")
	// Exhausted -> repeats last entry.
	Expect(t, m.Call()[0]).ToBe("b")

	// Sequence then fall back to Return default.
	m2 := NewMock("seq2")
	m2.Return("default").ReturnValues([]any{"once"})
	Expect(t, m2.Call()[0]).ToBe("once")
	Expect(t, m2.Call()[0]).ToBe("default")

	// Unconfigured mock returns nil results.
	m3 := NewMock("empty")
	Expect(t, m3.Call()).ToBeNil()
}

func TestTypedFn(t *testing.T) {
	fn0, m0 := Fn0[int]("zero")
	m0.Return(7)
	Expect(t, fn0()).ToBe(7)
	Expect(t, m0.CallCount()).ToBe(1)

	fn1, m1 := Fn1[int, string]("one")
	m1.Return("hi")
	Expect(t, fn1(5)).ToBe("hi")
	Expect(t, m1.CalledWith(5)).ToBeTrue()

	fn2, m2 := Fn2[int, int, int]("two")
	m2.ReturnValues([]any{10}, []any{20})
	Expect(t, fn2(1, 2)).ToBe(10)
	Expect(t, fn2(3, 4)).ToBe(20)
	Expect(t, m2.CallCount()).ToBe(2)

	// Unconfigured typed mock returns the zero value.
	fnZero, _ := Fn1[int, string]("zeroval")
	Expect(t, fnZero(1)).ToBe("")
}

func TestSpies(t *testing.T) {
	calls := 0
	s0, m0 := Spy0("greet", func() string { calls++; return "hello" })
	Expect(t, s0()).ToBe("hello")
	Expect(t, calls).ToBe(1)
	Expect(t, m0.CallCount()).ToBe(1)

	s1, m1 := Spy1("double", func(x int) int { return x * 2 })
	Expect(t, s1(21)).ToBe(42)
	Expect(t, m1.CalledWith(21)).ToBeTrue()
	last, _ := m1.LastCall()
	Expect(t, last.Results).ToEqual([]any{42})

	s2, m2 := Spy2("add", func(a, b int) int { return a + b })
	Expect(t, s2(2, 3)).ToBe(5)
	Expect(t, m2.CalledWith(2, 3)).ToBeTrue()
}

func TestCastResult(t *testing.T) {
	Expect(t, castResult[int]([]any{5}, 0)).ToBe(5)
	Expect(t, castResult[int]([]any{}, 0)).ToBe(0)        // missing
	Expect(t, castResult[int]([]any{"x"}, 0)).ToBe(0)     // wrong type
	Expect(t, castResult[string]([]any{nil}, 0)).ToBe("") // nil entry
	Expect(t, castResult[int]([]any{1, 2}, 5)).ToBe(0)    // out of range
}

func TestDescribeItHooks(t *testing.T) {
	var order []string
	Describe(t, "outer", func() {
		BeforeEach(func() { order = append(order, "before-outer") })
		AfterEach(func() { order = append(order, "after-outer") })

		It(t, "case one", func(t *testing.T) {
			order = append(order, "case-one")
			Expect(t, 1).ToBe(1)
		})

		Describe(t, "inner", func() {
			BeforeEach(func() { order = append(order, "before-inner") })
			It(t, "case two", func(t *testing.T) {
				order = append(order, "case-two")
			})
		})
	})

	joined := strings.Join(order, ",")
	want := "before-outer,case-one,after-outer,before-outer,before-inner,case-two,after-outer"
	if joined != want {
		t.Errorf("hook order wrong:\n got: %s\nwant: %s", joined, want)
	}
}

func TestTestAlias(t *testing.T) {
	ran := false
	Test(t, "alias runs", func(t *testing.T) { ran = true })
	Expect(t, ran).ToBeTrue()
}

func TestHooksOutsideScopePanic(t *testing.T) {
	Expect(t, func() { BeforeEach(func() {}) }).ToPanic("Describe")
	Expect(t, func() { AfterEach(func() {}) }).ToPanic("Describe")
}

func TestNegatedMessages(t *testing.T) {
	msg := expectFail(t, "negated msg", func(r TestReporter) { Expect(r, 5).Not().ToBe(5) })
	if !strings.Contains(msg, "NOT") {
		t.Errorf("expected NOT in negated message, got %s", msg)
	}
}
