package jest

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// Matcher is a fluent assertion builder produced by [Expect]. It carries the
// actual value under test, the reporter that failures are sent to, and a flag
// indicating whether the assertion has been negated with [Matcher.Not].
type Matcher[T any] struct {
	t       TestReporter
	actual  T
	negated bool
}

// Expect begins a fluent assertion for the given actual value, reporting any
// failures through t (typically a *testing.T).
func Expect[T any](t TestReporter, actual T) *Matcher[T] {
	return &Matcher[T]{t: t, actual: actual}
}

// Not returns a new matcher that inverts the sense of the assertion that
// follows it. For example jest.Expect(t, 1).Not().ToBe(2) passes.
func (m *Matcher[T]) Not() *Matcher[T] {
	return &Matcher[T]{t: m.t, actual: m.actual, negated: !m.negated}
}

// check evaluates a computed pass/fail result, applying negation, and reports a
// failure through the reporter when the assertion does not hold. expectation is
// a human-readable phrase describing what was expected, e.g.
// "to be 5, but got 3".
func (m *Matcher[T]) check(pass bool, expectation string) {
	m.t.Helper()
	countAssertion(m.t)
	if m.negated {
		pass = !pass
	}
	if pass {
		return
	}
	if m.negated {
		m.t.Errorf("assertion failed: expected value NOT %s", expectation)
	} else {
		m.t.Errorf("assertion failed: expected value %s", expectation)
	}
}

// ToBe asserts shallow equality using == semantics. For pointers this compares
// identity; for structs and other comparable values it compares by value.
// Non-comparable values fall back to a deep comparison.
func (m *Matcher[T]) ToBe(expected T) {
	m.t.Helper()
	m.check(shallowEqual(m.actual, expected),
		fmt.Sprintf("to be %s, but got %s", format(expected), format(m.actual)))
}

// ToEqual asserts deep equality using reflect.DeepEqual. On failure it includes
// a small structural diff between the expected and actual values.
func (m *Matcher[T]) ToEqual(expected T) {
	m.t.Helper()
	pass := asymEqual(expected, m.actual)
	expectation := fmt.Sprintf("to deeply equal %s, but got %s", format(expected), format(m.actual))
	if !pass && !m.negated && !hasAsymmetric(expected) {
		if d := diff(expected, m.actual); d != "" {
			expectation += "\n" + d
		}
	}
	m.check(pass, expectation)
}

// ToBeNil asserts that the value is nil, handling both untyped nil interfaces
// and typed nil pointers, slices, maps, channels and functions.
func (m *Matcher[T]) ToBeNil() {
	m.t.Helper()
	m.check(isNil(m.actual), fmt.Sprintf("to be nil, but got %s", format(m.actual)))
}

// ToBeTrue asserts that the value is the boolean true.
func (m *Matcher[T]) ToBeTrue() {
	m.t.Helper()
	b, ok := any(m.actual).(bool)
	m.check(ok && b, fmt.Sprintf("to be true, but got %s", format(m.actual)))
}

// ToBeFalse asserts that the value is the boolean false.
func (m *Matcher[T]) ToBeFalse() {
	m.t.Helper()
	b, ok := any(m.actual).(bool)
	m.check(ok && !b, fmt.Sprintf("to be false, but got %s", format(m.actual)))
}

// ToBeGreaterThan asserts that the numeric value is strictly greater than n.
func (m *Matcher[T]) ToBeGreaterThan(n T) {
	m.t.Helper()
	m.compare(n, "to be greater than", func(a, b float64) bool { return a > b })
}

// ToBeGreaterThanOrEqual asserts that the numeric value is greater than or
// equal to n.
func (m *Matcher[T]) ToBeGreaterThanOrEqual(n T) {
	m.t.Helper()
	m.compare(n, "to be greater than or equal to", func(a, b float64) bool { return a >= b })
}

// ToBeLessThan asserts that the numeric value is strictly less than n.
func (m *Matcher[T]) ToBeLessThan(n T) {
	m.t.Helper()
	m.compare(n, "to be less than", func(a, b float64) bool { return a < b })
}

// ToBeLessThanOrEqual asserts that the numeric value is less than or equal to n.
func (m *Matcher[T]) ToBeLessThanOrEqual(n T) {
	m.t.Helper()
	m.compare(n, "to be less than or equal to", func(a, b float64) bool { return a <= b })
}

