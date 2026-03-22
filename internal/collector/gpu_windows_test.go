//go:build windows

package collector_test

import (
	"strings"
	"testing"

	"github.com/mathrmm/watchdog-monitor/internal/collector"
)

// TestCollectGPU_Integration verifica que a query WMI é válida no Windows.
// Um erro de "no GPU found" é aceitável; um erro de query inválida indica bug.
func TestCollectGPU_Integration(t *testing.T) {
	result, err := collector.CollectGPU()
	if err != nil {
		lower := strings.ToLower(err.Error())
		if strings.Contains(lower, "invalid") || strings.Contains(lower, "inválid") || strings.Contains(lower, "exceção") {
			t.Fatalf("WMI query malformada — bug no coletor: %v", err)
		}
		// "no GPU found" é aceitável em VMs ou máquinas sem GPU detectável.
		t.Logf("GPU não disponível (aceitável): %v", err)
		return
	}

	if result == nil {
		t.Fatal("CollectGPU retornou (nil, nil) — estado inválido")
	}
	if result.DedicatedTotalBytes == 0 {
		t.Error("DedicatedTotalBytes deve ser > 0 quando GPU está presente")
	}
	if result.UsedPercent < 0 || result.UsedPercent > 100 {
		t.Errorf("UsedPercent fora do intervalo [0,100]: %f", result.UsedPercent)
	}
	t.Logf("GPU: %d bytes dedicados, %.1f%% uso", result.DedicatedTotalBytes, result.UsedPercent)
}
