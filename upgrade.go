package main

// upgrade.go: sysmon upgrade 子命令, 参照 dsh-cli 的 upgrade 实现。
// 检测 GitHub Releases 最新版本, 下载当前平台(darwin/linux × amd64/arm64)
// 对应产物, 原子替换当前可执行文件, 失败自动回滚。

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// upgradeRepo 默认发布仓库, 可用环境变量 SYSMON_REPO 覆盖。
const upgradeRepo = "jormin/sysmon"

// upgradeBase 下载基址, 可用环境变量 SYSMON_UPGRADE_BASE 覆盖(如镜像/GitHub 加速)。
const upgradeBaseEnv = "SYSMON_UPGRADE_BASE"

type githubRelease struct {
	TagName string `json:"tag_name"`
}

var versionRe = regexp.MustCompile(`^v?([0-9]+)\.([0-9]+)\.([0-9]+)`)

func parseVersion(v string) ([]int, error) {
	m := versionRe.FindStringSubmatch(strings.TrimSpace(v))
	if m == nil {
		return nil, fmt.Errorf("无法解析版本号: %s", v)
	}
	major, _ := strconv.Atoi(m[1])
	minor, _ := strconv.Atoi(m[2])
	patch, _ := strconv.Atoi(m[3])
	return []int{major, minor, patch}, nil
}

func compareVersions(a, b []int) int {
	for i := 0; i < 3; i++ {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	return 0
}

// upgradeRepoSlug 归一化仓库字符串为 owner/repo 形式。
func upgradeRepoSlug() string {
	slug := strings.TrimSpace(os.Getenv("SYSMON_REPO"))
	if slug == "" {
		slug = upgradeRepo
	}
	slug = strings.TrimSuffix(slug, "/")
	for _, prefix := range []string{"https://github.com/", "http://github.com/", "github.com/"} {
		slug = strings.TrimPrefix(slug, prefix)
	}
	return slug
}

// latestReleaseVersion 查询最新 release 的 tag_name(形如 v0.0.2)。
func latestReleaseVersion(repo string) (string, error) {
	url := "https://api.github.com/repos/" + repo + "/releases/latest"
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("查询 GitHub Release 失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return "", errors.New("该项目暂无 GitHub Release")
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 200))
		return "", fmt.Errorf("查询 GitHub Release 失败(HTTP %d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var rel githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "", err
	}
	return strings.TrimSpace(rel.TagName), nil
}

// askYN 交互询问, 默认否。
func askYN(prompt string) bool {
	fmt.Print(prompt + " [y/N] ")
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	return strings.EqualFold(line, "y") || strings.EqualFold(line, "yes")
}

// isExecutableMagic 通过文件头魔数判断是否为 ELF(Linux)或 Mach-O(macOS)可执行文件,
// 避免 HTML 错误页/代理干扰被当成二进制安装。
func isExecutableMagic(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	var head [4]byte
	if n, _ := io.ReadFull(f, head[:]); n < 4 {
		return false
	}
	return (head[0] == 0x7f && head[1] == 'E' && head[2] == 'L' && head[3] == 'F') ||
		(head[0] == 0xcf && head[1] == 0xfa && head[2] == 0xed && head[3] == 0xfe) ||
		(head[0] == 0xce && head[1] == 0xfa && head[2] == 0xed && head[3] == 0xfe)
}

