// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package jolokia implement Jolokia-based JVM metrics collector.
package jolokia

import (
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/influxdata/telegraf/plugins/common/tls"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

// --------------------------------------------------------------------
// --------------------------- jolokia agent --------------------------.
const (
	MaxGatherInterval = 30 * time.Minute
	MinGatherInterval = 1 * time.Second

	defaultFieldName = "value"
)

var log = logger.DefaultSLogger("jolokia")

type JolokiaAgent struct {
	DefaultFieldPrefix    string
	DefaultFieldSeparator string
	DefaultTagPrefix      string

	URLs            []string `toml:"urls"`
	Username        string
	Password        string
	ResponseTimeout time.Duration `toml:"response_timeout"`
	Interval        string        `toml:"interval"`
	Election        bool          `toml:"election"`

	tls.ClientConfig

	Metrics  []MetricConfig `toml:"metric"`
	gatherer *gatherer
	clients  []*jclient

	collectCache []*point.Point
	PluginName   string `toml:"-"`
	L            *logger.Logger

	Tags  map[string]string `toml:"-"`
	Types map[string]string `toml:"-"`

	SemStop *cliutils.Sem `toml:"-"` // start stop signal
	Feeder  dkio.Feeder   `toml:"-"`
	Tagger  datakit.GlobalTagger
	g       *goroutine.Group
	pause   bool
	pauseCh chan bool
}

func DefaultInput() *JolokiaAgent {
	return &JolokiaAgent{
		SemStop:  cliutils.NewSem(),
		pauseCh:  make(chan bool, inputs.ElectionPauseChannelLength),
		Election: true,
		Tagger:   datakit.DefaultGlobalTagger(),
		Feeder:   dkio.DefaultFeeder(),
	}
}

func (j *JolokiaAgent) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	defer tick.Stop()
	select {
	case j.pauseCh <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", j.PluginName)
	}
}

func (j *JolokiaAgent) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	defer tick.Stop()
	select {
	case j.pauseCh <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", j.PluginName)
	}
}

func (j *JolokiaAgent) Terminate() {
	if j.SemStop != nil {
		j.SemStop.Close()
	}
}

func (*JolokiaMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{}
}
