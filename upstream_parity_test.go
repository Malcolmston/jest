package jest

import (
	"math"
	"testing"
)

// The tests in this file encode concrete known-answer vectors taken directly
// from Jest's own test suite (jestjs/jest, packages/expect/src/__tests__/
// matchers.test.js and asymmetricMatchers.test.ts). Each case mirrors an
// assertion the upstream framework makes so the Go port stays faithful to
// Jest's documented matcher semantics. Every function is prefixed TestParity
// so they can be run in isolation with `go test -run Parity`.

// TestParityToBeCloseTo mirrors the `.toBeCloseTo` precision tables in
// matchers.test.js. Jest compares with |expected-received| < 10**-precision/2
// (default precision 2) and treats matching infinities as close. The port's
// faithful precision implementation is the CloseTo asymmetric matcher.
func TestParityToBeCloseTo(t *testing.T) {
	inf := math.Inf(1)
	// {pass: true} with default precision (2).
	passDefault := [][2]float64{
		{0, 0}, {0, 0.001}, {1.23, 1.229}, {1.23, 1.226},
		{1.23, 1.225}, {1.23, 1.234}, {inf, inf}, {-inf, -inf},
	}
	for _, c := range passDefault {
		if !CloseTo(c[1]).Matches(c[0]) {
			t.Errorf("expect(%g).toBeCloseTo(%g) should pass", c[0], c[1])
		}
	}
	// {pass: false} with default precision (2).
	failDefault := [][2]float64{
		{0, 0.01}, {1, 1.23}, {1.23, 1.2249999},
		{inf, -inf}, {inf, 1.23}, {-inf, -1.23},
	}
	for _, c := range failDefault {
		if CloseTo(c[1]).Matches(c[0]) {
			t.Errorf("expect(%g).toBeCloseTo(%g) should fail", c[0], c[1])
		}
	}
	// {pass: false} with explicit precision.
	type pc struct {
		n1, n2 float64
		p      int
	}
	for _, c := range []pc{{3.141592e-7, 3e-7, 8}, {56789, 51234, -4}} {
		if CloseTo(c.n2, c.p).Matches(c.n1) {
			t.Errorf("expect(%g).toBeCloseTo(%g, %d) should fail", c.n1, c.n2, c.p)
		}
	}
	// {pass: true} with explicit precision.
	for _, c := range []pc{{0, 0.1, 0}, {0, 0.0001, 3}, {0, 0.000004, 5}, {2.0000002, 2, 5}} {
		if !CloseTo(c.n2, c.p).Matches(c.n1) {
			t.Errorf("expect(%g).toBeCloseTo(%g, %d) should pass", c.n1, c.n2, c.p)
		}
	}
}

// TestParityToContain mirrors the `.toContain()` membership tables. Jest checks
// substrings for strings and element identity for collections.
func TestParityToContain(t *testing.T) {
	// Passing element/substring cases with Go analogs.
	expectPass(t, "int slice", func(r TestReporter) { Expect(r, []int{1, 2, 3, 4}).ToContain(1) })
	expectPass(t, "string slice", func(r TestReporter) { Expect(r, []string{"a", "b", "c", "d"}).ToContain("a") })
	expectPass(t, "substring abc", func(r TestReporter) { Expect(r, "abcdef").ToContain("abc") })
	expectPass(t, "substring digit", func(r TestReporter) { Expect(r, "11112111").ToContain("2") })
	expectPass(t, "map key", func(r TestReporter) { Expect(r, map[string]int{"abc": 1, "def": 2}).ToContain("abc") })
	// Failing cases.
	expectFail(t, "missing element", func(r TestReporter) { Expect(r, []int{1, 2, 3}).ToContain(4) })
	expectPass(t, "not contains", func(r TestReporter) { Expect(r, []int{1, 2, 3}).Not().ToContain(4) })
}

