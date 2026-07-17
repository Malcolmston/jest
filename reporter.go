package jest

// TestReporter is the minimal subset of *testing.T that the jest matchers need
// in order to report failures. It is satisfied by *testing.T, so in production
// tests you simply pass your test's t. In jest's own tests a fake reporter that
// records failures is passed instead, which makes it possible to exercise the
// failure branch of every matcher without failing the surrounding test.
type TestReporter interface {
	// Errorf reports a non-fatal failure, mirroring (*testing.T).Errorf.
	Errorf(format string, args ...any)
	// Fatalf reports a fatal failure, mirroring (*testing.T).Fatalf.
	Fatalf(format string, args ...any)
	// Helper marks the calling function as a test helper, mirroring
	// (*testing.T).Helper.
	Helper()
}
