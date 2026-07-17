package jest

import (
	"fmt"
	"sync"
)

// MatcherResult is returned by a [CustomMatcher] to report whether the
// assertion passed and to supply the message used on failure.
type MatcherResult struct {
	// Pass reports whether the assertion succeeded.
	Pass bool
	// Message is the failure message (used when the assertion fails, or when it
	// unexpectedly passes under negation).
	Message string
}

// CustomMatcher implements a user-defined matcher registered with [Extend] and
// invoked through [Matcher.To]. It receives the actual value under test and any
// extra arguments passed at the call site.
type CustomMatcher func(actual any, args ...any) MatcherResult

var (
	customMu       sync.RWMutex
	customMatchers = map[string]CustomMatcher{}
)

// Extend registers custom matchers by name, mirroring Jest's expect.extend.
// Registered matchers are invoked with [Matcher.To]. Calling Extend again with
// an existing name replaces that matcher.
func Extend(matchers map[string]CustomMatcher) {
	customMu.Lock()
	defer customMu.Unlock()
	for name, fn := range matchers {
		customMatchers[name] = fn
	}
}

// To invokes a custom matcher previously registered with [Extend], passing the
// value under test and any extra arguments. Negation via [Matcher.Not] is
// honored. It reports a failure if no matcher with the given name is registered.
func (m *Matcher[T]) To(name string, args ...any) {
	m.t.Helper()
	countAssertion(m.t)
	customMu.RLock()
	fn, ok := customMatchers[name]
	customMu.RUnlock()
	if !ok {
		m.t.Errorf("assertion failed: no custom matcher named %q is registered", name)
		return
	}
	res := fn(any(m.actual), args...)
	pass := res.Pass
	if m.negated {
		pass = !pass
	}
	if pass {
		return
	}
	msg := res.Message
	if msg == "" {
		msg = fmt.Sprintf("custom matcher %q failed for %s", name, format(m.actual))
	}
	m.t.Errorf("assertion failed: %s", msg)
}

// ---- assertion counting -----------------------------------------------------

type assertCounter struct {
	count    int
	expected int // -1 means "no exact expectation"
	atLeast  bool
	set      bool
}

var (
	assertMu       sync.Mutex
	assertCounters = map[TestReporter]*assertCounter{}
)

// countAssertion records that an assertion ran against t, for [Assertions] and
// [HasAssertions].
func countAssertion(t TestReporter) {
	assertMu.Lock()
	c := assertCounters[t]
	if c != nil {
		c.count++
	}
	assertMu.Unlock()
}

// cleanuper is the subset of *testing.T used to schedule verification of an
// expected assertion count when the test finishes.
type cleanuper interface {
	Helper()
	Cleanup(func())
	Errorf(format string, args ...any)
}

// Assertions declares that exactly n assertions are expected to run during the
// test, mirroring Jest's expect.assertions(n). The count is verified when the
// test finishes. t must provide a Cleanup method (as *testing.T does).
func Assertions(t cleanuper, n int) {
	t.Helper()
	registerCounter(t, n, false)
}

// HasAssertions declares that at least one assertion is expected to run during
// the test, mirroring Jest's expect.hasAssertions. The expectation is verified
// when the test finishes. t must provide a Cleanup method (as *testing.T does).
func HasAssertions(t cleanuper) {
	t.Helper()
	registerCounter(t, 1, true)
}

// registerCounter installs (or updates) an assertion counter for t and, on the
// first registration, schedules verification via t.Cleanup.
func registerCounter(t cleanuper, expected int, atLeast bool) {
	r, ok := t.(TestReporter)
	if !ok {
		return
	}
	assertMu.Lock()
	c := assertCounters[r]
	first := c == nil
	if first {
		c = &assertCounter{}
		assertCounters[r] = c
	}
	c.expected = expected
	c.atLeast = atLeast
	c.set = true
	assertMu.Unlock()

	if !first {
		return
	}
	t.Cleanup(func() {
		assertMu.Lock()
		cc := assertCounters[r]
		delete(assertCounters, r)
		assertMu.Unlock()
		if cc == nil || !cc.set {
			return
		}
		if cc.atLeast {
			if cc.count < cc.expected {
				t.Errorf("expected at least %d assertion(s), but %d ran", cc.expected, cc.count)
			}
			return
		}
		if cc.count != cc.expected {
			t.Errorf("expected exactly %d assertion(s), but %d ran", cc.expected, cc.count)
		}
	})
}
