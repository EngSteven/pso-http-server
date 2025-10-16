package metrics

import (
	"sort"
	"sync"
	"time"
)

type PoolMetrics struct {
	mu            sync.Mutex
	TotalProcessed int64
	TotalLatencyMs int64
	Samples       []int64 // ring buffer-like (append up to cap)
	maxSamples    int
}

func NewPoolMetrics(maxSamples int) *PoolMetrics {
	return &PoolMetrics{Samples: make([]int64, 0, maxSamples), maxSamples: maxSamples}
}

func (m *PoolMetrics) Record(latency time.Duration) {
	ms := latency.Milliseconds()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalProcessed++
	m.TotalLatencyMs += ms
	if len(m.Samples) < m.maxSamples {
		m.Samples = append(m.Samples, ms)
	} else {
		// simple replacement: drop oldest, append new (not circular for simplicity)
		m.Samples = append(m.Samples[1:], ms)
	}
}

func (m *PoolMetrics) AvgLatencyMs() float64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.TotalProcessed == 0 {
		return 0
	}
	return float64(m.TotalLatencyMs) / float64(m.TotalProcessed)
}

func (m *PoolMetrics) Percentile(p float64) int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.Samples) == 0 {
		return 0
	}
	cp := make([]int64, len(m.Samples))
	copy(cp, m.Samples)
	sort.Slice(cp, func(i, j int) bool { return cp[i] < cp[j] })
	idx := int((p/100.0)*float64(len(cp)-1) + 0.5)
	if idx < 0 {
		idx = 0
	}
	if idx >= len(cp) {
		idx = len(cp) - 1
	}
	return cp[idx]
}
