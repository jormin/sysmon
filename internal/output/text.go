package output

import (
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	"sysmon/internal/analyze"
	"sysmon/internal/collect"
)

// WriteText 输出明细行 + 汇总表格。
func WriteText(w io.Writer, samples []collect.Sample, s analyze.Summary) error {
	for _, sm := range samples {
		cpu, mem := cpuOf(sm), memOf(sm)
		fmt.Fprintf(w, "%s  cpu %5.1f%%  mem %5.1f%%\n",
			time.UnixMilli(sm.Timestamp).Format("15:04:05"),
			cpu.Usage, mem.UsagePct)
	}
	fmt.Fprintln(w)
	WriteSummaryText(w, s)
	return nil
}

// WriteSummaryText 输出汇总表格。
func WriteSummaryText(w io.Writer, s analyze.Summary) {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	fmt.Fprintf(tw, "采样 %d 条, 间隔 %.0fs, 持续 %.0fs\n", s.Samples, s.IntervalSec, s.DurationSec)
	row := func(name string, m analyze.MetricStats) {
		fmt.Fprintf(tw, "%s\tavg %.3f\tmin %.3f\tmax %.3f\tp50 %.3f\tp90 %.3f\tp95 %.3f\tp99 %.3f\ttrend %.3f/min\n",
			name, m.Avg, m.Min, m.Max, m.P50, m.P90, m.P95, m.P99, m.TrendPerMin)
	}
	if s.CPU != nil {
		fmt.Fprintln(tw, "\nCPU (%)")
		row("usage", s.CPU.Usage)
		row("user ", s.CPU.User)
		row("sys  ", s.CPU.System)
		row("iowait", s.CPU.IOWait)
		for k, v := range s.CPU.PerCore {
			row(k, v)
		}
	}
	if s.Mem != nil {
		fmt.Fprintln(tw, "\n内存")
		row("usage% ", s.Mem.UsagePct)
		row("usedMB ", s.Mem.UsedMB)
		row("availMB", s.Mem.AvailableMB)
		row("swap%  ", s.Mem.SwapPct)
	}
	tw.Flush()
}
