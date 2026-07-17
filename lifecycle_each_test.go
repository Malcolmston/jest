package jest

import (
	"strings"
	"testing"
)

func TestBeforeAllAfterAll(t *testing.T) {
	var order []string
	Describe(t, "suite", func() {
		BeforeAll(func() { order = append(order, "beforeAll") })
		AfterAll(func() { order = append(order, "afterAll") })
		BeforeEach(func() { order = append(order, "beforeEach") })

		It(t, "a", func(t *testing.T) { order = append(order, "a") })
		It(t, "b", func(t *testing.T) { order = append(order, "b") })
	})
	joined := strings.Join(order, ",")
	want := "beforeAll,beforeEach,a,beforeEach,b,afterAll"
	if joined != want {
		t.Errorf("lifecycle order wrong:\n got: %s\nwant: %s", joined, want)
	}
}

func TestBeforeAllNested(t *testing.T) {
	var order []string
	Describe(t, "outer", func() {
		BeforeAll(func() { order = append(order, "outer-all") })
		Describe(t, "inner", func() {
			BeforeAll(func() { order = append(order, "inner-all") })
			It(t, "case", func(t *testing.T) { order = append(order, "case") })
		})
	})
	if got := strings.Join(order, ","); got != "outer-all,inner-all,case" {
		t.Errorf("nested BeforeAll order = %s", got)
	}
}

func TestLifecycleHooksOutsideScopePanic(t *testing.T) {
	Expect(t, func() { BeforeAll(func() {}) }).ToPanic("Describe")
	Expect(t, func() { AfterAll(func() {}) }).ToPanic("Describe")
}

func TestItOnly(t *testing.T) {
	ran := map[string]bool{}
	Describe(t, "focus", func() {
		ItOnly(t, "focused", func(t *testing.T) { ran["focused"] = true })
		It(t, "plain", func(t *testing.T) { ran["plain"] = true })
	})
	Expect(t, ran["focused"]).ToBeTrue()
	Expect(t, ran["plain"]).ToBeFalse()
}

func TestItSkipAndTodo(t *testing.T) {
	ranSkip := false
	Describe(t, "skips", func() {
		ItSkip(t, "skipped", func(t *testing.T) { ranSkip = true })
		ItTodo(t, "todo item")
	})
	Expect(t, ranSkip).ToBeFalse()
}

func TestEach(t *testing.T) {
	type tc struct{ In, Out int }
	var seen []int
	Each(t, "square of %v", []tc{{2, 4}, {3, 9}}, func(t *testing.T, c tc) {
		seen = append(seen, c.In)
		Expect(t, c.In*c.In).ToBe(c.Out)
	})
	Expect(t, seen).ToEqual([]int{2, 3})
}

func TestEachPlainName(t *testing.T) {
	var count int
	Each(t, "runs", []int{10, 20, 30}, func(t *testing.T, n int) { count++ })
	Expect(t, count).ToBe(3)
	// Verify the generated name uses an index suffix when no verb is present.
	Expect(t, eachName("runs", 10, 0)).ToBe("runs [0]")
	Expect(t, eachName("val %d", 7, 0)).ToBe("val 7")
}

func TestDescribeEach(t *testing.T) {
	var count int
	DescribeEach(t, "group", []string{"x", "y"}, func(s string) {
		It(t, "runs a case", func(t *testing.T) { count++ })
	})
	Expect(t, count).ToBe(2)
}
