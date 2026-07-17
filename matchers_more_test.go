package jest

import (
	"math"
	"strings"
	"testing"
)

func TestToMatchObject(t *testing.T) {
	actual := map[string]any{"name": "bob", "age": 30, "extra": true}
	expectPass(t, "subset map", func(r TestReporter) {
		Expect[any](r, actual).ToMatchObject(map[string]any{"name": "bob"})
	})
	expectPass(t, "subset matcher", func(r TestReporter) {
		Expect[any](r, actual).ToMatchObject(map[string]any{"age": Any(0)})
	})
	expectFail(t, "missing key", func(r TestReporter) {
		Expect[any](r, actual).ToMatchObject(map[string]any{"nope": 1})
	})
	expectFail(t, "wrong value", func(r TestReporter) {
		Expect[any](r, actual).ToMatchObject(map[string]any{"name": "al"})
	})

	type addr struct{ City string }
	type person struct {
		Name string
		Addr addr
	}
	p := person{Name: "bob", Addr: addr{City: "NYC"}}
	// A struct pattern compares every exported field (Go has no "undefined"),
	// so partial matching against a struct value is expressed with a map subset.
	expectPass(t, "struct via map subset", func(r TestReporter) {
		Expect[any](r, p).ToMatchObject(map[string]any{"Name": "bob"})
	})
	expectPass(t, "nested struct via map", func(r TestReporter) {
		Expect[any](r, p).ToMatchObject(map[string]any{"Addr": map[string]any{"City": "NYC"}})
	})
	expectFail(t, "struct pattern all fields", func(r TestReporter) {
		Expect[any](r, p).ToMatchObject(person{Name: "bob"})
	})
	expectPass(t, "not match", func(r TestReporter) {
		Expect[any](r, actual).Not().ToMatchObject(map[string]any{"name": "zzz"})
	})
}

func TestToStrictEqual(t *testing.T) {
	expectPass(t, "strict slice", func(r TestReporter) { Expect(r, []int{1, 2}).ToStrictEqual([]int{1, 2}) })
	expectFail(t, "strict slice fail", func(r TestReporter) { Expect(r, []int{1, 2}).ToStrictEqual([]int{1, 3}) })

	type a struct{ X int }
	expectPass(t, "strict struct", func(r TestReporter) { Expect(r, a{1}).ToStrictEqual(a{1}) })

	// Different dynamic types behind any are not strictly equal.
	expectFail(t, "strict type mismatch", func(r TestReporter) {
		Expect[any](r, int32(1)).ToStrictEqual(any(int64(1)))
	})
	expectPass(t, "strict matcher", func(r TestReporter) {
		Expect[any](r, 5).ToStrictEqual(any(Any(0)))
	})
}

func TestToHaveProperty(t *testing.T) {
	data := map[string]any{
		"user": map[string]any{
			"name":  "bob",
			"roles": []any{"admin", "dev"},
		},
	}
	expectPass(t, "nested present", func(r TestReporter) { Expect[any](r, data).ToHaveProperty("user.name") })
	expectPass(t, "nested value", func(r TestReporter) { Expect[any](r, data).ToHaveProperty("user.name", "bob") })
	expectPass(t, "indexed", func(r TestReporter) { Expect[any](r, data).ToHaveProperty("user.roles[0]", "admin") })
	expectPass(t, "indexed matcher", func(r TestReporter) {
		Expect[any](r, data).ToHaveProperty("user.roles[1]", StringContaining("de"))
	})
	expectFail(t, "missing", func(r TestReporter) { Expect[any](r, data).ToHaveProperty("user.email") })
	expectFail(t, "wrong value", func(r TestReporter) { Expect[any](r, data).ToHaveProperty("user.name", "alice") })
	expectFail(t, "index oob", func(r TestReporter) { Expect[any](r, data).ToHaveProperty("user.roles[5]") })

	type inner struct{ V int }
	type outer struct{ In inner }
	expectPass(t, "struct path", func(r TestReporter) { Expect[any](r, outer{inner{42}}).ToHaveProperty("In.V", 42) })
}

func TestToBeInstanceOf(t *testing.T) {
	expectPass(t, "same type", func(r TestReporter) { Expect[any](r, "hi").ToBeInstanceOf("") })
	expectFail(t, "diff type", func(r TestReporter) { Expect[any](r, "hi").ToBeInstanceOf(0) })

	type widget struct{ ID int }
	expectPass(t, "struct type", func(r TestReporter) { Expect[any](r, widget{1}).ToBeInstanceOf(widget{}) })
	expectPass(t, "not instance", func(r TestReporter) { Expect[any](r, 5).Not().ToBeInstanceOf("") })
}

