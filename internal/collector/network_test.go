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

// TestNetworkCollector_AtLeastOneInterface verifica que o sistema possui pelo menos
// uma interface de rede detectável.
func TestNetworkCollector_AtLeastOneInterface(t *testing.T) {
	nc := collector.NewNetworkCollector()
	interfaces, err := nc.Collect()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(interfaces) == 0 {
		t.Error("expected at least one network interface")
	}
}

// TestNetworkCollector_InterfacesHaveNames verifica que cada interface retornada
// possui um Name não vazio.
func TestNetworkCollector_InterfacesHaveNames(t *testing.T) {
	nc := collector.NewNetworkCollector()
	interfaces, err := nc.Collect()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, iface := range interfaces {
		if iface.Name == "" {
			t.Errorf("interface[%d]: Name is empty", i)
		}
	}
}

// TestNetworkCollector_SecondCallSucceeds verifies that a second Collect() call
// succeeds and returns at least as many interfaces as the first.
// Non-negativity of deltas is guaranteed by the uint64 type (BytesSentDelta, BytesRecvDelta).
func TestNetworkCollector_SecondCallSucceeds(t *testing.T) {
	nc := collector.NewNetworkCollector()
	first, err := nc.Collect()
	if err != nil {
		t.Fatalf("first Collect() error: %v", err)
	}
	time.Sleep(100 * time.Millisecond)
	second, err := nc.Collect()
	if err != nil {
		t.Fatalf("second Collect() error: %v", err)
	}
	if len(second) < len(first) {
		t.Errorf("second call returned fewer interfaces (%d) than first (%d)", len(second), len(first))
	}
}
