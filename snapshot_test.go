package jest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// withSnapshotDir redirects the snapshot store to a temporary directory for the
// duration of a test and resets the in-memory store cache.
func withSnapshotDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	old := snapshotDir
	snapshotDir = dir
	t.Cleanup(func() { snapshotDir = old })
	return dir
}

func TestSnapshotWriteThenMatch(t *testing.T) {
	withSnapshotDir(t)
	val := map[string]any{"a": 1, "b": []any{"x", "y"}}

	// First run writes the snapshot and passes.
	expectPass(t, "first write", func(r TestReporter) { Expect[any](r, val).ToMatchSnapshot("case") })
	// Matching value passes.
	expectPass(t, "match", func(r TestReporter) { Expect[any](r, val).ToMatchSnapshot("case") })
	// Different value fails.
	expectFail(t, "mismatch", func(r TestReporter) {
		Expect[any](r, map[string]any{"a": 2}).ToMatchSnapshot("case")
	})
}

func TestSnapshotPersistedToDisk(t *testing.T) {
	dir := withSnapshotDir(t)
	expectPass(t, "write", func(r TestReporter) { Expect(r, "hello").ToMatchSnapshot("greeting") })

	data, err := os.ReadFile(filepath.Join(dir, "snapshots.snap"))
	if err != nil {
		t.Fatalf("snapshot file not written: %v", err)
	}
	if !strings.Contains(string(data), "=== greeting ===") {
		t.Errorf("snapshot file missing key: %s", data)
	}
	if !strings.Contains(string(data), `"hello"`) {
		t.Errorf("snapshot file missing value: %s", data)
	}
}

func TestSnapshotUpdateMode(t *testing.T) {
	withSnapshotDir(t)
	expectPass(t, "initial", func(r TestReporter) { Expect(r, "v1").ToMatchSnapshot("k") })
	// Without update mode, a changed value fails.
	expectFail(t, "changed fails", func(r TestReporter) { Expect(r, "v2").ToMatchSnapshot("k") })

	SetUpdateSnapshots(true)
	defer SetUpdateSnapshots(false)
	// In update mode, the new value is written and passes.
	expectPass(t, "update writes", func(r TestReporter) { Expect(r, "v2").ToMatchSnapshot("k") })
	SetUpdateSnapshots(false)
	expectPass(t, "now matches v2", func(r TestReporter) { Expect(r, "v2").ToMatchSnapshot("k") })
}

func TestSnapshotAutoName(t *testing.T) {
	withSnapshotDir(t)
	// A reporter exposing Name() drives an auto-generated, counter-based key.
	f := &fakeT{name: "TestSnapshotAutoName"}
	Expect(f, 1).ToMatchSnapshot()
	Expect(f, 2).ToMatchSnapshot()
	if f.failed() {
		t.Errorf("auto-name snapshot should pass on first write: %s", f.message())
	}
}

func TestSnapshotSerializeStable(t *testing.T) {
	a := snapshotSerialize(map[string]int{"b": 2, "a": 1, "c": 3})
	b := snapshotSerialize(map[string]int{"c": 3, "a": 1, "b": 2})
	if a != b {
		t.Errorf("map serialization not stable:\n%s\n---\n%s", a, b)
	}

	type point struct {
		X, Y int
	}
	s := snapshotSerialize(point{1, 2})
	if !strings.Contains(s, "point {") || !strings.Contains(s, "X: 1") {
		t.Errorf("struct serialization unexpected: %s", s)
	}
	if got := snapshotSerialize(nil); got != "nil" {
		t.Errorf("nil serialize = %q", got)
	}
}

func TestParseSnapshotsRoundTrip(t *testing.T) {
	m := map[string]string{}
	parseSnapshots("=== a ===\nline1\nline2\n=== end ===\n\n=== b ===\nx\n=== end ===\n", m)
	if m["a"] != "line1\nline2" {
		t.Errorf("parse a = %q", m["a"])
	}
	if m["b"] != "x" {
		t.Errorf("parse b = %q", m["b"])
	}
}
