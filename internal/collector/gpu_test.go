package collector_test

import (
	"testing"

	"github.com/mathrmm/watchdog-monitor/internal/collector"
)

// TestCollectGPU_NoPanic verifies that CollectGPU never panics.
// On non-Windows or systems without WMI support, it must return (nil, err).
func TestCollectGPU_NoPanic(t *testing.T) {
	result, err := collector.CollectGPU()
	if err != nil {
		if result != nil {
			t.Error("expected nil result when err is non-nil (RN02)")
		}
		// Error is acceptable — WMI unavailable or non-Windows platform.
		return
	}
	// GPU available: validate field ranges.
	if result.UsedPercent < 0 || result.UsedPercent > 100 {
		t.Errorf("UsedPercent out of range: %f", result.UsedPercent)
	}
}
