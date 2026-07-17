package jest

import (
	"reflect"
	"sync"
)

// Spy wraps a function value (typically stored in a package-level variable or an
// exported struct field) so that calls through it are recorded, while still
// delegating to the original implementation. The embedded [Mock] provides the
// full inspection API ([Mock.CallCount], [Mock.CalledWith], and so on), and the
// original function is reinstated with [Spy.Restore] (or [RestoreAllMocks]).
type Spy struct {
	*Mock
	restore func()
}

// mock exposes the backing mock to the mock-oriented matchers, so a *Spy can be
// passed directly to jest.Expect for ToHaveBeenCalled and friends.
func (s *Spy) mock() *Mock { return s.Mock }

// Restore reinstates the original function value that was replaced by [SpyOn].
func (s *Spy) Restore() { s.restore() }

var (
	spyRegistryMu sync.Mutex
	spyRegistry   []*Spy
)

// SpyOn replaces the function value pointed to by target with a recording
// wrapper of the same type and returns a [Spy] for inspection and restoration.
// target must point to a variable or addressable struct field of func type;
// SpyOn panics otherwise. By default the wrapper delegates to the original
// implementation; set a replacement with [Mock.MockImplementation] (or the
// one-shot / return-value variants) on the returned spy's mock. Every spy is
// registered so that [RestoreAllMocks] can undo it.
//
//	var greet = func(name string) string { return "hi " + name }
//	spy := jest.SpyOn(&greet)
//	defer spy.Restore()
//	greet("bob")
//	jest.Expect(t, spy).ToHaveBeenCalledWith("bob")
func SpyOn[T any](target *T) *Spy {
	orig := *target
	fv := reflect.ValueOf(orig)
	if fv.Kind() != reflect.Func {
		panic("jest: SpyOn requires a pointer to a func value")
	}
	ft := fv.Type()
	m := NewMock("spy")
	// The default implementation delegates to the original function.
	m.impl = func(args ...any) []any { return callReflect(fv, ft, args) }
	wrapper := reflect.MakeFunc(ft, func(in []reflect.Value) []reflect.Value {
		// For a variadic function reflect passes the final parameter as a single
		// slice; spread it so recorded arguments (and CalledWith matching) see
		// the individual values.
		var args []any
		for i, v := range in {
			if ft.IsVariadic() && i == len(in)-1 {
				for j := 0; j < v.Len(); j++ {
					args = append(args, v.Index(j).Interface())
				}
				continue
			}
			args = append(args, v.Interface())
		}
		results := m.Call(args...)
		return resultsToValues(ft, results)
	})
	*target = wrapper.Interface().(T)

	s := &Spy{Mock: m, restore: func() { *target = orig }}
	spyRegistryMu.Lock()
	spyRegistry = append(spyRegistry, s)
	spyRegistryMu.Unlock()
	return s
}

// callReflect invokes the original function via reflection with the recorded
// argument list, returning its results as a []any.
func callReflect(fv reflect.Value, ft reflect.Type, args []any) []any {
	in := make([]reflect.Value, len(args))
	for i := range args {
		if args[i] == nil {
			in[i] = reflect.Zero(paramType(ft, i))
		} else {
			in[i] = reflect.ValueOf(args[i])
		}
	}
	out := fv.Call(in)
	res := make([]any, len(out))
	for i, v := range out {
		res[i] = v.Interface()
	}
	return res
}

// paramType returns the type of the i-th parameter of ft, accounting for
// variadic functions.
func paramType(ft reflect.Type, i int) reflect.Type {
	if ft.IsVariadic() && i >= ft.NumIn()-1 {
		return ft.In(ft.NumIn() - 1).Elem()
	}
	if i < ft.NumIn() {
		return ft.In(i)
	}
	return ft.In(ft.NumIn() - 1)
}

// resultsToValues converts a []any result set into reflect.Values matching the
// function's return types, substituting typed zero values for missing or
// mismatched entries so a configured mock implementation need not be exact.
func resultsToValues(ft reflect.Type, results []any) []reflect.Value {
	out := make([]reflect.Value, ft.NumOut())
	for i := 0; i < ft.NumOut(); i++ {
		rt := ft.Out(i)
		if i >= len(results) || results[i] == nil {
			out[i] = reflect.Zero(rt)
			continue
		}
		rv := reflect.ValueOf(results[i])
		if rv.Type().AssignableTo(rt) {
			out[i] = rv
		} else if rv.Type().ConvertibleTo(rt) {
			out[i] = rv.Convert(rt)
		} else {
			out[i] = reflect.Zero(rt)
		}
	}
	return out
}

// RestoreAllMocks reinstates the original function values for every spy created
// with [SpyOn] since the last call, then clears the registry.
func RestoreAllMocks() {
	spyRegistryMu.Lock()
	spies := spyRegistry
	spyRegistry = nil
	spyRegistryMu.Unlock()
	for _, s := range spies {
		s.restore()
	}
}
