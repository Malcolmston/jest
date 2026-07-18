package jest

import (
	"fmt"
	"reflect"
	"strings"
)

// ToMatchInlineSnapshot compares the serialized form of the value against an
// inline expected snapshot supplied at the call site, mirroring Jest's
// toMatchInlineSnapshot. Both the stored and the actual serializations are
// compared after trimming surrounding whitespace, so the expected literal may be
// written across multiple indented lines. The value is serialized with the same
// deterministic, sorted-key serializer used by [Matcher.ToMatchSnapshot].
func (m *Matcher[T]) ToMatchInlineSnapshot(expected string) {
	m.t.Helper()
	countAssertion(m.t)
	got := snapshotSerialize(m.actual)
	if strings.TrimSpace(got) == strings.TrimSpace(expected) {
		return
	}
	if m.negated {
		return
	}
	m.t.Errorf("assertion failed: inline snapshot did not match\n--- expected ---\n%s\n--- actual ---\n%s",
		strings.TrimSpace(expected), strings.TrimSpace(got))
}

// ToThrowMatchingSnapshot calls the actual value (which must be a func with no
// arguments), captures the string form of the panic it raises, and compares it
// against an on-disk snapshot, mirroring Jest's toThrowErrorMatchingSnapshot. On
// the first run (or in snapshot update mode) the message is written to the
// snapshot store; on later runs it is compared. If the function does not panic,
// the assertion fails. An optional explicit name overrides the snapshot key.
func (m *Matcher[T]) ToThrowMatchingSnapshot(name ...string) {
	m.t.Helper()
	countAssertion(m.t)
	fn := reflect.ValueOf(m.actual)
	if !fn.IsValid() || fn.Kind() != reflect.Func || fn.Type().NumIn() != 0 {
		m.t.Errorf("assertion failed: ToThrowMatchingSnapshot requires a func() value, but got %s",
			format(m.actual))
		return
	}
	panicked, recovered := callAndRecover(fn)
	if !panicked {
		m.t.Errorf("assertion failed: expected value to panic, but it did not")
		return
	}
	got := fmt.Sprintf("%v", recovered)

	base := snapshotName(m.t, name...)
	store := storeFor(snapshotDir)
	key := store.key(base, len(name) > 0)
	existing, ok := store.get(key)
	if snapshotUpdate() || !ok {
		if err := store.set(key, got); err != nil {
			m.t.Errorf("assertion failed: could not write snapshot %q: %v", key, err)
		}
		return
	}
	if existing != got {
		m.t.Errorf("assertion failed: thrown-error snapshot %q did not match\n--- stored ---\n%s\n--- actual ---\n%s",
			key, existing, got)
	}
}
