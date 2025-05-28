// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package phpfpm collect host phpfpm metrics.
package phpfpm

import (
	"fmt"
	"strconv"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"

	pfpm "github.com/hipages/php-fpm_exporter/phpfpm"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/export/doc"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	minInterval = time.Second
	maxInterval = time.Minute
	inputName   = "phpfpm"
	metricName  = inputName
)

var (
	_ inputs.ElectionInput = (*Input)(nil)
	_ inputs.Singleton     = (*Input)(nil)
	l                      = logger.DefaultSLogger(inputName)
)

type Input struct {
	Interval time.Duration

	StatusURL    string            `toml:"status_url"`
	UseFastCGI   bool              `toml:"use_fastcgi"`
	Tags         map[string]string `toml:"tags"`
	collectCache []*point.Point
	feeder       dkio.Feeder
	mergedTags   map[string]string
	tagger       datakit.GlobalTagger

	Election bool `toml:"election"`
	pause    bool
	pauseCh  chan bool

	semStop *cliutils.Sem
	alignTS int64
}

func (ipt *Input) ElectionEnabled() bool {
	return ipt.Election
}

func (ipt *Input) Run() {
	ipt.setup()

	tick := time.NewTicker(ipt.Interval)
	defer tick.Stop()

	lastTS := time.Now()
	for {
		if ipt.pause {
			l.Debugf("not leader, skipped")
		} else {
			ipt.alignTS = lastTS.UnixNano()

			start := time.Now()
			if err := ipt.collect(); err != nil {
				l.Errorf("collect: %s", err)
				ipt.feeder.FeedLastError(err.Error(),
					metrics.WithLastErrorInput(inputName),
					metrics.WithLastErrorCategory(point.Metric),
				)
			}

			if len(ipt.collectCache) > 0 {
				if err := ipt.feeder.FeedV2(point.Metric, ipt.collectCache,
					dkio.WithCollectCost(time.Since(start)),
					dkio.WithElection(ipt.Election),
					dkio.WithInputName(metricName)); err != nil {
					ipt.feeder.FeedLastError(err.Error(),
						metrics.WithLastErrorInput(inputName),
						metrics.WithLastErrorCategory(point.Metric),
					)
					l.Errorf("feed measurement: %s", err)
				}
			}
		}

		select {
		case tt := <-tick.C:
			nextts := inputs.AlignTimeMillSec(tt, lastTS.UnixMilli(), ipt.Interval.Milliseconds())
			lastTS = time.UnixMilli(nextts)
		case <-datakit.Exit.Wait():
			l.Infof("%s input exit", inputName)
			return
		case <-ipt.semStop.Wait():
			l.Infof("%s input return", inputName)
			return
		case ipt.pause = <-ipt.pauseCh:
			// nil
		}
	}
}

func (ipt *Input) setup() {
	l = logger.SLogger(inputName)
	pfpm.SetLogger(&loggerAdapter{l: l})

	l.Infof("%s input started", inputName)
	ipt.Interval = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval)
	ipt.mergedTags = inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, "")
	l.Debugf("merged tags: %+#v", ipt.mergedTags)
}

func (ipt *Input) collect() error {
	ipt.collectCache = make([]*point.Point, 0)

	poolsPts, err := ipt.collectPoolsPts()
	if err != nil {
		l.Errorf("collect failed: %s", err.Error())
	}
	ipt.collectCache = poolsPts

	return nil
}

func (*Input) Singleton() {}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}
func (*Input) Catalog() string      { return inputName }
func (*Input) SampleConfig() string { return sampleCfg }
func (*Input) AvailableArchs() []string {
	return datakit.AllOSWithElection
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&baseMeasurement{},
	}
}

func (ipt *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	defer tick.Stop()
	select {
	case ipt.pauseCh <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (ipt *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	defer tick.Stop()
	select {
	case ipt.pauseCh <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}

func defaultInput() *Input {
	ipt := &Input{
		Interval:   time.Second * 10,
		Tags:       make(map[string]string),
		feeder:     dkio.DefaultFeeder(),
		semStop:    cliutils.NewSem(),
		tagger:     datakit.DefaultGlobalTagger(),
		Election:   true,
		pauseCh:    make(chan bool, inputs.ElectionPauseChannelLength),
		mergedTags: make(map[string]string),
	}
	return ipt
}

func (ipt *Input) GetENVDoc() []*inputs.ENVInfo {
	// nolint:lll
	infos := []*inputs.ENVInfo{
		{FieldName: "Interval", Type: doc.TimeDuration, Default: "`10s`", Desc: "Collect interval", DescZh: "采集器重复间隔时长"},
		{FieldName: "StatusURL", Type: doc.String, Default: `http://localhost/status`, Desc: "URL to fetch PHP-FPM pool metrics (HTTP or FastCGI)", DescZh: "用于获取 PHP-FPM 池指标的 URL（支持 HTTP 或 FastCGI）"},
		{FieldName: "UseFastCGI", Type: doc.Boolean, Default: "`false`", Desc: "Use FastCGI protocol instead of HTTP to fetch metrics", DescZh: "使用 FastCGI 协议而非 HTTP 获取指标"},
	}

	return doc.SetENVDoc("ENV_INPUT_PHPFPM_", infos)
}

// ReadEnv support envs：
//
//	ENV_INPUT_PHPFPM_INTERVAL : time.Duration
//	ENV_INPUT_PHPFPM_STATUS_URL : string
//	ENV_INPUT_PHPFPM_USE_FASTCGI : bool
func (ipt *Input) ReadEnv(envs map[string]string) {
	if tagsStr, ok := envs["ENV_INPUT_PHPFPM_TAGS"]; ok {
		tags := config.ParseGlobalTags(tagsStr)
		for k, v := range tags {
			ipt.Tags[k] = v
		}
	}

	//   ENV_INPUT_PHPFPM_INTERVAL : time.Duration
	if str, ok := envs["ENV_INPUT_PHPFPM_INTERVAL"]; ok {
		da, err := time.ParseDuration(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_PHPFPM_INTERVAL to time.Duration: %s, ignore", err)
		} else {
			ipt.Interval = config.ProtectedInterval(minInterval,
				maxInterval,
				da)
		}
	}

	//   ENV_INPUT_PHPFPM_STATUS_URL : string
	if statusURL, ok := envs["ENV_INPUT_PHPFPM_STATUS_URL"]; ok {
		ipt.StatusURL = statusURL
	}

	// ENV_INPUT_PHPFPM_USE_FASTCGI : bool
	if useFastCGI, ok := envs["ENV_INPUT_PHPFPM_USE_FASTCGI"]; ok {
		flag, err := strconv.ParseBool(useFastCGI)
		if err != nil {
			l.Warnf("parse ENV_INPUT_PHPFPM_USE_FASTCGI: %s, ignore", err)
		} else {
			ipt.UseFastCGI = flag
		}
	}
}

type loggerAdapter struct {
	l *logger.Logger
}

func (la *loggerAdapter) Info(args ...interface{}) {
	la.l.Info(args...)
}

func (la *loggerAdapter) Infof(format string, args ...interface{}) {
	la.l.Infof(format, args...)
}

func (la *loggerAdapter) Debug(args ...interface{}) {
	la.l.Debug(args...)
}

func (la *loggerAdapter) Debugf(format string, args ...interface{}) {
	la.l.Debugf(format, args...)
}

func (la *loggerAdapter) Error(args ...interface{}) {
	la.l.Error(args...)
}

func (la *loggerAdapter) Errorf(format string, args ...interface{}) {
	la.l.Errorf(format, args...)
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
