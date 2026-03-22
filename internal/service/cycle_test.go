package service_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	natsgo "github.com/nats-io/nats.go"
	natsserver "github.com/nats-io/nats-server/v2/server"
	"google.golang.org/protobuf/proto"

	metrics "github.com/mathrmm/watchdog-monitor/gen/metrics"
	"github.com/mathrmm/watchdog-monitor/internal/collector"
	"github.com/mathrmm/watchdog-monitor/internal/publisher"
	"github.com/mathrmm/watchdog-monitor/internal/service"
)

// =============================================================================
// Test helpers
// =============================================================================

// mockPublisher captures the last Publish call for assertions.
type mockPublisher struct {
	err     error
	subject string
	data    []byte
	called  bool
}

func (m *mockPublisher) Publish(subject string, data []byte) error {
	m.called = true
	m.subject = subject
	m.data = data
	return m.err
}

// blockingPublisher blocks inside Publish until released — simulates a slow publish.
type blockingPublisher struct {
	once    sync.Once
	entered chan struct{} // closed on first Publish entry
	block   chan struct{} // close to unblock Publish
}

func newBlockingPublisher() *blockingPublisher {
	return &blockingPublisher{
		entered: make(chan struct{}),
		block:   make(chan struct{}),
	}
}

func (b *blockingPublisher) Publish(_ string, _ []byte) error {
	b.once.Do(func() { close(b.entered) })
	<-b.block
	return nil
}

// noopCollectors returns instantly with minimal valid data — for shutdown/timing tests.
func noopCollectors() service.Collectors {
	return service.Collectors{
		CPU:     func() (*metrics.CpuMetrics, error) { return &metrics.CpuMetrics{}, nil },
		Memory:  func() (*metrics.MemoryMetrics, error) { return &metrics.MemoryMetrics{}, nil },
		GPU:     func() (*metrics.GpuMetrics, error) { return nil, errors.New("no gpu stub") },
		Network: func() ([]*metrics.NetworkInterface, error) { return nil, nil },
		Process: func() ([]*metrics.ProcessInfo, error) { return nil, nil },
	}
}

// realCollectors returns service.Collectors wired to the real system collectors.
func realCollectors() service.Collectors {
	nc := collector.NewNetworkCollector()
	return service.Collectors{
		CPU:     collector.CollectCPU,
		Memory:  collector.CollectMemory,
		GPU:     collector.CollectGPU,
		Network: nc.Collect,
		Process: collector.CollectProcesses,
	}
}

// startEmbeddedNATS starts an embedded NATS server and registers cleanup.
func startEmbeddedNATS(t *testing.T) *natsserver.Server {
	t.Helper()
	opts := &natsserver.Options{Port: -1}
	s, err := natsserver.NewServer(opts)
	if err != nil {
		t.Fatalf("embedded NATS: %v", err)
	}
	go s.Start()
	if !s.ReadyForConnections(2 * time.Second) {
		t.Fatal("embedded NATS not ready")
	}
	t.Cleanup(s.Shutdown)
	return s
}

// =============================================================================
// Hostname sanitization (RN04)
// =============================================================================

func TestSanitizeHostname_DotToHyphen(t *testing.T) {
	if got := service.SanitizeHostname("MY-PC.local"); got != "MY-PC-local" {
		t.Errorf("expected MY-PC-local, got %s", got)
	}
}

func TestSanitizeHostname_SpaceToHyphen(t *testing.T) {
	if got := service.SanitizeHostname("My PC"); got != "My-PC" {
		t.Errorf("expected My-PC, got %s", got)
	}
}

func TestSanitizeHostname_CollapseConsecutiveHyphens(t *testing.T) {
	if got := service.SanitizeHostname("pc--lab"); got != "pc-lab" {
		t.Errorf("expected pc-lab, got %s", got)
	}
}

// =============================================================================
// RunCycle — comportamento dos coletores
// =============================================================================

