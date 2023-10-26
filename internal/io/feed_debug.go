// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"

	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
)

// debugFeederOutput send feeder data to terminal.
type debugOutput struct{}

var _ FeederOutputer = new(debugOutput)

func (fo *debugOutput) Reader(cat point.Category) <-chan *iodata {
	return nil
}

func (fo *debugOutput) Write(data *iodata) error {
	for _, pt := range data.points {
		cp.Output("%s\n", pt.LineProto())
	}

	now := time.Now()
	date := fmt.Sprintf("%d/%02d/%02d %02d:%02d:%02d",
		now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())

	cp.Infof("# [%s] %d points(%q) from %s, cost %s | Ctrl+c to exit.\n",
		date, len(data.points), data.category.Alias(), data.from, data.opt.CollectCost)

	return nil
}

func (fo *debugOutput) WriteLastError(err string, opts ...LastErrorOption) {
	le := newLastError()

	for _, opt := range opts {
		if opt != nil {
			opt(le)
		}
	}

	cp.Errorf("[E] get error from input = %s, source = %s: %s", le.Input, le.Source, err)
	cp.Infof(" | Ctrl+c to exit.\n")
}

func NewDebugOutput() *debugOutput {
	return &debugOutput{}
}
