package jest_test

import (
	"fmt"
	"time"

	"github.com/malcolmston/jest"
)

// Example demonstrates configuring and inspecting a mock. Because mocks report
// through their own methods rather than *testing.T, they can be exercised in a
// runnable example with verified output.
func Example() {
	m := jest.NewMock("greeter")
	m.Return("hello")

	fmt.Println(m.Call("world")[0])
	fmt.Println(m.CallCount())
	fmt.Println(m.CalledWith("world"))

	fn, mock := jest.Fn1[int, int]("square")
	mock.ReturnValues([]any{1}, []any{4}, []any{9})
	fmt.Println(fn(1), fn(2), fn(3))
	fmt.Println(mock.CallCount())

	// Output:
	// hello
	// 1
	// true
	// 1 4 9
	// 3
}

// ExampleClock demonstrates driving virtual time with a fake clock.
func ExampleClock() {
	c := jest.NewClock()
	ticks := 0
	c.SetInterval(time.Second, func() { ticks++ })

	c.AdvanceTimersByTime(3 * time.Second)
	fmt.Println("ticks:", ticks)
	fmt.Println("now:", c.Now().Sub(time.Unix(0, 0)))

	// Output:
	// ticks: 3
	// now: 3s
}

// ExampleMock_mockImplementationOnce demonstrates one-shot implementations
// falling back to a base implementation.
func ExampleMock_mockImplementationOnce() {
	m := jest.NewMock("fetch")
	m.MockImplementation(func(_ ...any) []any { return []any{"live"} })
	m.MockReturnValueOnce("cached")

	fmt.Println(m.Call()[0])
	fmt.Println(m.Call()[0])

	// Output:
	// cached
	// live
}