func TestToBeDefinedUndefined(t *testing.T) {
	expectPass(t, "defined", func(r TestReporter) { Expect(r, 5).ToBeDefined() })
	expectFail(t, "defined fail", func(r TestReporter) { Expect[any](r, nil).ToBeDefined() })
	expectPass(t, "undefined", func(r TestReporter) { Expect[any](r, nil).ToBeUndefined() })
	expectFail(t, "undefined fail", func(r TestReporter) { Expect(r, 5).ToBeUndefined() })
	var p *int
	expectPass(t, "typed nil undefined", func(r TestReporter) { Expect(r, p).ToBeUndefined() })
}

func TestToBeNaN(t *testing.T) {
	expectPass(t, "nan", func(r TestReporter) { Expect(r, math.NaN()).ToBeNaN() })
	expectFail(t, "not nan", func(r TestReporter) { Expect(r, 1.0).ToBeNaN() })
	expectFail(t, "non numeric nan", func(r TestReporter) { Expect(r, "x").ToBeNaN() })
	expectPass(t, "not nan negated", func(r TestReporter) { Expect(r, 1.0).Not().ToBeNaN() })
}

func TestMockCallMatchers(t *testing.T) {
	m := NewMock("svc")
	m.Return("ok")
	m.Call(1, "a")
	m.Call(2, "b")

	expectPass(t, "called", func(r TestReporter) { Expect(r, m).ToHaveBeenCalled() })
	expectPass(t, "called times", func(r TestReporter) { Expect(r, m).ToHaveBeenCalledTimes(2) })
	expectFail(t, "called times fail", func(r TestReporter) { Expect(r, m).ToHaveBeenCalledTimes(5) })
	expectPass(t, "called with", func(r TestReporter) { Expect(r, m).ToHaveBeenCalledWith(1, "a") })
	expectPass(t, "called with matcher", func(r TestReporter) { Expect(r, m).ToHaveBeenCalledWith(Any(0), "b") })
	expectFail(t, "called with fail", func(r TestReporter) { Expect(r, m).ToHaveBeenCalledWith(9, "z") })
	expectPass(t, "nth called", func(r TestReporter) { Expect(r, m).ToHaveBeenNthCalledWith(1, 1, "a") })
	expectFail(t, "nth called fail", func(r TestReporter) { Expect(r, m).ToHaveBeenNthCalledWith(2, 1, "a") })
	expectFail(t, "nth oob", func(r TestReporter) { Expect(r, m).ToHaveBeenNthCalledWith(9, 1) })
	expectPass(t, "last called", func(r TestReporter) { Expect(r, m).ToHaveBeenLastCalledWith(2, "b") })
	expectFail(t, "last called fail", func(r TestReporter) { Expect(r, m).ToHaveBeenLastCalledWith(1, "a") })

	// A non-mock actual reports a distinct error.
	msg := expectFail(t, "non-mock", func(r TestReporter) { Expect(r, 5).ToHaveBeenCalled() })
	if !strings.Contains(msg, "jest.Mock") {
		t.Errorf("expected mock requirement error, got %s", msg)
	}

	empty := NewMock("empty")
	expectFail(t, "not called", func(r TestReporter) { Expect(r, empty).ToHaveBeenCalled() })
}

func TestMockReturnMatchers(t *testing.T) {
	m := NewMock("r")
	m.Return(42)
	m.Call()
	expectPass(t, "returned", func(r TestReporter) { Expect(r, m).ToHaveReturned() })
	expectPass(t, "returned with", func(r TestReporter) { Expect(r, m).ToHaveReturnedWith(42) })
	expectFail(t, "returned with fail", func(r TestReporter) { Expect(r, m).ToHaveReturnedWith(99) })

	// A panicking implementation records a non-returning call.
	p := NewMock("p")
	p.MockImplementation(func(_ ...any) []any { panic("boom") })
	Expect(t, func() { p.Call() }).ToPanic("boom")
	expectFail(t, "did not return", func(r TestReporter) { Expect(r, p).ToHaveReturned() })
	expectPass(t, "not returned", func(r TestReporter) { Expect(r, p).Not().ToHaveReturned() })
}
