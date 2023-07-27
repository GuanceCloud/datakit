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
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
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
	Tagger  dkpt.GlobalTagger
}

func (i *Input) ElectionEnabled() bool {
	return i.Election
}

func (i *Input) appendM(p *point.Point) {
	i.m.Lock()
	i.collectCache = append(i.collectCache, p)
	i.m.Unlock()
}

func (i *Input) Catalog() string {
	return "db"
}

func (i *Input) SampleConfig() string {
	return sampleConfig
}

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&SolrCache{},
		&SolrRequestTimes{},
		&SolrSearcher{},
	}
}

func (i *Input) RunPipeline() {
	if i.Log == nil || len(i.Log.Files) == 0 {
		return
	}

	opt := &tailer.Option{
		Source:            inputName,
		Service:           inputName,
		Pipeline:          i.Log.Pipeline,
		GlobalTags:        i.Tags,
		IgnoreStatus:      i.Log.IgnoreStatus,
		CharacterEncoding: i.Log.CharacterEncoding,
		MultilinePatterns: []string{i.Log.MultilineMatch},
		Done:              i.semStop.Wait(),
	}

	var err error
	i.tail, err = tailer.NewTailer(i.Log.Files, opt)
	if err != nil {
		l.Error(err)
		i.feeder.FeedLastError(inputName, err.Error())
		return
	}

	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_solr"})
	g.Go(func(ctx context.Context) error {
		i.tail.Start()
		return nil
	})
}

func (*Input) PipelineConfig() map[string]string {
	pipelineMap := map[string]string{
		inputName: pipelineCfg,
	}
	return pipelineMap
}

func (i *Input) LogExamples() map[string]map[string]string {
	return map[string]map[string]string{
		inputName: {
			"Solr log": `2013-10-01 12:33:08.319 INFO (org.apache.solr.core.SolrCore) [collection1] webapp.reporter`,
		},
	}
}

func (i *Input) GetPipeline() []*tailer.Option {
	return []*tailer.Option{
		{
			Source:  inputName,
			Service: inputName,
			Pipeline: func() string {
				if i.Log != nil {
					return i.Log.Pipeline
				}
				return ""
			}(),
		},
	}
}

func (i *Input) AvailableArchs() []string {
	return datakit.AllOSWithElection
}

func (i *Input) Run() {
	l = logger.SLogger(inputName)
	l.Infof("solr input started")
	i.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, i.Interval.Duration)

	tick := time.NewTicker(i.Interval.Duration)
	defer tick.Stop()
	for {
		select {
		case <-datakit.Exit.Wait():
			i.exit()
			l.Infof("solr input exit")
			return

		case <-i.semStop.Wait():
			i.exit()
			l.Infof("solr input return")
			return

		case <-tick.C:
			if i.pause {
				l.Debugf("not leader, skipped")
				continue
			}
			start := time.Now()
			if err := i.Collect(); err == nil {
				if err := i.feeder.Feed(
					inputName,
					point.Metric,
					i.collectCache,
					&dkio.Option{CollectCost: time.Since((start))},
				); err != nil {
					i.logError(err)
				}
			} else {
				i.logError(err)
			}
			i.collectCache = make([]*point.Point, 0)

		case i.pause = <-i.pauseCh:
			// nil
		}
	}
}

func (i *Input) exit() {
	if i.tail != nil {
		i.tail.Close()
		l.Info("solr log exit")
	}
}

func (i *Input) Terminate() {
	if i.semStop != nil {
		i.semStop.Close()
	}
}

func (i *Input) Collect() error {
	i.collectCache = make([]*point.Point, 0)
	if i.client == nil {
		i.client = createHTTPClient(i.HTTPTimeout)
	}
	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_solr"})
	for _, s := range i.Servers {
		func(s string) {
			g.Go(func(ctx context.Context) error {
				ts := time.Now()
				resp := Response{}
				if err := i.gatherData(i, URLAll(s), &resp); err != nil {
					i.logError(err)
					return nil
				}
				for gKey, gValue := range resp.Metrics {
					// common tags
					commTag := map[string]string{}
					for kTag, vTag := range i.Tags { // user-defined
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
							i.logError(i.gatherSolrCache(k, s, v, commTag, ts))
						case "requesttimes":
							i.logError(i.gatherSolrRequestTimes(k, s, v, commTag, ts))
						case "searcher":
							i.logError(i.gatherSolrSearcher(k, v, fieldSearcher))
						default:
							continue
						}
					}

					if i.Election {
						tagsSearcher = inputs.MergeTags(i.Tagger.ElectionTags(), tagsSearcher, s)
					} else {
						tagsSearcher = inputs.MergeTags(i.Tagger.HostTags(), tagsSearcher, s)
					}

					// append searcher stats
					if len(fieldSearcher) > 0 {
						metric := &SolrSearcher{
							fields: fieldSearcher,
							tags:   tagsSearcher,
							name:   metricNameSearcher,
							ts:     ts,
						}
						i.appendM(metric.Point())
					}
				}
				return nil
			})
		}(s)
	}
	_ = g.Wait()
	return nil
}

func (i *Input) logError(err error) {
	if err != nil {
		l.Error(err)
		i.feeder.FeedLastError(inputName, err.Error())
	}
}

func (i *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	defer tick.Stop()
	select {
	case i.pauseCh <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (i *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	defer tick.Stop()
	select {
	case i.pauseCh <- false:
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
		Tagger:  dkpt.DefaultGlobalTagger(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
