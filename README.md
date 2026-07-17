# jest

Testing, assertions and mocking framework for Go — a Jest-style fluent
assertion and mocking layer on top of the standard `testing` package. Pure Go
standard library, no third-party dependencies.

## Install

```sh
go get github.com/malcolmston/jest
```

Requires Go 1.24 or newer.

## Quick start

```go
package mypkg

import (
	"testing"

	"github.com/malcolmston/jest"
)

func TestExample(t *testing.T) {
	// Fluent expectations
	jest.Expect(t, 2+2).ToBe(4)
	jest.Expect(t, []int{1, 2, 3}).ToEqual([]int{1, 2, 3})
	jest.Expect(t, "hello world").ToContain("world")
	jest.Expect(t, 3.14159).ToBeCloseTo(3.14, 0.01)
	jest.Expect(t, 10).ToBeGreaterThan(5)
	jest.Expect(t, func() { panic("boom") }).ToPanic("boom")

	// Negate any matcher
	jest.Expect(t, 5).Not().ToBe(6)
}
```

### Matchers

| Matcher | Meaning |
| --- | --- |
| `ToBe(v)` | shallow equality (`==` semantics; pointer identity) |
| `ToEqual(v)` | deep equality (`reflect.DeepEqual`) with a diff on failure |
| `ToBeNil()` | nil check (typed and untyped nils) |
| `ToBeTrue()` / `ToBeFalse()` | boolean checks |
| `ToBeGreaterThan(n)` / `ToBeGreaterThanOrEqual(n)` | numeric ordering |
| `ToBeLessThan(n)` / `ToBeLessThanOrEqual(n)` | numeric ordering |
| `ToContain(x)` | substring / slice element / map key |
| `ToHaveLen(n)` | length of string, slice, array, map or channel |
| `ToMatch(pattern)` | regular-expression match |
| `ToBeCloseTo(v, eps...)` | float comparison within an epsilon (default `1e-9`) |
| `ToThrow(msg...)` / `ToPanic(msg...)` | assert a `func()` panics, optionally matching a message |
| `.Not()` | inverts the matcher that follows |

### Mocks and spies

```go
func TestMock(t *testing.T) {
	m := jest.NewMock("adder")
	m.Return(42)

	m.Call(1, 2) // -> [42]

	jest.Expect(t, m.CallCount()).ToBe(1)
	jest.Expect(t, m.CalledWith(1, 2)).ToBeTrue()

	// A sequence of results, one per call
	m.ReturnValues([]any{1}, []any{2}, []any{3})

	// Type-safe function mocks
	fn, mock := jest.Fn1[int, string]("stringify")
	mock.Return("hi")
	_ = fn(7) // "hi"
	jest.Expect(t, mock.CalledWith(7)).ToBeTrue()

	// Spies wrap a real implementation while recording calls
	double, spy := jest.Spy1("double", func(x int) int { return x * 2 })
	_ = double(21) // 42, recorded
	jest.Expect(t, spy.CalledWith(21)).ToBeTrue()
}
```

Inspection API: `CallCount()`, `Called()`, `CalledWith(...)`, `LastCall()`,
`NthCall(i)`, `Calls()`, `Reset()`. Configuration: `Return(...)`,
`ReturnValues(...)`. Typed helpers: `Fn0`/`Fn1`/`Fn2` and `Spy0`/`Spy1`/`Spy2`.

### Test organization

```go
func TestSuite(t *testing.T) {
	var counter int
	jest.Describe(t, "counter", func() {
		jest.BeforeEach(func() { counter = 0 })
		jest.AfterEach(func() { /* teardown */ })

		jest.It(t, "starts at zero", func(t *testing.T) {
			jest.Expect(t, counter).ToBe(0)
		})
		jest.It(t, "increments", func(t *testing.T) {
			counter++
			jest.Expect(t, counter).ToBe(1)
		})
	})
}
```

`Describe`/`It`/`Test` use `t.Run` under the hood, so subtests, `-run`
filtering and `go test -v` output all work as usual. `BeforeEach`/`AfterEach`
hooks are scoped to their enclosing `Describe` block and compose across nested
blocks.

## How failure reporting works

Matchers report through a small `TestReporter` interface (`Errorf`, `Fatalf`,
`Helper`) that `*testing.T` satisfies, so assertions plug straight into
`go test`. This indirection also lets the library test its own failure paths by
passing a fake reporter that records failures instead of failing the run.

## License

See repository.
