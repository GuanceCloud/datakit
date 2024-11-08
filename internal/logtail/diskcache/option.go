// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package diskcache

import (
	"github.com/GuanceCloud/cliutils/point"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
)

type Option func(*config)

type Handler func(*PBData) ([]*point.Point, *PBConfig, error)

type config struct {
	path        string
	capacity    int
	consumerNum int
	handle      Handler
	feeder      dkio.Feeder
}

func defaultConfig() *config {
	return &config{
		path:        "./logtail",
		capacity:    1024 * 5, // MiB
		consumerNum: 8,
		handle:      DecodeToPoint,
		feeder:      dkio.DefaultFeeder(),
	}
}

func WithFeeder(feeder dkio.Feeder) Option { return func(c *config) { c.feeder = feeder } }
func WithPath(s string) Option             { return func(c *config) { c.path = s } }
func WithCapacity(n int) Option            { return func(c *config) { c.capacity = n } }
func WithHandle(fn Handler) Option         { return func(c *config) { c.handle = fn } }
func WithConsumerNum(n int) Option         { return func(c *config) { c.consumerNum = n } }
