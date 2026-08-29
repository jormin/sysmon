#!/bin/sh
#
# sysmon 一键安装脚本
#
# 用法:
#   curl -fsSL https://raw.githubusercontent.com/jormin/sysmon/master/install.sh | sh
#   或下载后本地执行: sh install.sh [选项]
#
# 自动识别操作系统(macOS/Linux)与 CPU 架构(amd64/arm64),
# 从 GitHub Releases 下载对应二进制安装到 /usr/local/bin
# (安装目录不可写时自动通过 sudo 提示授权)。
#
# 可选环境变量:
#   SYSMON_VERSION     指定版本, 如 v0.0.1(默认自动获取最新 release)
#   SYSMON_OS          手动指定系统: darwin|linux
#   SYSMON_ARCH        手动指定架构: amd64|arm64
#   SYSMON_INSTALL_DIR 安装目录(默认 /usr/local/bin)
#   SYSMON_REPO        发布仓库, 如 jormin/sysmon(默认 jormin/sysmon)
#   SYSMON_BASE_URL    下载基础地址(默认 https://github.com/<repo>/releases/download, 一般无需设置)
#   SYSMON_CHECKSUM    可选: 期望的 SHA-256 校验和(64 位 hex)

set -eu

REPO="${SYSMON_REPO:-jormin/sysmon}"
INSTALL_DIR="${SYSMON_INSTALL_DIR:-/usr/local/bin}"
BASE_URL="${SYSMON_BASE_URL:-https://github.com/${REPO}/releases/download}"

info() { printf '%s\n' "$*"; }
warn() { printf '警告: %s\n' "$*" >&2; }
die()  { printf '错误: %s\n' "$*" >&2; exit 1; }

have_cmd() { command -v "$1" >/dev/null 2>&1; }

usage() {
	cat <<'EOF'
sysmon 一键安装脚本

用法:
  curl -fsSL https://raw.githubusercontent.com/jormin/sysmon/master/install.sh | sh
  sh install.sh [选项]

选项:
  -h, --help       显示本帮助
  -u, --uninstall  卸载 sysmon(目录不可写时自动使用 sudo)

环境变量:
  SYSMON_VERSION      指定版本, 如 v0.0.1(默认最新 release)
  SYSMON_INSTALL_DIR  安装目录(默认 /usr/local/bin)
  SYSMON_OS           手动指定系统: darwin|linux
  SYSMON_ARCH         手动指定架构: amd64|arm64
  SYSMON_CHECKSUM     可选 SHA-256 校验和(64 位 hex)
EOF
}

# 检测操作系统: macOS -> darwin, Linux -> linux, 其余不支持
detect_os() {
	if [ -n "${SYSMON_OS:-}" ]; then
		os="$SYSMON_OS"
	else
		os=$(uname -s 2>/dev/null || echo unknown)
		case "$os" in
			Darwin) os=darwin ;;
			Linux)  os=linux ;;
			*) die "不支持的操作系统: $os(仅支持 macOS/Linux)" ;;
		esac
	fi
	case "$os" in
		darwin|linux) ;;
		*) die "SYSMON_OS 仅支持 darwin|linux, 当前值: $os" ;;
	esac
	printf '%s\n' "$os"
}

# 检测 CPU 架构: x86_64 -> amd64, arm64/aarch64 -> arm64
detect_arch() {
	if [ -n "${SYSMON_ARCH:-}" ]; then
		arch="$SYSMON_ARCH"
	else
		arch=$(uname -m 2>/dev/null || echo unknown)
		case "$arch" in
			x86_64|amd64|x64)    arch=amd64 ;;
			arm64|aarch64|armv8l) arch=arm64 ;;
			*) die "不支持的 CPU 架构: $arch(仅支持 amd64/arm64)" ;;
		esac
		# Apple Silicon 在 Rosetta 下 uname -m 返回 x86_64, 用 sysctl 纠正为 arm64
		if [ "$arch" = amd64 ] && [ "$(uname -s)" = Darwin ] && have_cmd sysctl \
			&& [ "$(sysctl -n hw.optional.arm64 2>/dev/null)" = 1 ]; then
			arch=arm64
		fi
	fi
	case "$arch" in
		amd64|arm64) ;;
		*) die "SYSMON_ARCH 仅支持 amd64|arm64, 当前值: $arch" ;;
	esac
	printf '%s\n' "$arch"
}

# 获取最新 release 的版本号, 返回形如 v0.0.1
latest_release_version() {
	# 优先用 GitHub API 获取最新 release 的 tag_name
	if have_cmd curl; then
		resp=$(curl -fsSL --max-time 20 "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null || true)
		v=$(printf '%s\n' "$resp" | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p')
		[ -n "$v" ] && { printf '%s\n' "$v"; return; }
	fi
	# 回退: 利用 /releases/latest 的 302 跳转地址解析版本(不消耗 API 配额)
	if have_cmd curl; then
		loc=$(curl -s -o /dev/null --max-time 20 -w '%{redirect_url}' "https://github.com/${REPO}/releases/latest" 2>/dev/null || true)
		v=$(printf '%s\n' "$loc" | sed -n 's#.*/tag/\(v[^/]*\)$#\1#p')
		[ -n "$v" ] && { printf '%s\n' "$v"; return; }
	fi
	die "无法获取最新版本号, 请稍后重试或设置 SYSMON_VERSION=vX.Y.Z 指定版本"
}

# 解析最终安装版本, 统一补 v 前缀
resolve_version() {
	if [ -n "${SYSMON_VERSION:-}" ]; then
		version="$SYSMON_VERSION"
	else
		version=$(latest_release_version)
	fi
	case "$version" in
		v*) ;;
		*) version="v${version}" ;;
	esac
	printf '%s\n' "$version"
}

