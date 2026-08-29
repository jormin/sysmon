package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sysmon/internal/analyze"
	"sysmon/internal/collect"
	"sysmon/internal/output"
)

// version 由构建脚本通过 -ldflags "-X main.version=vX.Y.Z" 注入。
var version = "dev"

func main() {
	run(os.Args[1:])
}

func run(args []string) {
	fs := flag.NewFlagSet("sysmon", flag.ExitOnError)
	interval := fs.Duration("interval", time.Second, "采样间隔")
	duration := fs.Duration("duration", 0, "采样总时长（0 = 直到 Ctrl+C）")
	count := fs.Int("count", 0, "采样条数（0 = 使用 duration）")
	out := fs.String("out", "sysmon.csv", "输出文件路径，'-' 表示标准输出")
	format := fs.String("format", "csv", "输出格式：csv | json | text")
	quiet := fs.Bool("quiet", false, "不在终端打印实时采样行与汇总")
	cpu := fs.Bool("cpu", true, "采集 CPU")
	cores := fs.Bool("cores", false, "采集每核 CPU")
	mem := fs.Bool("mem", true, "采集内存")
	showVersion := fs.Bool("version", false, "显示版本号")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "用法: sysmon [flags]\n")
		fs.PrintDefaults()
	}
	fs.Parse(args)

	if *showVersion {
		fmt.Println(version)
		return
	}

	if *interval <= 0 {
		fmt.Fprintln(os.Stderr, "sysmon: -interval 必须大于 0")
		os.Exit(2)
	}
	switch *format {
	case "csv", "json", "text":
	default:
		fmt.Fprintln(os.Stderr, "sysmon: -format 必须是 csv/json/text")
		os.Exit(2)
	}
	if !*cpu && !*mem {
		fmt.Fprintln(os.Stderr, "sysmon: 至少启用 -cpu 或 -mem 之一")
		os.Exit(2)
	}

	opts := collect.Options{CPU: *cpu, PerCore: *cores, Mem: *mem}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	col := collect.New(opts)
	col.Baseline()

	var samples []collect.Sample
	deadline := time.Time{}
	if *duration > 0 {
		deadline = time.Now().Add(*duration)
	}
	ticker := time.NewTicker(*interval)
	defer ticker.Stop()
	fmt.Fprintf(os.Stderr, "sysmon: 开始采样, 间隔 %s, 输出 %s (Ctrl+C 提前结束)\n", *interval, *out)
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case now := <-ticker.C:
			s := col.Sample(now)
			samples = append(samples, s)
			if !*quiet && *out != "-" {
				printLive(s, opts)
			}
			if *count > 0 && len(samples) >= *count {
				break loop
			}
			if !deadline.IsZero() && time.Now().After(deadline) {
				break loop
			}
		}
	}

	summary := analyze.SummaryOf(samples, interval.Seconds())
	writeSampleOutput(*format, *out, samples, summary)
	if !*quiet && *out != "-" {
		output.WriteSummaryText(os.Stdout, summary)
	} else if *out == "-" && !*quiet {
		output.WriteSummaryText(os.Stderr, summary)
	}
	fmt.Fprintf(os.Stderr, "sysmon: 采集 %d 条采样\n", len(samples))
}

func writeSampleOutput(format, out string, samples []collect.Sample, summary analyze.Summary) {
	var err error
	switch format {
	case "json":
		if out == "-" {
			err = output.WriteReportJSON(os.Stdout, samples, summary)
		} else {
			var f *os.File
			if f, err = os.Create(out); err == nil {
				defer f.Close()
				err = output.WriteReportJSON(f, samples, summary)
			}
		}
	case "text":
		if out == "-" {
			err = output.WriteText(os.Stdout, samples, summary)
		} else {
			var f *os.File
			if f, err = os.Create(out); err == nil {
				defer f.Close()
				err = output.WriteText(f, samples, summary)
			}
			if err == nil {
				err = writeSummarySidecar(out, summary)
			}
		}
	default: // csv
		if out == "-" {
			err = output.WriteCSV(os.Stdout, samples)
		} else {
			var f *os.File
			if f, err = os.Create(out); err == nil {
				defer f.Close()
				err = output.WriteCSV(f, samples)
			}
			if err == nil {
				err = writeSummarySidecar(out, summary)
			}
		}
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "sysmon: 写输出失败: %v\n", err)
		os.Exit(1)
	}
}

func writeSummarySidecar(out string, summary analyze.Summary) error {
	f, err := os.Create(out + ".summary.json")
	if err != nil {
		return err
	}
	defer f.Close()
	return output.WriteSummaryJSON(f, summary)
}

func printLive(s collect.Sample, opts collect.Options) {
	fmt.Printf("%s ", time.UnixMilli(s.Timestamp).Format("15:04:05"))
	if opts.CPU && s.CPU != nil {
		fmt.Printf("cpu %5.1f%% ", s.CPU.Usage)
	}
	if opts.Mem && s.Mem != nil {
		fmt.Printf("mem %5.1f%% ", s.Mem.UsagePct)
	}
	fmt.Println()
}
