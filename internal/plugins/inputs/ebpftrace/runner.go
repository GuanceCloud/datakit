// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !(windows && 386)
// +build !windows !386

// Package ebpftrace connect span
package ebpftrace

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/ebpftrace/spans"
)

func NewMRRunner(ipt *Input) error {
	if ipt.mrrunner == nil {
		sp := spans.NewSpanDB2(ipt.Window, ipt.SQLitePath)
		ipt.mrrunner = spans.NewMRRunner(spans.DefaultGenTraceID,
			sp, ipt.Window, ipt.UseAppTraceID, ipt.SamplingRate)
	}
	return nil
}
