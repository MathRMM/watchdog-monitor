package service

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"

	metrics "github.com/mathrmm/watchdog-monitor/gen/metrics"
	"github.com/mathrmm/watchdog-monitor/internal/collector"
	"github.com/mathrmm/watchdog-monitor/internal/logger"
)

// Publisher defines the interface for publishing serialized payloads.
// Defined here (consumer side) to avoid import cycles with the publisher package.
// publisher.Publisher satisfies this interface via duck typing.
type Publisher interface {
	Publish(subject string, data []byte) error
}

// CycleRunner collects metrics each tick, builds the Protobuf payload, and publishes it.
type CycleRunner struct {
	hostname      string
	subject       string
	pub           Publisher
	netCollector  *collector.NetworkCollector
	gpuErrActive  bool // true while GPU collector is failing (log-once suppression)
}

// NewCycleRunner creates a CycleRunner ready to run.
// hostname must already be sanitized via SanitizeHostname.
func NewCycleRunner(hostname string, pub Publisher) *CycleRunner {
	return &CycleRunner{
		hostname:     hostname,
		subject:      fmt.Sprintf("watchdog.%s.metrics", hostname),
		pub:          pub,
		netCollector: collector.NewNetworkCollector(),
	}
}

// RunCycle executes a single collection cycle:
// collect all metrics → build SystemMetrics → proto.Marshal → Publish.
//
// Collector failures result in zero/nil fields — the publish still happens (RN02).
// A Publish failure is logged and silently discarded — the cycle continues (RN03).
func (r *CycleRunner) RunCycle() {
	ts := time.Now().UnixMilli() // capture timestamp before collection (RF12)

	cpuMetrics, err := collector.CollectCPU()
	if err != nil {
		logger.Error("cpu collect: %v", err)
	}

	memMetrics, err := collector.CollectMemory()
	if err != nil {
		logger.Error("memory collect: %v", err)
	}

	// GPU failure → nil field in payload, not a skipped publish (RN02, RNF09).
	// Log-once suppression: first failure is logged; subsequent identical failures
	// are silenced until the collector recovers (Alternativo C do PRD, Fase 5).
	gpuMetrics, err := collector.CollectGPU()
	if err != nil {
		if !r.gpuErrActive {
			logger.Error("gpu collect: %v", err)
			r.gpuErrActive = true
		}
		gpuMetrics = nil
	} else if r.gpuErrActive {
		logger.Info("gpu collector recovered")
		r.gpuErrActive = false
	}

	netInterfaces, err := r.netCollector.Collect()
	if err != nil {
		logger.Error("network collect: %v", err)
	}

	procList, err := collector.CollectProcesses()
	if err != nil {
		logger.Error("process collect: %v", err)
	}

	payload := &metrics.SystemMetrics{
		Hostname:     r.hostname,
		TimestampMs:  ts,
		Cpu:          cpuMetrics,
		Memory:       memMetrics,
		Gpu:          gpuMetrics,
		Network:      netInterfaces,
		TopProcesses: procList,
	}

	data, err := proto.Marshal(payload)
	if err != nil {
		logger.Error("proto marshal: %v", err)
		return
	}

	if err := r.pub.Publish(r.subject, data); err != nil {
		logger.Error("nats publish: %v", err) // discard silently (RN03)
	}
}

// Run starts the absolute ticker loop and calls RunCycle on each tick (RF07, RN06).
// Uses time.NewTicker — not time.Sleep — so ticks are absolute and drift-free.
// Blocks until stopCh is closed.
func (r *CycleRunner) Run(stopCh <-chan struct{}) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			r.RunCycle()
		}
	}
}

// SanitizeHostname replaces any character outside [a-zA-Z0-9-] with a hyphen,
// collapses consecutive hyphens, and trims leading/trailing hyphens (RN04).
func SanitizeHostname(hostname string) string {
	reInvalid := regexp.MustCompile(`[^a-zA-Z0-9-]+`)
	reDouble := regexp.MustCompile(`-{2,}`)

	s := reInvalid.ReplaceAllString(hostname, "-")
	s = reDouble.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}
