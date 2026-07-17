package jest

import "testing"

func TestMockImplementation(t *testing.T) {
	m := NewMock("impl")
	m.MockImplementation(func(args ...any) []any {
		return []any{args[0].(int) * 2}
	})
	Expect(t, m.Call(5)[0]).ToBe(10)
	Expect(t, m.Call(6)[0]).ToBe(12)
	Expect(t, m.CallCount()).ToBe(2)
}

func TestMockImplementationOnce(t *testing.T) {
	m := NewMock("once")
	m.MockImplementation(func(_ ...any) []any { return []any{"base"} })
	m.MockImplementationOnce(func(_ ...any) []any { return []any{"first"} }).
		MockImplementationOnce(func(_ ...any) []any { return []any{"second"} })
	Expect(t, m.Call()[0]).ToBe("first")
	Expect(t, m.Call()[0]).ToBe("second")
	Expect(t, m.Call()[0]).ToBe("base")
}

func TestMockReturnValueOnce(t *testing.T) {
	m := NewMock("rvo")
	m.Return("default")
	m.MockReturnValueOnce("a").MockReturnValueOnce("b")
	Expect(t, m.Call()[0]).ToBe("a")
	Expect(t, m.Call()[0]).ToBe("b")
	Expect(t, m.Call()[0]).ToBe("default")
}

func TestMockResolvedRejected(t *testing.T) {
	ok := NewMock("resolve")
	ok.MockResolvedValue(42)
	res := ok.Call()
	Expect(t, res[0]).ToBe(42)
	Expect(t, res[1]).ToBeNil()

	bad := NewMock("reject")
	err := errString("boom")
	bad.MockRejectedValue(err)
	res = bad.Call()
	Expect(t, res[0]).ToBeNil()
	Expect(t, res[1].(error).Error()).ToBe("boom")
}

func TestMockResults(t *testing.T) {
	m := NewMock("results")
	m.ReturnValues([]any{1}, []any{2})
	m.Call()
	m.Call()
	got := m.Results()
	Expect(t, len(got)).ToBe(2)
	Expect(t, got[0]).ToEqual([]any{1})
	Expect(t, got[1]).ToEqual([]any{2})
}

// errString is a trivial error used to avoid importing errors in this file.
type errString string

func (e errString) Error() string { return string(e) }
