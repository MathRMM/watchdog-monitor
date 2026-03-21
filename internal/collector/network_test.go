package collector_test

import (
	"testing"
	"time"

	"github.com/mathrmm/watchdog-monitor/internal/collector"
)

// TestNetworkCollector_FirstCallDeltasZero verifies that on the first Collect()
// call all deltas are 0 — no prior baseline exists.
func TestNetworkCollector_FirstCallDeltasZero(t *testing.T) {
	nc := collector.NewNetworkCollector()
	interfaces, err := nc.Collect()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, iface := range interfaces {
		if iface.BytesSentDelta != 0 || iface.BytesRecvDelta != 0 {
			t.Errorf("expected 0 deltas on first call for %s: sent=%d recv=%d",
				iface.Name, iface.BytesSentDelta, iface.BytesRecvDelta)
		}
	}
}

// TestNetworkCollector_SecondCallDeltasNonNegative verifies that deltas computed
// between two consecutive Collect() calls are non-negative.
func TestNetworkCollector_SecondCallDeltasNonNegative(t *testing.T) {
	nc := collector.NewNetworkCollector()
	if _, err := nc.Collect(); err != nil {
		t.Fatalf("first Collect() error: %v", err)
	}
	time.Sleep(100 * time.Millisecond)
	interfaces, err := nc.Collect()
	if err != nil {
		t.Fatalf("second Collect() error: %v", err)
	}
	for _, iface := range interfaces {
		if iface.BytesSentDelta < 0 || iface.BytesRecvDelta < 0 {
			t.Errorf("negative delta for %s: sent=%d recv=%d",
				iface.Name, iface.BytesSentDelta, iface.BytesRecvDelta)
		}
	}
}
