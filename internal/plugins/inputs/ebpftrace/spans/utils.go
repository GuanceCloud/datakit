// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package spans connect span
package spans

import (
	"github.com/GuanceCloud/cliutils/logger"
)

var log = logger.DefaultSLogger("ebpftrace-span")

func Init() {
	log = logger.SLogger("ebpftrace-span")
}