// TestRunCycle_AllCollectorsSuccess verifica que um ciclo completo publica um payload válido.
func TestRunCycle_AllCollectorsSuccess(t *testing.T) {
	pub := &mockPublisher{}
	runner := service.NewCycleRunnerWith("testhost", pub, realCollectors(), 5*time.Second)
	runner.RunCycle()

	if !pub.called {
		t.Fatal("expected Publish to be called")
	}
	if pub.subject != "watchdog.testhost.metrics" {
		t.Errorf("unexpected subject: %s", pub.subject)
	}

	var payload metrics.SystemMetrics
	if err := proto.Unmarshal(pub.data, &payload); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if payload.Hostname != "testhost" {
		t.Errorf("expected hostname testhost, got %s", payload.Hostname)
	}
	if payload.TimestampMs == 0 {
		t.Error("expected non-zero TimestampMs")
	}
	if payload.Cpu == nil {
		t.Error("expected non-nil Cpu field")
	}
	if payload.Memory == nil {
		t.Error("expected non-nil Memory field")
	}
}

// TestRunCycle_GPUError_PublishesWithNilGPU verifica que quando GPU falha,
// o payload ainda é publicado com gpu == nil (RN02, RNF09).
// A falha do GPU é injetada via stub — independente de plataforma ou hardware.
func TestRunCycle_GPUError_PublishesWithNilGPU(t *testing.T) {
	pub := &mockPublisher{}
	c := realCollectors()
	c.GPU = func() (*metrics.GpuMetrics, error) {
		return nil, errors.New("gpu injected failure")
	}
	runner := service.NewCycleRunnerWith("testhost", pub, c, 5*time.Second)
	runner.RunCycle()

	if !pub.called {
		t.Fatal("expected Publish to be called even when GPU fails")
	}

	var payload metrics.SystemMetrics
	if err := proto.Unmarshal(pub.data, &payload); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if payload.Gpu != nil {
		t.Errorf("expected nil Gpu when GPU collector fails (RN02), got %+v", payload.Gpu)
	}
	// Outros campos devem continuar populados.
	if payload.Cpu == nil {
		t.Error("expected non-nil Cpu even when GPU fails")
	}
}