// TestParityToHaveLength mirrors the `.toHaveLength` tables: strings, slices and
// empty collections report their element/character count.
func TestParityToHaveLength(t *testing.T) {
	expectPass(t, "two ints", func(r TestReporter) { Expect(r, []int{1, 2}).ToHaveLength(2) })
	expectPass(t, "empty slice", func(r TestReporter) { Expect(r, []int{}).ToHaveLength(0) })
	expectPass(t, "two strings", func(r TestReporter) { Expect(r, []string{"a", "b"}).ToHaveLength(2) })
	expectPass(t, "abc", func(r TestReporter) { Expect(r, "abc").ToHaveLength(3) })
	expectPass(t, "empty string", func(r TestReporter) { Expect(r, "").ToHaveLength(0) })
	// {pass: false}
	expectFail(t, "wrong len slice", func(r TestReporter) { Expect(r, []int{1, 2}).ToHaveLength(3) })
	expectFail(t, "wrong len empty", func(r TestReporter) { Expect(r, []int{}).ToHaveLength(1) })
	expectFail(t, "wrong len abc", func(r TestReporter) { Expect(r, "abc").ToHaveLength(66) })
}

// TestParityToBeTruthyFalsy mirrors the `.toBeTruthy(), .toBeFalsy()` value
// tables. Jest treats {}, [], true, 1, "a", 0.5 and Infinity as truthy and
// false, null, NaN, 0, "" and undefined as falsy.
func TestParityToBeTruthyFalsy(t *testing.T) {
	truthy := []any{map[string]int{}, []int{}, true, 1, "a", 0.5, math.Inf(1)}
	for _, v := range truthy {
		v := v
		expectPass(t, "truthy", func(r TestReporter) { Expect[any](r, v).ToBeTruthy() })
		expectPass(t, "not falsy", func(r TestReporter) { Expect[any](r, v).Not().ToBeFalsy() })
	}
	falsy := []any{false, nil, math.NaN(), 0, ""}
	for _, v := range falsy {
		v := v
		expectPass(t, "falsy", func(r TestReporter) { Expect[any](r, v).ToBeFalsy() })
		expectPass(t, "not truthy", func(r TestReporter) { Expect[any](r, v).Not().ToBeTruthy() })
	}
}

// TestParityToMatch mirrors the `.toMatch()` tables: a literal substring and a
// case-insensitive regular expression pass, unrelated patterns fail.
func TestParityToMatch(t *testing.T) {
	expectPass(t, "literal foo", func(r TestReporter) { Expect(r, "foo").ToMatch("foo") })
	expectPass(t, "regex ci", func(r TestReporter) { Expect(r, "Foo bar").ToMatch("(?i)^foo") })
	expectFail(t, "no match literal", func(r TestReporter) { Expect(r, "bar").ToMatch("foo") })
	expectFail(t, "no match regex", func(r TestReporter) { Expect(r, "bar").ToMatch("foo") })
}

// TestParityToHaveProperty mirrors the `.toHaveProperty()` deep-path table,
// including dotted paths, bracket indices and asymmetric expected values.
func TestParityToHaveProperty(t *testing.T) {
	m := func(kv ...any) map[string]any {
		out := map[string]any{}
		for i := 0; i < len(kv); i += 2 {
			out[kv[i].(string)] = kv[i+1]
		}
		return out
	}
	// {a: {b: {c: {d: 1}}}}, 'a.b.c.d' -> 1
	expectPass(t, "nested d", func(r TestReporter) {
		Expect[any](r, m("a", m("b", m("c", m("d", 1))))).ToHaveProperty("a.b.c.d", 1)
	})
	// {a: {b: [1,2,3]}}, 'a.b[1]' -> 2
	expectPass(t, "index", func(r TestReporter) {
		Expect[any](r, m("a", m("b", []any{1, 2, 3}))).ToHaveProperty("a.b[1]", 2)
	})
	// {a: {b: [1,2,3]}}, 'a.b[1]' -> Any(Number)
	expectPass(t, "index any number", func(r TestReporter) {
		Expect[any](r, m("a", m("b", []any{1, 2, 3}))).ToHaveProperty("a.b[1]", Any(0))
	})
	// {a: {b: [{c: [{d: 1}]}]}}, 'a.b[0].c[0].d' -> 1
	expectPass(t, "deep index", func(r TestReporter) {
		Expect[any](r, m("a", m("b", []any{m("c", []any{m("d", 1)})}))).ToHaveProperty("a.b[0].c[0].d", 1)
	})
	// {a: 0}, 'a' -> 0
	expectPass(t, "zero value", func(r TestReporter) { Expect[any](r, m("a", 0)).ToHaveProperty("a", 0) })
	// missing path fails
	expectFail(t, "missing", func(r TestReporter) { Expect[any](r, m("a", 1)).ToHaveProperty("b", 1) })
}

