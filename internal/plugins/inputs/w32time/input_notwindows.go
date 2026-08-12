// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows || !amd64
// +build !windows !amd64

package w32time

import "github.com/GuanceCloud/cliutils/logger"

func (*Input) Run() {
	l = logger.SLogger(inputName)
	l.Warnf("%s input is only supported on Windows amd64", inputName)
}
