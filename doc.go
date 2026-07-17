// Package jest provides a Jest-style assertion and mocking framework layered
// on top of Go's standard testing package. It aims to give Go tests the same
// expressive, fluent feel that the JavaScript Jest framework offers, while
// remaining 100% standard-library only and fully integrated with `go test`.
//
// # Expectations
//
// The entry point for assertions is [Expect], a generic function that returns a
// fluent [Matcher]. Every matcher reports failures through the supplied
// [TestReporter] (satisfied by *testing.T), so assertions plug straight into
// the normal `go test` flow:
//
//	func TestNumbers(t *testing.T) {
//	    jest.Expect(t, 2+2).ToBe(4)
//	    jest.Expect(t, []int{1, 2, 3}).ToHaveLen(3)
//	    jest.Expect(t, "hello world").ToContain("world")
//	    jest.Expect(t, 3.14159).ToBeCloseTo(3.14, 0.01)
//	}
//
// Available matchers:
//
//   - ToBe(expected)                    – shallow equality (== semantics)
//   - ToEqual(expected)                 – deep equality (reflect.DeepEqual) with a diff
//   - ToBeNil()                         – nil check (typed and untyped nils)
//   - ToBeTrue() / ToBeFalse()          – boolean checks
//   - ToBeGreaterThan(n)                – numeric ordering (>)
//   - ToBeGreaterThanOrEqual(n)         – numeric ordering (>=)
//   - ToBeLessThan(n)                   – numeric ordering (<)
//   - ToBeLessThanOrEqual(n)            – numeric ordering (<=)
//   - ToContain(item)                   – substring / slice element / map key
//   - ToHaveLen(n)                      – length of string, slice, array, map or channel
//   - ToMatch(pattern)                  – regular-expression match on a string
//   - ToBeCloseTo(expected, epsilon...) – float comparison within an epsilon
//   - ToThrow(msg...) / ToPanic(msg...) – assert that a func value panics
//
// Any matcher can be inverted with [Matcher.Not]:
//
//	jest.Expect(t, 5).Not().ToBe(6)
//	jest.Expect(t, "abc").Not().ToContain("z")
//
// # Mocks and spies
//
// [Mock] records the calls made to it (arguments and call count) and can be
// configured to return canned values, either a single fixed result set via
// [Mock.Return] or a sequence via [Mock.ReturnValues]. Recorded calls are
// inspected with [Mock.CallCount], [Mock.CalledWith] and [Mock.LastCall].
//
//	m := jest.NewMock("adder")
//	m.Return(42)
//	m.Call(1, 2) // -> [42]
//	jest.Expect(t, m.CallCount()).ToBe(1)
//	jest.Expect(t, m.CalledWith(1, 2)).ToBeTrue()
//
// Type-safe function mocks are produced with the generic [Fn0], [Fn1] and
// [Fn2] helpers, which return a real Go function plus the backing [Mock]:
//
//	fn, m := jest.Fn1[int, string]("stringify")
//	m.Return("hi")
//	got := fn(7) // "hi"
//	jest.Expect(t, m.CalledWith(7)).ToBeTrue()
//
// Spies ([Spy0], [Spy1], [Spy2]) wrap an existing function, recording calls
// while delegating to the real implementation.
//
// # Test organization
//
// [Describe], [It] (and its alias [Test]) organize tests using t.Run under the
// hood, and [BeforeEach]/[AfterEach] register setup/teardown hooks scoped to
// the enclosing Describe block:
//
//	func TestSuite(t *testing.T) {
//	    var counter int
//	    jest.Describe(t, "counter", func() {
//	        jest.BeforeEach(func() { counter = 0 })
//	        jest.It(t, "starts at zero", func(t *testing.T) {
//	            jest.Expect(t, counter).ToBe(0)
//	        })
//	        jest.It(t, "increments", func(t *testing.T) {
//	            counter++
//	            jest.Expect(t, counter).ToBe(1)
//	        })
//	    })
//	}
package jest
