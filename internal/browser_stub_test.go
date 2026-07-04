package internal

// No test may launch a real browser: openBrowser defaults to a no-op for the
// whole test binary. Tests that assert browser behavior stub it per-test (see
// stubBrowser), restoring to this no-op.
func init() {
	openBrowser = func(string) error { return nil }
}
