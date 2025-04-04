// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//nolint:lll
package cmds

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/monitor"
)

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

func runMonitorFlags() error {
	if *flagMonitorRefreshInterval < time.Second {
		*flagMonitorRefreshInterval = time.Second
	}

	// default
	to := config.Cfg.HTTPAPI.Listen

	// load from datakit.conf
	if x := loadLocalDatakitConf(); x != "" {
		to = x
	}

	// use command line host if specified
	if *flagMonitorTo != "" {
		to = *flagMonitorTo
	}

	schema := "http"
	if config.Cfg.HTTPAPI.HTTPSEnabled() {
		schema = "https"
	}
	monitor.Start(
		monitor.WithHost(schema, to),
		monitor.WithDumMetrics(*flagDumpMetrics),
		monitor.WithSource(*flagMonitorFilePath),
		monitor.WithTimestampMS(*flagMonitorTimestamp),
		monitor.WithMaxTableWidth(*flagMonitorMaxTableWidth),
		monitor.WithOnlyInputs(*flagMonitorOnlyInputs),
		monitor.WithOnlyModules(*flagMonitorModule),
		monitor.WithRefresh(*flagMonitorRefreshInterval),
		monitor.WithVerbose(*flagMonitorVerbose),
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
