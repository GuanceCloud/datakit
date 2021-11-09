// Package prom scrape prometheus exportor metrics.
package prom

import (
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/prom"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	_ inputs.ElectionInput = (*Input)(nil)
	_ inputs.Instance      = (*Input)(nil)
)

const (
	inputName = "prom"
	catalog   = "prom"
)

// defaultMaxFileSize is the default maximum response body size, in bytes.
// If the response body is over this size, we will simply discard its content instead of writing it to disk.
// 32 MB.
const defaultMaxFileSize int64 = 32 * 1024 * 1024

var l = logger.DefaultSLogger(inputName)

type Input struct {
	Source   string `toml:"source"`
	Interval string `toml:"interval"`

	URL               string      `toml:"url"`
	MetricTypes       []string    `toml:"metric_types"`
	MetricNameFilter  []string    `toml:"metric_name_filter"`
	MeasurementPrefix string      `toml:"measurement_prefix"`
	MeasurementName   string      `toml:"measurement_name"`
	Measurements      []prom.Rule `json:"measurements"`
	Output            string      `toml:"output"`
	maxFileSize       int64       `toml:"max_file_size"`

	TLSOpen    bool   `toml:"tls_open"`
	CacertFile string `toml:"tls_ca"`
	CertFile   string `toml:"tls_cert"`
	KeyFile    string `toml:"tls_key"`

	Tags           map[string]string `toml:"tags"`
	TagsIgnore     []string          `toml:"tags_ignore"`
	DeprecatedAuth map[string]string `toml:"auth"`

	pm      *prom.Prom
	chPause chan bool
	pause   bool

	stopCh chan interface{}

	semStop *cliutils.Sem // start stop signal
}

func (*Input) SampleConfig() string { return sampleCfg }

func (*Input) SampleMeasurement() []inputs.Measurement { return nil }

func (*Input) AvailableArchs() []string { return datakit.AllArch }

func (*Input) Catalog() string { return catalog }

func (ipt *Input) Stop() { ipt.stopCh <- nil }

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)

	if ipt.setup() {
		return
	}

	tick := time.NewTicker(ipt.pm.Option().GetIntervalDuration())
	defer tick.Stop()

	source := ipt.pm.Option().GetSource()

	l.Info("prom start")

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("prom exit")
			return

		case <-ipt.semStop.Wait():
			l.Info("prom return")
			return

		case <-ipt.stopCh:
			l.Info("prom stop")
			return

		case <-tick.C:
			if ipt.pause {
				l.Debugf("not leader, skipped")
				continue
			}
			l.Debugf("collect URL %s", ipt.pm.Option().URL)

			// If Output is configured, data is written to local file specified by Output.
			// Data will no more be written to datakit io.
			if ipt.Output != "" {
				err := ipt.pm.WriteFile()
				if err != nil {
					l.Debugf(err.Error())
				}
				continue
			}

			start := time.Now()
			pts, err := ipt.pm.Collect()
			if err != nil {
				l.Errorf("Collect: %s", err)

				io.FeedLastError(source, err.Error())
				continue
			}

			if len(pts) == 0 {
				l.Debug("len(points) is zero")
				continue
			}

			if err := io.Feed(source,
				datakit.Metric,
				pts,
				&io.Option{CollectCost: time.Since(start)}); err != nil {
				l.Errorf("Feed: %s", err)

				io.FeedLastError(source, err.Error())
			}

		case ipt.pause = <-ipt.chPause:
			// nil
		}
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (ipt *Input) setup() bool {
	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return true
		default:
			// nil
		}

		if err := ipt.Init(); err != nil {
			continue
		} else {
			break
		}
	}

	return false
}

func (ipt *Input) Init() error {
	// toml 不支持匿名字段的 marshal，JSON 支持
	opt := &prom.Option{
		Source:            ipt.Source,
		Interval:          ipt.Interval,
		URL:               ipt.URL,
		MetricTypes:       ipt.MetricTypes,
		MetricNameFilter:  ipt.MetricNameFilter,
		MeasurementPrefix: ipt.MeasurementPrefix,
		MeasurementName:   ipt.MeasurementName,
		Measurements:      ipt.Measurements,
		TLSOpen:           ipt.TLSOpen,
		CacertFile:        ipt.CacertFile,
		CertFile:          ipt.CertFile,
		KeyFile:           ipt.KeyFile,
		Tags:              ipt.Tags,
		TagsIgnore:        ipt.TagsIgnore,
		Output:            ipt.Output,
		MaxFileSize:       ipt.maxFileSize,
	}

	pm, err := prom.NewProm(opt)
	if err != nil {
		l.Error(err)
		return err
	}
	ipt.pm = pm
	return nil
}

func (ipt *Input) Collect() ([]*io.Point, error) {
	if ipt.pm == nil {
		return nil, nil
	}
	return ipt.pm.Collect()
}

func (ipt *Input) CollectFromFile() ([]*io.Point, error) {
	if ipt.pm == nil {
		return nil, nil
	}
	return ipt.pm.CollectFromFile()
}

func (ipt *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	select {
	case ipt.chPause <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (ipt *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	select {
	case ipt.chPause <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}

var maxPauseCh = inputs.ElectionPauseChannelLength

func NewProm() *Input {
	return &Input{
		stopCh:      make(chan interface{}, 1),
		chPause:     make(chan bool, maxPauseCh),
		maxFileSize: defaultMaxFileSize,

		semStop: cliutils.NewSem(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return NewProm()
	})
}
