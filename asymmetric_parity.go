package jest

import (
	"fmt"
	"math"
)

// numberCloseToMatcher matches numbers within a decimal precision of a target.
type numberCloseToMatcher struct {
	expected  float64
	precision int
}

// CloseTo returns an [AsymmetricMatcher] that matches any numeric value within
// half a unit of the last requested decimal place of expected, mirroring Jest's
// expect.closeTo. precision is the number of decimal digits checked and defaults
// to 2, so CloseTo(0.3) matches values within 0.005 of 0.3. It is useful for
// floating-point fields nested inside [Matcher.ToEqual] or [Matcher.ToMatchObject].
func CloseTo(expected float64, precision ...int) AsymmetricMatcher {
	p := 2
	if len(precision) > 0 {
		p = precision[0]
	}
	return numberCloseToMatcher{expected: expected, precision: p}
}

// Matches reports whether actual is a numeric value within the matcher's
// tolerance of the target.
func (n numberCloseToMatcher) Matches(actual any) bool {
	a, ok := toFloat(actual)
	if !ok {
		return false
	}
	if math.IsNaN(a) || math.IsNaN(n.expected) {
		return false
	}
	tol := 0.5 * math.Pow(10, -float64(n.precision))
	return math.Abs(a-n.expected) < tol
}

// String describes the matcher.
func (n numberCloseToMatcher) String() string {
	return fmt.Sprintf("CloseTo(%g, %d)", n.expected, n.precision)
}

// notMatcher inverts the sense of another [AsymmetricMatcher].
type notMatcher struct{ inner AsymmetricMatcher }

// Matches reports whether actual does not satisfy the wrapped matcher.
func (n notMatcher) Matches(actual any) bool { return !n.inner.Matches(actual) }

// String describes the matcher.
func (n notMatcher) String() string { return "Not(" + n.inner.String() + ")" }

// NotArrayContaining returns an [AsymmetricMatcher] that matches any slice or
// array that does NOT contain all of elems, mirroring Jest's
// expect.not.arrayContaining.
func NotArrayContaining(elems ...any) AsymmetricMatcher {
	return notMatcher{inner: ArrayContaining(elems...)}
}

// NotObjectContaining returns an [AsymmetricMatcher] that matches any map or
// struct that does NOT contain all of fields with matching values, mirroring
// Jest's expect.not.objectContaining.
func NotObjectContaining(fields map[string]any) AsymmetricMatcher {
	return notMatcher{inner: ObjectContaining(fields)}
}

// NotStringContaining returns an [AsymmetricMatcher] that matches any string
// that does NOT contain sub, mirroring Jest's expect.not.stringContaining. A
// non-string value also fails to match (it is not a string containing sub), so
// the negation matches it.
func NotStringContaining(sub string) AsymmetricMatcher {
	return notMatcher{inner: StringContaining(sub)}
}

// NotStringMatching returns an [AsymmetricMatcher] that matches any string that
// does NOT match the given regular expression, mirroring Jest's
// expect.not.stringMatching. It panics if pattern is not a valid regular
// expression.
func NotStringMatching(pattern string) AsymmetricMatcher {
	return notMatcher{inner: StringMatching(pattern)}
}
