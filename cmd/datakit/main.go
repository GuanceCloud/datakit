// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package main

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/datakit/core"
)

// injected during building: -X.
var (
	InputsReleaseType = ""
	ReleaseVersion    = ""
	Lite              = "false"
	ELinker           = "false"
)

func main() {
	app := core.New()
	if err := app.Initialize(ReleaseVersion, InputsReleaseType, Lite, ELinker); err != nil {
		panic(err)
	}
}
