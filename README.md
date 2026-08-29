# sysmon

轻量系统监控工具（Go + gopsutil）：采集 **CPU** 和 **内存** 使用量，输出明细数据与汇总统计。
专为 MacBook Pro 实际使用场景设计，用 WPS 表格即可查看曲线。

## 获取

**推荐在 Mac 本机构建**（macOS 的 CPU 采集依赖 cgo，本机 `go build` 默认开启）：

```bash
go build -o sysmon .
./sysmon -interval 5s -out mac.csv
```

也可以直接使用一键安装脚本（见下），安装 GitHub Releases 上由 CI 在 macOS runner 上原生构建的版本。

> ⚠️ **macOS 的 CPU 采集依赖 cgo**（gopsutil）：在非 macOS 主机上用 `CGO_ENABLED=0` 交叉编译出的
> darwin 二进制 CPU 不可用，运行时报 `sysmon: warning: cpu 不可用: not implemented yet`。
> Linux 目标不依赖 cgo，可在任意主机交叉编译：
>
> ```bash
> # Linux x86_64 / ARM64（任意宿主均可）
> GOOS=linux GOARCH=amd64 go build -o sysmon_linux_amd64 .
> GOOS=linux GOARCH=arm64 go build -o sysmon_linux_arm64 .
> ```

使用仓库自带的构建脚本（darwin 目标会强制在 macOS 宿主上以 cgo 构建，避免产出不可用的二进制）：

```bash
./build.sh                     # 构建当前宿主平台
./build.sh v0.0.4 all          # 完整矩阵: darwin/amd64 darwin/arm64 linux/amd64 linux/arm64
./build.sh v0.0.4 darwin/arm64 # 指定目标(darwin 仅限 macOS 宿主)
```

### 一键安装（macOS / Linux）

自动识别操作系统与 CPU 架构，从 GitHub Releases 下载对应版本安装到 `/usr/local/bin`（目录不可写时自动用 `sudo` 提示授权）：

```bash
curl -fsSL https://raw.githubusercontent.com/jormin/sysmon/master/install.sh | sh
```

常用方式：

```bash
# 指定版本安装（默认安装最新 release）
curl -fsSL https://raw.githubusercontent.com/jormin/sysmon/master/install.sh | SYSMON_VERSION=v0.0.1 sh

# 安装到用户目录（无需 sudo）
curl -fsSL https://raw.githubusercontent.com/jormin/sysmon/master/install.sh | SYSMON_INSTALL_DIR="$HOME/.local/bin" sh

# 卸载
curl -fsSL https://raw.githubusercontent.com/jormin/sysmon/master/install.sh | sh -s -- --uninstall
```

环境变量：`SYSMON_VERSION`（版本）、`SYSMON_INSTALL_DIR`（安装目录）、`SYSMON_OS` / `SYSMON_ARCH`（手动指定系统/架构，一般无需设置）、`SYSMON_CHECKSUM`（可选 SHA-256 校验）。

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

## 升级

```bash
sysmon upgrade         # 检测 GitHub Releases 最新版, 确认后自动替换自身二进制
sysmon upgrade -check  # 只检测, 不更新
```

- 自动识别当前平台（仅 darwin/linux × amd64/arm64 有预编译产物），下载替换并保留 `.old` 备份，失败自动回滚；
- 环境变量 `SYSMON_REPO` 可覆盖发布仓库（如镜像），`SYSMON_UPGRADE_BASE` 可覆盖下载基址（默认 `https://github.com`）。

## 汇总分析

每次采样结束自动生成 `<输出名>.summary.json`（csv/text 格式时；json 格式内嵌在报告里）。
每个指标含：`avg / min / max / p50 / p90 / p95 / p99 / stddev / trend_per_min`（趋势为最小二乘斜率，单位/分钟）。

## 用 WPS 表格查看 CSV

1. 双击 CSV 用 WPS 表格打开（乱码时选 UTF-8 编码、逗号分隔）；
2. 按住 Ctrl 选中「ts」列和「cpu_usage_pct」列 → 插入 → 图表 → 折线图；
3. 再加选「mem_usage_pct」列即可同图对比 CPU/内存曲线；
4. 现成统计看 `mac.csv.summary.json`，或 WPS 里用 =AVERAGE()、=MAX() 现算。

## 说明

- 跨平台（Linux/macOS/Windows 均可编译运行）；macOS 的 CPU 采集依赖 cgo，Release 产物由 CI 在
  macOS runner 上原生构建（`build.sh` 会在非 macOS 宿主上拒绝构建 darwin 目标）；
- CPU/内存任一子系统读取失败只警告一次、跳过该次采样，不中断整体运行；
- 默认只做 CPU + 内存；网络/磁盘/温度、Grafana/Prometheus 导出不在本期范围。
