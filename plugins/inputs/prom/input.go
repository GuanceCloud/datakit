// Package prom scrape prometheus exportor metrics.
package prom

import (
	"fmt"
	"net/url"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/prom"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var _ inputs.ElectionInput = (*Input)(nil)

const (
	inputName = "prom"
	catalog   = "prom"
)

// defaultMaxFileSize is the default maximum response body size, in bytes.
// If the response body is over i size, we will simply discard its content instead of writing it to disk.
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

	pm *prom.Prom

	chPause chan bool
	pause   bool

	url    *url.URL
	stopCh chan interface{}

	semStop *cliutils.Sem // start stop signal
}

func (*Input) SampleConfig() string { return sampleCfg }

func (*Input) SampleMeasurement() []inputs.Measurement { return nil }

func (*Input) AvailableArchs() []string { return datakit.AllArch }

func (*Input) Catalog() string { return catalog }

func (i *Input) Run() {
	l = logger.SLogger(inputName)

	if i.setup() {
		return
	}

	tick := time.NewTicker(i.pm.Option().GetIntervalDuration())
	defer tick.Stop()

	source := i.pm.Option().GetSource()

	l.Info("prom start")

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("prom exit")
			return

		case <-i.semStop.Wait():
			l.Info("prom return")
			return

		case <-i.stopCh:
			l.Info("prom stop")
			return

		case <-tick.C:
			if i.pause {
				l.Debugf("not leader, skipped")
				continue
			}
			l.Debugf("collect URL %s", i.pm.Option().URL)

			// If Output is configured, data is written to local file specified by Output.
			// Data will no more be written to datakit io.
			if i.Output != "" {
				err := i.pm.WriteFile()
				if err != nil {
					l.Debugf(err.Error())
				}
				continue
			}

			start := time.Now()
			pts, err := i.pm.Collect()
			if err != nil {
				l.Errorf("Collect: %s", err)
				io.FeedLastError(source, err.Error())

				// Try testing the connect
				if i.url != nil {
					if err := net.RawConnect(i.url.Hostname(), i.url.Port(), time.Second*3); err != nil {
						l.Errorf("failed to connect to %s:%s, %s, exit", i.url.Hostname(), i.url.Port(), err)
						return
					}
				}

				continue
			}

			if len(pts) == 0 {
				l.Debug("len(points) is 0")
				continue
			}

			if err := io.Feed(source,
				datakit.Metric,
				pts,
				&io.Option{CollectCost: time.Since(start)}); err != nil {
				l.Errorf("Feed: %s", err)

				io.FeedLastError(source, err.Error())
			}

		case i.pause = <-i.chPause:
			// nil
		}
	}
}

func (i *Input) Terminate() {
	if i.semStop != nil {
		i.semStop.Close()
	}
}

func (i *Input) setup() bool {
	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return true
		default:
			// nil
		}

		if err := i.Init(); err != nil {
			continue
		} else {
			break
		}
	}

	return false
}

func (i *Input) Init() error {
	u, err := url.Parse(i.URL)
	if err != nil {
		return err
	}
	i.url = u

	// toml 不支持匿名字段的 marshal，JSON 支持
	opt := &prom.Option{
		Source:            i.Source,
		Interval:          i.Interval,
		URL:               i.URL,
		MetricTypes:       i.MetricTypes,
		MetricNameFilter:  i.MetricNameFilter,
		MeasurementPrefix: i.MeasurementPrefix,
		MeasurementName:   i.MeasurementName,
		Measurements:      i.Measurements,
		TLSOpen:           i.TLSOpen,
		CacertFile:        i.CacertFile,
		CertFile:          i.CertFile,
		KeyFile:           i.KeyFile,
		Tags:              i.Tags,
		TagsIgnore:        i.TagsIgnore,
		Output:            i.Output,
		MaxFileSize:       i.maxFileSize,
	}

	pm, err := prom.NewProm(opt)
	if err != nil {
		l.Error(err)
		return err
	}
	i.pm = pm

	return nil
}

func (i *Input) Collect() ([]*io.Point, error) {
	if i.pm == nil {
		return nil, nil
	}
	return i.pm.Collect()
}

func (i *Input) CollectFromFile() ([]*io.Point, error) {
	if i.pm == nil {
		return nil, nil
	}
	return i.pm.CollectFromFile()
}

func (i *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	select {
	case i.chPause <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (i *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	select {
	case i.chPause <- false:
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
