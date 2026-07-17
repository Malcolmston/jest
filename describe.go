package jest

import (
	"sync"
	"testing"
)

// describeScope holds the setup/teardown hooks registered for a single Describe
// block, along with the *testing.T of that block so nested Describe and It calls
// attach beneath it.
type describeScope struct {
	t          *testing.T
	beforeEach []func()
	afterEach  []func()
}

var (
	scopeMu    sync.Mutex
	scopeStack []*describeScope
)

func currentScope() *describeScope {
	scopeMu.Lock()
	defer scopeMu.Unlock()
	if len(scopeStack) == 0 {
		return nil
	}
	return scopeStack[len(scopeStack)-1]
}

func pushScope(s *describeScope) {
	scopeMu.Lock()
	scopeStack = append(scopeStack, s)
	scopeMu.Unlock()
}

func popScope() {
	scopeMu.Lock()
	if len(scopeStack) > 0 {
		scopeStack = scopeStack[:len(scopeStack)-1]
	}
	scopeMu.Unlock()
}

// snapshotStack returns the current stack of scopes (outermost first) so that
// hooks from every enclosing Describe run for each It.
func snapshotStack() []*describeScope {
	scopeMu.Lock()
	defer scopeMu.Unlock()
	out := make([]*describeScope, len(scopeStack))
	copy(out, scopeStack)
	return out
}

// Describe groups related tests under a named subtest (via t.Run) and
// establishes a scope in which [BeforeEach] and [AfterEach] hooks may be
// registered. Describe blocks may be nested; inner blocks run under their
// parent's subtest.
func Describe(t *testing.T, name string, fn func()) {
	t.Helper()
	parent := t
	if sc := currentScope(); sc != nil {
		parent = sc.t
	}
	parent.Run(name, func(st *testing.T) {
		pushScope(&describeScope{t: st})
		defer popScope()
		fn()
	})
}

// It defines a single test case, run as a subtest (via t.Run) beneath the
// enclosing Describe block (if any). Any BeforeEach hooks from enclosing scopes
// run before fn, and AfterEach hooks run afterwards in reverse order.
func It(t *testing.T, name string, fn func(t *testing.T)) {
	t.Helper()
	parent := t
	if sc := currentScope(); sc != nil {
		parent = sc.t
	}
	stack := snapshotStack()
	parent.Run(name, func(st *testing.T) {
		for _, s := range stack {
			for _, before := range s.beforeEach {
				before()
			}
		}
		defer func() {
			for i := len(stack) - 1; i >= 0; i-- {
				for _, after := range stack[i].afterEach {
					after()
				}
			}
		}()
		fn(st)
	})
}

// Test is an alias for [It], provided for readability.
func Test(t *testing.T, name string, fn func(t *testing.T)) {
	t.Helper()
	It(t, name, fn)
}

// BeforeEach registers a setup function that runs before each [It] in the
// enclosing Describe scope. It panics if called outside of a Describe block.
func BeforeEach(fn func()) {
	sc := currentScope()
	if sc == nil {
		panic("jest: BeforeEach called outside of a Describe block")
	}
	scopeMu.Lock()
	sc.beforeEach = append(sc.beforeEach, fn)
	scopeMu.Unlock()
}

// AfterEach registers a teardown function that runs after each [It] in the
// enclosing Describe scope. It panics if called outside of a Describe block.
func AfterEach(fn func()) {
	sc := currentScope()
	if sc == nil {
		panic("jest: AfterEach called outside of a Describe block")
	}
	scopeMu.Lock()
	sc.afterEach = append(sc.afterEach, fn)
	scopeMu.Unlock()
}
