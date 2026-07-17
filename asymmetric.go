package jest

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"unsafe"
)

// AsymmetricMatcher is a value that matches a range of actual values rather than
// a single fixed one. Asymmetric matchers are placed on the expected side of
// [Matcher.ToEqual], [Matcher.ToMatchObject], [Matcher.ToHaveProperty] and the
// call-argument matchers, where they are consulted instead of a direct value
// comparison. The built-in matchers are produced by [Any], [Anything],
// [StringContaining], [StringMatching], [ArrayContaining] and [ObjectContaining].
type AsymmetricMatcher interface {
	// Matches reports whether actual satisfies the matcher.
	Matches(actual any) bool
	// String returns a human-readable description used in failure messages.
	String() string
}

// anyMatcher matches any value whose dynamic type is assignable to a target type.
type anyMatcher struct{ typ reflect.Type }

// Any returns an [AsymmetricMatcher] that matches any value whose dynamic type
// is assignable to the type of typ. typ may be a sample value (its type is
// used) or a [reflect.Type]. Any(nil) matches any non-nil value.
func Any(typ any) AsymmetricMatcher {
	if typ == nil {
		return anyMatcher{typ: nil}
	}
	if rt, ok := typ.(reflect.Type); ok {
		return anyMatcher{typ: rt}
	}
	return anyMatcher{typ: reflect.TypeOf(typ)}
}

// Matches reports whether actual's dynamic type is assignable to the target type.
func (a anyMatcher) Matches(actual any) bool {
	if a.typ == nil {
		return actual != nil && !isNil(actual)
	}
	at := reflect.TypeOf(actual)
	if at == nil {
		return false
	}
	if at == a.typ || at.AssignableTo(a.typ) {
		return true
	}
	return a.typ.Kind() == reflect.Interface && at.Implements(a.typ)
}

// String describes the matcher.
func (a anyMatcher) String() string {
	if a.typ == nil {
		return "Any(<non-nil>)"
	}
	return "Any(" + a.typ.String() + ")"
}

// anythingMatcher matches any non-nil value.
type anythingMatcher struct{}

// Anything returns an [AsymmetricMatcher] that matches any non-nil value.
func Anything() AsymmetricMatcher { return anythingMatcher{} }

// Matches reports whether actual is non-nil.
func (anythingMatcher) Matches(actual any) bool { return actual != nil && !isNil(actual) }

// String describes the matcher.
func (anythingMatcher) String() string { return "Anything()" }

// stringContainingMatcher matches strings containing a substring.
type stringContainingMatcher struct{ sub string }

// StringContaining returns an [AsymmetricMatcher] that matches any string
// value containing sub as a substring.
func StringContaining(sub string) AsymmetricMatcher { return stringContainingMatcher{sub} }

// Matches reports whether actual is a string containing the substring.
func (s stringContainingMatcher) Matches(actual any) bool {
	str, ok := actual.(string)
	return ok && strings.Contains(str, s.sub)
}

// String describes the matcher.
func (s stringContainingMatcher) String() string {
	return fmt.Sprintf("StringContaining(%q)", s.sub)
}

// stringMatchingMatcher matches strings against a regular expression.
type stringMatchingMatcher struct{ re *regexp.Regexp }

// StringMatching returns an [AsymmetricMatcher] that matches any string value
// matching the given regular expression. It panics if pattern is not a valid
// regular expression.
func StringMatching(pattern string) AsymmetricMatcher {
	return stringMatchingMatcher{re: regexp.MustCompile(pattern)}
}

// Matches reports whether actual is a string matching the pattern.
func (s stringMatchingMatcher) Matches(actual any) bool {
	str, ok := actual.(string)
	return ok && s.re.MatchString(str)
}

// String describes the matcher.
func (s stringMatchingMatcher) String() string {
	return fmt.Sprintf("StringMatching(/%s/)", s.re.String())
}

