package collect

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Options 选择要采集的子系统。
type Options struct {
	CPU, PerCore, Mem bool
}

// Collector 按间隔采样系统指标。必须先调用 Baseline 再调用 Sample。
type Collector struct {
	opts    Options
	last    time.Time
	prevCPU map[string]cpuRaw
	warned  map[string]bool
}

// New 创建采集器。
func New(opts Options) *Collector {
	return &Collector{opts: opts, warned: map[string]bool{}}
}

// warn 每个子系统只打印一次警告，避免刷屏。
func (c *Collector) warn(subsys string, err error) {
	if c.warned[subsys] {
		return
	}
	c.warned[subsys] = true
	fmt.Fprintf(os.Stderr, "sysmon: warning: %s 不可用: %v\n", subsys, err)
}

// Baseline 抓取初始快照，使第一次 Sample 就有有效速率。
func (c *Collector) Baseline() {
	c.last = time.Now()
	if c.opts.CPU {
		if m, err := fetchCPU(); err == nil {
			c.prevCPU = m
		} else {
			c.warn("cpu", err)
		}
	}
}

// Sample 采集一个采样点；now 为采样时刻。
func (c *Collector) Sample(now time.Time) Sample {
	elapsed := now.Sub(c.last).Seconds()
	c.last = now
	s := Sample{Timestamp: now.UnixMilli()}

	if c.opts.CPU {
		if m, err := fetchCPU(); err == nil {
			if c.prevCPU != nil {
				agg := cpuDelta(c.prevCPU["cpu"], m["cpu"], elapsed)
				s.CPU = &agg
				if c.opts.PerCore {
					for _, n := range sortedCPUKeys(m) {
						if n == "cpu" {
							continue
						}
						s.Cores = append(s.Cores, cpuDelta(c.prevCPU[n], m[n], elapsed))
					}
				}
			}
			c.prevCPU = m
		} else {
			c.warn("cpu", err)
		}
	}

	if c.opts.Mem {
		if m, err := fetchMem(); err == nil {
			s.Mem = m
		} else {
			c.warn("mem", err)
		}
	}

	return s
}

func sortedCPUKeys(m map[string]cpuRaw) []string {
	names := make([]string, 0, len(m))
	for k := range m {
		names = append(names, k)
	}
	sort.Slice(names, func(i, j int) bool {
		ai, aj := cpuIndex(names[i]), cpuIndex(names[j])
		if ai != aj {
			return ai < aj
		}
		return names[i] < names[j]
	})
	return names
}

// cpuIndex 把 "cpu"→-1、"cpuN"→N，保证 cpu0…cpu10 数字序。
func cpuIndex(name string) int {
	if name == "cpu" {
		return -1
	}
	v, err := strconv.Atoi(strings.TrimPrefix(name, "cpu"))
	if err != nil {
		return 0
	}
	return v
}
