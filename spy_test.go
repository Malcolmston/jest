package jest

import (
	"strings"
	"testing"
)

func TestSpyOnFuncVar(t *testing.T) {
	greet := func(name string) string { return "hi " + name }
	spy := SpyOn(&greet)
	defer spy.Restore()

	Expect(t, greet("bob")).ToBe("hi bob")
	Expect(t, spy.CallCount()).ToBe(1)
	Expect(t, spy).ToHaveBeenCalledWith("bob")
	Expect(t, spy).ToHaveReturnedWith("hi bob")

	// Override with a mock implementation.
	spy.MockImplementation(func(_ ...any) []any { return []any{"mocked"} })
	Expect(t, greet("x")).ToBe("mocked")

	spy.Restore()
	Expect(t, greet("bob")).ToBe("hi bob")
}

func TestSpyOnStructField(t *testing.T) {
	type service struct {
		Fetch func(id int) (string, error)
	}
	s := service{Fetch: func(id int) (string, error) { return "real", nil }}
	spy := SpyOn(&s.Fetch)
	defer spy.Restore()

	v, err := s.Fetch(7)
	Expect(t, v).ToBe("real")
	Expect(t, err).ToBeNil()
	Expect(t, spy).ToHaveBeenCalledWith(7)

	// One-shot override, then fall back to the real implementation.
	spy.MockReturnValueOnce("cached", nil)
	v, _ = s.Fetch(8)
	Expect(t, v).ToBe("cached")
	v, _ = s.Fetch(9)
	Expect(t, v).ToBe("real")
}

func TestSpyOnVariadic(t *testing.T) {
	sum := func(nums ...int) int {
		total := 0
		for _, n := range nums {
			total += n
		}
		return total
	}
	spy := SpyOn(&sum)
	defer spy.Restore()
	Expect(t, sum(1, 2, 3)).ToBe(6)
	Expect(t, spy).ToHaveBeenCalledWith(1, 2, 3)
}

func TestRestoreAllMocks(t *testing.T) {
	orig := func() string { return "original" }
	fn := orig
	spy := SpyOn(&fn)
	spy.MockImplementation(func(_ ...any) []any { return []any{"spied"} })
	Expect(t, fn()).ToBe("spied")

	RestoreAllMocks()
	Expect(t, fn()).ToBe("original")
}

func TestSpyOnNonFuncPanics(t *testing.T) {
	x := 5
	Expect(t, func() { SpyOn(&x) }).ToPanic("func")
}

func TestSpyOnPropagatesPanic(t *testing.T) {
	boom := func() { panic("kaboom") }
	fn := boom
	spy := SpyOn(&fn)
	defer spy.Restore()
	msg := ""
	func() {
		defer func() {
			if r := recover(); r != nil {
				msg = strings.TrimSpace(r.(string))
			}
		}()
		fn()
	}()
	Expect(t, msg).ToBe("kaboom")
	Expect(t, spy).Not().ToHaveReturned()
}
