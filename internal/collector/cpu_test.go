package collector_test

import (
	"testing"

	"github.com/mathrmm/watchdog-monitor/internal/collector"
)

func TestCollectCPU_ValidData(t *testing.T) {
	m, err := collector.CollectCPU()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.TotalPercent < 0 || m.TotalPercent > 100 {
		t.Errorf("TotalPercent out of range: %f", m.TotalPercent)
	}
	if len(m.PerCorePercent) == 0 {
		t.Error("expected at least 1 core in PerCorePercent")
	}
	for i, p := range m.PerCorePercent {
		if p < 0 || p > 100 {
			t.Errorf("core %d percent out of range: %f", i, p)
		}
	}
}
