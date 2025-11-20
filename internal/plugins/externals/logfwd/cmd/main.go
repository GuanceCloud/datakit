// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/logfwd"
)

var (
	log         = logger.DefaultSLogger("main")
	loggerLevel = os.Getenv("LOGFWD_LOG_LEVEL")
)

func main() {
	initLogger()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		cancel()
	}()

	if err := logfwd.StartLogForwarding(ctx); err != nil {
		log.Errorf("logfwd failed: %v", err)
		return
	}
}

func initLogger() {
	lopt := &logger.Option{
		Level: "info",
		Flags: (logger.OPT_DEFAULT | logger.OPT_STDOUT),
	}

	if loggerLevel == "debug" {
		lopt.Level = "debug"
	}

	if err := logger.InitRoot(lopt); err != nil {
		log.Errorf("failed to init logger: %s", err.Error())
		return
	}

	log = logger.SLogger("logfwd")
}
