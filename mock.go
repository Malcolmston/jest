package jest

import (
	"reflect"
	"sync"
)

// Call is a single recorded invocation of a [Mock]: the arguments it was called
// with and the results it returned.
type Call struct {
	Args    []any
	Results []any
}

// Mock records the calls made to it and can be configured with canned return
// values. It is safe for concurrent use.
type Mock struct {
	mu       sync.Mutex
	name     string
	calls    []Call
	fixed    []any   // default result set returned by every call
	hasFixed bool    // whether Return has been configured
	seq      [][]any // sequence of result sets consumed one per call
	seqIndex int
	impl     func(args ...any) []any // optional live implementation (spies)
}

// NewMock creates a new mock with the given descriptive name.
func NewMock(name string) *Mock {
	return &Mock{name: name}
}

// Name returns the mock's descriptive name.
func (m *Mock) Name() string { return m.name }

// Return configures the mock to return the given values from every call
// (unless a sequence configured with [Mock.ReturnValues] still has entries
// left). It returns the mock to allow chaining.
func (m *Mock) Return(values ...any) *Mock {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.fixed = values
	m.hasFixed = true
	return m
}

// ReturnValues configures a sequence of result sets, one consumed per call. Once
// the sequence is exhausted, subsequent calls fall back to the value configured
// with [Mock.Return], or to the last sequence entry if Return was never called.
// It returns the mock to allow chaining.
func (m *Mock) ReturnValues(sets ...[]any) *Mock {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.seq = sets
	m.seqIndex = 0
	return m
}

// Call records an invocation with the given arguments and returns the
// configured results.
func (m *Mock) Call(args ...any) []any {
	m.mu.Lock()
	var results []any
	switch {
	case m.impl != nil:
		// Live implementation (spy): release the lock while calling through so
		// re-entrant inspection does not deadlock.
		impl := m.impl
		m.mu.Unlock()
		results = impl(args...)
		m.mu.Lock()
	case m.seqIndex < len(m.seq):
		results = m.seq[m.seqIndex]
		m.seqIndex++
	case m.hasFixed:
		results = m.fixed
	case len(m.seq) > 0:
		results = m.seq[len(m.seq)-1]
	}
	m.calls = append(m.calls, Call{Args: args, Results: results})
	m.mu.Unlock()
	return results
}

// CallCount returns the number of times the mock has been called.
func (m *Mock) CallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.calls)
}

// Called reports whether the mock has been called at least once.
func (m *Mock) Called() bool { return m.CallCount() > 0 }

// Calls returns a copy of all recorded calls in invocation order.
func (m *Mock) Calls() []Call {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Call, len(m.calls))
	copy(out, m.calls)
	return out
}

// CalledWith reports whether the mock was ever called with exactly the given
// arguments (compared with reflect.DeepEqual).
func (m *Mock) CalledWith(args ...any) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, c := range m.calls {
		if argsEqual(c.Args, args) {
			return true
		}
	}
	return false
}

// LastCall returns the most recent recorded call and true, or a zero Call and
// false if the mock has not been called.
func (m *Mock) LastCall() (Call, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.calls) == 0 {
		return Call{}, false
	}
	return m.calls[len(m.calls)-1], true
}

// NthCall returns the call at the given zero-based index and true, or a zero
// Call and false if the index is out of range.
func (m *Mock) NthCall(i int) (Call, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if i < 0 || i >= len(m.calls) {
		return Call{}, false
	}
	return m.calls[i], true
}

// Reset clears the recorded call history without changing the configured return
// values.
func (m *Mock) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = nil
	m.seqIndex = 0
}

func argsEqual(a, b []any) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !reflect.DeepEqual(a[i], b[i]) {
			return false
		}
	}
	return true
}

// castResult returns the i-th value in results converted to type R, or the zero
// value of R when the entry is missing or of an incompatible type.
func castResult[R any](results []any, i int) R {
	var zero R
	if i < 0 || i >= len(results) {
		return zero
	}
	if results[i] == nil {
		return zero
	}
	if v, ok := results[i].(R); ok {
		return v
	}
	return zero
}

// Fn0 returns a type-safe mock function taking no arguments and returning R,
// together with the backing [Mock] used to configure and inspect it.
func Fn0[R any](name string) (func() R, *Mock) {
	m := NewMock(name)
	return func() R {
		return castResult[R](m.Call(), 0)
	}, m
}

// Fn1 returns a type-safe mock function taking one argument and returning R,
// together with the backing [Mock].
func Fn1[A any, R any](name string) (func(A) R, *Mock) {
	m := NewMock(name)
	return func(a A) R {
		return castResult[R](m.Call(a), 0)
	}, m
}

// Fn2 returns a type-safe mock function taking two arguments and returning R,
// together with the backing [Mock].
func Fn2[A any, B any, R any](name string) (func(A, B) R, *Mock) {
	m := NewMock(name)
	return func(a A, b B) R {
		return castResult[R](m.Call(a, b), 0)
	}, m
}

// Spy0 wraps a zero-argument function, recording each call (and its result)
// while delegating to fn. It returns the wrapper and the backing [Mock].
func Spy0[R any](name string, fn func() R) (func() R, *Mock) {
	m := NewMock(name)
	m.impl = func(_ ...any) []any { return []any{fn()} }
	return func() R {
		return castResult[R](m.Call(), 0)
	}, m
}

// Spy1 wraps a one-argument function, recording each call while delegating to
// fn. It returns the wrapper and the backing [Mock].
func Spy1[A any, R any](name string, fn func(A) R) (func(A) R, *Mock) {
	m := NewMock(name)
	m.impl = func(args ...any) []any { return []any{fn(args[0].(A))} }
	return func(a A) R {
		return castResult[R](m.Call(a), 0)
	}, m
}

// Spy2 wraps a two-argument function, recording each call while delegating to
// fn. It returns the wrapper and the backing [Mock].
func Spy2[A any, B any, R any](name string, fn func(A, B) R) (func(A, B) R, *Mock) {
	m := NewMock(name)
	m.impl = func(args ...any) []any { return []any{fn(args[0].(A), args[1].(B))} }
	return func(a A, b B) R {
		return castResult[R](m.Call(a, b), 0)
	}, m
}