# 下载文件: 优先 curl, 其次 wget
download() {
	url=$1
	dest=$2
	if have_cmd curl; then
		curl -fsSL --max-time 120 -o "$dest" "$url"
	elif have_cmd wget; then
		wget -q -O "$dest" "$url"
	else
		die "未找到 curl 或 wget, 无法下载"
	fi
}

# 校验下载内容确实是 ELF(Linux)或 Mach-O(macOS)可执行文件,
# 避免 HTML 错误页/代理干扰被当成二进制安装
check_binary() {
	f=$1
	[ -s "$f" ] || return 1
	if have_cmd file; then
		kind=$(file -b "$f")
		case "$kind" in
			*ELF*|*Mach-O*) return 0 ;;
			*) return 1 ;;
		esac
	fi
	magic=$(head -c 4 "$f" 2>/dev/null | od -An -tx1 2>/dev/null | tr -d ' \n') || return 1
	case "$magic" in
		7f454c46|cffaedfe|cefaedfe) return 0 ;; # ELF | Mach-O 64(小端/大端)
		*) return 1 ;;
	esac
}

# 可选: 校验 SHA-256
verify_checksum() {
	[ -n "${SYSMON_CHECKSUM:-}" ] || return 0
	if have_cmd sha256sum; then
		actual=$(sha256sum "$1" | awk '{print $1}')
	elif have_cmd shasum; then
		actual=$(shasum -a 256 "$1" | awk '{print $1}')
	else
		warn "未找到 sha256sum/shasum, 跳过校验和验证"
		return 0
	fi
	[ "$actual" = "$SYSMON_CHECKSUM" ] || die "校验和不匹配: 期望 $SYSMON_CHECKSUM, 实际 $actual"
	info "==> 校验和验证通过 ($actual)"
}

# 安装到目标目录, 不可写时自动用 sudo
install_to() {
	src=$1
	dest_dir=$2
	if [ ! -d "$dest_dir" ]; then
		if ! mkdir -p "$dest_dir" 2>/dev/null; then
			have_cmd sudo || die "无法创建安装目录 $dest_dir(且无 sudo)"
			sudo mkdir -p "$dest_dir" || die "无法创建安装目录 $dest_dir"
		fi
	fi
	if [ -w "$dest_dir" ]; then
		if have_cmd install; then
			install -m 0755 "$src" "${dest_dir}/sysmon"
		else
			cp "$src" "${dest_dir}/sysmon"
			chmod 0755 "${dest_dir}/sysmon"
		fi
	elif have_cmd sudo; then
		info "安装目录 $dest_dir 不可写, 使用 sudo 安装..."
		sudo install -m 0755 "$src" "${dest_dir}/sysmon"
	else
		die "安装目录 $dest_dir 不可写且没有 sudo; 可用 SYSMON_INSTALL_DIR 指定可写目录"
	fi
}

# 安装目录不在 PATH 中时提醒
check_path() {
	dest_dir=$1
	case ":${PATH}:" in
		*":${dest_dir}:"*) return 0 ;;
	esac
	warn "$dest_dir 不在 PATH 中, 使用时请写全路径 ${dest_dir}/sysmon 或将其加入 PATH"
}

# 卸载
uninstall() {
	target="${INSTALL_DIR}/sysmon"
	if [ ! -e "$target" ] && [ ! -L "$target" ]; then
		info "未找到 $target, 无需卸载"
		return 0
	fi
	if [ -w "$INSTALL_DIR" ]; then
		rm -f "$target"
	elif have_cmd sudo; then
		sudo rm -f "$target"
	else
		die "无法删除 $target: 目录不可写且没有 sudo"
	fi
	info "==> 已卸载 $target"
}

main() {
	case "${1:-}" in
		'') ;;
		-h|--help) usage; exit 0 ;;
		-u|--uninstall) uninstall; exit 0 ;;
		*) usage; exit 2 ;;
	esac

	os=$(detect_os)
	arch=$(detect_arch)
	version=$(resolve_version)
	case "$os" in
		darwin) os_name=macOS ;;
		linux)  os_name=Linux ;;
	esac

	asset="sysmon_${os}_${arch}_${version}"
	url="${BASE_URL}/${version}/${asset}"

	info "==> 检测到 ${os_name} / ${arch}, 安装 sysmon ${version}"
	info "==> 下载: ${url}"

	tmpdir=$(mktemp -d "${TMPDIR:-/tmp}/sysmon-install.XXXXXX") || die "创建临时目录失败"
	trap 'rm -rf "${tmpdir}"' EXIT HUP INT TERM
	tmpfile="${tmpdir}/sysmon"

	download "$url" "$tmpfile" || die "下载失败: ${url}(可用 SYSMON_VERSION 指定其他版本)"
	check_binary "$tmpfile" || die "下载内容不是有效的可执行文件, 可能该版本缺少 ${os}/${arch} 产物或网络被代理干扰: ${url}"
	verify_checksum "$tmpfile"

	install_to "$tmpfile" "$INSTALL_DIR"
	check_path "$INSTALL_DIR"

	installed=$("${INSTALL_DIR}/sysmon" -version 2>/dev/null || true)
	if [ -n "$installed" ]; then
		info "==> 安装完成: ${INSTALL_DIR}/sysmon (${installed})"
		if [ "$installed" != "$version" ]; then
			warn "安装的版本(${installed})与期望(${version})不一致"
		fi
	else
		warn "安装完成, 但版本自检失败, 请手动运行 ${INSTALL_DIR}/sysmon -version 检查"
	fi
	info "==> 使用: sysmon -interval 5s -out data.csv"
}

main "$@"