func (m *Matcher[T]) compare(n T, phrase string, cmp func(a, b float64) bool) {
	m.t.Helper()
	a, aok := toFloat(m.actual)
	b, bok := toFloat(n)
	if !aok || !bok {
		m.t.Errorf("assertion failed: %s requires numeric values, but got %s and %s",
			phrase, format(m.actual), format(n))
		return
	}
	m.check(cmp(a, b), fmt.Sprintf("%s %s, but got %s", phrase, format(n), format(m.actual)))
}

// ToContain asserts containment. For strings it checks for a substring; for
// slices and arrays it checks for an element (deep equality); for maps it
// checks for a key.
func (m *Matcher[T]) ToContain(item any) {
	m.t.Helper()
	pass, kind := contains(m.actual, item)
	m.check(pass, fmt.Sprintf("to contain %s (%s), but got %s", format(item), kind, format(m.actual)))
}

// ToHaveLen asserts that the value has the given length. It works for strings,
// slices, arrays, maps and channels.
func (m *Matcher[T]) ToHaveLen(n int) {
	m.t.Helper()
	l, ok := lengthOf(m.actual)
	if !ok {
		m.t.Errorf("assertion failed: ToHaveLen requires a string, slice, array, map or channel, but got %s",
			format(m.actual))
		return
	}
	m.check(l == n, fmt.Sprintf("to have length %d, but got length %d (%s)", n, l, format(m.actual)))
}

// ToMatch asserts that the value, rendered as a string, matches the given
// regular expression.
func (m *Matcher[T]) ToMatch(pattern string) {
	m.t.Helper()
	re, err := regexp.Compile(pattern)
	if err != nil {
		m.t.Errorf("assertion failed: ToMatch received an invalid pattern %q: %v", pattern, err)
		return
	}
	s := fmt.Sprintf("%v", m.actual)
	m.check(re.MatchString(s), fmt.Sprintf("to match /%s/, but got %q", pattern, s))
}

// ToBeCloseTo asserts that a floating-point value is within epsilon of expected.
// If epsilon is omitted it defaults to 1e-9.
func (m *Matcher[T]) ToBeCloseTo(expected float64, epsilon ...float64) {
	m.t.Helper()
	eps := 1e-9
	if len(epsilon) > 0 {
		eps = epsilon[0]
	}
	a, ok := toFloat(m.actual)
	if !ok {
		m.t.Errorf("assertion failed: ToBeCloseTo requires a numeric value, but got %s", format(m.actual))
		return
	}
	d := a - expected
	if d < 0 {
		d = -d
	}
	m.check(d <= eps, fmt.Sprintf("to be within %g of %g, but got %g (difference %g)", eps, expected, a, d))
}

// ToPanic calls the actual value (which must be a func with no arguments) and
// asserts that it panics. If a message argument is supplied, the panic value's
// string form must contain that substring.
func (m *Matcher[T]) ToPanic(msg ...string) {
	m.t.Helper()
	m.assertPanic("ToPanic", msg...)
}

// ToThrow is an alias for [Matcher.ToPanic] provided for familiarity with the
// Jest API.
func (m *Matcher[T]) ToThrow(msg ...string) {
	m.t.Helper()
	m.assertPanic("ToThrow", msg...)
}

func (m *Matcher[T]) assertPanic(name string, msg ...string) {
	m.t.Helper()
	fn := reflect.ValueOf(m.actual)
	if !fn.IsValid() || fn.Kind() != reflect.Func || fn.Type().NumIn() != 0 {
		m.t.Errorf("assertion failed: %s requires a func() value, but got %s", name, format(m.actual))
		return
	}
	panicked, recovered := callAndRecover(fn)
	if !panicked {
		m.check(false, "to panic, but it did not")
		return
	}
	// It panicked. If a substring was requested, verify it.
	if len(msg) > 0 && msg[0] != "" {
		got := fmt.Sprintf("%v", recovered)
		m.check(strings.Contains(got, msg[0]),
			fmt.Sprintf("to panic with a message containing %q, but got %q", msg[0], got))
		return
	}
	m.check(true, "to panic")
}

func callAndRecover(fn reflect.Value) (panicked bool, recovered any) {
	defer func() {
		if r := recover(); r != nil {
			panicked = true
			recovered = r
		}
	}()
	fn.Call(nil)
	return false, nil
}
