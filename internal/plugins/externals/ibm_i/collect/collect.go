// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux && amd64

// Package collect starts the IBM i external collector.
package collect

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/GuanceCloud/cliutils/logger"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ibm_i/collect/ccommon"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ibm_i/collect/cibmi"
)

const inputName = "ibm_i"

func Run(opt *ccommon.Option) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)
	defer signal.Stop(stop)

	ccommon.DatakitLastErrURL = ccommon.GetLastErrorURL(opt.DatakitHTTPHost, opt.DatakitHTTPPort)

	ipt, err := cibmi.NewInput(opt)
	if err != nil {
		l := logger.DefaultSLogger(inputName)
		ccommon.ReportErrorf(inputName, l, "initialize collector failed: %v", err)
		return
	}

	ipt.Run(stop)
}
