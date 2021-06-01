package solr

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
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
	sampleConfig           = `
[[inputs.solr]]
  ##(optional) collect interval, default is 10 seconds
  interval = '10s'

  ## specify a list of one or more Solr servers
  servers = ["http://localhost:8983"]

  ## Optional HTTP Basic Auth Credentials
  # username = "username"
  # password = "pa$$word"

  [inputs.solr.log]
    # files = []
    ## grok pipeline script path
    # pipeline = "solr.p"

 [inputs.solr.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"

`
	pipelineCfg = `
add_pattern("solrReporter","(?:[.\\w\\d]+)")
add_pattern("solrParams", "(?:[A-Za-z0-9$.+!*'|(){},~@#%&/=:;_?\\-\\[\\]<>]*)")
add_pattern("solrPath", "(?:%{PATH}|null)")
grok(_,"%{TIMESTAMP_ISO8601:time}%{SPACE}%{LOGLEVEL:status}%{SPACE}\\(%{NOTSPACE:thread}\\)%{SPACE}\\[%{SPACE}%{NOTSPACE}?\\]%{SPACE}%{solrReporter:reporter}%{SPACE}\\[%{NOTSPACE:core}\\]%{SPACE}webapp=%{NOTSPACE:webapp}%{SPACE}path=%{solrPath:path}%{SPACE}params=\\{%{solrParams:params}\\}(?:%{SPACE}hits=%{NUMBER:hits})?%{SPACE}status=%{NUMBER:qstatus}%{SPACE}QTime=%{NUMBER:qtime}")
grok(_,"%{TIMESTAMP_ISO8601:time}%{SPACE}%{LOGLEVEL:status}%{SPACE}\\(%{NOTSPACE:thread}\\)%{SPACE}\\[%{SPACE}%{NOTSPACE}?\\]%{SPACE}%{solrReporter:reporter}.*")
default_time(time,"UTC")
`
)

type Input struct {
	Cores    []string // deprecated
	Servers  []string
	Username string
	Password string
	Interval datakit.Duration

	Log  *inputs.TailerOption `toml:"log"`
	Tags map[string]string

	HTTPTimeout  datakit.Duration
	client       *http.Client
	collectCache []inputs.Measurement
	gatherData   GatherData
	m            sync.Mutex
}

func (i *Input) appendM(m inputs.Measurement) {
	i.m.Lock()
	i.collectCache = append(i.collectCache, m)
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

func (i *Input) PipelineConfig() map[string]string {
	pipelineMap := map[string]string{
		inputName: pipelineCfg,
	}
	return pipelineMap
}

func (i *Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (i *Input) Run() {
	l = logger.SLogger(inputName)
	l.Infof("solr input started")
	i.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, i.Interval.Duration)

	if i.Log != nil && len(i.Log.Files) > 0 {
		go i.gatherLog()
	}

	tick := time.NewTicker(i.Interval.Duration)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			start := time.Now()
			if err := i.Collect(); err == nil {
				if feedErr := inputs.FeedMeasurement(inputName, datakit.Metric, i.collectCache,
					&io.Option{CollectCost: time.Since((start))}); feedErr != nil {
					logError(feedErr)
				}
			} else {
				logError(err)
			}
			i.collectCache = make([]inputs.Measurement, 0)
		case <-datakit.Exit.Wait():
			l.Infof("solr input exit")
			return
		}
	}
}

func (i *Input) Collect() error {
	i.collectCache = make([]inputs.Measurement, 0)
	if i.client == nil {
		i.client = createHTTPClient(i.HTTPTimeout)
	}
	var wg sync.WaitGroup
	wg.Add(len(i.Servers))
	for _, s := range i.Servers {
		go func(s string) {
			defer wg.Done()
			ts := time.Now()
			resp := Response{}
			if err := i.gatherData(i, UrlAll(s), &resp); err != nil {
				logError(err)
				return
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
						logError(i.gatherSolrCache(k, v, commTag, ts))
					case "requesttimes":
						logError(i.gatherSolrRequestTimes(k, v, commTag, ts))
					case "searcher":
						logError(i.gatherSolrSearcher(k, v, fieldSearcher))
					default:
						continue
					}
				}
				// append searcher stats
				if len(fieldSearcher) > 0 {
					i.appendM(&SolrSearcher{
						fields: fieldSearcher,
						tags:   tagsSearcher,
						name:   metricNameSearcher,
						ts:     ts,
					})
				}
			}
		}(s)
	}
	wg.Wait()
	return nil
}

func (i *Input) gatherLog() {
	inputs.JoinPipelinePath(i.Log, inputName+".p")
	i.Log.Source = inputName
	i.Log.Tags = map[string]string{}
	for k, v := range i.Tags {
		i.Log.Tags[k] = v
	}
	if tail, err := inputs.NewTailer(i.Log); err != nil {
		return
	} else {
		defer tail.Close()
		tail.Run()
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			HTTPTimeout: datakit.Duration{Duration: time.Second * 5},
			Interval:    datakit.Duration{Duration: time.Second * 10},
			gatherData:  gatherDataFunc,
		}
	})
}