// TestParityToStrictEqual mirrors `.toStrictEqual()`: two instances of the same
// class with equal fields are strictly equal, but instances of different
// classes with the same field shape are not (the dynamic type must match).
func TestParityToStrictEqual(t *testing.T) {
	type TestClassA struct{ A, B int }
	type TestClassB struct{ A, B int }
	expectPass(t, "same type", func(r TestReporter) {
		Expect[any](r, TestClassA{1, 2}).ToStrictEqual(TestClassA{1, 2})
	})
	expectFail(t, "different type", func(r TestReporter) {
		Expect[any](r, TestClassA{1, 2}).ToStrictEqual(TestClassB{1, 2})
	})
	// Wrapped inside a map, mirroring {test: new TestClassA(1, 2)}.
	expectPass(t, "wrapped same", func(r TestReporter) {
		Expect[any](r, map[string]any{"test": TestClassA{1, 2}}).
			ToStrictEqual(map[string]any{"test": TestClassA{1, 2}})
	})
}

// TestParityToBeNaN mirrors `.toBeNaN()`: NaN, sqrt(-1), Infinity-Infinity and
// 0/0 are all NaN; a finite value is not.
func TestParityToBeNaN(t *testing.T) {
	nans := []float64{math.NaN(), math.Sqrt(-1), math.Inf(1) - math.Inf(1), 0.0 / func() float64 { return 0 }()}
	for _, v := range nans {
		v := v
		expectPass(t, "nan", func(r TestReporter) { Expect(r, v).ToBeNaN() })
		expectPass(t, "not nan negated", func(r TestReporter) { Expect(r, v).Not().Not().ToBeNaN() })
	}
	expectFail(t, "finite", func(r TestReporter) { Expect(r, 1.0).ToBeNaN() })
}

// TestParityToBeInstanceOf mirrors `.toBeInstanceOf()`: a value is an instance
// of its own dynamic type and fails against an unrelated one.
func TestParityToBeInstanceOf(t *testing.T) {
	expectPass(t, "string", func(r TestReporter) { Expect[any](r, "a").ToBeInstanceOf("") })
	expectPass(t, "number", func(r TestReporter) { Expect[any](r, 1).ToBeInstanceOf(0) })
	expectPass(t, "bool", func(r TestReporter) { Expect[any](r, true).ToBeInstanceOf(false) })
	expectFail(t, "string not number", func(r TestReporter) { Expect[any](r, "a").ToBeInstanceOf(0) })
	expectFail(t, "nil not string", func(r TestReporter) { Expect[any](r, nil).ToBeInstanceOf("") })
}

// TestParityAsymmetricMatchers mirrors asymmetricMatchers.test.ts vectors for
// arrayContaining, objectContaining, stringContaining and stringMatching used
// via ToEqual.
func TestParityAsymmetricMatchers(t *testing.T) {
	// expect.arrayContaining([1,2]) matches [1,2,3,4]
	expectPass(t, "array containing", func(r TestReporter) {
		Expect[any](r, []any{1, 2, 3, 4}).ToEqual(ArrayContaining(1, 2))
	})
	expectFail(t, "array containing missing", func(r TestReporter) {
		Expect[any](r, []any{1, 2, 3}).ToEqual(ArrayContaining(1, 2, 3, 4))
	})
	// expect.objectContaining({foo: 'foo'})
	expectPass(t, "object containing", func(r TestReporter) {
		Expect[any](r, map[string]any{"foo": "foo", "bar": "bar"}).
			ToEqual(ObjectContaining(map[string]any{"foo": "foo"}))
	})
	// expect.stringContaining('en*')
	expectPass(t, "string containing", func(r TestReporter) {
		Expect[any](r, "queen").ToEqual(StringContaining("een"))
	})
	// expect.stringMatching(/en/)
	expectPass(t, "string matching", func(r TestReporter) {
		Expect[any](r, "queen").ToEqual(StringMatching("en"))
	})
	// expect.any(Number)
	expectPass(t, "any number", func(r TestReporter) {
		Expect[any](r, 42).ToEqual(Any(0))
	})
	// expect.anything() matches non-nil, not nil.
	expectPass(t, "anything", func(r TestReporter) { Expect[any](r, "x").ToEqual(Anything()) })
	expectFail(t, "anything nil", func(r TestReporter) { Expect[any](r, nil).ToEqual(Anything()) })
}