// arrayContainingMatcher matches slices/arrays that contain every element.
type arrayContainingMatcher struct{ elems []any }

// ArrayContaining returns an [AsymmetricMatcher] that matches any slice or
// array containing every one of elems (each compared with asymmetric-aware
// equality, so elems may themselves be asymmetric matchers).
func ArrayContaining(elems ...any) AsymmetricMatcher { return arrayContainingMatcher{elems} }

// Matches reports whether actual is a slice/array containing each expected element.
func (a arrayContainingMatcher) Matches(actual any) bool {
	rv := reflect.ValueOf(actual)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return false
	}
	for _, want := range a.elems {
		found := false
		for i := 0; i < rv.Len(); i++ {
			if asymEqual(want, rv.Index(i).Interface()) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// String describes the matcher.
func (a arrayContainingMatcher) String() string {
	return fmt.Sprintf("ArrayContaining(%v)", a.elems)
}

// objectContainingMatcher matches maps/structs that contain every key/value.
type objectContainingMatcher struct{ fields map[string]any }

// ObjectContaining returns an [AsymmetricMatcher] that matches any map with
// string keys, or any struct, that contains every key in fields with a value
// satisfying the corresponding expected value (compared with asymmetric-aware
// equality). Extra keys or fields on the actual value are ignored.
func ObjectContaining(fields map[string]any) AsymmetricMatcher {
	return objectContainingMatcher{fields: fields}
}

// Matches reports whether actual contains each expected key/value.
func (o objectContainingMatcher) Matches(actual any) bool {
	for key, want := range o.fields {
		got, ok := propertyByName(actual, key)
		if !ok || !asymEqual(want, got) {
			return false
		}
	}
	return true
}

// String describes the matcher.
func (o objectContainingMatcher) String() string {
	return fmt.Sprintf("ObjectContaining(%v)", o.fields)
}

// propertyByName reads a single named property from a map (string keys) or a
// struct (exported field), reporting whether it was present.
func propertyByName(container any, name string) (any, bool) {
	rv := reflect.ValueOf(container)
	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return nil, false
		}
		rv = rv.Elem()
	}
	switch rv.Kind() {
	case reflect.Map:
		if rv.Type().Key().Kind() != reflect.String {
			return nil, false
		}
		v := rv.MapIndex(reflect.ValueOf(name).Convert(rv.Type().Key()))
		if !v.IsValid() {
			return nil, false
		}
		return v.Interface(), true
	case reflect.Struct:
		f := rv.FieldByName(name)
		if !f.IsValid() {
			return nil, false
		}
		if !f.CanInterface() {
			f = forceInterface(rv, name)
			if !f.IsValid() {
				return nil, false
			}
		}
		return f.Interface(), true
	default:
		return nil, false
	}
}

// forceInterface returns an interfaceable copy of an unexported struct field.
func forceInterface(structVal reflect.Value, name string) reflect.Value {
	c := addrCopy(structVal)
	f := c.FieldByName(name)
	if !f.IsValid() || !f.CanAddr() {
		return reflect.Value{}
	}
	return reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
}

// asymEqual reports whether actual matches expected, honoring any
// [AsymmetricMatcher] embedded in expected. When no asymmetric matchers are
// present it agrees with reflect.DeepEqual.
func asymEqual(expected, actual any) bool {
	if am, ok := expected.(AsymmetricMatcher); ok {
		return am.Matches(actual)
	}
	if reflect.DeepEqual(expected, actual) {
		return true
	}
	return asymEqualVal(reflect.ValueOf(expected), reflect.ValueOf(actual))
}

