package jest

import "testing"

func TestParityCloseTo(t *testing.T) {
	cases := []struct {
		name      string
		expected  float64
		precision []int
		actual    any
		match     bool
	}{
		{"default within", 0.3, nil, 0.3001, true},
		{"default outside", 0.3, nil, 0.31, false},
		{"precision0 within", 5, []int{0}, 5.4, true},
		{"precision0 outside", 5, []int{0}, 5.6, false},
		{"precision3 within", 1.234, []int{3}, 1.2342, true},
		{"integer actual", 5, []int{0}, 5, true},
		{"non numeric", 1, nil, "x", false},
	}
	for _, c := range cases {
		c := c
		got := CloseTo(c.expected, c.precision...).Matches(c.actual)
		if got != c.match {
			t.Errorf("%s: CloseTo(%g).Matches(%v)=%v, want %v", c.name, c.expected, c.actual, got, c.match)
		}
	}
	// Usable inside ToEqual.
	expectPass(t, "in equal", func(r TestReporter) {
		Expect[any](r, map[string]any{"pi": 3.14159}).ToEqual(map[string]any{"pi": CloseTo(3.14, 2)})
	})
	if CloseTo(1, 2).String() == "" {
		t.Error("String should be non-empty")
	}
}

func TestParityNotMatchers(t *testing.T) {
	if !NotArrayContaining("z").Matches([]string{"a", "b"}) {
		t.Error("NotArrayContaining should match a slice lacking z")
	}
	if NotArrayContaining("a").Matches([]string{"a", "b"}) {
		t.Error("NotArrayContaining should not match a slice containing a")
	}
	if !NotObjectContaining(map[string]any{"x": 1}).Matches(map[string]any{"y": 2}) {
		t.Error("NotObjectContaining should match object lacking x")
	}
	if NotObjectContaining(map[string]any{"x": 1}).Matches(map[string]any{"x": 1}) {
		t.Error("NotObjectContaining should not match object with x")
	}
	if !NotStringContaining("z").Matches("abc") {
		t.Error("NotStringContaining should match abc")
	}
	if NotStringContaining("b").Matches("abc") {
		t.Error("NotStringContaining should not match abc")
	}
	if !NotStringMatching(`\d`).Matches("abc") {
		t.Error("NotStringMatching should match non-digit string")
	}
	if NotStringMatching(`\d`).Matches("a1c") {
		t.Error("NotStringMatching should not match string with digit")
	}
	if NotStringContaining("z").String() != `Not(StringContaining("z"))` {
		t.Errorf("unexpected String: %s", NotStringContaining("z").String())
	}
}
