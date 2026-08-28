package collect

import (
	"math"
	"testing"
)

func TestCPUDelta(t *testing.T) {
	prev := cpuRaw{user: 10, system: 5, idle: 85, iowait: 0, total: 100}
	cur := cpuRaw{user: 20, system: 10, idle: 170, iowait: 0, total: 200}
	got := cpuDelta(prev, cur, 1)
	if math.Abs(got.Usage-15) > 1e-9 {
		t.Fatalf("usage = %v, want 15", got.Usage)
	}
	if math.Abs(got.User-10) > 1e-9 || math.Abs(got.System-5) > 1e-9 || math.Abs(got.Idle-85) > 1e-9 {
		t.Fatalf("bad split: %+v", got)
	}
}

func TestCPUDeltaZeroElapsed(t *testing.T) {
	got := cpuDelta(cpuRaw{total: 100}, cpuRaw{total: 200}, 0)
	if got.Usage != 0 {
		t.Fatalf("usage = %v, want 0", got.Usage)
	}
}

func TestCPUDeltaCounterReset(t *testing.T) {
	got := cpuDelta(cpuRaw{total: 200}, cpuRaw{total: 100}, 1)
	if got.Usage != 0 {
		t.Fatalf("usage = %v, want 0 on reset", got.Usage)
	}
}
