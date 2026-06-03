// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package main

import (
	cobracmd "gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/datakit/cmd"
)

// injected during building: -X.
var (
	InputsReleaseType = ""
	ReleaseVersion    = ""
	Lite              = "false"
	ELinker           = "false"
)

func main() {
	cobracmd.SetBuildInfo(ReleaseVersion, InputsReleaseType, Lite, ELinker)
	if err := cobracmd.Execute(); err != nil {
		panic(err)
	}
}
