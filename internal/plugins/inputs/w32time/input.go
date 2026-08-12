// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package w32time collects Windows Time Service metrics.
package w32time

import (
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	inputName  = "w32time"
	metricName = inputName

	defaultInterval = time.Minute
	minInterval     = 10 * time.Second
	maxInterval     = 10 * time.Minute
)

var (
	l = logger.DefaultSLogger(inputName)

	_ inputs.InputV2   = (*Input)(nil)
	_ inputs.Singleton = (*Input)(nil)
)

type Input struct {
	Interval time.Duration     `toml:"interval"`
	Tags     map[string]string `toml:"tags"`

	semStop    *cliutils.Sem
	feeder     dkio.Feeder
	tagger     datakit.GlobalTagger
	mergedTags map[string]string
}

func (*Input) Catalog() string      { return inputName }
func (*Input) SampleConfig() string { return sampleCfg }
func (*Input) Singleton()           {}

func (*Input) AvailableArchs() []string {
	return []string{datakit.OSLabelWindows}
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&docMeasurement{}}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func newDefaultInput() *Input {
	return &Input{
		Interval: defaultInterval,
		Tags:     map[string]string{},
		semStop:  cliutils.NewSem(),
		feeder:   dkio.DefaultFeeder(),
		tagger:   datakit.DefaultGlobalTagger(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return newDefaultInput()
	})
}
