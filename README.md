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
| `ToMatchObject(subset)` | recursive subset match against a map or struct |
| `ToStrictEqual(v)` | deep equality that is also strict about dynamic types |
| `ToHaveProperty(path, v...)` | dotted / indexed property lookup (`"a.b[0].c"`) |
| `ToBeInstanceOf(type)` | dynamic-type or interface-implementation check |
| `ToBeDefined()` / `ToBeUndefined()` | non-nil / nil checks |
| `ToBeNaN()` | floating-point `NaN` check |
| `ToMatchSnapshot(name...)` | compare against an on-disk snapshot |
| `.Not()` | inverts the matcher that follows |

### Asymmetric matchers

Asymmetric matchers match a range of values and can be nested anywhere on the
expected side of `ToEqual`, `ToMatchObject`, `ToHaveProperty` and the
call-argument matchers:

```go
jest.Expect[any](t, resp).ToEqual(map[string]any{
	"id":    jest.Any(0),                    // any int
	"name":  jest.StringContaining("bo"),    // substring
	"email": jest.StringMatching(`@`),       // regexp
	"roles": jest.ArrayContaining("admin"),  // slice contains
	"meta":  jest.ObjectContaining(map[string]any{"ok": true}),
	"extra": jest.Anything(),                // any non-nil
})
```

### Snapshots

```go
jest.Expect(t, render()).ToMatchSnapshot()
```

Snapshots are written under `__snapshots__/` on first run and compared on
subsequent runs. Refresh them with `JEST_UPDATE_SNAPSHOTS=1 go test ./...` (or
`jest.SetUpdateSnapshots(true)`).

### Fake timers

```go
c := jest.NewClock()
c.SetTimeout(time.Second, func() { /* ... */ })
c.SetInterval(time.Second, tick)
c.AdvanceTimersByTime(3 * time.Second) // fires due timers deterministically
c.RunAllTimers()
```

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
`NthCall(i)`, `Calls()`, `Results()`, `Reset()`. Configuration: `Return(...)`,
`ReturnValues(...)`, `MockImplementation(...)`, `MockImplementationOnce(...)`,
`MockReturnValueOnce(...)`, `MockResolvedValue(...)`, `MockRejectedValue(...)`.
Typed helpers: `Fn0`/`Fn1`/`Fn2` and `Spy0`/`Spy1`/`Spy2`.

Mock-oriented matchers work directly on a `*Mock` (or a spy):

```go
jest.Expect(t, m).ToHaveBeenCalledTimes(2)
jest.Expect(t, m).ToHaveBeenCalledWith(jest.Any(0), "b")
jest.Expect(t, m).ToHaveBeenLastCalledWith(2, "b")
jest.Expect(t, m).ToHaveReturnedWith(42)
```

`SpyOn` replaces a function variable or struct field in place, recording calls
while delegating to the original; `RestoreAllMocks()` reinstates every original:

```go
spy := jest.SpyOn(&client.Fetch)
defer spy.Restore()
client.Fetch(7)
jest.Expect(t, spy).ToHaveBeenCalledWith(7)
```

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
blocks. `BeforeAll`/`AfterAll` run once per block. `ItSkip`, `ItOnly` and
`ItTodo` mark individual cases (focus applies within a `Describe` block).

### Parameterized tests

```go
jest.Each(t, "square of %v", []struct{ In, Out int }{{2, 4}, {3, 9}},
	func(t *testing.T, tc struct{ In, Out int }) {
		jest.Expect(t, tc.In*tc.In).ToBe(tc.Out)
	})
```

`DescribeEach` runs a whole `Describe` block once per case.

### Custom matchers and assertion counts

```go
jest.Extend(map[string]jest.CustomMatcher{
	"ToBeEven": func(actual any, _ ...any) jest.MatcherResult {
		n, ok := actual.(int)
		return jest.MatcherResult{Pass: ok && n%2 == 0, Message: "expected an even number"}
	},
})
jest.Expect(t, 4).To("ToBeEven")

jest.Assertions(t, 2) // fail unless exactly 2 assertions run
jest.HasAssertions(t) // fail unless at least 1 runs
```

## How failure reporting works

Matchers report through a small `TestReporter` interface (`Errorf`, `Fatalf`,
`Helper`) that `*testing.T` satisfies, so assertions plug straight into
`go test`. This indirection also lets the library test its own failure paths by
passing a fake reporter that records failures instead of failing the run.

## License

See repository.
