package gitlab

import (
	"net/http"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/election"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "gitlab"

	sampleCfg = `
[[inputs.gitlab]]
    ## param type: string - default: http://127.0.0.1:80/-/metrics
    prometheus_url = "http://127.0.0.1:80/-/metrics"

    ## param type: string - optional: time units are "ms", "s", "m", "h" - default: 10s
    interval = "10s"

    ## param type: map object, string to string
    [inputs.gitlab.tags]
    #tag1 = "value1"
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
}

func newInput() *Input {
	return &Input{
		Tags:     make(map[string]string),
		duration: datakit.IntervalDuration,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (*Input) SampleConfig() string {
	return sampleCfg
}

func (*Input) Catalog() string {
	return "gitlab"
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
			if election.CurrentStats().IsLeader() {
				this.gather()
			}
		}
	}
}

func (this *Input) loadCfg() {
	dur, err := time.ParseDuration(this.Interval)
	if err != nil {
		l.Warnf("parse interval error (use default %s): %s", datakit.IntervalDuration, err)
		return
	}
	this.duration = dur
}

func (this *Input) gather() {
	startTime := time.Now()

	pts, err := this.gatherMetrics()
	if err != nil {
		l.Error(err)
		return
	}

	cost := time.Since(startTime)

	if err := io.Feed(inputName, datakit.Metric, pts, &io.Option{CollectCost: cost}); err != nil {
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
		for k, v := range this.Tags {
			m.tags[k] = v
		}

		point, err := io.MakePoint(m.name, m.tags, m.fields)
		if err != nil {
			l.Warn(err)
			continue
		}
		points = append(points, point)
	}

	return points, nil
}

func (*Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&gitlabTransactionDBCountTotal{},
		&gitlabTransactionCacheReadMissCountTotal{},
		&gitlabRackRequestsTotal{},
		&gitlabCacheOperationsTotal{},
		&gitlabTransactionViewDurationTotal{},
		&gitlabTransactionNewRedisConnectionsTotal{},
		&gitlabSQLDurationSeconds{},
		&gitlabCacheOperationsDurationSeconds{},
		&gitlabRedisClientRequestsDurationSeconds{},
		&gitlabHTTPRequestDurationSeconds{},
		&gitlabRedisClientRequestsTotal{},
		&gitlabTransactionCacheReadHitCountTotal{},
		&gitlabTransactionDurationSeconds{},
		&gitlabHTTPHealthRequestsTotal{},
		&gitlabBanzaiCachelessRenderRealDurationSeconds{},
		&gitlabRubyGCDurationSeconds{},
		&gitlabRubySamplerDurationSecondsTotal{},
		&gitlabRailsQueueDurationSeconds{},
		&gitlabTransactionDBCachedCountTotal{},
		&gitlabCacheMissesTotal{},
	}
}
