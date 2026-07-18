package jest

import (
	"fmt"
	"math"
	"reflect"
)

// parityTruthy reports whether v is "truthy" in the JavaScript sense that Jest's
// toBeTruthy/toBeFalsy use: nil, the boolean false, a numeric zero, an empty
// string and a floating-point NaN are falsy; every other value is truthy.
func parityTruthy(v any) bool {
	if v == nil || isNil(v) {
		return false
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Bool:
		return rv.Bool()
	case reflect.String:
		return rv.Len() != 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return rv.Uint() != 0
	case reflect.Float32, reflect.Float64:
		f := rv.Float()
		return f != 0 && !math.IsNaN(f)
	default:
		return true
	}
}

// ToBeTruthy asserts that the value is truthy in the JavaScript sense used by
// Jest: everything except nil, the boolean false, a numeric zero, the empty
// string and NaN is considered truthy.
func (m *Matcher[T]) ToBeTruthy() {
	m.t.Helper()
	m.check(parityTruthy(m.actual), fmt.Sprintf("to be truthy, but got %s", format(m.actual)))
}

// ToBeFalsy asserts that the value is falsy in the JavaScript sense used by
// Jest: nil, the boolean false, a numeric zero, the empty string and NaN are
// the falsy values.
func (m *Matcher[T]) ToBeFalsy() {
	m.t.Helper()
	m.check(!parityTruthy(m.actual), fmt.Sprintf("to be falsy, but got %s", format(m.actual)))
}

// ToBeNull asserts that the value is nil, the Go analogue of Jest's toBeNull.
// It handles both untyped nil interfaces and typed nil pointers, slices, maps,
// channels and functions.
func (m *Matcher[T]) ToBeNull() {
	m.t.Helper()
	m.check(isNil(m.actual), fmt.Sprintf("to be null, but got %s", format(m.actual)))
}

// ToContainEqual asserts that the actual value contains an element deeply equal
// to item. Unlike [Matcher.ToContain], the comparison is asymmetric-aware, so
// item may itself be an [AsymmetricMatcher]. It works for slices and arrays
// (element membership) and for maps (value membership).
func (m *Matcher[T]) ToContainEqual(item any) {
	m.t.Helper()
	rv := reflect.ValueOf(m.actual)
	found := false
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < rv.Len(); i++ {
			if asymEqual(item, rv.Index(i).Interface()) {
				found = true
			}
		}
	case reflect.Map:
		for _, k := range rv.MapKeys() {
			if asymEqual(item, rv.MapIndex(k).Interface()) {
				found = true
			}
		}
	default:
		m.t.Errorf("assertion failed: ToContainEqual requires a slice, array or map, but got %s",
			format(m.actual))
		return
	}
	m.check(found, fmt.Sprintf("to contain an element deeply equal to %s, but got %s",
		format(item), format(m.actual)))
}

// ToHaveLength asserts that the value has the given length, mirroring Jest's
// toHaveLength. It works for strings, slices, arrays, maps and channels and is
// an alias, using the Jest spelling, of [Matcher.ToHaveLen].
func (m *Matcher[T]) ToHaveLength(n int) {
	m.t.Helper()
	l, ok := lengthOf(m.actual)
	if !ok {
		m.t.Errorf("assertion failed: ToHaveLength requires a string, slice, array, map or channel, but got %s",
			format(m.actual))
		return
	}
	m.check(l == n, fmt.Sprintf("to have length %d, but got length %d (%s)", n, l, format(m.actual)))
}

// ToBeOneOf asserts that the value is deeply equal to one of the supplied
// candidates, mirroring Jest's toBeOneOf. Each candidate is compared with
// asymmetric-aware equality, so candidates may be asymmetric matchers.
func (m *Matcher[T]) ToBeOneOf(candidates ...any) {
	m.t.Helper()
	found := false
	for _, c := range candidates {
		if asymEqual(c, m.actual) {
			found = true
			break
		}
	}
	m.check(found, fmt.Sprintf("to be one of %s, but got %s", format(candidates), format(m.actual)))
}

// ---- return-value matchers --------------------------------------------------

