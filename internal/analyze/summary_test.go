package analyze

import (
	"testing"

	"sysmon/internal/collect"
)

func makeSample(cpuUsage float64) collect.Sample {
	return collect.Sample{
		CPU:   &collect.CPUSample{Usage: cpuUsage, User: 1, System: 1, Idle: 100 - cpuUsage},
		Cores: []collect.CPUSample{{Usage: cpuUsage / 2}},
		Mem:   &collect.MemSample{UsagePct: 50, UsedMB: 1000, AvailableMB: 1000, SwapPct: 1},
	}
}

func TestSummaryOf(t *testing.T) {
	samples := []collect.Sample{makeSample(10), makeSample(20), makeSample(30)}
	for i := range samples {
		samples[i].Timestamp = int64(i) * 1000
	}
	s := SummaryOf(samples, 1)
	if s.Samples != 3 || s.DurationSec != 2 {
		t.Fatalf("samples=%d dur=%v", s.Samples, s.DurationSec)
	}
	if s.CPU == nil || s.CPU.Usage.Avg != 20 || s.CPU.Usage.Max != 30 {
		t.Fatalf("cpu summary wrong: %+v", s.CPU)
	}
	if s.CPU.PerCore["core0"].Avg != 10 {
		t.Fatalf("per-core wrong: %+v", s.CPU.PerCore)
	}
	if s.Mem.UsagePct.Avg != 50 || s.Mem.UsedMB.Avg != 1000 {
		t.Fatalf("mem wrong: %+v", s.Mem)
	}
}
