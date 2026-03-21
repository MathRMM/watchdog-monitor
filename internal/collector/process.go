package collector

import (
	"fmt"
	"sort"

	metrics "github.com/mathrmm/watchdog-monitor/gen/metrics"
	"github.com/shirou/gopsutil/v3/process"
)

const maxTopProcesses = 5

// CollectProcesses returns the top 5 processes ordered by CPU usage descending (RF05, RN05).
// Processes that exit during collection are silently ignored (case from PRD).
// Processes that return Access Denied on any field are also silently skipped.
//
// Note: CPUPercent() uses a delta since the last call per process instance.
// On the first cycle, all processes may return 0% CPU — this is expected behavior.
func CollectProcesses() ([]*metrics.ProcessInfo, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, fmt.Errorf("list processes: %w", err)
	}

	var collected []*metrics.ProcessInfo
	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			// Process exited or access denied — skip silently.
			continue
		}

		cpuPct, err := p.CPUPercent()
		if err != nil {
			continue
		}

		mi, err := p.MemoryInfo()
		if err != nil {
			continue
		}

		var memBytes uint64
		if mi != nil {
			memBytes = mi.RSS
		}

		collected = append(collected, &metrics.ProcessInfo{
			Name:        name,
			Pid:         uint32(p.Pid),
			CpuPercent:  cpuPct,
			MemoryBytes: memBytes,
		})
	}

	sort.Slice(collected, func(i, j int) bool {
		return collected[i].CpuPercent > collected[j].CpuPercent
	})

	if len(collected) > maxTopProcesses {
		collected = collected[:maxTopProcesses]
	}

	return collected, nil
}
