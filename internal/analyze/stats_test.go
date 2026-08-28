package analyze

import (
	"math"
	"testing"
)

func TestStatsBasics(t *testing.T) {
	got := Stats([]float64{1, 2, 3, 4, 5}, 1)
	if math.Abs(got.Avg-3) > 1e-9 || math.Abs(got.Min-1) > 1e-9 || math.Abs(got.Max-5) > 1e-9 {
		t.Fatalf("bad basics: %+v", got)
	}
	if math.Abs(got.P50-3) > 1e-9 {
		t.Fatalf("p50 = %v, want 3", got.P50)
	}
	// 样本标准差 = sqrt(2.5)
	if math.Abs(got.StdDev-math.Sqrt(2.5)) > 1e-9 {
		t.Fatalf("stddev = %v, want %v", got.StdDev, math.Sqrt(2.5))
	}
	// 斜率 1/s → 每分钟 60
	if math.Abs(got.TrendPerMin-60) > 1e-9 {
		t.Fatalf("trend = %v, want 60", got.TrendPerMin)
	}
}

func TestPercentile(t *testing.T) {
	sorted := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	cases := []struct {
		p    float64
		want float64
	}{
		{50, 5.5},
		{90, 9.1},
		{100, 10},
		{0, 1},
	}
	for _, c := range cases {
		if got := percentile(sorted, c.p); math.Abs(got-c.want) > 1e-9 {
			t.Fatalf("p%v = %v, want %v", c.p, got, c.want)
		}
	}
	if got := percentile([]float64{42}, 90); got != 42 {
		t.Fatalf("single = %v, want 42", got)
	}
	if got := percentile(nil, 90); got != 0 {
		t.Fatalf("empty = %v, want 0", got)
	}
}

func TestStatsEmptyAndSingle(t *testing.T) {
	if got := Stats(nil, 1); got.Avg != 0 || got.P50 != 0 || got.TrendPerMin != 0 {
		t.Fatalf("empty stats not zero: %+v", got)
	}
	got := Stats([]float64{7}, 1)
	if got.Avg != 7 || got.Min != 7 || got.Max != 7 || got.P50 != 7 || got.StdDev != 0 {
		t.Fatalf("single stats wrong: %+v", got)
	}
}
