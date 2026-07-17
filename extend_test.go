package jest

import (
	"fmt"
	"testing"
)

// fakeT extends fakeReporter with the Cleanup/Name surface of *testing.T so the
// assertion-counting helpers can be exercised deterministically.
type fakeT struct {
	fakeReporter
	name     string
	cleanups []func()
}

func (f *fakeT) Cleanup(fn func()) { f.cleanups = append(f.cleanups, fn) }
func (f *fakeT) Name() string      { return f.name }
func (f *fakeT) runCleanups() {
	for i := len(f.cleanups) - 1; i >= 0; i-- {
		f.cleanups[i]()
	}
}

func TestExtendCustomMatcher(t *testing.T) {
	Extend(map[string]CustomMatcher{
		"ToBeEven": func(actual any, _ ...any) MatcherResult {
			n, ok := actual.(int)
			return MatcherResult{Pass: ok && n%2 == 0, Message: fmt.Sprintf("expected %v to be even", actual)}
		},
		"ToBeWithin": func(actual any, args ...any) MatcherResult {
			n := actual.(int)
			lo, hi := args[0].(int), args[1].(int)
			return MatcherResult{Pass: n >= lo && n <= hi, Message: fmt.Sprintf("expected %d within [%d,%d]", n, lo, hi)}
		},
	})

	expectPass(t, "even", func(r TestReporter) { Expect(r, 4).To("ToBeEven") })
	expectFail(t, "odd", func(r TestReporter) { Expect(r, 3).To("ToBeEven") })
	expectPass(t, "not odd", func(r TestReporter) { Expect(r, 3).Not().To("ToBeEven") })
	expectPass(t, "within", func(r TestReporter) { Expect(r, 5).To("ToBeWithin", 1, 10) })
	expectFail(t, "outside", func(r TestReporter) { Expect(r, 50).To("ToBeWithin", 1, 10) })

	msg := expectFail(t, "unknown", func(r TestReporter) { Expect(r, 1).To("NoSuchMatcher") })
	if msg == "" {
		t.Error("expected error for unknown matcher")
	}
}

func TestAssertionsExact(t *testing.T) {
	// Correct count passes.
	f := &fakeT{name: "TestAssertionsExact/ok"}
	Assertions(f, 2)
	Expect(f, 1).ToBe(1)
	Expect(f, 2).ToBe(2)
	f.runCleanups()
	if f.failed() {
		t.Errorf("expected no failure, got: %s", f.message())
	}

	// Wrong count fails at cleanup.
	f2 := &fakeT{name: "TestAssertionsExact/bad"}
	Assertions(f2, 3)
	Expect(f2, 1).ToBe(1)
	f2.runCleanups()
	if !f2.failed() {
		t.Error("expected failure for wrong assertion count")
	}
}

func TestHasAssertions(t *testing.T) {
	f := &fakeT{name: "TestHasAssertions/ok"}
	HasAssertions(f)
	Expect(f, 1).ToBe(1)
	f.runCleanups()
	if f.failed() {
		t.Errorf("expected no failure, got: %s", f.message())
	}

	f2 := &fakeT{name: "TestHasAssertions/none"}
	HasAssertions(f2)
	f2.runCleanups()
	if !f2.failed() {
		t.Error("expected failure when no assertions ran")
	}
}
