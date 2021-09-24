package gitlab

import (
	"fmt"
	"net/http"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var _ inputs.ElectionInput = (*Input)(nil)

const (
	inputName = "gitlab"
	catalog   = "gitlab"

	sampleCfg = `
[[inputs.gitlab]]
    ## param type: string - default: http://127.0.0.1:80/-/metrics
    prometheus_url = "http://127.0.0.1:80/-/metrics"

    ## param type: string - optional: time units are "ms", "s", "m", "h" - default: 10s
    interval = "10s"

    [inputs.gitlab.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
`
)

var l = logger.DefaultSLogger(inputName)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return newInput()
	})
}

type Input struct {
	URL      string            `toml:"prometheus_url"`
	Interval string            `toml:"interval"`
	Tags     map[string]string `toml:"tags"`

	httpClient *http.Client
	duration   time.Duration

	pause   bool
	pauseCh chan bool
}

func newInput() *Input {
	return &Input{
		Tags:       make(map[string]string),
		pauseCh:    make(chan bool, 1),
		duration:   time.Second * 10,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (this *Input) Run() {
	l = logger.SLogger(inputName)

	this.loadCfg()

	ticker := time.NewTicker(this.duration)
	defer ticker.Stop()

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return

		case <-ticker.C:
			if this.pause {
				l.Debugf("not leader, skipped")
				continue
			}
			this.gather()

		case this.pause = <-this.pauseCh:
			// nil
		}
	}
}

func (this *Input) Pause() error {
	tick := time.NewTicker(time.Second * 5)
	defer tick.Stop()
	select {
	case this.pauseCh <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (this *Input) Resume() error {
	tick := time.NewTicker(time.Second * 5)
	defer tick.Stop()
	select {
	case this.pauseCh <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}

func (this *Input) loadCfg() {
	dur, err := time.ParseDuration(this.Interval)
	if err != nil {
		l.Warnf("parse interval error (use default 10s): %s", err)
		return
	}
	this.duration = dur
}

func (this *Input) gather() {
	start := time.Now()

	pts, err := this.gatherMetrics()
	if err != nil {
		l.Error(err)
		return
	}

	if err := io.Feed(inputName, datakit.Metric, pts, &io.Option{CollectCost: time.Since(start)}); err != nil {
		l.Error(err)
	}
}

func (this *Input) gatherMetrics() ([]*io.Point, error) {
	resp, err := this.httpClient.Get(this.URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	metrics, err := promTextToMetrics(resp.Body)
	if err != nil {
		return nil, err
	}

	var points []*io.Point

	for _, m := range metrics {
		measurement := inputName

		// 非常粗暴的筛选方式
		if len(m.tags) == 0 {
			measurement = inputName + "_base"
		}
		if _, ok := m.tags["method"]; ok {
			measurement = inputName + "_http"
		}

		for k, v := range this.Tags {
			m.tags[k] = v
		}

		point, err := io.MakePoint(measurement, m.tags, m.fields)
		if err != nil {
			l.Warn(err)
			continue
		}
		points = append(points, point)
	}

	return points, nil
}

func (*Input) SampleConfig() string { return sampleCfg }

func (*Input) Catalog() string { return catalog }

func (*Input) AvailableArchs() []string { return datakit.AllArch }

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&gitlabMeasurement{},
		&gitlabBaseMeasurement{},
		&gitlabHTTPMeasurement{},
	}
}
