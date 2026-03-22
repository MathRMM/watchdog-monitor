package collector_test

import (
	"testing"

	"github.com/mathrmm/watchdog-monitor/internal/collector"
)

// TestCollectGPU_NoPanic verifica que CollectGPU nunca panics.
// Em plataformas não-Windows retorna (nil, err) — isso é comportamento esperado (RN02).
// Em Windows, a validação real da query está em gpu_windows_test.go.
func TestCollectGPU_NoPanic(t *testing.T) {
	result, err := collector.CollectGPU()
	if err != nil {
		if result != nil {
			t.Error("esperado result == nil quando err != nil (RN02)")
		}
		return
	}
	if result == nil {
		t.Fatal("CollectGPU retornou (nil, nil) — estado inválido")
	}
	if result.UsedPercent < 0 || result.UsedPercent > 100 {
		t.Errorf("UsedPercent fora do intervalo [0,100]: %f", result.UsedPercent)
	}
}
