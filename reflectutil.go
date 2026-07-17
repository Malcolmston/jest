package jest

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// shallowEqual reports whether a and b are equal using == semantics. When the
// underlying type is not comparable it falls back to reflect.DeepEqual.
func shallowEqual(a, b any) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)
	if va.Type() != vb.Type() {
		return false
	}
	if !va.Comparable() || !vb.Comparable() {
		return reflect.DeepEqual(a, b)
	}
	return va.Interface() == vb.Interface()
}

// isNil reports whether v is nil, handling untyped nil interfaces as well as
// typed nil pointers, slices, maps, channels, funcs and interfaces.
func isNil(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func, reflect.Interface:
		return rv.IsNil()
	default:
		return false
	}
}

// toFloat converts any numeric value to a float64, reporting whether the value
// was numeric.
func toFloat(v any) (float64, bool) {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(rv.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return float64(rv.Uint()), true
	case reflect.Float32, reflect.Float64:
		return rv.Float(), true
	default:
		return 0, false
	}
}

// contains reports whether container holds item. It returns the containment
// relationship checked ("substring", "element" or "key") so failure messages
// can describe what was searched.
func contains(container, item any) (bool, string) {
	rv := reflect.ValueOf(container)
	switch rv.Kind() {
	case reflect.String:
		sub := fmt.Sprintf("%v", item)
		return strings.Contains(rv.String(), sub), "substring"
	case reflect.Slice, reflect.Array:
		for i := 0; i < rv.Len(); i++ {
			if reflect.DeepEqual(rv.Index(i).Interface(), item) {
				return true, "element"
			}
		}
		return false, "element"
	case reflect.Map:
		iv := reflect.ValueOf(item)
		if iv.IsValid() && iv.Type().AssignableTo(rv.Type().Key()) {
			return rv.MapIndex(iv).IsValid(), "key"
		}
		return false, "key"
	default:
		return false, "element"
	}
}

// lengthOf returns the length of v when it is a string, slice, array, map or
// channel.
func lengthOf(v any) (int, bool) {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.String, reflect.Slice, reflect.Array, reflect.Map, reflect.Chan:
		return rv.Len(), true
	default:
		return 0, false
	}
}

// format renders a value for inclusion in failure messages, quoting strings and
// using Go-syntax representation for composite values.
func format(v any) string {
	if v == nil {
		return "<nil>"
	}
	switch x := v.(type) {
	case string:
		return fmt.Sprintf("%q", x)
	case error:
		return fmt.Sprintf("%q", x.Error())
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map, reflect.Struct, reflect.Ptr:
		return fmt.Sprintf("%+v", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// diff produces a small human-readable structural diff between expected and
// actual, used by ToEqual failures. It understands maps, slices/arrays and
// structs, and falls back to a two-line expected/actual dump otherwise.
func diff(expected, actual any) string {
	ev := reflect.ValueOf(expected)
	av := reflect.ValueOf(actual)
	if ev.IsValid() && av.IsValid() && ev.Type() == av.Type() {
		switch ev.Kind() {
		case reflect.Map:
			return diffMap(ev, av)
		case reflect.Slice, reflect.Array:
			return diffSeq(ev, av)
		case reflect.Struct:
			return diffStruct(ev, av)
		}
	}
	return fmt.Sprintf("  diff:\n    expected: %s\n    actual:   %s", format(expected), format(actual))
}

func diffMap(ev, av reflect.Value) string {
	keys := map[string]reflect.Value{}
	order := []string{}
	add := func(v reflect.Value) {
		for _, k := range v.MapKeys() {
			ks := fmt.Sprintf("%v", k.Interface())
			if _, ok := keys[ks]; !ok {
				keys[ks] = k
				order = append(order, ks)
			}
		}
	}
	add(ev)
	add(av)
	sort.Strings(order)
	var b strings.Builder
	b.WriteString("  diff:\n")
	for _, ks := range order {
		k := keys[ks]
		e := ev.MapIndex(k)
		a := av.MapIndex(k)
		switch {
		case !a.IsValid():
			fmt.Fprintf(&b, "    - [%s]: %s (missing)\n", ks, format(e.Interface()))
		case !e.IsValid():
			fmt.Fprintf(&b, "    + [%s]: %s (unexpected)\n", ks, format(a.Interface()))
		case !reflect.DeepEqual(e.Interface(), a.Interface()):
			fmt.Fprintf(&b, "    ~ [%s]: expected %s, actual %s\n", ks, format(e.Interface()), format(a.Interface()))
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

func diffSeq(ev, av reflect.Value) string {
	var b strings.Builder
	b.WriteString("  diff:\n")
	n := ev.Len()
	if av.Len() > n {
		n = av.Len()
	}
	for i := 0; i < n; i++ {
		switch {
		case i >= av.Len():
			fmt.Fprintf(&b, "    - [%d]: %s (missing)\n", i, format(ev.Index(i).Interface()))
		case i >= ev.Len():
			fmt.Fprintf(&b, "    + [%d]: %s (unexpected)\n", i, format(av.Index(i).Interface()))
		case !reflect.DeepEqual(ev.Index(i).Interface(), av.Index(i).Interface()):
			fmt.Fprintf(&b, "    ~ [%d]: expected %s, actual %s\n", i,
				format(ev.Index(i).Interface()), format(av.Index(i).Interface()))
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

func diffStruct(ev, av reflect.Value) string {
	var b strings.Builder
	b.WriteString("  diff:\n")
	t := ev.Type()
	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).PkgPath != "" {
			continue // unexported
		}
		e := ev.Field(i).Interface()
		a := av.Field(i).Interface()
		if !reflect.DeepEqual(e, a) {
			fmt.Fprintf(&b, "    ~ %s: expected %s, actual %s\n", t.Field(i).Name, format(e), format(a))
		}
	}
	return strings.TrimRight(b.String(), "\n")
}