// runUpgrade 执行 upgrade 子命令, 返回进程退出码。
func runUpgrade(args []string) int {
	fs := flag.NewFlagSet("sysmon upgrade", flag.ExitOnError)
	checkOnly := fs.Bool("check", false, "只检测最新版本, 不更新")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "用法: sysmon upgrade [-check]\n")
		fs.PrintDefaults()
	}
	fs.Parse(args)

	repo := upgradeRepoSlug()
	current := strings.TrimPrefix(version, "v")
	fmt.Println("当前版本: v" + current)
	fmt.Println("检查仓库:", repo)

	latest, err := latestReleaseVersion(repo)
	if err != nil {
		fmt.Fprintln(os.Stderr, "错误: "+err.Error())
		return 1
	}
	latest = strings.TrimPrefix(latest, "v")

	lv, err := parseVersion(latest)
	if err != nil {
		fmt.Fprintln(os.Stderr, "错误: "+err.Error())
		return 1
	}
	cv, err := parseVersion(current)
	if err != nil {
		// 本地 dev 构建(未注入版本号)视为旧版本
		fmt.Println("当前为未注入版本号的构建(" + version + "), 视为旧版本")
		cv = []int{0, 0, 0}
	}

	if compareVersions(lv, cv) <= 0 {
		fmt.Println("已是最新版本: v" + current)
		return 0
	}

	fmt.Println("发现新版本: v" + latest + " (当前 v" + current + ")")
	if *checkOnly {
		fmt.Println("运行 sysmon upgrade 可自动更新。")
		return 0
	}
	if !askYN("是否下载并更新?") {
		fmt.Println("已取消。")
		return 0
	}

	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, "错误: 无法定位当前可执行文件: "+err.Error())
		return 1
	}

	// 平台限制: build.sh 只发布 darwin/linux × amd64/arm64
	switch runtime.GOOS + "/" + runtime.GOARCH {
	case "darwin/amd64", "darwin/arm64", "linux/amd64", "linux/arm64":
	default:
		fmt.Fprintf(os.Stderr, "错误: 当前平台 %s/%s 无预编译产物(仅 darwin/linux × amd64/arm64), 请用 go build 自行编译\n", runtime.GOOS, runtime.GOARCH)
		return 1
	}

	base := strings.TrimSuffix(os.Getenv(upgradeBaseEnv), "/")
	if base == "" {
		base = "https://github.com"
	}
	asset := fmt.Sprintf("sysmon_%s_%s_v%s", runtime.GOOS, runtime.GOARCH, latest)
	url := fmt.Sprintf("%s/%s/releases/download/v%s/%s", base, repo, latest, asset)
	fmt.Println("下载: " + url)

	tmp, err := os.CreateTemp("", "sysmon-upgrade-*")
	if err != nil {
		fmt.Fprintln(os.Stderr, "错误: "+err.Error())
		return 1
	}
	defer os.Remove(tmp.Name())

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Fprintln(os.Stderr, "错误: 下载失败: "+err.Error())
		return 1
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "错误: 下载失败(HTTP %d): %s\n", resp.StatusCode, url)
		return 1
	}
	if _, err := io.Copy(tmp, resp.Body); err != nil {
		fmt.Fprintln(os.Stderr, "错误: 下载中断: "+err.Error())
		return 1
	}
	if err := tmp.Chmod(0o755); err != nil {
		fmt.Fprintln(os.Stderr, "错误: "+err.Error())
		return 1
	}
	if err := tmp.Close(); err != nil {
		fmt.Fprintln(os.Stderr, "错误: "+err.Error())
		return 1
	}
	if !isExecutableMagic(tmp.Name()) {
		fmt.Fprintln(os.Stderr, "错误: 下载内容不是有效的可执行文件: "+url)
		return 1
	}

	// 原子替换: 当前二进制 -> .old, 新二进制 -> 当前路径; 失败回滚
	backup := exe + ".old"
	_ = os.Remove(backup)
	if err := os.Rename(exe, backup); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无法移动当前二进制(可能需要 sudo 或对该目录有写权限): %v\n", err)
		return 1
	}
	if err := os.Rename(tmp.Name(), exe); err != nil {
		_ = os.Rename(backup, exe)
		fmt.Fprintf(os.Stderr, "错误: 替换可执行文件失败(已回滚): %v\n", err)
		return 1
	}
	_ = os.Remove(backup)
	fmt.Println("已更新到 v" + latest + ", 请重新运行 sysmon(当前进程退出后新版本生效)。")
	return 0
}
