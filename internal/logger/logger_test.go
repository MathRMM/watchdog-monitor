package logger_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mathrmm/watchdog-monitor/internal/logger"
)

// captureLog sets up the logger writing to a strings.Builder for inspection.
// Returns a function that reads accumulated output.
func captureLog(t *testing.T) func() string {
	t.Helper()
	var buf strings.Builder
	logger.SetupWriter(&buf)
	t.Cleanup(func() { logger.SetupWriter(nil) })
	return buf.String
}

// TestErrorDedup_SameMessageSuppressed verifies that the same error message
// logged 3 consecutive times produces only 1 log entry (supressão de duplicatas).
func TestErrorDedup_SameMessageSuppressed(t *testing.T) {
	getLog := captureLog(t)

	logger.Error("gpu collect: wmi unavailable")
	logger.Error("gpu collect: wmi unavailable")
	logger.Error("gpu collect: wmi unavailable")

	got := getLog()
	count := strings.Count(got, "gpu collect: wmi unavailable")
	if count != 1 {
		t.Errorf("expected 1 log entry for repeated error, got %d\nlog output:\n%s", count, got)
	}
}

// TestErrorDedup_DifferentMessageLogged verifies that after a message is suppressed,
// a different error message is logged normally (forward from suppressed state).
func TestErrorDedup_DifferentMessageLogged(t *testing.T) {
	getLog := captureLog(t)

	logger.Error("gpu collect: wmi unavailable")
	logger.Error("gpu collect: wmi unavailable") // suppressed
	logger.Error("gpu collect: access denied")   // different → must appear

	got := getLog()
	if !strings.Contains(got, "access denied") {
		t.Errorf("expected different error to be logged after suppression\nlog output:\n%s", got)
	}
	// The second unique message should appear exactly once.
	count := strings.Count(got, "gpu collect: access denied")
	if count != 1 {
		t.Errorf("expected 1 entry for new error, got %d", count)
	}
}

// TestInfo_LogsMessage verifica que Info() escreve o prefixo [INFO] e a mensagem formatada.
func TestInfo_LogsMessage(t *testing.T) {
	getLog := captureLog(t)

	logger.Info("version=%s started", "1.0.0")

	got := getLog()
	if !strings.Contains(got, "[INFO]") {
		t.Errorf("expected [INFO] prefix in log output, got: %s", got)
	}
	if !strings.Contains(got, "version=1.0.0 started") {
		t.Errorf("expected formatted message in log output, got: %s", got)
	}
}

// TestSetup_WritesToFile verifica que Setup() cria o arquivo de log e escreve nele.
func TestSetup_WritesToFile(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "watchdog-test.log")

	logger.Setup(logPath)
	t.Cleanup(func() { logger.SetupWriter(nil) })

	logger.Info("hello from file test")

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("log file not created at %s: %v", logPath, err)
	}
	if !strings.Contains(string(data), "hello from file test") {
		t.Errorf("expected message in log file, got: %q", string(data))
	}
}

// TestErrorDedup_InfoResetsDedup verifies that an Info call between repeated errors
// does not break deduplication (Info and Error are independent levels).
func TestErrorDedup_InfoDoesNotInterfere(t *testing.T) {
	getLog := captureLog(t)

	logger.Error("gpu collect: wmi unavailable")
	logger.Info("cycle complete")
	logger.Error("gpu collect: wmi unavailable") // same error after Info → still suppressed

	got := getLog()
	count := strings.Count(got, "gpu collect: wmi unavailable")
	if count != 1 {
		t.Errorf("expected 1 entry (Info should not reset dedup), got %d\nlog:\n%s", count, got)
	}
}
