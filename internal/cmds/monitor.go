// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//nolint:lll
package cmds

import (
	"fmt"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/monitor"
)

type MonitorOptions struct {
	To            string
	MaxTableWidth int
	LogPath       string
	Refresh       time.Duration
	Verbose       bool
	Module        string
	OnlyInputs    string
	FilePath      string
	TimestampMS   int64
	DumpMetrics   bool
	Quantile      string
}

var moduleMap = map[string]string{
	"G":   "goroutine",
	"B":   "basic",
	"R":   "runtime",
	"F":   "filter",
	"H":   "http",
	"In":  "inputs",
	"P":   "pipeline",
	"IO":  "io_stats",
	"W":   "dataway",
	"WAL": "wal",
}

// loadLocalDatakitConf try to find where local datakit listen.
func loadLocalDatakitConf() string {
	if err := config.Cfg.LoadMainTOML(datakit.MainConfPath); err != nil {
		return ""
	}

	return config.Cfg.HTTPAPI.Listen
}

func RunMonitor(opts MonitorOptions) error {
	if opts.Module != "" {
		nomodule := existsModule(strings.Split(opts.Module, ","))
		if len(nomodule) != 0 {
			return fmt.Errorf("has no module:%+v,check please", nomodule)
		}
	}

	ConfigureCommandLog(opts.LogPath)

	if opts.Refresh < time.Second {
		opts.Refresh = time.Second
	}

	// default
	to := config.Cfg.HTTPAPI.Listen

	// load from datakit.conf
	if x := loadLocalDatakitConf(); x != "" {
		to = x
	}

	// use command line host if specified
	if opts.To != "" {
		to = opts.To
	}

	schema := "http"
	if config.Cfg.HTTPAPI.HTTPSEnabled() {
		schema = "https"
	}
	monitor.Start(
		monitor.WithHost(schema, to),
		monitor.WithQuantile(opts.Quantile),
		monitor.WithDumpMetrics(opts.DumpMetrics),
		monitor.WithSource(opts.FilePath),
		monitor.WithTimestampMS(opts.TimestampMS),
		monitor.WithMaxTableWidth(opts.MaxTableWidth),
		monitor.WithOnlyInputs(opts.OnlyInputs),
		monitor.WithOnlyModules(opts.Module),
		monitor.WithRefresh(opts.Refresh),
		monitor.WithVerbose(opts.Verbose),
		monitor.WithProxy(config.Cfg.Dataway.HTTPProxy),
	)
	return nil
}

func existsModule(str []string) []string {
	wrong := []string{}
	for _, s := range str {
		exsist := false
		for k, v := range moduleMap {
			if s == k || s == v {
				exsist = true
				break
			}
		}
		if !exsist {
			wrong = append(wrong, s)
		}
	}

	return wrong
}
