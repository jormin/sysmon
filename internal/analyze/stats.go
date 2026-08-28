package analyze

import (
	"math"
	"sort"
)

// MetricStats 是单个指标跨采样的统计汇总。
type MetricStats struct {
	Avg         float64 `json:"avg"`
	Min         float64 `json:"min"`
	Max         float64 `json:"max"`
	P50         float64 `json:"p50"`
	P90         float64 `json:"p90"`
	P95         float64 `json:"p95"`
	P99         float64 `json:"p99"`
	StdDev      float64 `json:"stddev"`
	TrendPerMin float64 `json:"trend_per_min"`
}

// Stats 计算统计汇总；dt 是采样间隔（秒），用于趋势的斜率。
func Stats(vals []float64, dt float64) MetricStats {
	if len(vals) == 0 {
		return MetricStats{}
	}
	sum, min, max := 0.0, vals[0], vals[0]
	for _, v := range vals {
		sum += v
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	mean := sum / float64(len(vals))
	var ss float64
	for _, v := range vals {
		ss += (v - mean) * (v - mean)
	}
	stddev := 0.0
	if len(vals) > 1 {
		stddev = math.Sqrt(ss / float64(len(vals)-1))
	}
	sorted := append([]float64(nil), vals...)
	sort.Float64s(sorted)
	return MetricStats{
		Avg: mean, Min: min, Max: max,
		P50: percentile(sorted, 50), P90: percentile(sorted, 90),
		P95: percentile(sorted, 95), P99: percentile(sorted, 99),
		StdDev: stddev, TrendPerMin: slope(vals, dt) * 60,
	}
}

// percentile 用最近秩线性插值计算分位数（p 为 0-100）。
func percentile(sorted []float64, p float64) float64 {
	n := len(sorted)
	if n == 0 {
		return 0
	}
	if n == 1 {
		return sorted[0]
	}
	rank := p / 100 * float64(n-1)
	lo := int(math.Floor(rank))
	hi := int(math.Ceil(rank))
	if lo == hi {
		return sorted[lo]
	}
	return sorted[lo] + (sorted[hi]-sorted[lo])*(rank-float64(lo))
}

// slope 返回最小二乘线性拟合的斜率（每秒）。
func slope(vals []float64, dt float64) float64 {
	n := len(vals)
	if n < 2 || dt <= 0 {
		return 0
	}
	var sx, sy, sxy, sxx float64
	for i, v := range vals {
		x := float64(i) * dt
		sx += x
		sy += v
		sxy += x * v
		sxx += x * x
	}
	denom := float64(n)*sxx - sx*sx
	if denom == 0 {
		return 0
	}
	return (float64(n)*sxy - sx*sy) / denom
}
