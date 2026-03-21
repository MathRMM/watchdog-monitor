package collector

import (
	"fmt"

	metrics "github.com/mathrmm/watchdog-monitor/gen/metrics"
	"github.com/shirou/gopsutil/v3/cpu"
)

// CollectCPU returns CPU usage metrics (total and per-core).
// Interval 0 uses gopsutil's internal delta since the last call.
// The first invocation may return 0% — this is expected and documented behavior.
func CollectCPU() (*metrics.CpuMetrics, error) {
	total, err := cpu.Percent(0, false)
	if err != nil {
		return nil, fmt.Errorf("cpu total percent: %w", err)
	}
	if len(total) == 0 {
		return nil, fmt.Errorf("cpu total percent returned empty slice")
	}

	perCore, err := cpu.Percent(0, true)
	if err != nil {
		return nil, fmt.Errorf("cpu per-core percent: %w", err)
	}

	return &metrics.CpuMetrics{
		TotalPercent:   total[0],
		PerCorePercent: perCore,
	}, nil
}
