package collect

import "github.com/shirou/gopsutil/v3/cpu"

// fetchCPU 读取所有 CPU 的原始计数，并保证返回的 map 里有 "cpu" 聚合键。
// gopsutil 的 cpu.Times(true) 只返回每核数据（不含聚合行），因此聚合值由各核求和得到。
// 某些平台（如 darwin 无 cgo 交叉编译）gopsutil 不支持时降级到 fetchCPUFallback。
func fetchCPU() (map[string]cpuRaw, error) {
	stats, err := cpu.Times(true)
	if err == nil {
		return cpuRawFromTimes(stats), nil
	}
	if m, ferr := fetchCPUFallback(); ferr == nil {
		return m, nil
	}
	return nil, err
}

func cpuRawFromTimes(stats []cpu.TimesStat) map[string]cpuRaw {
	out := make(map[string]cpuRaw, len(stats)+1)
	var agg cpuRaw
	haveAgg := false
	for _, t := range stats {
		r := cpuRaw{
			user:   t.User,
			system: t.System + t.Irq + t.Softirq + t.Steal,
			idle:   t.Idle,
			iowait: t.Iowait,
			total:  t.User + t.Nice + t.System + t.Idle + t.Iowait +
				t.Irq + t.Softirq + t.Steal + t.Guest + t.GuestNice,
		}
		switch t.CPU {
		case "cpu", "cpu-total", "cpu-all", "all":
			if !haveAgg {
				out["cpu"] = r
				haveAgg = true
			}
			continue
		}
		out[t.CPU] = r
		agg.user += r.user
		agg.system += r.system
		agg.idle += r.idle
		agg.iowait += r.iowait
		agg.total += r.total
	}
	if !haveAgg {
		out["cpu"] = agg
	}
	return out
}
