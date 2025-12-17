// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/flameshot"
)

var configPath = flag.String("config", "/flameshot/flameshot.conf", "log file path")

func main() {
	flag.Parse()

	config := flameshot.InitConfig(*configPath)
	if config == nil {
		return
	}
	m := flameshot.NewMonitor(config) // {config: config, cs: make([]*processM, 0), csChan: make(chan *processM, 2)}

	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt)

	m.Start(sigChan)
}
