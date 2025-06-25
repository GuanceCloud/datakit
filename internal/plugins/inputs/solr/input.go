// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package solr collects solr metrics.
package solr

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
)

const (
	minInterval = time.Second * 5
	maxInterval = time.Minute * 10
)

var (
	inputName              = "solr"
	metricNameCache        = "solr_cache"
	metricNameRequestTimes = "solr_request_times"
	metricNameSearcher     = "solr_searcher"

	l                      = logger.DefaultSLogger("solr")
	_ inputs.ElectionInput = (*Input)(nil)

	sampleConfig = `
[[inputs.solr]]
  ##(optional) collect interval, default is 10 seconds
  interval = '10s'

  ## specify a list of one or more Solr servers
  servers = ["http://localhost:8983"]

  ## Optional HTTP Basic Auth Credentials
  # username = "username"
  # password = "pa$$word"

  ## Set true to enable election
  election = true

  # [inputs.solr.log]
  # files = []
  # #grok pipeline script path
  # pipeline = "solr.p"

  [inputs.solr.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"

`
	//nolint:lll
	pipelineCfg = `
add_pattern("solrReporter","(?:[.\\w\\d]+)")
add_pattern("solrParams", "(?:[A-Za-z0-9$.+!*'|(){},~@#%&/=:;_?\\-\\[\\]<>]*)")
add_pattern("solrPath", "(?:%{PATH}|null)")
grok(_,"%{TIMESTAMP_ISO8601:time}%{SPACE}%{LOGLEVEL:status}%{SPACE}\\(%{NOTSPACE:thread}\\)%{SPACE}\\[%{SPACE}%{NOTSPACE}?\\]%{SPACE}%{solrReporter:reporter}%{SPACE}\\[%{NOTSPACE:core}\\]%{SPACE}webapp=%{NOTSPACE:webapp}%{SPACE}path=%{solrPath:path}%{SPACE}params=\\{%{solrParams:params}\\}(?:%{SPACE}hits=%{NUMBER:hits})?%{SPACE}status=%{NUMBER:qstatus}%{SPACE}QTime=%{NUMBER:qtime}")
grok(_,"%{TIMESTAMP_ISO8601:time}%{SPACE}%{LOGLEVEL:status}%{SPACE}\\(%{NOTSPACE:thread}\\)%{SPACE}\\[%{SPACE}%{NOTSPACE}?\\]%{SPACE}%{solrReporter:reporter}.*")
default_time(time,"UTC")
`
)

type solrlog struct {
	Files             []string `toml:"files"`
	Pipeline          string   `toml:"pipeline"`
	IgnoreStatus      []string `toml:"ignore"`
	CharacterEncoding string   `toml:"character_encoding"`
	MultilineMatch    string   `toml:"multiline_match"`
}

type Input struct {
	Cores    []string // deprecated
	Servers  []string
	Username string
	Password string
	Interval datakit.Duration

	Log  *solrlog `toml:"log"`
	Tags map[string]string

	HTTPTimeout  datakit.Duration
	client       *http.Client
	tail         *tailer.Tailer
	collectCache []*point.Point
	gatherData   GatherData
	m            sync.Mutex

	Election bool `toml:"election"`
	pause    bool
	pauseCh  chan bool

	semStop *cliutils.Sem // start stop signal
	feeder  dkio.Feeder
	Tagger  datakit.GlobalTagger
	ptsTime time.Time
}

func (ipt *Input) ElectionEnabled() bool {
	return ipt.Election
}

func (ipt *Input) appendM(p *point.Point) {
	ipt.m.Lock()
	ipt.collectCache = append(ipt.collectCache, p)
	ipt.m.Unlock()
}

func (*Input) Catalog() string {
	return "db"
}

func (*Input) SampleConfig() string {
	return sampleConfig
}

func (ipt *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&SolrCache{},
		&SolrRequestTimes{},
		&SolrSearcher{},
	}
}

func (ipt *Input) RunPipeline() {
	if ipt.Log == nil || len(ipt.Log.Files) == 0 {
		return
	}

	opts := []tailer.Option{
		tailer.WithSource(inputName),
		tailer.WithService(inputName),
		tailer.WithPipeline(ipt.Log.Pipeline),
		tailer.WithIgnoreStatus(ipt.Log.IgnoreStatus),
		tailer.WithCharacterEncoding(ipt.Log.CharacterEncoding),
		tailer.EnableMultiline(true),
		tailer.WithMaxMultilineLength(int64(float64(config.Cfg.Dataway.MaxRawBodySize) * 0.8)),
		tailer.WithMultilinePatterns([]string{ipt.Log.MultilineMatch}),
		tailer.WithGlobalTags(inputs.MergeTags(ipt.Tagger.HostTags(), ipt.Tags, "")),
		tailer.EnableDebugFields(config.Cfg.EnableDebugFields),
	}

	var err error
	ipt.tail, err = tailer.NewTailer(ipt.Log.Files, opts...)
	if err != nil {
		l.Error(err)
		ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(inputName),
		)
		return
	}

	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_solr"})
	g.Go(func(ctx context.Context) error {
		ipt.tail.Start()
		return nil
	})
}