// ToHaveReturnedTimes asserts that the mock under test returned (completed a
// call without panicking) exactly n times, mirroring Jest's
// toHaveReturnedTimes.
func (m *Matcher[T]) ToHaveReturnedTimes(n int) {
	m.t.Helper()
	mk, ok := m.mockActual("ToHaveReturnedTimes")
	if !ok {
		return
	}
	got := 0
	for _, c := range mk.Calls() {
		if !c.Panicked {
			got++
		}
	}
	m.check(got == n, fmt.Sprintf("%q to have returned %d time(s), but it returned %d time(s)",
		mk.Name(), n, got))
}

// ToHaveLastReturnedWith asserts that the mock's most recent call returned a
// first result matching value (asymmetric-aware), mirroring Jest's
// toHaveLastReturnedWith. A most-recent call that panicked fails the assertion.
func (m *Matcher[T]) ToHaveLastReturnedWith(value any) {
	m.t.Helper()
	mk, ok := m.mockActual("ToHaveLastReturnedWith")
	if !ok {
		return
	}
	c, exists := mk.LastCall()
	pass := exists && !c.Panicked && len(c.Results) > 0 && asymEqual(value, c.Results[0])
	m.check(pass, fmt.Sprintf("%q last call to have returned %s, but it did not",
		mk.Name(), format(value)))
}

// ToHaveNthReturnedWith asserts that the mock's n-th call (1-based, mirroring
// Jest) returned a first result matching value (asymmetric-aware), mirroring
// Jest's toHaveNthReturnedWith. A missing or panicking call fails the assertion.
func (m *Matcher[T]) ToHaveNthReturnedWith(n int, value any) {
	m.t.Helper()
	mk, ok := m.mockActual("ToHaveNthReturnedWith")
	if !ok {
		return
	}
	c, exists := mk.NthCall(n - 1)
	pass := exists && !c.Panicked && len(c.Results) > 0 && asymEqual(value, c.Results[0])
	m.check(pass, fmt.Sprintf("call #%d of %q to have returned %s, but it did not",
		n, mk.Name(), format(value)))
}

// ---- Jest alias spellings ---------------------------------------------------

// ToBeCalled is an alias for [Matcher.ToHaveBeenCalled], provided for
// familiarity with Jest's toBeCalled spelling.
func (m *Matcher[T]) ToBeCalled() {
	m.t.Helper()
	m.ToHaveBeenCalled()
}

// ToBeCalledTimes is an alias for [Matcher.ToHaveBeenCalledTimes], provided for
// familiarity with Jest's toBeCalledTimes spelling.
func (m *Matcher[T]) ToBeCalledTimes(n int) {
	m.t.Helper()
	m.ToHaveBeenCalledTimes(n)
}

// ToBeCalledWith is an alias for [Matcher.ToHaveBeenCalledWith], provided for
// familiarity with Jest's toBeCalledWith spelling.
func (m *Matcher[T]) ToBeCalledWith(args ...any) {
	m.t.Helper()
	m.ToHaveBeenCalledWith(args...)
}

// LastCalledWith is an alias for [Matcher.ToHaveBeenLastCalledWith], provided
// for familiarity with Jest's lastCalledWith spelling.
func (m *Matcher[T]) LastCalledWith(args ...any) {
	m.t.Helper()
	m.ToHaveBeenLastCalledWith(args...)
}

// NthCalledWith is an alias for [Matcher.ToHaveBeenNthCalledWith], provided for
// familiarity with Jest's nthCalledWith spelling.
func (m *Matcher[T]) NthCalledWith(n int, args ...any) {
	m.t.Helper()
	m.ToHaveBeenNthCalledWith(n, args...)
}

// ToReturn is an alias for [Matcher.ToHaveReturned], provided for familiarity
// with Jest's toReturn spelling.
func (m *Matcher[T]) ToReturn() {
	m.t.Helper()
	m.ToHaveReturned()
}

// ToReturnWith is an alias for [Matcher.ToHaveReturnedWith], provided for
// familiarity with Jest's toReturnWith spelling.
func (m *Matcher[T]) ToReturnWith(value any) {
	m.t.Helper()
	m.ToHaveReturnedWith(value)
}
