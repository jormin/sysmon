package collect

// CPUSample 是一次 CPU 使用率采样，单位 %（0-100）。
type CPUSample struct {
	Usage  float64 `json:"usage_pct"`
	User   float64 `json:"user_pct"`
	System float64 `json:"system_pct"`
	IOWait float64 `json:"iowait_pct"`
	Idle   float64 `json:"idle_pct"`
}

// MemSample 是一次内存采样，容量单位 MB。
type MemSample struct {
	TotalMB     float64 `json:"total_mb"`
	UsedMB      float64 `json:"used_mb"`
	AvailableMB float64 `json:"available_mb"`
	BuffersMB   float64 `json:"buffers_mb"`
	CachedMB    float64 `json:"cached_mb"`
	UsagePct    float64 `json:"usage_pct"`
	SwapTotalMB float64 `json:"swap_total_mb"`
	SwapUsedMB  float64 `json:"swap_used_mb"`
	SwapPct     float64 `json:"swap_pct"`
}

// Sample 是一个完整采样点。
type Sample struct {
	Timestamp int64       `json:"ts"` // Unix 毫秒
	CPU       *CPUSample  `json:"cpu,omitempty"`
	Cores     []CPUSample `json:"cores,omitempty"`
	Mem       *MemSample  `json:"mem,omitempty"`
}
