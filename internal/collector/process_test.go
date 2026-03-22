package collector_test

import (
	"testing"

	"github.com/mathrmm/watchdog-monitor/internal/collector"
)

// TestCollectProcesses_AtMostFive verifies the top-5 cap (RF05).
func TestCollectProcesses_AtMostFive(t *testing.T) {
	processes, err := collector.CollectProcesses()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(processes) > 5 {
		t.Errorf("expected at most 5 processes, got %d", len(processes))
	}
}

// TestCollectProcesses_SortedByCPUDescending verifies ordering (RN05).
func TestCollectProcesses_SortedByCPUDescending(t *testing.T) {
	processes, err := collector.CollectProcesses()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i := 1; i < len(processes); i++ {
		if processes[i].CpuPercent > processes[i-1].CpuPercent {
			t.Errorf("processes not sorted by CPU descending at index %d: %f > %f",
				i, processes[i].CpuPercent, processes[i-1].CpuPercent)
		}
	}
}

// TestCollectProcesses_AtLeastOne verifies system has at least 1 visible process.
func TestCollectProcesses_AtLeastOne(t *testing.T) {
	processes, err := collector.CollectProcesses()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(processes) == 0 {
		t.Error("expected at least 1 process")
	}
}

// TestCollectProcesses_FieldsValid verifica que cada processo retornado tem
// Name não vazio e Pid válido (> 0).
func TestCollectProcesses_FieldsValid(t *testing.T) {
	processes, err := collector.CollectProcesses()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, p := range processes {
		if p.Name == "" {
			t.Errorf("process[%d]: Name is empty", i)
		}
		if p.Pid == 0 {
			t.Errorf("process[%d] (%s): Pid is 0", i, p.Name)
		}
	}
}
