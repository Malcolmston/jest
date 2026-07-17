package jest

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
)

// snapshotDir is the directory (relative to the test's working directory, i.e.
// the package directory under `go test`) in which snapshot files are stored. It
// is a variable so tests can redirect it to a temporary directory.
var snapshotDir = "__snapshots__"

// updateSnapshots forces snapshots to be (re)written rather than compared. It is
// set by [SetUpdateSnapshots] and defaults to true when the JEST_UPDATE_SNAPSHOTS
// environment variable is a non-empty value other than "0" or "false".
var updateSnapshots = envUpdate()

// envUpdate reports whether the update-snapshots environment variable is set.
func envUpdate() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("JEST_UPDATE_SNAPSHOTS")))
	return v != "" && v != "0" && v != "false"
}

// SetUpdateSnapshots enables or disables snapshot update mode. In update mode
// [Matcher.ToMatchSnapshot] (re)writes the stored snapshot instead of comparing
// against it. Update mode is also enabled by setting the JEST_UPDATE_SNAPSHOTS
// environment variable.
func SetUpdateSnapshots(on bool) { updateSnapshots = on }

// snapshotUpdate reports whether snapshots should be written rather than compared.
func snapshotUpdate() bool { return updateSnapshots }

// ToMatchSnapshot compares the actual value against a snapshot stored on disk.
// On the first run (or in update mode) the serialized value is written to the
// snapshot store and the assertion passes; on subsequent runs the value is
// compared against the stored snapshot and a failure is reported on any
// difference. An optional explicit name overrides the snapshot key (which
// otherwise derives from the test name plus a per-test counter), making the
// snapshot stable regardless of test-execution order.
func (m *Matcher[T]) ToMatchSnapshot(name ...string) {
	m.t.Helper()
	countAssertion(m.t)
	base := snapshotName(m.t, name...)
	store := storeFor(snapshotDir)
	key := store.key(base, len(name) > 0)
	got := snapshotSerialize(m.actual)

	existing, ok := store.get(key)
	if snapshotUpdate() || !ok {
		if err := store.set(key, got); err != nil {
			m.t.Errorf("assertion failed: could not write snapshot %q: %v", key, err)
		}
		return
	}
	if existing != got {
		m.t.Errorf("assertion failed: snapshot %q did not match\n--- stored ---\n%s\n--- actual ---\n%s",
			key, existing, got)
	}
}

// snapshotName resolves the snapshot key base from an explicit name or the
// reporter's test name, falling back to "snapshot".
func snapshotName(t TestReporter, explicit ...string) string {
	if len(explicit) > 0 && explicit[0] != "" {
		return explicit[0]
	}
	if n, ok := t.(interface{ Name() string }); ok {
		if name := n.Name(); name != "" {
			return name
		}
	}
	return "snapshot"
}

// ---- store ------------------------------------------------------------------

type snapStore struct {
	mu       sync.Mutex
	dir      string
	loaded   bool
	entries  map[string]string
	counters map[string]int
}

var (
	storesMu sync.Mutex
	stores   = map[string]*snapStore{}
)

// storeFor returns the snapshot store for a directory, creating it on first use.
func storeFor(dir string) *snapStore {
	storesMu.Lock()
	defer storesMu.Unlock()
	s := stores[dir]
	if s == nil {
		s = &snapStore{dir: dir, entries: map[string]string{}, counters: map[string]int{}}
		stores[dir] = s
	}
	return s
}

