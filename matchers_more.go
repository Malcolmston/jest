package jest

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

// ToMatchObject asserts that the actual value (a map with string keys or a
// struct) contains every key/field present in expected, each satisfying the
// corresponding value with asymmetric-aware equality. Extra keys or fields on
// the actual value are ignored, and nested maps/structs are matched recursively.
// Asymmetric matchers (see [AsymmetricMatcher]) may appear anywhere in expected.
func (m *Matcher[T]) ToMatchObject(expected any) {
	m.t.Helper()
	pass := matchObject(expected, m.actual)
	m.check(pass, fmt.Sprintf("to match object %s, but got %s", format(expected), format(m.actual)))
}

// matchObject reports whether actual contains the subset described by expected.
func matchObject(expected, actual any) bool {
	if am, ok := expected.(AsymmetricMatcher); ok {
		return am.Matches(actual)
	}
	ev := reflect.ValueOf(expected)
	for ev.Kind() == reflect.Ptr || ev.Kind() == reflect.Interface {
		if ev.IsNil() {
			return actual == nil || isNil(actual)
		}
		ev = ev.Elem()
	}
	switch ev.Kind() {
	case reflect.Map:
		for _, k := range ev.MapKeys() {
			got, ok := propertyByName(actual, fmt.Sprintf("%v", k.Interface()))
			if !ok || !matchObject(ev.MapIndex(k).Interface(), got) {
				return false
			}
		}
		return true
	case reflect.Struct:
		for i := 0; i < ev.NumField(); i++ {
			if ev.Type().Field(i).PkgPath != "" {
				continue
			}
			name := ev.Type().Field(i).Name
			got, ok := propertyByName(actual, name)
			if !ok || !matchObject(ev.Field(i).Interface(), got) {
				return false
			}
		}
		return true
	default:
		return asymEqual(expected, actual)
	}
}

// ToStrictEqual asserts deep equality that is also strict about dynamic types:
// like [Matcher.ToEqual] it walks the value recursively and honors asymmetric
// matchers, but it additionally requires the actual and expected values to have
// identical types at every level (so, for example, two structurally identical
// but differently named struct types are not considered equal).
func (m *Matcher[T]) ToStrictEqual(expected T) {
	m.t.Helper()
	pass := strictEqual(expected, m.actual)
	expectation := fmt.Sprintf("to strictly equal %s, but got %s", format(expected), format(m.actual))
	if !pass && !m.negated && !hasAsymmetric(expected) {
		if d := diff(expected, m.actual); d != "" {
			expectation += "\n" + d
		}
	}
	m.check(pass, expectation)
}

// strictEqual is asymEqual with an added top-level dynamic-type check.
func strictEqual(expected, actual any) bool {
	if _, ok := expected.(AsymmetricMatcher); ok {
		return asymEqual(expected, actual)
	}
	et, at := reflect.TypeOf(expected), reflect.TypeOf(actual)
	if et != at {
		return false
	}
	return asymEqual(expected, actual)
}

// ToHaveProperty asserts that the actual value has a property reachable by path
// (a dotted path such as "a.b.c", with numeric segments or "[i]" indexing into
// slices and arrays, e.g. "items[0].name"). If a value argument is supplied the
// property must additionally satisfy it with asymmetric-aware equality.
func (m *Matcher[T]) ToHaveProperty(path string, value ...any) {
	m.t.Helper()
	got, ok := resolvePath(m.actual, path)
	if !ok {
		m.check(false, fmt.Sprintf("to have property %q, but it was not found in %s", path, format(m.actual)))
		return
	}
	if len(value) == 0 {
		m.check(true, fmt.Sprintf("to have property %q", path))
		return
	}
	m.check(asymEqual(value[0], got),
		fmt.Sprintf("to have property %q equal to %s, but got %s", path, format(value[0]), format(got)))
}

// resolvePath walks a dotted/indexed property path through maps, structs,
// slices, arrays and pointers.
func resolvePath(root any, path string) (any, bool) {
	cur := root
	for _, seg := range splitPath(path) {
		next, ok := resolveSegment(cur, seg)
		if !ok {
			return nil, false
		}
		cur = next
	}
	return cur, true
}

// splitPath splits "a.b[0].c" into ["a","b","0","c"].
func splitPath(path string) []string {
	path = strings.ReplaceAll(path, "[", ".")
	path = strings.ReplaceAll(path, "]", "")
	parts := strings.Split(path, ".")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// resolveSegment reads a single path segment from container.
func resolveSegment(container any, seg string) (any, bool) {
	rv := reflect.ValueOf(container)
	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return nil, false
		}
		rv = rv.Elem()
	}
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		i, err := strconv.Atoi(seg)
		if err != nil || i < 0 || i >= rv.Len() {
			return nil, false
		}
		return rv.Index(i).Interface(), true
	default:
		return propertyByName(rv.Interface(), seg)
	}
}

// ToBeInstanceOf asserts that the actual value's dynamic type is assignable to
// the type of target (which may be a sample value or a [reflect.Type]). For an
// interface target it checks that the actual value implements the interface.
func (m *Matcher[T]) ToBeInstanceOf(target any) {
	m.t.Helper()
	pass := Any(target).Matches(m.actual)
	m.check(pass, fmt.Sprintf("to be an instance of %s, but got %s (%T)",
		typeName(target), format(m.actual), m.actual))
}

// typeName renders the target type for ToBeInstanceOf messages.
func typeName(target any) string {
	if rt, ok := target.(reflect.Type); ok {
		return rt.String()
	}
	if t := reflect.TypeOf(target); t != nil {
		return t.String()
	}
	return "<nil>"
}

