package jest

import (
	"sync"
	"testing"
)

// describeScope holds the setup/teardown hooks registered for a single Describe
// block, along with the *testing.T of that block so nested Describe and It calls
// attach beneath it.
type describeScope struct {
	t             *testing.T
	beforeEach    []func()
	afterEach     []func()
	beforeAll     []func()
	afterAll      []func()
	beforeAllDone bool
	focused       bool // an ItOnly in this block focuses it, skipping plain Its
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
		sc := &describeScope{t: st}
		pushScope(sc)
		defer popScope()
		fn()
		// Every It in this block runs synchronously via t.Run during fn(), so by
		// the time fn returns all tests have completed and AfterAll hooks can run
		// (in reverse registration order).
		for i := len(sc.afterAll) - 1; i >= 0; i-- {
			sc.afterAll[i]()
		}
	})
}

// itMode selects the execution behavior of an It variant.
type itMode int

const (
	itNormal itMode = iota
	itSkip
	itOnly
	itTodo
)

// It defines a single test case, run as a subtest (via t.Run) beneath the
// enclosing Describe block (if any). BeforeAll hooks from enclosing scopes run
// once before the first test, BeforeEach hooks run before fn, and AfterEach
// hooks run afterwards in reverse order.
func It(t *testing.T, name string, fn func(t *testing.T)) {
	t.Helper()
	itImpl(t, name, fn, itNormal)
}

// ItSkip registers a test case that is always skipped, mirroring Jest's it.skip.
func ItSkip(t *testing.T, name string, fn func(t *testing.T)) {
	t.Helper()
	itImpl(t, name, fn, itSkip)
}

// ItOnly focuses a test case within its enclosing Describe block: when at least
// one ItOnly is declared before the other tests in a block, the plain [It] cases
// in that block are skipped, mirroring Jest's it.only. Focus applies within a
// Describe block; top-level ItOnly simply runs.
func ItOnly(t *testing.T, name string, fn func(t *testing.T)) {
	t.Helper()
	itImpl(t, name, fn, itOnly)
}

// ItTodo registers a placeholder test that is reported as skipped, mirroring
// Jest's it.todo. The function argument may be nil.
func ItTodo(t *testing.T, name string) {
	t.Helper()
	itImpl(t, name, func(*testing.T) {}, itTodo)
}

// itImpl is the shared implementation behind It and its variants.
func itImpl(t *testing.T, name string, fn func(t *testing.T), mode itMode) {
	t.Helper()
	sc := currentScope()
	parent := t
	if sc != nil {
		parent = sc.t
	}
	if mode == itOnly && sc != nil {
		scopeMu.Lock()
		sc.focused = true
		scopeMu.Unlock()
	}
	stack := snapshotStack()
	if mode == itNormal && focusActive(stack) {
		parent.Run(name, func(st *testing.T) { st.Skip("skipped: focused test(s) present in this block") })
		return
	}
	parent.Run(name, func(st *testing.T) {
		switch mode {
		case itSkip:
			st.Skip("skipped")
			return
		case itTodo:
			st.Skip("todo")
			return
		}
		runBeforeAll(stack)
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

// focusActive reports whether any scope in the stack has been focused by an
// ItOnly, so plain It cases should be skipped.
func focusActive(stack []*describeScope) bool {
	scopeMu.Lock()
	defer scopeMu.Unlock()
	for _, s := range stack {
		if s.focused {
			return true
		}
	}
	return false
}

// runBeforeAll runs, once per scope, the BeforeAll hooks of every scope in the
// stack (outermost first) before the first test that needs them.
func runBeforeAll(stack []*describeScope) {
	for _, s := range stack {
		scopeMu.Lock()
		if s.beforeAllDone {
			scopeMu.Unlock()
			continue
		}
		s.beforeAllDone = true
		hooks := append([]func(){}, s.beforeAll...)
		scopeMu.Unlock()
		for _, h := range hooks {
			h()
		}
	}
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

// BeforeAll registers a setup function that runs once, before the first [It] in
// the enclosing Describe scope. Declare it before the tests it prepares. It
// panics if called outside of a Describe block.
func BeforeAll(fn func()) {
	sc := currentScope()
	if sc == nil {
		panic("jest: BeforeAll called outside of a Describe block")
	}
	scopeMu.Lock()
	sc.beforeAll = append(sc.beforeAll, fn)
	scopeMu.Unlock()
}

// AfterAll registers a teardown function that runs once, after the last [It] in
// the enclosing Describe scope. It panics if called outside of a Describe block.
func AfterAll(fn func()) {
	sc := currentScope()
	if sc == nil {
		panic("jest: AfterAll called outside of a Describe block")
	}
	scopeMu.Lock()
	sc.afterAll = append(sc.afterAll, fn)
	scopeMu.Unlock()
}
