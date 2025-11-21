package logger_test

import (
	"io"
	"os"
	"testing"

	logpkg "github.com/garnizeh/luckyfive/internal/logger"
)

func TestNewLoggerReturnsLogger(t *testing.T) {
	l := logpkg.New("DEBUG")
	if l == nil {
		t.Fatalf("expected New to return a logger, got nil")
	}
}

func TestLoggerMethodsDoNotPanic(t *testing.T) {
	l := logpkg.New("INFO")
	// Ensure basic logging methods don't panic
	// We don't assert on output here, only that calls succeed
	l.Info("test-info", "key", "value")
	l.Warn("test-warn")
	l.Error("test-error")
}

// captureStdout captures stdout while fn runs and returns the captured output.
func captureStdout(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	_ = w.Close()
	os.Stdout = old
	out, _ := io.ReadAll(r)
	return string(out)
}

func TestLoggerLevelWarnSuppressesInfo(t *testing.T) {
	out := captureStdout(func() {
		l := logpkg.New("WARN")
		l.Info("info-msg")
		l.Warn("warn-msg")
		l.Error("error-msg")
	})

	// continue with assertions

	if contains := (func() bool { return containsSubstring(out, "info-msg") })(); contains {
		t.Fatalf("expected INFO messages to be suppressed at WARN level, but found info-msg in output: %s", out)
	}
	if !containsSubstring(out, "warn-msg") {
		t.Fatalf("expected WARN messages to appear at WARN level, output: %s", out)
	}
	if !containsSubstring(out, "error-msg") {
		t.Fatalf("expected ERROR messages to appear at WARN level, output: %s", out)
	}
}

func TestLoggerLevelErrorOnlyShowsError(t *testing.T) {
	out := captureStdout(func() {
		l := logpkg.New("ERROR")
		l.Warn("warn-msg")
		l.Error("error-msg")
	})

	if containsSubstring(out, "warn-msg") {
		t.Fatalf("expected WARN messages to be suppressed at ERROR level, but found warn-msg in output: %s", out)
	}
	if !containsSubstring(out, "error-msg") {
		t.Fatalf("expected ERROR messages to appear at ERROR level, output: %s", out)
	}
}

func TestLoggerDefaultLevelIsInfo(t *testing.T) {
	out := captureStdout(func() {
		l := logpkg.New("SOMETHING_UNKOWN")
		l.Debug("debug-msg")
		l.Info("info-msg")
	})

	if containsSubstring(out, "debug-msg") {
		t.Fatalf("expected DEBUG messages to be suppressed at default INFO level, but found debug-msg in output: %s", out)
	}
	if !containsSubstring(out, "info-msg") {
		t.Fatalf("expected INFO messages to appear at default level, output: %s", out)
	}
}

// containsSubstring is a small helper to check substring presence
func containsSubstring(s, sub string) bool {
	return len(s) >= len(sub) && (func() bool { return (len(sub) == 0) || (stringIndex(s, sub) >= 0) })()
}

// stringIndex returns the index of substr in s or -1 if not found.
// We implement a tiny wrapper to avoid importing strings in multiple tests.
func stringIndex(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
