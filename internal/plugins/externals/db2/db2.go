// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux && amd64

package main

import (
	"fmt"

	"github.com/jessevdk/go-flags"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/db2/collect"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/db2/collect/ccommon"
)

var opt ccommon.Option

func main() {
	// input := bufio.NewScanner(os.Stdin)
	// input.Scan()
	// fmt.Println(input.Text())

	if _, err := flags.Parse(&opt); err != nil {
		fmt.Println("flags.Parse error:", err.Error())
		return
	}

	collect.PrintInfof("election: %t", opt.Election)

	collect.PrintInfof("Datakit: host=%s, port=%d", opt.DatakitHTTPHost, opt.DatakitHTTPPort)

	collect.Run(&opt)

	fmt.Println("exiting...")
}
