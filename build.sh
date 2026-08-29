#!/usr/bin/env bash
#
# sysmon 构建脚本
#
# 用法:
#   ./build.sh [版本]             构建当前宿主平台的发布产物(VERSION 默认 dev)
#   ./build.sh [版本] all         构建完整发布矩阵: darwin/amd64 darwin/arm64 linux/amd64 linux/arm64
#   ./build.sh [版本] os/arch...  构建指定目标, 如 ./build.sh v0.0.4 darwin/arm64 linux/amd64
#
# 说明:
#   - darwin(macOS) 产物必须在 macOS 主机(Xcode/CommandLineTools, 含 clang)上以 cgo 构建:
#     gopsutil 在 macOS 上采集 CPU 依赖 cgo, 无 cgo 的交叉编译产物 CPU 不可用
#     (运行时报 "cpu 不可用: not implemented yet")。脚本在非 macOS 宿主上构建 darwin 会直接报错。
#   - linux 产物不依赖 cgo, 可在任意宿主机(含 macOS)交叉编译。
set -euo pipefail

VERSION="${1:-dev}"
if [ "$#" -gt 0 ]; then shift; fi

HOST_OS=$(uname -s)
case "$HOST_OS" in
	Darwin) HOST_OS=darwin ;;
	Linux)  HOST_OS=linux ;;
	*) echo "错误: 不支持的宿主系统 $HOST_OS(仅支持 macOS/Linux)" >&2; exit 1 ;;
esac

HOST_ARCH=$(uname -m)
case "$HOST_ARCH" in
	x86_64|amd64) HOST_ARCH=amd64 ;;
	arm64|aarch64) HOST_ARCH=arm64 ;;
	*) echo "错误: 不支持的宿主架构 $HOST_ARCH" >&2; exit 1 ;;
esac

ALL_TARGETS="darwin/amd64 darwin/arm64 linux/amd64 linux/arm64"

if [ "$#" -eq 0 ]; then
	TARGETS="${HOST_OS}/${HOST_ARCH}"
elif [ "$1" = "all" ]; then
	TARGETS="$ALL_TARGETS"
else
	TARGETS="$*"
fi

for t in $TARGETS; do
	case "$t" in
		darwin/amd64|darwin/arm64|linux/amd64|linux/arm64) ;;
		*) echo "错误: 未知目标 $t(支持: darwin/amd64 darwin/arm64 linux/amd64 linux/arm64)" >&2; exit 1 ;;
	esac
done

rm -rf target && mkdir -p target
LDFLAGS="-X main.version=${VERSION}"

build_one() {
	local os="$1" arch="$2"
	local out="target/sysmon_${os}_${arch}_${VERSION}"

	if [ "$os" = "darwin" ]; then
		if [ "$HOST_OS" != "darwin" ]; then
			echo "错误: darwin 产物需要 cgo(CPU 采集依赖), 只能在 macOS 主机上构建(当前宿主: $HOST_OS)" >&2
			echo "      请在 Mac 本机或 CI 的 macos runner 上执行; linux 目标不受影响" >&2
			exit 1
		fi
		echo "==> 构建 darwin/$arch (原生 cgo, CPU 可用)"
		CGO_ENABLED=1 GOOS=darwin GOARCH="$arch" go build -ldflags "${LDFLAGS}" -o "$out" .
	else
		echo "==> 构建 $os/$arch (无 cgo 交叉编译)"
		CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build -ldflags "${LDFLAGS}" -o "$out" .
	fi
}

for t in $TARGETS; do
	build_one "${t%/*}" "${t#*/}"
done

echo
echo "==> 构建完成, 产物:"
ls -lh target/