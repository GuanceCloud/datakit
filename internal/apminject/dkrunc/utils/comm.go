// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package utils contains utils
package utils

import (
	"path/filepath"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

var (
	DirInject          = filepath.Join(datakit.InstallDir, "apm_inject/")
	DirInjectSubInject = filepath.Join(datakit.InstallDir, "apm_inject/inject")
	DirInjectSubLib    = filepath.Join(datakit.InstallDir, "apm_inject/lib")
	DirInjectSubLog    = filepath.Join(datakit.InstallDir, "apm_inject/log")

	InjectSubInject = "inject"
	InjectSubLib    = "lib"
	InjectSubLog    = "log"
)
