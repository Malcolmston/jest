// Library content for the jest documentation site. Mirrors the shape used by
// the malcolmston/go landing site's data.ts so the sibling sites stay in sync.
export interface Lib {
  id: string; name: string; icon: string; accent: string; pkg: string; node: string;
  repo: string; docs: string; tagline: string; blurb: string; tags: string[];
  features: string[]; node_code: string; go_code: string; integrate: string;
}

export const NODE_ACCENT = '#8cc84b';

export const JEST: Lib = {
  id:"jest", name:"Jest", icon:'<i class="fa-solid fa-vial-circle-check"></i>', accent:"#e64b3c",
  pkg:"github.com/malcolmston/jest", node:"jestjs/jest",
  repo:"https://github.com/malcolmston/jest", docs:"https://malcolmston.github.io/jest/",
  tagline:"Jest-style fluent assertions and mocking for Go.",
  blurb:"A standard-library-only Go framework that brings the expressive, fluent feel of JavaScript's Jest to "+
    "Go tests, layered directly on top of the standard testing package. The entry point is a generic "+
    "Expect[T] that returns a fluent Matcher[T] with a full family of matchers — ToBe, ToEqual, ToContain, "+
    "ToHaveLen, ToMatch, ToBeCloseTo, numeric ordering and ToPanic — each invertible with .Not(). On top of "+
    "that sit configurable Mocks (Return/ReturnValues plus CallCount/CalledWith/LastCall inspection), "+
    "type-safe Fn0/Fn1/Fn2 and Spy helpers, and Describe/It/BeforeEach/AfterEach that run over t.Run through "+
    "a small TestReporter interface satisfied by *testing.T. No cgo, no third-party dependencies — assertions "+
    "plug straight into go test.",
  tags:["Expect[T]","Matcher[T]",".Not()","ToBe / ToEqual","ToContain","ToMatch","ToBeCloseTo","mocks","spies","Fn1 / Fn2","Describe / It","stdlib-only"],
  features:[
    "Generic fluent assertions — <code>Expect[T]</code> returns a <code>Matcher[T]</code> that reports through a <code>TestReporter</code> satisfied by <code>*testing.T</code>",
    "Equality &amp; nil — <code>ToBe</code> (shallow ==), <code>ToEqual</code> (deep equal with a diff), <code>ToBeNil</code>, <code>ToBeTrue</code>/<code>ToBeFalse</code>",
    "Containment &amp; shape — <code>ToContain</code> (substring / slice element / map key), <code>ToHaveLen</code>, <code>ToMatch</code> (regexp)",
    "Numbers — <code>ToBeCloseTo</code> with an epsilon plus <code>ToBeGreaterThan</code>/<code>ToBeGreaterThanOrEqual</code>/<code>ToBeLessThan</code>/<code>ToBeLessThanOrEqual</code>",
    "Panics &amp; negation — <code>ToPanic</code>/<code>ToThrow</code> assert a <code>func()</code> panics, and any matcher inverts with <code>.Not()</code>",
    "Mocks — <code>NewMock</code> with <code>Return</code>/<code>ReturnValues</code> and <code>CallCount</code>/<code>Called</code>/<code>CalledWith</code>/<code>LastCall</code>/<code>NthCall</code>/<code>Calls</code>/<code>Reset</code> inspection",
    "Type-safe function doubles — <code>Fn0</code>/<code>Fn1</code>/<code>Fn2</code> mock functions and <code>Spy0</code>/<code>Spy1</code>/<code>Spy2</code> that record while delegating to a real impl",
    "Test organization — <code>Describe</code>/<code>It</code> (alias <code>Test</code>) over <code>t.Run</code> with scoped, composable <code>BeforeEach</code>/<code>AfterEach</code> hooks",
    "Zero dependencies — pure Go standard library, fully integrated with <code>go test</code>, <code>-run</code> filtering and <code>-v</code> output"
  ],
  node_code:
`test('math and mocks', () => {
  expect(2 + 2).toBe(4);
  expect([1, 2, 3]).toContain(2);
  expect('hello world').toMatch(/world/);
  expect(5).not.toBe(6);

  const fn = jest.fn().mockReturnValue('hi');
  expect(fn(7)).toBe('hi');
  expect(fn).toHaveBeenCalledWith(7);
});`,
  go_code:
`import "github.com/malcolmston/jest"

func TestMathAndMocks(t *testing.T) {
	jest.Expect(t, 2+2).ToBe(4)
	jest.Expect(t, []int{1, 2, 3}).ToContain(2)
	jest.Expect(t, "hello world").ToMatch("world")
	jest.Expect(t, 5).Not().ToBe(6)

	fn, mock := jest.Fn1[int, string]("stringify")
	mock.Return("hi")
	jest.Expect(t, fn(7)).ToBe("hi")
	jest.Expect(t, mock.CalledWith(7)).ToBeTrue()
}`,
  integrate:
`func TestCounter(t *testing.T) {
	var counter int

	<span class="tok-c">// Describe/It run over t.Run, so subtests, -run and -v all work.</span>
	jest.Describe(t, "counter", func() {
		<span class="tok-c">// Hooks are scoped to this block and compose across nesting.</span>
		jest.BeforeEach(func() { counter = 0 })

		jest.It(t, "starts at zero", func(t *testing.T) {
			jest.Expect(t, counter).ToBe(0)
		})
		jest.It(t, "increments", func(t *testing.T) {
			counter++
			jest.Expect(t, counter).ToBe(1)
		})
	})

	<span class="tok-c">// Mocks record every call for later inspection.</span>
	m := jest.NewMock("adder")
	m.Return(42)
	m.Call(1, 2)

	jest.Expect(t, m.CallCount()).ToBe(1)
	jest.Expect(t, m.CalledWith(1, 2)).ToBeTrue()
	if last, ok := m.LastCall(); ok {
		jest.Expect(t, last.Args).ToEqual([]any{1, 2})
	}
}`
};
