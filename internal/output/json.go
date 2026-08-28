package output

import (
	"encoding/json"
	"io"

	"sysmon/internal/analyze"
	"sysmon/internal/collect"
)

// Report 是 JSON 格式的完整报告：明细 + 汇总。
type Report struct {
	Samples []collect.Sample `json:"samples"`
	Summary analyze.Summary  `json:"summary"`
}

// WriteReportJSON 输出完整报告。
func WriteReportJSON(w io.Writer, samples []collect.Sample, s analyze.Summary) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(Report{Samples: samples, Summary: s})
}

// WriteSummaryJSON 只输出汇总（csv/text 格式的 sidecar 文件）。
func WriteSummaryJSON(w io.Writer, s analyze.Summary) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(s)
}