func (*Input) PipelineConfig() map[string]string {
	pipelineMap := map[string]string{
		inputName: pipelineCfg,
	}
	return pipelineMap
}

func (ipt *Input) LogExamples() map[string]map[string]string {
	return map[string]map[string]string{
		inputName: {
			"Solr log": `2013-10-01 12:33:08.319 INFO (org.apache.solr.core.SolrCore) [collection1] webapp.reporter`,
		},
	}
}

func (ipt *Input) GetPipeline() []tailer.Option {
	opts := []tailer.Option{
		tailer.WithSource(inputName),
		tailer.WithService(inputName),
	}
	if ipt.Log != nil {
		opts = append(opts, tailer.WithPipeline(ipt.Log.Pipeline))
	}
	return opts
}

func (ipt *Input) AvailableArchs() []string {
	return datakit.AllOSWithElection
}

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)
	l.Infof("solr input started")
	ipt.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval.Duration)

	tick := time.NewTicker(ipt.Interval.Duration)
	defer tick.Stop()

	ipt.ptsTime = ntp.Now()
	for {
		if ipt.pause {
			l.Debugf("not leader, skipped")
		} else {
			collectStart := time.Now()
			if err := ipt.Collect(); err == nil {
				if err := ipt.feeder.Feed(point.Metric, ipt.collectCache,
					dkio.WithCollectCost(time.Since(collectStart)),
					dkio.WithElection(ipt.Election),
					dkio.WithSource(inputName),
				); err != nil {
					ipt.logError(err)
				}
			} else {
				ipt.logError(err)
			}
			ipt.collectCache = ipt.collectCache[:0]
		}

		select {
		case <-datakit.Exit.Wait():
			ipt.exit()
			l.Infof("solr input exit")
			return

		case <-ipt.semStop.Wait():
			ipt.exit()
			l.Infof("solr input return")
			return

		case tt := <-tick.C:
			ipt.ptsTime = inputs.AlignTime(tt, ipt.ptsTime, ipt.Interval.Duration)

		case ipt.pause = <-ipt.pauseCh:
			// nil
		}
	}
}

func (ipt *Input) exit() {
	if ipt.tail != nil {
		ipt.tail.Close()
		l.Info("solr log exit")
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (ipt *Input) Collect() error {
	ipt.collectCache = make([]*point.Point, 0)
	if ipt.client == nil {
		ipt.client = createHTTPClient(ipt.HTTPTimeout)
	}
	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_solr"})
	for _, s := range ipt.Servers {
		func(s string) {
			g.Go(func(ctx context.Context) error {
				ts := time.Now()
				resp := Response{}
				if err := ipt.gatherData(ipt, URLAll(s), &resp); err != nil {
					ipt.logError(err)
					return nil
				}
				for gKey, gValue := range resp.Metrics {
					// common tags
					commTag := map[string]string{}
					for kTag, vTag := range ipt.Tags { // user-defined
						commTag[kTag] = vTag
					}
					keySplit := strings.Split(gKey, ".")
					commTag["group"] = keySplit[1]
					if commTag["group"] == "core" {
						commTag["core"] = keySplit[2]
					}
					if instcName, err := instanceName(s); err == nil { // generated based on server address
						commTag["instance"] = instcName
					}
					// searcher stats tags and fields
					tagsSearcher := map[string]string{}
					for kTag, vTag := range commTag {
						tagsSearcher[kTag] = vTag
					}
					tagsSearcher["category"] = "SEARCHER" // searcher metric category
					fieldSearcher := map[string]interface{}{}
					// gather stats
					for k, v := range gValue {
						switch whichMesaurement(k) {
						case "cache":
							ipt.logError(ipt.gatherSolrCache(k, s, v, commTag, ts))
						case "requesttimes":
							ipt.logError(ipt.gatherSolrRequestTimes(k, s, v, commTag, ts))
						case "searcher":
							ipt.logError(ipt.gatherSolrSearcher(k, v, fieldSearcher))
						default:
							continue
						}
					}

					if ipt.Election {
						tagsSearcher = inputs.MergeTags(ipt.Tagger.ElectionTags(), tagsSearcher, s)
					} else {
						tagsSearcher = inputs.MergeTags(ipt.Tagger.HostTags(), tagsSearcher, s)
					}

					// append searcher stats
					if len(fieldSearcher) > 0 {
						metric := &SolrSearcher{
							fields: fieldSearcher,
							tags:   tagsSearcher,
							name:   metricNameSearcher,
							ts:     ipt.ptsTime.UnixNano(),
						}
						ipt.appendM(metric.Point())
					}
				}
				return nil
			})
		}(s)
	}

	return g.Wait()
}

func (ipt *Input) logError(err error) {
	if err != nil {
		l.Error(err)
		ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(inputName),
		)
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
	return &Input{
		HTTPTimeout: datakit.Duration{Duration: time.Second * 5},
		Interval:    datakit.Duration{Duration: time.Second * 10},
		gatherData:  gatherDataFunc,
		pauseCh:     make(chan bool, inputs.ElectionPauseChannelLength),
		Election:    true,

		semStop: cliutils.NewSem(),
		feeder:  dkio.DefaultFeeder(),
		Tagger:  datakit.DefaultGlobalTagger(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
