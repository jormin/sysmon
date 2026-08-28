package collect

import (
	"encoding/binary"
	"testing"
)

func u32(v uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, v)
	return b
}

func concat(parts ...[]byte) []byte {
	var out []byte
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}

func TestParseDarwinCPTimes(t *testing.T) {
	// BSD CP_* 顺序：每核 [user, nice, system, idle]
	bsd := concat(u32(10), u32(2), u32(3), u32(85), u32(20), u32(4), u32(6), u32(170))
	m, err := parseDarwinCPTimes(bsd)
	if err != nil {
		t.Fatal(err)
	}
	if len(m) != 3 {
		t.Fatalf("want cpu+cpu0+cpu1, got %v", m)
	}
	if m["cpu"].total != 300 || m["cpu0"].user != 10 || m["cpu1"].idle != 170 {
		t.Fatalf("bad parse (bsd order): %+v", m)
	}
	// mach CPU_STATE 顺序：每核 [user, system, idle, nice]——数值语义应解析出一致结果
	mach := concat(u32(10), u32(3), u32(85), u32(2), u32(20), u32(6), u32(170), u32(4))
	m2, err := parseDarwinCPTimes(mach)
	if err != nil {
		t.Fatal(err)
	}
	if m2["cpu"].total != 300 || m2["cpu0"].user != 10 || m2["cpu1"].idle != 170 {
		t.Fatalf("bad parse (mach order): %+v", m2)
	}
	if _, err := parseDarwinCPTimes([]byte{1, 2, 3}); err == nil {
		t.Fatal("bad length must error")
	}
}
