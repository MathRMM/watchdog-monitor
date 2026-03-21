//go:build !windows

package collector

import (
	"errors"

	metrics "github.com/mathrmm/watchdog-monitor/gen/metrics"
)

// CollectGPU is not supported on non-Windows platforms.
// Returns (nil, err) as per RN02 — the cycle will omit the gpu field in the payload.
func CollectGPU() (*metrics.GpuMetrics, error) {
	return nil, errors.New("GPU collection via WMI is only supported on Windows")
}
