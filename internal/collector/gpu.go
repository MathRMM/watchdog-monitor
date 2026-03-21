//go:build windows

package collector

import (
	"fmt"

	metrics "github.com/mathrmm/watchdog-monitor/gen/metrics"
	"github.com/yusufpapurcu/wmi"
)

type win32VideoController struct {
	Name         string
	AdapterRAM   uint32
	CurrentUsage uint32
}

// CollectGPU queries Win32_VideoController via WMI for GPU metrics (RF03).
// Returns (nil, err) if WMI is unavailable or the GPU doesn't expose usage data.
// This is expected behavior — the cycle will omit the gpu field in the payload (RN02).
//
// Known limitation: AMD RX 6600 may not expose CurrentUsage via Win32_VideoController.
// AdapterRAM is capped at uint32 (~4 GB) due to the WMI property type.
// Temperature via MSAcpi_ThermalZoneTemperature is attempted but silently omitted on failure.
func CollectGPU() (*metrics.GpuMetrics, error) {
	var controllers []win32VideoController
	if err := wmi.Query("SELECT Name, AdapterRAM, CurrentUsage FROM Win32_VideoController", &controllers); err != nil {
		return nil, fmt.Errorf("WMI query Win32_VideoController: %w", err)
	}
	if len(controllers) == 0 {
		return nil, fmt.Errorf("no GPU found via Win32_VideoController")
	}

	c := controllers[0]
	m := &metrics.GpuMetrics{
		UsedPercent:         float64(c.CurrentUsage),
		DedicatedTotalBytes: uint64(c.AdapterRAM),
	}

	// Temperature: attempt root\wmi MSAcpi_ThermalZoneTemperature.
	// Not all systems expose GPU temperature via this interface.
	// Failure is silent — TemperatureCelsius remains nil (optional field, RNF09).
	if temp, err := queryGPUTemperature(); err == nil {
		m.TemperatureCelsius = &temp
	}

	return m, nil
}

type thermalZone struct {
	CurrentTemperature uint32
}

// queryGPUTemperature attempts to read temperature via root\wmi namespace.
// Returns the temperature in Celsius, or error if unsupported.
func queryGPUTemperature() (float64, error) {
	var zones []thermalZone
	if err := wmi.QueryNamespace("SELECT CurrentTemperature FROM MSAcpi_ThermalZoneTemperature", &zones, "root\\wmi"); err != nil {
		return 0, fmt.Errorf("thermal zone WMI query: %w", err)
	}
	if len(zones) == 0 {
		return 0, fmt.Errorf("no thermal zone data")
	}
	// WMI returns temperature in tenths of Kelvin; convert to Celsius.
	tempCelsius := float64(zones[0].CurrentTemperature)/10.0 - 273.15
	return tempCelsius, nil
}
