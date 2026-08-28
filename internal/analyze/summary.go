package analyze

import (
	"fmt"
	"time"

	"sysmon/internal/collect"
)

// CPUSummary 汇总 CPU 指标，含每核。
type CPUSummary struct {
	Usage   MetricStats            `json:"usage_pct"`
	User    MetricStats            `json:"user_pct"`
	System  MetricStats            `json:"system_pct"`
	IOWait  MetricStats            `json:"iowait_pct"`
	PerCore map[string]MetricStats `json:"per_core,omitempty"`
}

// MemSummary 汇总内存指标。
type MemSummary struct {
	UsagePct    MetricStats `json:"usage_pct"`
	UsedMB      MetricStats `json:"used_mb"`
	AvailableMB MetricStats `json:"available_mb"`
	SwapPct     MetricStats `json:"swap_pct"`
}

// Summary 是整轮采样的统计汇总。
type Summary struct {
	StartedAt   time.Time   `json:"started_at"`
	EndedAt     time.Time   `json:"ended_at"`
	IntervalSec float64     `json:"interval_sec"`
	Samples     int         `json:"samples"`
	DurationSec float64     `json:"duration_sec"`
	CPU         *CPUSummary `json:"cpu,omitempty"`
	Mem         *MemSummary `json:"mem,omitempty"`
}

// SummaryOf 把采样序列聚合为 Summary；dt 为采样间隔（秒）。
func SummaryOf(samples []collect.Sample, dt float64) Summary {
	s := Summary{IntervalSec: dt, Samples: len(samples)}
	if len(samples) == 0 {
		return s
	}
	s.StartedAt = time.UnixMilli(samples[0].Timestamp)
	s.EndedAt = time.UnixMilli(samples[len(samples)-1].Timestamp)
	s.DurationSec = float64(samples[len(samples)-1].Timestamp-samples[0].Timestamp) / 1000

	var cpuUsage, cpuUser, cpuSystem, cpuIowait []float64
	coreVals := map[int][]float64{}
	var memUsage, memUsed, memAvail, memSwap []float64

	for _, sm := range samples {
		if sm.CPU != nil {
			cpuUsage = append(cpuUsage, sm.CPU.Usage)
			cpuUser = append(cpuUser, sm.CPU.User)
			cpuSystem = append(cpuSystem, sm.CPU.System)
			cpuIowait = append(cpuIowait, sm.CPU.IOWait)
			for i, c := range sm.Cores {
				coreVals[i] = append(coreVals[i], c.Usage)
			}
		}
		if sm.Mem != nil {
			memUsage = append(memUsage, sm.Mem.UsagePct)
			memUsed = append(memUsed, sm.Mem.UsedMB)
			memAvail = append(memAvail, sm.Mem.AvailableMB)
			memSwap = append(memSwap, sm.Mem.SwapPct)
		}
	}

	if len(cpuUsage) > 0 {
		s.CPU = &CPUSummary{
			Usage: Stats(cpuUsage, dt), User: Stats(cpuUser, dt),
			System: Stats(cpuSystem, dt), IOWait: Stats(cpuIowait, dt),
		}
		if len(coreVals) > 0 {
			s.CPU.PerCore = map[string]MetricStats{}
			for i, v := range coreVals {
				s.CPU.PerCore[fmt.Sprintf("core%d", i)] = Stats(v, dt)
			}
		}
	}
	if len(memUsage) > 0 {
		s.Mem = &MemSummary{
			UsagePct: Stats(memUsage, dt), UsedMB: Stats(memUsed, dt),
			AvailableMB: Stats(memAvail, dt), SwapPct: Stats(memSwap, dt),
		}
	}
	return s
}
