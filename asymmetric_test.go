package jest

import (
	"reflect"
	"strings"
	"testing"
)

func TestAsymmetricAny(t *testing.T) {
	expectPass(t, "any int", func(r TestReporter) { Expect[any](r, 5).ToEqual(Any(0)) })
	expectFail(t, "any int fail", func(r TestReporter) { Expect[any](r, "x").ToEqual(Any(0)) })
	expectPass(t, "any string", func(r TestReporter) { Expect[any](r, "hi").ToEqual(Any("")) })
	expectPass(t, "any reflect.Type", func(r TestReporter) { Expect[any](r, 3).ToEqual(Any(reflect.TypeOf(0))) })
	expectPass(t, "any nil non-nil", func(r TestReporter) { Expect[any](r, 3).ToEqual(Any(nil)) })
	expectFail(t, "any nil with nil", func(r TestReporter) { Expect[any](r, nil).ToEqual(Any(nil)) })

	// Interface target: match by implementing the interface.
	var errType = reflect.TypeOf((*error)(nil)).Elem()
	expectPass(t, "any error iface", func(r TestReporter) {
		Expect[any](r, strings.NewReplacer()).Not().ToEqual(Any(errType))
	})
}

func TestAsymmetricAnything(t *testing.T) {
	expectPass(t, "anything", func(r TestReporter) { Expect[any](r, 5).ToEqual(Anything()) })
	expectFail(t, "anything nil", func(r TestReporter) { Expect[any](r, nil).ToEqual(Anything()) })
	if got := Anything().String(); got != "Anything()" {
		t.Errorf("Anything().String() = %q", got)
	}
}

func TestAsymmetricStrings(t *testing.T) {
	expectPass(t, "contains", func(r TestReporter) { Expect[any](r, "hello world").ToEqual(StringContaining("wor")) })
	expectFail(t, "contains fail", func(r TestReporter) { Expect[any](r, "hello").ToEqual(StringContaining("z")) })
	expectFail(t, "contains non-string", func(r TestReporter) { Expect[any](r, 5).ToEqual(StringContaining("5")) })

	expectPass(t, "matching", func(r TestReporter) { Expect[any](r, "abc123").ToEqual(StringMatching(`\d+`)) })
	expectFail(t, "matching fail", func(r TestReporter) { Expect[any](r, "abc").ToEqual(StringMatching(`\d+`)) })
	if !strings.Contains(StringMatching(`\d`).String(), "StringMatching") {
		t.Error("StringMatching String() wrong")
	}
	if !strings.Contains(StringContaining("x").String(), "StringContaining") {
		t.Error("StringContaining String() wrong")
	}
}

func TestAsymmetricArrayContaining(t *testing.T) {
	expectPass(t, "array contains", func(r TestReporter) {
		Expect[any](r, []int{1, 2, 3}).ToEqual(ArrayContaining(2, 3))
	})
	expectFail(t, "array missing", func(r TestReporter) {
		Expect[any](r, []int{1, 2, 3}).ToEqual(ArrayContaining(9))
	})
	expectPass(t, "array with matcher elem", func(r TestReporter) {
		Expect[any](r, []any{"a", 2}).ToEqual(ArrayContaining(Any(0)))
	})
	expectFail(t, "array containing non-slice", func(r TestReporter) {
		Expect[any](r, "abc").ToEqual(ArrayContaining("a"))
	})
	if !strings.Contains(ArrayContaining(1).String(), "ArrayContaining") {
		t.Error("ArrayContaining String() wrong")
	}
}

func TestAsymmetricObjectContaining(t *testing.T) {
	obj := map[string]any{"name": "bob", "age": 30, "extra": true}
	expectPass(t, "object subset", func(r TestReporter) {
		Expect[any](r, obj).ToEqual(ObjectContaining(map[string]any{"name": "bob"}))
	})
	expectPass(t, "object subset matcher", func(r TestReporter) {
		Expect[any](r, obj).ToEqual(ObjectContaining(map[string]any{"age": Any(0)}))
	})
	expectFail(t, "object missing key", func(r TestReporter) {
		Expect[any](r, obj).ToEqual(ObjectContaining(map[string]any{"missing": 1}))
	})
	expectFail(t, "object wrong value", func(r TestReporter) {
		Expect[any](r, obj).ToEqual(ObjectContaining(map[string]any{"name": "alice"}))
	})

	type person struct {
		Name string
		Age  int
	}
	expectPass(t, "struct subset", func(r TestReporter) {
		Expect[any](r, person{"bob", 30}).ToEqual(ObjectContaining(map[string]any{"Name": "bob"}))
	})
	if !strings.Contains(ObjectContaining(nil).String(), "ObjectContaining") {
		t.Error("ObjectContaining String() wrong")
	}
}

func TestAsymmetricNested(t *testing.T) {
	actual := map[string]any{
		"user": map[string]any{"id": 7, "name": "bob"},
		"tags": []any{"a", "b"},
	}
	expectPass(t, "nested matchers", func(r TestReporter) {
		Expect[any](r, actual).ToEqual(map[string]any{
			"user": map[string]any{"id": Any(0), "name": StringContaining("bo")},
			"tags": ArrayContaining("a"),
		})
	})
	expectFail(t, "nested fail", func(r TestReporter) {
		Expect[any](r, actual).ToEqual(map[string]any{
			"user": map[string]any{"id": Any(""), "name": "bob"},
			"tags": ArrayContaining("a"),
		})
	})
}

func TestAsymEqualUnexportedFields(t *testing.T) {
	type rec struct {
		ID     any
		secret int
	}
	// Exported field satisfied by a matcher; unexported fields equal.
	expectPass(t, "unexported equal + matcher", func(r TestReporter) {
		Expect(r, rec{ID: 5, secret: 1}).ToEqual(rec{ID: Any(0), secret: 1})
	})
	// Unexported fields differ -> not equal.
	expectFail(t, "unexported differ", func(r TestReporter) {
		Expect(r, rec{ID: 5, secret: 1}).ToEqual(rec{ID: Any(0), secret: 2})
	})
}

func TestAsymEqualAgreesWithDeepEqual(t *testing.T) {
	cases := []struct{ a, b any }{
		{[]int{1, 2}, []int{1, 2}},
		{map[string]int{"a": 1}, map[string]int{"a": 1}},
		{[]int{1, 2}, []int{1, 3}},
		{struct{ X int }{1}, struct{ X int }{2}},
		{nil, nil},
	}
	for i, c := range cases {
		if got, want := asymEqual(c.a, c.b), reflect.DeepEqual(c.a, c.b); got != want {
			t.Errorf("case %d: asymEqual=%v deepEqual=%v", i, got, want)
		}
	}
}
