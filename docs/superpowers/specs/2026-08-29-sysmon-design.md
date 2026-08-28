# sysmon — 系统监控采集工具设计（简化版）

日期：2026-08-29
状态：已批准（仅 CPU + 内存）

## 目标

用 Golang 开发工具，检测 MacBook Pro 实际使用过程中的 CPU 和内存使用量，生成可用于
WPS 表格查看与分析的数据。轻量：单二进制、低开销、可长期常驻。

## 范围（YAGNI）

- 不做：网络、磁盘、温度、serve 模式、Grafana/Prometheus（无需安装任何监控软件）。

## 用法

| 命令 | 说明 |
|---|---|
| `./sysmon -interval 5s -out mac.csv` | 每 5 秒一条，Ctrl+C 停止 |
| `./sysmon -interval 2s -duration 10m -out mac.csv` | 采 10 分钟自动停止 |
| `./sysmon -count 60 -cores -format json -out mac.json` | 60 条，JSON 明细 + 汇总 |

## 指标（gopsutil，跨平台）

- CPU：总体 + 每核（-cores）usage/user/system/iowait/idle（%）；
- 内存：total/used/available/buffers/cached/swap。

## 输出与分析

- 明细：CSV（默认）/ JSON / text；
- 汇总 `<out>.summary.json`：每指标 avg/min/max/p50/p90/p95/p99/stddev/trend_per_min（/分钟）；
- macOS 无 cgo（交叉编译）时 CPU 走 sysctl kern.cp_times 兜底。

## 结构

```
main.go             CLI + 信号 + 输出分发
internal/collect/   采样核心（含 darwin 兜底）
internal/analyze/   统计汇总
internal/output/    csv/json/text
README.md           一页使用文档（含 WPS 方法）
```

## 测试

统计/速率/CSV 列对齐单测 + 本机实跑 + macOS 交叉编译产物验证。
