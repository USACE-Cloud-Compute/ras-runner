package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDataDir returns the root of the test fixtures. It honors the
// RAS_RUNNER_TEST_DATA env var and defaults to ./testData (repo-relative).
func TestDataDir() string {
	if v := os.Getenv("RAS_RUNNER_TEST_DATA"); v != "" {
		return v
	}
	return "./testData"
}

// FixturePath returns an absolute path to a fixture under the test data root.
func FixturePath(t *testing.T, rel ...string) string {
	t.Helper()
	return filepath.Join(TestDataDir(), filepath.Join(rel...))
}

// SkipIfNoTestData skips the test if the fixture root does not exist.
func SkipIfNoTestData(t *testing.T) {
	t.Helper()
	dir := TestDataDir()
	if _, err := os.Stat(dir); err != nil {
		t.Skipf("test data not found at %s — set RAS_RUNNER_TEST_DATA to run this test", dir)
	}
}

// SkipIfNoFile skips the test if the named fixture file does not exist under
// the test data root.
func SkipIfNoFile(t *testing.T, rel ...string) {
	t.Helper()
	p := FixturePath(t, rel...)
	if _, err := os.Stat(p); err != nil {
		t.Skipf("fixture not found at %s", p)
	}
}

// SkipIfNoEnv skips the test if any named env var is empty.
func SkipIfNoEnv(t *testing.T, vars ...string) {
	t.Helper()
	for _, v := range vars {
		if os.Getenv(v) == "" {
			t.Skipf("env var %s not set — skipping", v)
		}
	}
}
