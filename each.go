package jest

import (
	"fmt"
	"strings"
	"testing"
)

// Each runs a table-driven test: for every entry in cases it invokes fn as a
// separate [It] subtest, mirroring Jest's it.each. If name contains a fmt verb
// (a '%'), the subtest name is formatted with the case value; otherwise the
// case index is appended.
//
//	jest.Each(t, "square of %v", []struct{ In, Out int }{{2, 4}, {3, 9}},
//	    func(t *testing.T, tc struct{ In, Out int }) {
//	        jest.Expect(t, tc.In*tc.In).ToBe(tc.Out)
//	    })
func Each[C any](t *testing.T, name string, cases []C, fn func(t *testing.T, tc C)) {
	t.Helper()
	for i, c := range cases {
		cc := c
		It(t, eachName(name, cc, i), func(st *testing.T) { fn(st, cc) })
	}
}

// DescribeEach runs a parameterized [Describe] block once per case, mirroring
// Jest's describe.each. The supplied fn receives the case value and registers
// tests inside the block.
func DescribeEach[C any](t *testing.T, name string, cases []C, fn func(tc C)) {
	t.Helper()
	for i, c := range cases {
		cc := c
		Describe(t, eachName(name, cc, i), func() { fn(cc) })
	}
}

// eachName builds a subtest name for a table case.
func eachName[C any](name string, c C, i int) string {
	if strings.Contains(name, "%") {
		return fmt.Sprintf(name, c)
	}
	return fmt.Sprintf("%s [%d]", name, i)
}
