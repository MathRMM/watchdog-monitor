package collector_test

import (
	"testing"

	"github.com/mathrmm/watchdog-monitor/internal/collector"
)

func TestCollectMemory_ValidData(t *testing.T) {
	m, err := collector.CollectMemory()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.TotalBytes == 0 {
		t.Error("TotalBytes should be > 0")
	}
	if m.AvailableBytes > m.TotalBytes {
		t.Errorf("AvailableBytes (%d) > TotalBytes (%d)", m.AvailableBytes, m.TotalBytes)
	}
	if m.UsedPercent < 0 || m.UsedPercent > 100 {
		t.Errorf("UsedPercent out of range: %f", m.UsedPercent)
	}
}
