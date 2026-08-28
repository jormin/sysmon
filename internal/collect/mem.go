package collect

import "github.com/shirou/gopsutil/v3/mem"

// fetchMem 读取内存与 swap 使用情况。
func fetchMem() (*MemSample, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	s := &MemSample{
		TotalMB:     float64(v.Total) / 1e6,
		UsedMB:      float64(v.Used) / 1e6,
		AvailableMB: float64(v.Available) / 1e6,
		BuffersMB:   float64(v.Buffers) / 1e6,
		CachedMB:    float64(v.Cached) / 1e6,
		UsagePct:    v.UsedPercent,
		SwapTotalMB: float64(v.SwapTotal) / 1e6,
		SwapUsedMB:  float64(v.SwapTotal-v.SwapFree) / 1e6,
	}
	if s.SwapTotalMB > 0 {
		s.SwapPct = s.SwapUsedMB / s.SwapTotalMB * 100
	}
	return s, nil
}
