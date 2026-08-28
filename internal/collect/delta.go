package collect

// cpuRaw 是相邻两次 cpu.TimesStat 之间做差所需的原始计数（秒）。
type cpuRaw struct {
	user, system, idle, iowait float64
	total                      float64
}

// cpuDelta 计算两次原始计数之间的 CPU 百分比。elapsed 为秒；计数回退或间隔非法时返回零值。
func cpuDelta(prev, cur cpuRaw, elapsed float64) CPUSample {
	var out CPUSample
	if elapsed <= 0 || cur.total <= prev.total {
		return out
	}
	d := cur.total - prev.total
	if d == 0 {
		return out
	}
	out.User = (cur.user - prev.user) / d * 100
	out.System = (cur.system - prev.system) / d * 100
	out.IOWait = (cur.iowait - prev.iowait) / d * 100
	out.Idle = (cur.idle - prev.idle) / d * 100
	out.Usage = 100 - out.Idle
	return out
}
