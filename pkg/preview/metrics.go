package preview

import (
	"runtime"
	"time"
)

// RenderMetrics is an opt-in diagnostic for one synchronous render. Durations
// accumulate across operations. Heap values are boundary samples, not RSS peaks
// or an attribution of all live memory to the named stage.
type RenderMetrics struct {
	Stages map[string]StageMetric `json:"stages"`
}

type StageMetric struct {
	Milliseconds   float64 `json:"milliseconds"`
	Operations     int     `json:"operations"`
	MaxSampledHeap uint64  `json:"max_sampled_heap_bytes"`
}

func (m *RenderMetrics) start() time.Time {
	if m == nil {
		return time.Time{}
	}
	return time.Now()
}

// Record permits callers to include DAT/scenario reads, grading and encoding in
// the same report as the renderer's asset and compositing stages.
func (m *RenderMetrics) Record(stage string, start time.Time) {
	if m == nil {
		return
	}
	duration := time.Since(start)
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	if m.Stages == nil {
		m.Stages = map[string]StageMetric{}
	}
	s := m.Stages[stage]
	s.Milliseconds += float64(duration) / float64(time.Millisecond)
	s.Operations++
	s.MaxSampledHeap = max(s.MaxSampledHeap, mem.HeapAlloc)
	m.Stages[stage] = s
}
