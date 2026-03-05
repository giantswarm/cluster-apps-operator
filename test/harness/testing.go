package harness

// TestingT is the subset of testing.T used by the harness.
type TestingT interface {
	Helper()
	Fatalf(format string, args ...interface{})
	Cleanup(func())
	Logf(format string, args ...interface{})
	Name() string
}
