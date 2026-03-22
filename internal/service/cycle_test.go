package service_test

import (
	"errors"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"

	metrics "github.com/mathrmm/watchdog-monitor/gen/metrics"
	"github.com/mathrmm/watchdog-monitor/internal/service"
)

// mockPublisher captures Publish calls for assertions.
type mockPublisher struct {
	err     error
	subject string
	data    []byte
	called  bool
}

func (m *mockPublisher) Publish(subject string, data []byte) error {
	m.called = true
	m.subject = subject
	m.data = data
	return m.err
}

// --- Hostname sanitization tests (RN04) ---

func TestSanitizeHostname_DotToHyphen(t *testing.T) {
	if got := service.SanitizeHostname("MY-PC.local"); got != "MY-PC-local" {
		t.Errorf("expected MY-PC-local, got %s", got)
	}
}

func TestSanitizeHostname_SpaceToHyphen(t *testing.T) {
	if got := service.SanitizeHostname("My PC"); got != "My-PC" {
		t.Errorf("expected My-PC, got %s", got)
	}
}

func TestSanitizeHostname_CollapseConsecutiveHyphens(t *testing.T) {
	if got := service.SanitizeHostname("pc--lab"); got != "pc-lab" {
		t.Errorf("expected pc-lab, got %s", got)
	}
}

// --- RunCycle tests ---

// TestRunCycle_AllCollectorsSuccess verifies that a full cycle publishes a valid payload.
func TestRunCycle_AllCollectorsSuccess(t *testing.T) {
	pub := &mockPublisher{}
	runner := service.NewCycleRunner("testhost", pub)
	runner.RunCycle()

	if !pub.called {
		t.Fatal("expected Publish to be called")
	}
	if pub.subject != "watchdog.testhost.metrics" {
		t.Errorf("unexpected subject: %s", pub.subject)
	}

	var payload metrics.SystemMetrics
	if err := proto.Unmarshal(pub.data, &payload); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if payload.Hostname != "testhost" {
		t.Errorf("expected hostname testhost, got %s", payload.Hostname)
	}
	if payload.TimestampMs == 0 {
		t.Error("expected non-zero TimestampMs")
	}
	if payload.Cpu == nil {
		t.Error("expected non-nil Cpu field")
	}
	if payload.Memory == nil {
		t.Error("expected non-nil Memory field")
	}
	// GPU is nil on non-Windows — no assertion here.
}

// TestRunCycle_GPUError_PublishesWithNilGPU verifies that when GPU fails,
// the payload is still published with gpu == nil (RN02, RNF09).
func TestRunCycle_GPUError_PublishesWithNilGPU(t *testing.T) {
	pub := &mockPublisher{}
	runner := service.NewCycleRunner("testhost", pub)
	runner.RunCycle()

	if !pub.called {
		t.Fatal("expected Publish to be called even when GPU fails")
	}

	var payload metrics.SystemMetrics
	if err := proto.Unmarshal(pub.data, &payload); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	// On Linux, GPU collector always returns error → gpu must be nil.
	if payload.Gpu != nil {
		t.Error("expected nil Gpu when GPU collector fails (RN02)")
	}
}

// TestRunCycle_PublishError_NoPanic verifies that a Publish failure is silently
// discarded — the cycle continues without panic (RN03).
func TestRunCycle_PublishError_NoPanic(t *testing.T) {
	pub := &mockPublisher{err: errors.New("NATS unavailable")}
	runner := service.NewCycleRunner("testhost", pub)
	runner.RunCycle() // must not panic
}

// TestRunCycle_TimestampUTC verifies that timestamp_ms reflects the collection
// time in Unix milliseconds UTC (RF12).
func TestRunCycle_TimestampUTC(t *testing.T) {
	pub := &mockPublisher{}
	beforeMs := time.Now().UnixMilli()
	runner := service.NewCycleRunner("testhost", pub)
	runner.RunCycle()
	afterMs := time.Now().UnixMilli()

	var payload metrics.SystemMetrics
	if err := proto.Unmarshal(pub.data, &payload); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if payload.TimestampMs < beforeMs || payload.TimestampMs > afterMs {
		t.Errorf("timestamp %d out of range [%d, %d]", payload.TimestampMs, beforeMs, afterMs)
	}
}
