package metrics_test

import (
	"testing"
	"time"

	pb "github.com/mathrmm/watchdog-monitor/gen/metrics"
	"google.golang.org/protobuf/proto"
)

// dado SystemMetrics com todos os campos preenchidos (incluindo GPU),
// quando serializado e desserializado, então todos os valores são preservados.
func TestMarshal_FullPayload(t *testing.T) {
	temp := float64(72.5)
	original := &pb.SystemMetrics{
		Hostname:    "test-host",
		TimestampMs: time.Now().UnixMilli(),
		Cpu: &pb.CpuMetrics{
			TotalPercent:   45.5,
			PerCorePercent: []float64{40.0, 51.0, 43.0, 47.0},
		},
		Memory: &pb.MemoryMetrics{
			TotalBytes:     16 * 1024 * 1024 * 1024,
			AvailableBytes: 8 * 1024 * 1024 * 1024,
			UsedPercent:    50.0,
		},
		Gpu: &pb.GpuMetrics{
			UsedPercent:         65.0,
			DedicatedUsedBytes:  4 * 1024 * 1024 * 1024,
			DedicatedTotalBytes: 8 * 1024 * 1024 * 1024,
			TemperatureCelsius:  &temp,
		},
		Network: []*pb.NetworkInterface{
			{Name: "Ethernet", BytesSentDelta: 1024, BytesRecvDelta: 2048},
		},
		TopProcesses: []*pb.ProcessInfo{
			{Name: "chrome.exe", Pid: 1234, CpuPercent: 12.5, MemoryBytes: 500 * 1024 * 1024},
		},
	}

	data, err := proto.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	restored := &pb.SystemMetrics{}
	if err := proto.Unmarshal(data, restored); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if restored.Hostname != original.Hostname {
		t.Errorf("Hostname: got %q, want %q", restored.Hostname, original.Hostname)
	}
	if restored.TimestampMs != original.TimestampMs {
		t.Errorf("TimestampMs: got %d, want %d", restored.TimestampMs, original.TimestampMs)
	}
	if restored.Cpu.TotalPercent != original.Cpu.TotalPercent {
		t.Errorf("Cpu.TotalPercent: got %f, want %f", restored.Cpu.TotalPercent, original.Cpu.TotalPercent)
	}
	if len(restored.Cpu.PerCorePercent) != len(original.Cpu.PerCorePercent) {
		t.Errorf("Cpu.PerCorePercent len: got %d, want %d", len(restored.Cpu.PerCorePercent), len(original.Cpu.PerCorePercent))
	}
	if restored.Memory.TotalBytes != original.Memory.TotalBytes {
		t.Errorf("Memory.TotalBytes: got %d, want %d", restored.Memory.TotalBytes, original.Memory.TotalBytes)
	}
	if restored.Gpu == nil {
		t.Fatal("Gpu: expected non-nil, got nil")
	}
	if restored.Gpu.UsedPercent != original.Gpu.UsedPercent {
		t.Errorf("Gpu.UsedPercent: got %f, want %f", restored.Gpu.UsedPercent, original.Gpu.UsedPercent)
	}
	if restored.Gpu.TemperatureCelsius == nil {
		t.Fatal("Gpu.TemperatureCelsius: expected non-nil, got nil")
	}
	if *restored.Gpu.TemperatureCelsius != *original.Gpu.TemperatureCelsius {
		t.Errorf("Gpu.TemperatureCelsius: got %f, want %f", *restored.Gpu.TemperatureCelsius, *original.Gpu.TemperatureCelsius)
	}
}

// dado SystemMetrics sem o campo gpu (nil),
// quando serializado, então a desserialização resulta em campo GPU nil (não zero-value).
func TestMarshal_GpuNil(t *testing.T) {
	original := &pb.SystemMetrics{
		Hostname:    "test-host",
		TimestampMs: 1000000,
		Cpu:         &pb.CpuMetrics{TotalPercent: 10.0},
		Memory:      &pb.MemoryMetrics{TotalBytes: 1024},
		Gpu:         nil,
	}

	data, err := proto.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	restored := &pb.SystemMetrics{}
	if err := proto.Unmarshal(data, restored); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if restored.Gpu != nil {
		t.Errorf("Gpu: expected nil, got %+v", restored.Gpu)
	}
}

// dado GpuMetrics sem temperature_celsius,
// quando serializado, então temperatura é distinguível de 0.0 após desserialização.
func TestMarshal_GpuTemperatureNil(t *testing.T) {
	gpu := &pb.GpuMetrics{
		UsedPercent:         50.0,
		DedicatedUsedBytes:  1024,
		DedicatedTotalBytes: 2048,
		TemperatureCelsius:  nil, // não fornecida
	}

	data, err := proto.Marshal(gpu)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	restored := &pb.GpuMetrics{}
	if err := proto.Unmarshal(data, restored); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if restored.TemperatureCelsius != nil {
		t.Errorf("TemperatureCelsius: expected nil (distinguível de 0.0), got %v", *restored.TemperatureCelsius)
	}
}

// dado timestamp_ms com valor Unix milliseconds UTC,
// quando serializado e desserializado, então valor é exato.
func TestMarshal_TimestampExact(t *testing.T) {
	tsMs := time.Now().UnixMilli()
	original := &pb.SystemMetrics{
		Hostname:    "ts-test",
		TimestampMs: tsMs,
	}

	data, err := proto.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	restored := &pb.SystemMetrics{}
	if err := proto.Unmarshal(data, restored); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if restored.TimestampMs != tsMs {
		t.Errorf("TimestampMs: got %d, want %d", restored.TimestampMs, tsMs)
	}
}

// payload com lista top_processes vazia serializa e desserializa sem erro.
func TestMarshal_EmptyTopProcesses(t *testing.T) {
	original := &pb.SystemMetrics{
		Hostname:     "test-host",
		TimestampMs:  1000000,
		TopProcesses: []*pb.ProcessInfo{},
	}

	data, err := proto.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	restored := &pb.SystemMetrics{}
	if err := proto.Unmarshal(data, restored); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if len(restored.TopProcesses) != 0 {
		t.Errorf("TopProcesses: expected empty, got %d elements", len(restored.TopProcesses))
	}
}
