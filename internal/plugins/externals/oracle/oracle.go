// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux && amd64

package main

import (
	"fmt"
	"path/filepath"

	"github.com/GuanceCloud/cliutils/logger"
	_ "github.com/godror/godror"
	"github.com/jessevdk/go-flags"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/oracle/collect"
)

var (
	opt collect.Option
	l   = logger.DefaultSLogger("oracle")
)

func main() {
	_, err := flags.Parse(&opt)
	if err != nil {
		fmt.Println("Parse error:", err)
		return
	}

	if opt.Log == "" {
		opt.Log = filepath.Join(datakit.InstallDir, "externals", "oracle.log")
	}

	if err := logger.InitRoot(&logger.Option{
		Path:  opt.Log,
		Level: opt.LogLevel,
		Flags: logger.OPT_DEFAULT,
	}); err != nil {
		l.Errorf("set root log failed: %s", err.Error())
	}

	if opt.InstanceDesc != "" { // add description to logger
		l = logger.SLogger("oracle-" + opt.InstanceDesc)
	} else {
		l = logger.SLogger("oracle")
	}

	l.Debugf("election: %t", opt.Election)

	l.Infof("datakit: host=%s, port=%d", opt.DatakitHTTPHost, opt.DatakitHTTPPort)

	collect.Set(&opt, l)

	m := collect.NewMonitor()
	m.Run()
}
