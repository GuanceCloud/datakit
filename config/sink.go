// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sender"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink/sinkcommon"
)

func (c *Config) setupSinks() error {
	var writeFunc func(string, []*point.Point) (*sinkcommon.Failed, error)

	if c.DataWay != nil {
		if dw, ok := c.DataWay.(sender.Writer); ok {
			writeFunc = dw.Write
		}
	}

	return sink.Init(c.Sinks.Sink, writeFunc)
}
