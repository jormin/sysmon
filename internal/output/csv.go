package output

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"time"

	"sysmon/internal/collect"
)

func cpuOf(s collect.Sample) collect.CPUSample {
	if s.CPU != nil {
		return *s.CPU
	}
	return collect.CPUSample{}
}

func memOf(s collect.Sample) collect.MemSample {
	if s.Mem != nil {
		return *s.Mem
	}
	return collect.MemSample{}
}

func f(v float64) string { return strconv.FormatFloat(v, 'f', 3, 64) }

// WriteCSV 输出一行一条采样。列集合由第一条采样决定（每核动态扩列）。
func WriteCSV(w io.Writer, samples []collect.Sample) error {
	cw := csv.NewWriter(w)
	header := []string{
		"ts",
		"cpu_usage_pct", "cpu_user_pct", "cpu_system_pct", "cpu_iowait_pct", "cpu_idle_pct",
		"mem_total_mb", "mem_used_mb", "mem_available_mb", "mem_usage_pct",
		"mem_buffers_mb", "mem_cached_mb", "swap_used_mb", "swap_usage_pct",
	}
	coreN := 0
	if len(samples) > 0 {
		coreN = len(samples[0].Cores)
		for i := 0; i < coreN; i++ {
			header = append(header, fmt.Sprintf("core%d_usage_pct", i))
		}
	}
	if err := cw.Write(header); err != nil {
		return err
	}
	for _, s := range samples {
		cpu, mem := cpuOf(s), memOf(s)
		row := []string{
			time.UnixMilli(s.Timestamp).Format(time.RFC3339),
			f(cpu.Usage), f(cpu.User), f(cpu.System), f(cpu.IOWait), f(cpu.Idle),
			f(mem.TotalMB), f(mem.UsedMB), f(mem.AvailableMB), f(mem.UsagePct),
			f(mem.BuffersMB), f(mem.CachedMB), f(mem.SwapUsedMB), f(mem.SwapPct),
		}
		// 每核列固定 coreN 个：按索引查表，缺失补 0，保证列对齐。
		core := map[int]collect.CPUSample{}
		for i, c := range s.Cores {
			core[i] = c
		}
		for i := 0; i < coreN; i++ {
			if c, ok := core[i]; ok {
				row = append(row, f(c.Usage))
			} else {
				row = append(row, "0.000")
			}
		}
		if err := cw.Write(row); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}
