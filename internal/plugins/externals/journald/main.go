// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package main

import (
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/spf13/pflag"

	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/journald/collect"
)

var opt collect.Option

func run(opt *collect.Option) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)

	// If singleton, use function pointer's list: "[]func(string) collect.IInput"
	// If multiple instance, use interface's list: "[]collect.IInput"
	//
	// For now, we use function pointer's list because of performance improving,
	// e.g. if not specified, it should not be initialized.
	var definedInputList []func(*collect.Option) collect.IInput
	definedInputList = append(definedInputList, collect.NewInput)

	ch := make(chan struct{}, len(definedInputList))

	for _, inputFunc := range definedInputList {
		input := inputFunc(opt)
		if input != nil {
			go func() {
				input.Run()
				ch <- struct{}{}
			}()
		}
	}

	count := 0
	select {
	case <-ch:
		count++
		if count == len(definedInputList) {
			close(ch)
		}
	case <-sigs:
		return
	}
}

func initFlags() {
	pflag.StringVar(&opt.DatakitHTTPEndpoint, "datakit-http-endpoint", "http://localhost:9529", "Datakit HTTP endpoint")

	pflag.StringVar(&opt.LogPath, "log-path", filepath.Join(datakit.InstallDir, "externals", "journald.log"), "Log file path")
	pflag.StringVar(&opt.LogLevel, "log-level", "info", "Log file level(debug/info/warn)")
	pflag.StringVar(&opt.Tags, "tags", "", "Extra tags, format: key1=value1,key2=value2,...")

	pflag.StringVar(&opt.Paths, "paths", "/var/log/journal,/run/log/journal", "Journal paths, comma separated")
	pflag.StringVar(&opt.Units, "units", "", "Filter by systemd units, comma separated")
	pflag.StringVar(&opt.Priorities, "priorities", "", "Filter by priority levels, comma separated")
	pflag.StringVar(&opt.ExcludeFields, "exclude-fields",
		"", "Fields to exclude, comma separated, for example _BOOT_ID,_MACHINE_ID,__MONOTONIC_TIMESTAMP")

	pflag.BoolVar(&opt.TailOnly, "tail-only", false, "Only read new entries(default false)")
	pflag.IntVar(&opt.MaxEntries, "max-entries", 1000, "Maximum entries per batch")
	pflag.BoolVar(&opt.SaveCursor, "save-cursor", false, "Save cursor position")
	pflag.StringVar(&opt.CursorFile, "cursor-file", "", "Cursor file path")

	pflag.CommandLine.SortFlags = false

	// Add help flag
	pflag.BoolP("help", "h", false, "Show help")

	pflag.Parse()

	// Check if help was requested
	if help, _ := pflag.CommandLine.GetBool("help"); help {
		cp.Printf("Journald Collector - Collect systemd journal entries and send to DataKit\n\n")
		cp.Printf("Usage: journald [OPTIONS]\n\n")
		cp.Printf("Options:\n")
		pflag.PrintDefaults()
		os.Exit(0)
	}
}

func main() {
	initFlags()

	// Initialize logger
	lo := &logger.Option{
		Path:  opt.LogPath,
		Level: opt.LogLevel,
		Flags: logger.OPT_DEFAULT,
	}

	if err := logger.InitRoot(lo); err != nil {
		cp.Errorf("init looger failed: %s\n", err)
		return
	}

	l := logger.SLogger("journald")

	cp.Infof("Datakit endpoint: %s", opt.DatakitHTTPEndpoint)

	run(&opt)

	l.Info("exiting...")
}
