package collector

import (
	"fmt"

	metrics "github.com/mathrmm/watchdog-monitor/gen/metrics"
	"github.com/shirou/gopsutil/v3/mem"
)

// CollectMemory returns RAM usage metrics (RF02).
func CollectMemory() (*metrics.MemoryMetrics, error) {
	vm, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("virtual memory: %w", err)
	}

	return &metrics.MemoryMetrics{
		TotalBytes:     vm.Total,
		AvailableBytes: vm.Available,
		UsedPercent:    vm.UsedPercent,
	}, nil
}
