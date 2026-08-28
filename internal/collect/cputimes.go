package collect

import (
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
)

// parseDarwinCPTimes 解析 macOS sysctl kern.cp_times 的原始字节。
// 每核 4 个 uint32（小端）。不同资料对状态顺序描述不一
// （BSD CP_*: user,nice,system,idle；mach CPU_STATE: user,system,idle,nice），
// 因此运行时按数值语义识别：user 固定为第 1 个；idle 取剩余三态中最大，
// nice 取最小，system 取中间。累计时间上 idle 必然占绝对多数，该识别稳定可靠，
// 保证 CPU 使用率（100-idle%）不受顺序争议影响。
func parseDarwinCPTimes(buf []byte) (map[string]cpuRaw, error) {
	if len(buf)%16 != 0 {
		return nil, fmt.Errorf("kern.cp_times 长度 %d 不是 16 的倍数", len(buf))
	}
	n := len(buf) / 16
	out := make(map[string]cpuRaw, n+1)
	var tUser, tNice, tSystem, tIdle float64
	for i := 0; i < n; i++ {
		off := i * 16
		user := float64(binary.LittleEndian.Uint32(buf[off : off+4]))
		v1 := float64(binary.LittleEndian.Uint32(buf[off+4 : off+8]))
		v2 := float64(binary.LittleEndian.Uint32(buf[off+8 : off+12]))
		v3 := float64(binary.LittleEndian.Uint32(buf[off+12 : off+16]))
		idle := math.Max(v1, math.Max(v2, v3))
		nice := math.Min(v1, math.Min(v2, v3))
		system := v1 + v2 + v3 - idle - nice
		total := user + nice + system + idle
		out["cpu"+strconv.Itoa(i)] = cpuRaw{user: user, system: system, idle: idle, total: total}
		tUser += user
		tNice += nice
		tSystem += system
		tIdle += idle
	}
	out["cpu"] = cpuRaw{user: tUser, system: tSystem, idle: tIdle, total: tUser + tNice + tSystem + tIdle}
	return out, nil
}
