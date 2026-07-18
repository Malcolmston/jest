package jest

import "testing"

func TestParityToBeTruthyFalsyExtra(t *testing.T) {
	truthy := []any{true, 1, -3, 3.14, "x", []int{0}, map[string]int{}, struct{}{}}
	for _, v := range truthy {
		v := v
		expectPass(t, "truthy", func(r TestReporter) { Expect[any](r, v).ToBeTruthy() })
		expectFail(t, "not falsy", func(r TestReporter) { Expect[any](r, v).ToBeFalsy() })
	}
	falsy := []any{false, 0, 0.0, "", nil}
	for _, v := range falsy {
		v := v
		expectPass(t, "falsy", func(r TestReporter) { Expect[any](r, v).ToBeFalsy() })
		expectFail(t, "not truthy", func(r TestReporter) { Expect[any](r, v).ToBeTruthy() })
	}
	// A typed nil pointer is falsy.
	var p *int
	expectPass(t, "nil ptr falsy", func(r TestReporter) { Expect[any](r, p).ToBeFalsy() })
}

func TestParityToBeNull(t *testing.T) {
	expectPass(t, "nil", func(r TestReporter) { Expect[any](r, nil).ToBeNull() })
	var m map[string]int
	expectPass(t, "nil map", func(r TestReporter) { Expect[any](r, m).ToBeNull() })
	expectFail(t, "value", func(r TestReporter) { Expect(r, 5).ToBeNull() })
	expectPass(t, "not null", func(r TestReporter) { Expect(r, 5).Not().ToBeNull() })
}

func TestParityToContainEqual(t *testing.T) {
	type pt struct{ X, Y int }
	pts := []pt{{1, 2}, {3, 4}}
	expectPass(t, "struct member", func(r TestReporter) { Expect(r, pts).ToContainEqual(pt{3, 4}) })
	expectFail(t, "struct missing", func(r TestReporter) { Expect(r, pts).ToContainEqual(pt{9, 9}) })
	expectPass(t, "asymmetric", func(r TestReporter) {
		Expect[any](r, pts).ToContainEqual(ObjectContaining(map[string]any{"X": 1}))
	})
	m := map[string]int{"a": 1, "b": 2}
	expectPass(t, "map value", func(r TestReporter) { Expect[any](r, m).ToContainEqual(2) })
	expectFail(t, "map value missing", func(r TestReporter) { Expect[any](r, m).ToContainEqual(9) })
	expectFail(t, "wrong kind", func(r TestReporter) { Expect(r, 5).ToContainEqual(5) })
}

func TestParityToHaveLengthExtra(t *testing.T) {
	expectPass(t, "slice", func(r TestReporter) { Expect(r, []int{1, 2, 3}).ToHaveLength(3) })
	expectPass(t, "string", func(r TestReporter) { Expect(r, "abcd").ToHaveLength(4) })
	expectFail(t, "wrong", func(r TestReporter) { Expect(r, "ab").ToHaveLength(3) })
	expectFail(t, "wrong kind", func(r TestReporter) { Expect(r, 5).ToHaveLength(1) })
}

func TestParityToBeOneOf(t *testing.T) {
	expectPass(t, "match", func(r TestReporter) { Expect(r, 2).ToBeOneOf(1, 2, 3) })
	expectFail(t, "no match", func(r TestReporter) { Expect(r, 5).ToBeOneOf(1, 2, 3) })
	expectPass(t, "asymmetric", func(r TestReporter) {
		Expect[any](r, "hello").ToBeOneOf(StringContaining("ell"), 0)
	})
	expectPass(t, "not one of", func(r TestReporter) { Expect(r, 9).Not().ToBeOneOf(1, 2) })
}

func TestParityReturnMatchers(t *testing.T) {
	m := NewMock("f")
	m.ReturnValues([]any{1}, []any{2}, []any{3})
	m.Call()
	m.Call()
	m.Call()
	expectPass(t, "returned times", func(r TestReporter) { Expect(r, m).ToHaveReturnedTimes(3) })
	expectFail(t, "returned times wrong", func(r TestReporter) { Expect(r, m).ToHaveReturnedTimes(2) })
	expectPass(t, "last returned", func(r TestReporter) { Expect(r, m).ToHaveLastReturnedWith(3) })
	expectFail(t, "last returned wrong", func(r TestReporter) { Expect(r, m).ToHaveLastReturnedWith(1) })
	expectPass(t, "nth returned", func(r TestReporter) { Expect(r, m).ToHaveNthReturnedWith(2, 2) })
	expectFail(t, "nth returned wrong", func(r TestReporter) { Expect(r, m).ToHaveNthReturnedWith(1, 9) })
	expectFail(t, "nth out of range", func(r TestReporter) { Expect(r, m).ToHaveNthReturnedWith(9, 1) })
}

func TestParityReturnTimesWithPanic(t *testing.T) {
	m := NewMock("boom")
	m.MockImplementationOnce(func(_ ...any) []any { panic("x") })
	m.MockImplementationOnce(func(_ ...any) []any { return []any{7} })
	func() {
		defer func() { recover() }()
		m.Call()
	}()
	m.Call()
	// One panicking call and one returning call.
	expectPass(t, "one return", func(r TestReporter) { Expect(r, m).ToHaveReturnedTimes(1) })
}

func TestParityJestAliases(t *testing.T) {
	m := NewMock("f")
	m.Return(42)
	m.Call(1, 2)
	expectPass(t, "toBeCalled", func(r TestReporter) { Expect(r, m).ToBeCalled() })
	expectPass(t, "toBeCalledTimes", func(r TestReporter) { Expect(r, m).ToBeCalledTimes(1) })
	expectPass(t, "toBeCalledWith", func(r TestReporter) { Expect(r, m).ToBeCalledWith(1, 2) })
	expectPass(t, "lastCalledWith", func(r TestReporter) { Expect(r, m).LastCalledWith(1, 2) })
	expectPass(t, "nthCalledWith", func(r TestReporter) { Expect(r, m).NthCalledWith(1, 1, 2) })
	expectPass(t, "toReturn", func(r TestReporter) { Expect(r, m).ToReturn() })
	expectPass(t, "toReturnWith", func(r TestReporter) { Expect(r, m).ToReturnWith(42) })
	expectFail(t, "alias fails", func(r TestReporter) { Expect(r, m).ToBeCalledTimes(9) })
}
