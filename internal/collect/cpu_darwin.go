//go:build darwin
// +build darwin

package collect

import "golang.org/x/sys/unix"

// fetchCPUFallback 在 darwin 无 cgo 时用 sysctl kern.cp_times 读取 CPU 计数。
func fetchCPUFallback() (map[string]cpuRaw, error) {
	buf, err := unix.SysctlRaw("kern.cp_times")
	if err != nil {
		return nil, err
	}
	return parseDarwinCPTimes(buf)
}
