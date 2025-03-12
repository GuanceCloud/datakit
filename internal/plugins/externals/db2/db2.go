// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux && amd64

package main

import (
	"github.com/jessevdk/go-flags"

	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/db2/collect"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/db2/collect/ccommon"
)

var opt ccommon.Option

func main() {
	if _, err := flags.Parse(&opt); err != nil {
		cp.Println("flags.Parse error:", err.Error())
		return
	}

	collect.PrintInfof("election: %t", opt.Election)

	collect.PrintInfof("Datakit: host=%s, port=%d", opt.DatakitHTTPHost, opt.DatakitHTTPPort)

	collect.Run(&opt)

	cp.Println("exiting...")
}