// TestRunCycle_CPUError_StillPublishes verifica que falha no CPU não suprime o publish (RN02).
func TestRunCycle_CPUError_StillPublishes(t *testing.T) {
	pub := &mockPublisher{}
	c := noopCollectors()
	c.CPU = func() (*metrics.CpuMetrics, error) {
		return nil, errors.New("cpu injected failure")
	}
	runner := service.NewCycleRunnerWith("testhost", pub, c, 5*time.Second)
	runner.RunCycle()

	if !pub.called {
		t.Fatal("expected Publish to be called even when CPU fails")
	}

	var payload metrics.SystemMetrics
	if err := proto.Unmarshal(pub.data, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if payload.Cpu != nil {
		t.Errorf("expected nil Cpu when collector fails, got %+v", payload.Cpu)
	}
}

// TestRunCycle_PublishError_NoPanic verifica que falha no Publish é descartada
// silenciosamente — o ciclo não panics (RN03).
func TestRunCycle_PublishError_NoPanic(t *testing.T) {
	pub := &mockPublisher{err: errors.New("NATS unavailable")}
	runner := service.NewCycleRunnerWith("testhost", pub, noopCollectors(), 5*time.Second)
	runner.RunCycle() // must not panic
}

// TestRunCycle_TimestampUTC verifica que timestamp_ms reflete o momento da coleta
// em Unix milliseconds UTC (RF12).
func TestRunCycle_TimestampUTC(t *testing.T) {
	pub := &mockPublisher{}
	beforeMs := time.Now().UnixMilli()
	runner := service.NewCycleRunnerWith("testhost", pub, noopCollectors(), 5*time.Second)
	runner.RunCycle()
	afterMs := time.Now().UnixMilli()

	var payload metrics.SystemMetrics
	if err := proto.Unmarshal(pub.data, &payload); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if payload.TimestampMs < beforeMs || payload.TimestampMs > afterMs {
		t.Errorf("timestamp %d out of range [%d, %d]", payload.TimestampMs, beforeMs, afterMs)
	}
}

// =============================================================================
// Run — comportamento do loop e shutdown gracioso
// =============================================================================

// TestRun_StopsOnClosedChannel verifica que Run() retorna quando stopCh é fechado
// sem goroutine leak ou deadlock (Fase 5 shutdown gracioso).
func TestRun_StopsOnClosedChannel(t *testing.T) {
	pub := &mockPublisher{}
	runner := service.NewCycleRunnerWith("testhost", pub, noopCollectors(), 10*time.Millisecond)

	stopCh := make(chan struct{})
	done := make(chan struct{})

	go func() {
		runner.Run(stopCh)
		close(done)
	}()

	close(stopCh)

	select {
	case <-done:
		// OK — goroutine encerrou
	case <-time.After(time.Second):
		t.Fatal("Run() did not return after stopCh was closed (goroutine leak or deadlock)")
	}
}

// TestRun_ShutdownDuringCycle verifica que fechar stopCh enquanto RunCycle está
// em execução não causa deadlock — Run() retorna após o ciclo concluir (Fase 5 borda).
func TestRun_ShutdownDuringCycle(t *testing.T) {
	// blockingPublisher garante que a goroutine está dentro de Publish quando stopCh fecha.
	blockPub := newBlockingPublisher()
	runner := service.NewCycleRunnerWith("testhost", blockPub, noopCollectors(), 10*time.Millisecond)

	stopCh := make(chan struct{})
	done := make(chan struct{})

	go func() {
		runner.Run(stopCh)
		close(done)
	}()

	// Aguarda até o primeiro Publish ser chamado (ciclo em andamento).
	select {
	case <-blockPub.entered:
	case <-time.After(3 * time.Second):
		t.Fatal("timeout: Publish was never called")
	}

	// Sinaliza stop com ciclo em andamento.
	close(stopCh)
	// Desbloqueia o Publish para que RunCycle possa concluir e Run retornar.
	close(blockPub.block)

	select {
	case <-done:
		// OK
	case <-time.After(3 * time.Second):
		t.Fatal("Run() did not return after stop signal during active cycle (deadlock)")
	}
}

// =============================================================================
// End-to-end — stack completo: coletores → NATS real → payload verificado
// =============================================================================

// TestEndToEnd_RunCycleDeliversParsablePayload verifica o stack completo:
// coletores reais → proto.Marshal → NATS publish → NATS receive → proto.Unmarshal.
func TestEndToEnd_RunCycleDeliversParsablePayload(t *testing.T) {
	s := startEmbeddedNATS(t)

	pub, err := publisher.Connect(s.ClientURL())
	if err != nil {
		t.Fatalf("publisher.Connect: %v", err)
	}
	defer pub.Close()

	nc, err := natsgo.Connect(s.ClientURL())
	if err != nil {
		t.Fatalf("subscriber connect: %v", err)
	}
	defer nc.Close()

	received := make(chan []byte, 1)
	if _, err := nc.Subscribe("watchdog.e2ehost.metrics", func(msg *natsgo.Msg) {
		received <- msg.Data
	}); err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	if err := nc.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}

	runner := service.NewCycleRunnerWith("e2ehost", pub, realCollectors(), 5*time.Second)
	runner.RunCycle()

	select {
	case data := <-received:
		var payload metrics.SystemMetrics
		if err := proto.Unmarshal(data, &payload); err != nil {
			t.Fatalf("proto.Unmarshal: %v", err)
		}
		if payload.Hostname != "e2ehost" {
			t.Errorf("Hostname: got %q, want e2ehost", payload.Hostname)
		}
		if payload.TimestampMs == 0 {
			t.Error("TimestampMs is 0")
		}
		if payload.Cpu == nil {
			t.Error("Cpu field is nil")
		}
		if payload.Memory == nil {
			t.Error("Memory field is nil")
		}
		if len(payload.Network) == 0 {
			t.Error("Network is empty — expected at least one interface")
		}
		if len(payload.TopProcesses) == 0 {
			t.Error("TopProcesses is empty — expected at least one process")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout: no payload received via NATS")
	}
}
