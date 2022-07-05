package metrics

import (
	"sort"
	"time"
)

type LatencyReport struct {
	P50 time.Duration
	P90 time.Duration
	P99 time.Duration
}

type LatencyMetrics struct {
	name      string
	latencies []time.Duration
}

func NewLatencyMetrics(name string) *LatencyMetrics {
	return &LatencyMetrics{
		name:      name,
		latencies: make([]time.Duration, 0),
	}
}

func (m *LatencyMetrics) AddLatencyMetrics(newLatency time.Duration) {
	m.latencies = append(m.latencies, newLatency)
}

func (m *LatencyMetrics) Len() int {
	return len(m.latencies)
}

func (m *LatencyMetrics) Less(i, j int) bool {
	return m.latencies[i] < m.latencies[j]
}

func (m *LatencyMetrics) Swap(i, j int) {
	m.latencies[i], m.latencies[j] = m.latencies[j], m.latencies[i]
}

func (m *LatencyMetrics) GetSummary() *LatencyReport {
	// sort
	sort.Sort(m)
	count := len(m.latencies)
	return &LatencyReport{
		P50: m.latencies[count/2-1],
		P90: m.latencies[count-count/10-1],
		P99: m.latencies[count-count/100-1],
	}
}
