package jest

import "testing"

func TestToMatchInlineSnapshot(t *testing.T) {
	expectPass(t, "string", func(r TestReporter) { Expect(r, "hi").ToMatchInlineSnapshot(`"hi"`) })
	expectPass(t, "int", func(r TestReporter) { Expect(r, 42).ToMatchInlineSnapshot(`42`) })
	expectPass(t, "slice multiline", func(r TestReporter) {
		Expect(r, []int{1, 2}).ToMatchInlineSnapshot(`
[
  1,
  2,
]
`)
	})
	expectFail(t, "mismatch", func(r TestReporter) { Expect(r, "hi").ToMatchInlineSnapshot(`"bye"`) })
	expectPass(t, "negated mismatch", func(r TestReporter) {
		Expect(r, "hi").Not().ToMatchInlineSnapshot(`"bye"`)
	})
}

func TestToThrowMatchingSnapshot(t *testing.T) {
	withSnapshotDir(t)
	boom := func() { panic("kaboom") }
	// First run writes and passes.
	expectPass(t, "write", func(r TestReporter) { Expect[any](r, boom).ToThrowMatchingSnapshot("err") })
	// Same message matches.
	expectPass(t, "match", func(r TestReporter) { Expect[any](r, boom).ToThrowMatchingSnapshot("err") })
	// Different message fails.
	other := func() { panic("different") }
	expectFail(t, "mismatch", func(r TestReporter) { Expect[any](r, other).ToThrowMatchingSnapshot("err") })
	// A function that does not panic fails.
	noop := func() {}
	expectFail(t, "no panic", func(r TestReporter) { Expect[any](r, noop).ToThrowMatchingSnapshot("noerr") })
	// A non-func value fails.
	expectFail(t, "not func", func(r TestReporter) { Expect(r, 5).ToThrowMatchingSnapshot("x") })
}