// asymEqualVal is the recursive core of [asymEqual]. It is always called with
// interfaceable values (unexported struct fields are converted before descent).
func asymEqualVal(ev, av reflect.Value) bool {
	if !ev.IsValid() || !av.IsValid() {
		return ev.IsValid() == av.IsValid()
	}
	if ev.CanInterface() {
		if am, ok := ev.Interface().(AsymmetricMatcher); ok {
			var actual any
			if av.CanInterface() {
				actual = av.Interface()
			}
			return am.Matches(actual)
		}
	}
	if ev.Type() != av.Type() {
		return false
	}
	switch ev.Kind() {
	case reflect.Ptr:
		if ev.IsNil() || av.IsNil() {
			return ev.IsNil() == av.IsNil()
		}
		return asymEqualVal(ev.Elem(), av.Elem())
	case reflect.Interface:
		return asymEqualVal(ev.Elem(), av.Elem())
	case reflect.Struct:
		ev = addrCopy(ev)
		av = addrCopy(av)
		for i := 0; i < ev.NumField(); i++ {
			ef, af := ev.Field(i), av.Field(i)
			if !ef.CanInterface() {
				ef = reflect.NewAt(ef.Type(), unsafe.Pointer(ef.UnsafeAddr())).Elem()
				af = reflect.NewAt(af.Type(), unsafe.Pointer(af.UnsafeAddr())).Elem()
			}
			if !asymEqualVal(ef, af) {
				return false
			}
		}
		return true
	case reflect.Map:
		if ev.IsNil() || av.IsNil() {
			return ev.IsNil() == av.IsNil()
		}
		if ev.Len() != av.Len() {
			return false
		}
		for _, k := range ev.MapKeys() {
			avv := av.MapIndex(k)
			if !avv.IsValid() {
				return false
			}
			if !asymEqualVal(ev.MapIndex(k), avv) {
				return false
			}
		}
		return true
	case reflect.Slice:
		if ev.IsNil() || av.IsNil() {
			return ev.IsNil() == av.IsNil()
		}
		fallthrough
	case reflect.Array:
		if ev.Len() != av.Len() {
			return false
		}
		for i := 0; i < ev.Len(); i++ {
			if !asymEqualVal(ev.Index(i), av.Index(i)) {
				return false
			}
		}
		return true
	default:
		if !ev.CanInterface() || !av.CanInterface() {
			return false
		}
		return reflect.DeepEqual(ev.Interface(), av.Interface())
	}
}

// addrCopy returns an addressable copy of v (or v itself if already
// addressable) so that unexported fields can be read via unsafe.
func addrCopy(v reflect.Value) reflect.Value {
	if v.CanAddr() {
		return v
	}
	c := reflect.New(v.Type()).Elem()
	c.Set(v)
	return c
}

// hasAsymmetric reports whether expected contains any [AsymmetricMatcher],
// used to suppress the structural diff (which cannot meaningfully render a
// matcher) on a failed asymmetric comparison.
func hasAsymmetric(v any) bool {
	if v == nil {
		return false
	}
	if _, ok := v.(AsymmetricMatcher); ok {
		return true
	}
	return hasAsymmetricVal(reflect.ValueOf(v), 0)
}

func hasAsymmetricVal(rv reflect.Value, depth int) bool {
	if !rv.IsValid() || depth > 8 {
		return false
	}
	if rv.CanInterface() {
		if _, ok := rv.Interface().(AsymmetricMatcher); ok {
			return true
		}
	}
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface:
		return !rv.IsNil() && hasAsymmetricVal(rv.Elem(), depth+1)
	case reflect.Slice, reflect.Array:
		for i := 0; i < rv.Len(); i++ {
			if hasAsymmetricVal(rv.Index(i), depth+1) {
				return true
			}
		}
	case reflect.Map:
		for _, k := range rv.MapKeys() {
			if hasAsymmetricVal(rv.MapIndex(k), depth+1) {
				return true
			}
		}
	case reflect.Struct:
		for i := 0; i < rv.NumField(); i++ {
			if rv.Field(i).CanInterface() && hasAsymmetricVal(rv.Field(i), depth+1) {
				return true
			}
		}
	}
	return false
}
