package jest_test

import (
	"fmt"

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