// ToBeDefined asserts that the value is not nil (Go's closest analogue to a
// defined JavaScript value).
func (m *Matcher[T]) ToBeDefined() {
	m.t.Helper()
	m.check(!isNil(m.actual), fmt.Sprintf("to be defined, but got %s", format(m.actual)))
}

// ToBeUndefined asserts that the value is nil (Go's closest analogue to an
// undefined JavaScript value).
func (m *Matcher[T]) ToBeUndefined() {
	m.t.Helper()
	m.check(isNil(m.actual), fmt.Sprintf("to be undefined, but got %s", format(m.actual)))
}

// ToBeNaN asserts that the value is a floating-point NaN.
func (m *Matcher[T]) ToBeNaN() {
	m.t.Helper()
	f, ok := toFloat(m.actual)
	m.check(ok && math.IsNaN(f), fmt.Sprintf("to be NaN, but got %s", format(m.actual)))
}

// ---- mock call/return matchers ----------------------------------------------

// mockActual extracts the *Mock underlying the value under test. The value may
// be a *Mock directly, or a value exposing a `Mock() *Mock` accessor (as the
// spy references do).
func (m *Matcher[T]) mockActual(name string) (*Mock, bool) {
	switch v := any(m.actual).(type) {
	case *Mock:
		return v, true
	case interface{ mock() *Mock }:
		return v.mock(), true
	default:
		m.t.Errorf("assertion failed: %s requires a *jest.Mock value, but got %s", name, format(m.actual))
		return nil, false
	}
}

// ToHaveBeenCalled asserts that the mock under test was called at least once.
func (m *Matcher[T]) ToHaveBeenCalled() {
	m.t.Helper()
	mk, ok := m.mockActual("ToHaveBeenCalled")
	if !ok {
		return
	}
	m.check(mk.CallCount() > 0, fmt.Sprintf("to have been called, but %q was not called", mk.Name()))
}

// ToHaveBeenCalledTimes asserts that the mock under test was called exactly n times.
func (m *Matcher[T]) ToHaveBeenCalledTimes(n int) {
	m.t.Helper()
	mk, ok := m.mockActual("ToHaveBeenCalledTimes")
	if !ok {
		return
	}
	m.check(mk.CallCount() == n,
		fmt.Sprintf("to have been called %d time(s), but %q was called %d time(s)", n, mk.Name(), mk.CallCount()))
}

// ToHaveBeenCalledWith asserts that the mock under test was called at least once
// with arguments matching args (compared with asymmetric-aware equality, so
// asymmetric matchers may be used as arguments).
func (m *Matcher[T]) ToHaveBeenCalledWith(args ...any) {
	m.t.Helper()
	mk, ok := m.mockActual("ToHaveBeenCalledWith")
	if !ok {
		return
	}
	pass := false
	for _, c := range mk.Calls() {
		if argsMatch(args, c.Args) {
			pass = true
			break
		}
	}
	m.check(pass, fmt.Sprintf("to have been called with %s, but %q was not", format(args), mk.Name()))
}

// ToHaveBeenNthCalledWith asserts that the mock's n-th call (1-based, mirroring
// Jest) was made with arguments matching args (asymmetric-aware).
func (m *Matcher[T]) ToHaveBeenNthCalledWith(n int, args ...any) {
	m.t.Helper()
	mk, ok := m.mockActual("ToHaveBeenNthCalledWith")
	if !ok {
		return
	}
	c, exists := mk.NthCall(n - 1)
	m.check(exists && argsMatch(args, c.Args),
		fmt.Sprintf("call #%d to have arguments %s, but %q had %s", n, format(args), mk.Name(), nthArgs(c, exists)))
}

// ToHaveBeenLastCalledWith asserts that the mock's most recent call was made
// with arguments matching args (asymmetric-aware).
func (m *Matcher[T]) ToHaveBeenLastCalledWith(args ...any) {
	m.t.Helper()
	mk, ok := m.mockActual("ToHaveBeenLastCalledWith")
	if !ok {
		return
	}
	c, exists := mk.LastCall()
	m.check(exists && argsMatch(args, c.Args),
		fmt.Sprintf("last call to have arguments %s, but %q had %s", format(args), mk.Name(), nthArgs(c, exists)))
}

// ToHaveReturned asserts that the mock under test returned (completed without
// panicking) on at least one call.
func (m *Matcher[T]) ToHaveReturned() {
	m.t.Helper()
	mk, ok := m.mockActual("ToHaveReturned")
	if !ok {
		return
	}
	pass := false
	for _, c := range mk.Calls() {
		if !c.Panicked {
			pass = true
			break
		}
	}
	m.check(pass, fmt.Sprintf("%q to have returned, but it never did", mk.Name()))
}

// ToHaveReturnedWith asserts that the mock returned a first result matching
// value (asymmetric-aware) on at least one non-panicking call.
func (m *Matcher[T]) ToHaveReturnedWith(value any) {
	m.t.Helper()
	mk, ok := m.mockActual("ToHaveReturnedWith")
	if !ok {
		return
	}
	pass := false
	for _, c := range mk.Calls() {
		if c.Panicked || len(c.Results) == 0 {
			continue
		}
		if asymEqual(value, c.Results[0]) {
			pass = true
			break
		}
	}
	m.check(pass, fmt.Sprintf("%q to have returned %s, but it did not", mk.Name(), format(value)))
}

// argsMatch reports whether a recorded argument list matches an expected one,
// element by element, with asymmetric-aware equality.
func argsMatch(expected, actual []any) bool {
	if len(expected) != len(actual) {
		return false
	}
	for i := range expected {
		if !asymEqual(expected[i], actual[i]) {
			return false
		}
	}
	return true
}

// nthArgs renders a call's arguments for a failure message.
func nthArgs(c Call, exists bool) string {
	if !exists {
		return "no such call"
	}
	return format(c.Args)
}