// key derives a stable snapshot key. When the caller supplied an explicit name
// the base is used verbatim; otherwise a per-base counter disambiguates
// multiple snapshots taken within the same test.
func (s *snapStore) key(base string, explicit bool) string {
	if explicit {
		return base
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counters[base]++
	return fmt.Sprintf("%s %d", base, s.counters[base])
}

// filePath is the on-disk location of the snapshot file.
func (s *snapStore) filePath() string { return filepath.Join(s.dir, "snapshots.snap") }

// load reads the snapshot file into memory once.
func (s *snapStore) load() {
	if s.loaded {
		return
	}
	s.loaded = true
	data, err := os.ReadFile(s.filePath())
	if err != nil {
		return
	}
	parseSnapshots(string(data), s.entries)
}

// get returns a stored snapshot value.
func (s *snapStore) get(key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.load()
	v, ok := s.entries[key]
	return v, ok
}

// set stores a snapshot value and persists the whole store to disk.
func (s *snapStore) set(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.load()
	s.entries[key] = value
	return s.persist()
}

// persist writes every entry to the snapshot file, sorted by key.
func (s *snapStore) persist() error {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}
	keys := make([]string, 0, len(s.entries))
	for k := range s.entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString("// jest snapshot file — do not edit by hand.\n\n")
	for _, k := range keys {
		fmt.Fprintf(&b, "=== %s ===\n%s\n=== end ===\n\n", k, s.entries[k])
	}
	return os.WriteFile(s.filePath(), []byte(b.String()), 0o644)
}

// parseSnapshots reads the "=== key ===\nvalue\n=== end ===" format into m.
func parseSnapshots(data string, m map[string]string) {
	lines := strings.Split(data, "\n")
	i := 0
	for i < len(lines) {
		line := lines[i]
		if strings.HasPrefix(line, "=== ") && strings.HasSuffix(line, " ===") && line != "=== end ===" {
			key := strings.TrimSuffix(strings.TrimPrefix(line, "=== "), " ===")
			i++
			var body []string
			for i < len(lines) && lines[i] != "=== end ===" {
				body = append(body, lines[i])
				i++
			}
			m[key] = strings.Join(body, "\n")
		}
		i++
	}
}

// ---- serialization ----------------------------------------------------------

// snapshotSerialize renders a value to a stable, human-readable string. Map keys
// are sorted so the output is deterministic across runs.
func snapshotSerialize(v any) string {
	var b strings.Builder
	serializeValue(&b, reflect.ValueOf(v), 0)
	return b.String()
}

func serializeValue(b *strings.Builder, rv reflect.Value, depth int) {
	if !rv.IsValid() {
		b.WriteString("nil")
		return
	}
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface:
		if rv.IsNil() {
			b.WriteString("nil")
			return
		}
		serializeValue(b, rv.Elem(), depth)
	case reflect.String:
		fmt.Fprintf(b, "%q", rv.String())
	case reflect.Slice, reflect.Array:
		if rv.Kind() == reflect.Slice && rv.IsNil() {
			b.WriteString("nil")
			return
		}
		b.WriteString("[\n")
		for i := 0; i < rv.Len(); i++ {
			indent(b, depth+1)
			serializeValue(b, rv.Index(i), depth+1)
			b.WriteString(",\n")
		}
		indent(b, depth)
		b.WriteString("]")
	case reflect.Map:
		if rv.IsNil() {
			b.WriteString("nil")
			return
		}
		type kv struct {
			ks string
			k  reflect.Value
		}
		pairs := make([]kv, 0, rv.Len())
		for _, k := range rv.MapKeys() {
			pairs = append(pairs, kv{fmt.Sprintf("%v", k.Interface()), k})
		}
		sort.Slice(pairs, func(i, j int) bool { return pairs[i].ks < pairs[j].ks })
		b.WriteString("{\n")
		for _, p := range pairs {
			indent(b, depth+1)
			fmt.Fprintf(b, "%q: ", p.ks)
			serializeValue(b, rv.MapIndex(p.k), depth+1)
			b.WriteString(",\n")
		}
		indent(b, depth)
		b.WriteString("}")
	case reflect.Struct:
		t := rv.Type()
		fmt.Fprintf(b, "%s {\n", t.Name())
		for i := 0; i < rv.NumField(); i++ {
			if t.Field(i).PkgPath != "" {
				continue
			}
			indent(b, depth+1)
			fmt.Fprintf(b, "%s: ", t.Field(i).Name)
			serializeValue(b, rv.Field(i), depth+1)
			b.WriteString(",\n")
		}
		indent(b, depth)
		b.WriteString("}")
	default:
		fmt.Fprintf(b, "%v", rv.Interface())
	}
}

func indent(b *strings.Builder, depth int) {
	for i := 0; i < depth; i++ {
		b.WriteString("  ")
	}
}
