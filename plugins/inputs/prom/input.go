package prom

import (
	"fmt"
	"time"

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
}

func (*Input) SampleConfig() string { return sampleCfg }

func (*Input) SampleMeasurement() []inputs.Measurement { return nil }

func (*Input) AvailableArchs() []string { return datakit.AllArch }

func (*Input) Catalog() string { return catalog }

func (i *Input) Stop() { i.stopCh <- nil }

func (i *Input) getSource() string {
	source := inputName
	if i.Source != "" {
		source = i.Source
	}
	return source
}

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

		case <-i.stopCh:
			l.Info("prom stop")
			return

		case <-tick.C:
			if i.pause {
				continue
			}
			l.Info(i.pm.Option().URL)

			start := time.Now()
			pts, err := i.pm.Collect()
			if err != nil {
				io.FeedLastError(source, err.Error())
				l.Error(err)
				continue
			}

			if len(pts) == 0 {
				continue
			}

			if err := io.Feed(source, datakit.Metric, pts, &io.Option{CollectCost: time.Since(start)}); err != nil {
				io.FeedLastError(source, err.Error())
				l.Error(err)
			}

		case i.pause = <-i.chPause:
			// nil
		}
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

func (i *Input) DebugCollect() ([]*io.Point, error) {
	if i.pm == nil {
		return nil, nil
	}
	return i.pm.DebugCollect()
}

func (i *Input) Pause() error {
	tick := time.NewTicker(time.Second * 5)
	select {
	case i.chPause <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (i *Input) Resume() error {
	tick := time.NewTicker(time.Second * 5)
	select {
	case i.chPause <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}

func NewProm() *Input {
	return &Input{
		chPause: make(chan bool, 1),
		stopCh:  make(chan interface{}, 1),
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return NewProm()
	})
}
