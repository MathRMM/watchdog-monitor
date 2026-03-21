package collector

import (
	"fmt"

	metrics "github.com/mathrmm/watchdog-monitor/gen/metrics"
	psnet "github.com/shirou/gopsutil/v3/net"
)

// NetworkCollector is stateful — it tracks the previous IOCounters reading
// to calculate per-cycle byte deltas (RF04).
type NetworkCollector struct {
	lastCounters map[string]psnet.IOCountersStat
}

// NewNetworkCollector creates a NetworkCollector with no prior baseline.
// The first Collect() call will return deltas of 0 for all interfaces.
func NewNetworkCollector() *NetworkCollector {
	return &NetworkCollector{}
}

// Collect returns delta bytes sent/received per network interface since the last call.
// Interfaces that disappear between cycles are silently ignored (case from PRD).
// Interfaces with zero deltas in both directions are included if they were active.
func (c *NetworkCollector) Collect() ([]*metrics.NetworkInterface, error) {
	counters, err := psnet.IOCounters(true)
	if err != nil {
		return nil, fmt.Errorf("net io counters: %w", err)
	}

	current := make(map[string]psnet.IOCountersStat, len(counters))
	for _, ct := range counters {
		current[ct.Name] = ct
	}

	var result []*metrics.NetworkInterface
	for name, ct := range current {
		var sentDelta, recvDelta uint64
		if c.lastCounters != nil {
			if prev, ok := c.lastCounters[name]; ok {
				// Guard against counter reset (e.g. interface restart).
				if ct.BytesSent >= prev.BytesSent {
					sentDelta = ct.BytesSent - prev.BytesSent
				}
				if ct.BytesRecv >= prev.BytesRecv {
					recvDelta = ct.BytesRecv - prev.BytesRecv
				}
			}
		}
		result = append(result, &metrics.NetworkInterface{
			Name:           name,
			BytesSentDelta: sentDelta,
			BytesRecvDelta: recvDelta,
		})
	}

	c.lastCounters = current
	return result, nil
}
