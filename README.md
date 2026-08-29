# sysmon

轻量系统监控工具（Go + gopsutil）：采集 **CPU** 和 **内存** 使用量，输出明细数据与汇总统计。
专为 MacBook Pro 实际使用场景设计，用 WPS 表格即可查看曲线。

## 获取

在任意机器上交叉编译出 macOS 可执行文件（或直接在 Mac 上 `go build -o sysmon .`）：

```bash
# Apple Silicon (M1/M2/M3...)
GOOS=darwin GOARCH=arm64 go build -o sysmon_mac_arm64 .
# Intel MacBook Pro
GOOS=darwin GOARCH=amd64 go build -o sysmon_mac_amd64 .
```

把二进制拷到 Mac，终端运行：`./sysmon_mac_arm64 -interval 5s -out mac.csv`。

## 用法

```bash
# 每 5 秒一条，直到 Ctrl+C，输出 mac.csv + mac.csv.summary.json
./sysmon -interval 5s -out mac.csv

# 采 10 分钟自动停止
./sysmon -interval 2s -duration 10m -out mac.csv

# 每核 CPU + 内存，JSON 明细
./sysmon -interval 2s -count 60 -cores -format json -out mac.json

# 只采内存
./sysmon -interval 5s -cpu=false -out mem_only.csv
```

参数：`-interval`（默认 1s）、`-duration`（默认 0=直到 Ctrl+C）、`-count`、`-out`（默认 sysmon.csv，'-'=stdout）、`-format`（csv|json|text）、`-quiet`、`-cpu`（默认开）、`-cores`（每核，默认关）、`-mem`（默认开）、`-version`（打印版本号）。

## 汇总分析

每次采样结束自动生成 `<输出名>.summary.json`（csv/text 格式时；json 格式内嵌在报告里）。
每个指标含：`avg / min / max / p50 / p90 / p95 / p99 / stddev / trend_per_min`（趋势为最小二乘斜率，单位/分钟）。

## 用 WPS 表格查看 CSV

1. 双击 CSV 用 WPS 表格打开（乱码时选 UTF-8 编码、逗号分隔）；
2. 按住 Ctrl 选中「ts」列和「cpu_usage_pct」列 → 插入 → 图表 → 折线图；
3. 再加选「mem_usage_pct」列即可同图对比 CPU/内存曲线；
4. 现成统计看 `mac.csv.summary.json`，或 WPS 里用 =AVERAGE()、=MAX() 现算。

## 说明

- 跨平台（Linux/macOS/Windows 均可编译运行）；macOS 上即使交叉编译（无 cgo）也能采集 CPU
  （自动降级 sysctl kern.cp_times），本地构建则走 gopsutil 原生路径；
- CPU/内存任一子系统读取失败只警告一次、跳过该次采样，不中断整体运行；
- 默认只做 CPU + 内存；网络/磁盘/温度、Grafana/Prometheus 导出不在本期范围。
