package output

import (
	"encoding/csv"
	"strings"
	"testing"

	"sysmon/internal/collect"
)

// s1 1 核、s2 2 核：第二条采样核数更多，但表头以第一条为准，行必须保持列数一致。
func TestWriteCSVColumnAlignment(t *testing.T) {
	s1 := collect.Sample{CPU: &collect.CPUSample{Usage: 10, User: 2, System: 3, Idle: 90},
		Cores: []collect.CPUSample{{Usage: 10}},
		Mem:   &collect.MemSample{UsedMB: 1000, AvailableMB: 1000, UsagePct: 50, BuffersMB: 1, CachedMB: 2, SwapUsedMB: 3, SwapPct: 4}}
	s2 := collect.Sample{CPU: &collect.CPUSample{Usage: 20, User: 3, System: 4, Idle: 80},
		Cores: []collect.CPUSample{{Usage: 10}, {Usage: 20}},
		Mem:   &collect.MemSample{UsedMB: 1200, AvailableMB: 800, UsagePct: 60, BuffersMB: 1, CachedMB: 2, SwapUsedMB: 3, SwapPct: 4}}

	var sb strings.Builder
	if err := WriteCSV(&sb, []collect.Sample{s1, s2}); err != nil {
		t.Fatal(err)
	}
	r := csv.NewReader(strings.NewReader(sb.String()))
	rows, err := r.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 {
		t.Fatalf("want header+2 rows, got %d", len(rows))
	}
	n := len(rows[0])
	for i, row := range rows {
		if len(row) != n {
			t.Fatalf("row %d has %d cols, want %d (row: %v)", i, len(row), n, row)
		}
	}
	header := strings.Join(rows[0], ",")
	for _, want := range []string{"ts", "cpu_usage_pct", "mem_usage_pct", "core0_usage_pct"} {
		if !strings.Contains(header, want) {
			t.Fatalf("header missing %q: %s", want, header)
		}
	}
	if strings.Contains(header, "core1_usage_pct") {
		t.Fatalf("header must NOT contain core1 (column set from first sample): %s", header)
	}
	// core0 列位置从表头动态查找（前 14 列为固定指标），行 2 的 core0 = 10
	core0Idx := -1
	for i, h := range rows[0] {
		if h == "core0_usage_pct" {
			core0Idx = i
			break
		}
	}
	if core0Idx < 0 {
		t.Fatalf("header missing core0_usage_pct: %v", rows[0])
	}
	if rows[2][core0Idx] != "10.000" {
		t.Fatalf("row2 core0 = %q, want 10.000", rows[2][core0Idx])
	}
}
