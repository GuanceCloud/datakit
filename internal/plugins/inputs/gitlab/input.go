// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package gitlab collect GitLab metrics
package gitlab

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var (
	_ inputs.ElectionInput = (*Input)(nil)
	l                      = logger.DefaultSLogger(inputName)
	g                      = datakit.G("inputs_gitlab")
)

const (
	inputName = "gitlab"
	catalog   = "gitlab"

	gitlabEventHeader = "X-Gitlab-Event"
	pipelineHook      = "Pipeline Hook"
	jobHook           = "Job Hook"

	sampleCfg = `
[[inputs.gitlab]]
    ## set true if you need to collect metric from url below
    enable_collect = true

    ## param type: string - default: http://127.0.0.1:80/-/metrics
    prometheus_url = "http://127.0.0.1:80/-/metrics"

    ## param type: string - optional: time units are "ms", "s", "m", "h" - default: 10s
    interval = "10s"

    ## datakit can listen to gitlab ci data at /v1/gitlab when enabled
    enable_ci_visibility = true

    ## Set true to enable election
    election = true

    ## extra tags for gitlab-ci data.
    ## these tags will not overwrite existing tags.
    [inputs.gitlab.ci_extra_tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"

    ## extra tags for gitlab metrics
    [inputs.gitlab.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
`
)

type Input struct {
	EnableCollect bool              `toml:"enable_collect"`
	URL           string            `toml:"prometheus_url"`
	Interval      string            `toml:"interval"`
	Tags          map[string]string `toml:"tags"`

	EnableCIVisibility bool              `toml:"enable_ci_visibility"`
	CIExtraTags        map[string]string `toml:"ci_extra_tags"`

	httpClient *http.Client
	duration   time.Duration

	Election bool `toml:"election"`
	pause    bool
	pauseCh  chan bool

	semStop *cliutils.Sem // start stop signal
	reqMemo requestMemo
	feeder  dkio.Feeder
	Tagger  datakit.GlobalTagger
}

func (ipt *Input) ElectionEnabled() bool {
	return ipt.Election
}

func (ipt *Input) RegHTTPHandler() {
	if ipt.EnableCIVisibility {
		l.Infof("start listening to gitlab pipeline/job webhooks")
		g.Go(func(ctx context.Context) error {
			ipt.reqMemo.memoMaintainer(time.Second * 30)
			return nil
		})
		httpapi.RegHTTPHandler("POST", "/v1/gitlab", httpapi.ProtectedHandlerFunc(ipt.ServeHTTP, l))
	}
}

var maxPauseCh = inputs.ElectionPauseChannelLength

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)

	if !ipt.EnableCollect {
		l.Infof("metric collecting is disabled, gitlab exited")
		return
	}

	ipt.loadCfg()

	ticker := time.NewTicker(ipt.duration)
	defer ticker.Stop()
	start := ntp.Now()

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("gitlab exited")
			return

		case <-ipt.semStop.Wait():
			l.Info("gitlab returned")
			return

		case tt := <-ticker.C:
			if ipt.pause {
				l.Debugf("not leader, skipped")
				continue
			}
			start = inputs.AlignTime(tt, start, ipt.duration)
			ipt.gather(start.UnixNano())

		case ipt.pause = <-ipt.pauseCh:
			// nil
		}
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}

	if ipt.EnableCIVisibility {
		httpapi.RemoveHTTPRoute("POST", "/v1/gitlab")
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

func (ipt *Input) loadCfg() {
	dur, err := time.ParseDuration(ipt.Interval)
	if err != nil {
		l.Warnf("parse interval error (use default 10s): %s", err)
		return
	}
	ipt.duration = dur
}

func (ipt *Input) gather(ptTS int64) {
	start := time.Now()

	pts, err := ipt.gatherMetrics(ptTS)
	if err != nil {
		l.Error(err)
		return
	}

	if len(pts) > 0 {
		if err := ipt.feeder.FeedV2(point.Metric, pts,
			dkio.WithCollectCost(time.Since(start)),
			dkio.WithElection(ipt.Election),
			dkio.WithInputName(inputName)); err != nil {
			l.Errorf("feed measurement: %s", err)
		}
	}
}

func (ipt *Input) gatherMetrics(ptTS int64) ([]*point.Point, error) {
	resp, err := ipt.httpClient.Get(ipt.URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck

	metrics, err := promTextToMetrics(resp.Body)
	if err != nil {
		return nil, err
	}

	var points []*point.Point

	for _, m := range metrics {
		measurement := inputName

		// 非常粗暴的筛选方式
		if len(m.tags) == 0 {
			measurement = inputName + "_base"
		}
		if _, ok := m.tags["method"]; ok {
			measurement = inputName + "_http"
		}

		setHostTagIfNotLoopback(m.tags, ipt.URL)
		for k, v := range ipt.Tags {
			m.tags[k] = v
		}

		if ipt.Election {
			m.tags = inputs.MergeTags(ipt.Tagger.ElectionTags(), m.tags, ipt.URL)
		} else {
			m.tags = inputs.MergeTags(ipt.Tagger.HostTags(), m.tags, ipt.URL)
		}

		opts := point.DefaultMetricOptions()
		opts = append(opts, point.WithTimestamp(ptTS))
		points = append(points, point.NewPointV2(measurement,
			append(point.NewTags(m.tags), point.NewKVs(m.fields)...), opts...),
		)
	}

	return points, nil
}

func (*Input) SampleConfig() string { return sampleCfg }

func (*Input) Catalog() string { return catalog }

func (*Input) AvailableArchs() []string { return datakit.AllOSWithElection }

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&gitlabMeasurement{},
		&gitlabBaseMeasurement{},
		&gitlabHTTPMeasurement{},
		&gitlabPipelineMeasurement{},
		&gitlabJobMeasurement{},
	}
}

func setHostTagIfNotLoopback(tags map[string]string, u string) {
	uu, err := url.Parse(u)
	if err != nil {
		l.Errorf("parse url: %v", err)
		return
	}
	host, _, err := net.SplitHostPort(uu.Host)
	if err != nil {
		l.Errorf("split host and port: %v", err)
		return
	}
	if host != "localhost" && !net.ParseIP(host).IsLoopback() {
		tags["host"] = host
	}
}

func defaultInput() *Input {
	sem := cliutils.NewSem()
	return &Input{
		EnableCollect: true,
		Tags:          make(map[string]string),
		pauseCh:       make(chan bool, maxPauseCh),
		Election:      true,
		duration:      time.Second * 10,
		httpClient:    &http.Client{Timeout: 5 * time.Second},

		semStop: sem,
		feeder:  dkio.DefaultFeeder(),
		Tagger:  datakit.DefaultGlobalTagger(),

		EnableCIVisibility: true,
		CIExtraTags:        make(map[string]string),
		reqMemo: requestMemo{
			memoMap:     map[[16]byte]time.Time{},
			hasReqCh:    make(chan hasRequest),
			addReqCh:    make(chan [16]byte),
			removeReqCh: make(chan [16]byte),
			semStop:     sem,
		},
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
