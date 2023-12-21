// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package elasticsearch Collect ElasticSearch metrics.
package elasticsearch

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
)

var _ inputs.ElectionInput = (*Input)(nil)

var mask = regexp.MustCompile(`https?:\/\/\S+:\S+@`)

const (
	statsPath      = "/_nodes/stats"
	statsPathLocal = "/_nodes/_local/stats"
)

type nodeStat struct {
	Host       string            `json:"host"`
	Name       string            `json:"name"`
	Roles      []string          `json:"roles"`
	Attributes map[string]string `json:"attributes"`
	Indices    interface{}       `json:"indices"`
	OS         interface{}       `json:"os"`
	Process    interface{}       `json:"process"`
	JVM        interface{}       `json:"jvm"`
	ThreadPool interface{}       `json:"thread_pool"`
	FS         interface{}       `json:"fs"`
	Transport  interface{}       `json:"transport"`
	HTTP       interface{}       `json:"http"`
	Breakers   interface{}       `json:"breakers"`
}

type clusterHealth struct {
	ActivePrimaryShards         int                    `json:"active_primary_shards"`
	ActiveShards                int                    `json:"active_shards"`
	ActiveShardsPercentAsNumber float64                `json:"active_shards_percent_as_number"`
	ClusterName                 string                 `json:"cluster_name"`
	DelayedUnassignedShards     int                    `json:"delayed_unassigned_shards"`
	InitializingShards          int                    `json:"initializing_shards"`
	NumberOfDataNodes           int                    `json:"number_of_data_nodes"`
	NumberOfInFlightFetch       int                    `json:"number_of_in_flight_fetch"`
	NumberOfNodes               int                    `json:"number_of_nodes"`
	NumberOfPendingTasks        int                    `json:"number_of_pending_tasks"`
	RelocatingShards            int                    `json:"relocating_shards"`
	Status                      string                 `json:"status"`
	TaskMaxWaitingInQueueMillis int                    `json:"task_max_waiting_in_queue_millis"`
	TimedOut                    bool                   `json:"timed_out"`
	UnassignedShards            int                    `json:"unassigned_shards"`
	Indices                     map[string]indexHealth `json:"indices"`
}

type indexState struct {
	Indices map[string]struct {
		Managed bool   `json:"managed"`
		Step    string `json:"step"`
	} `json:"indices"`
}

type indexHealth struct {
	ActivePrimaryShards int    `json:"active_primary_shards"`
	ActiveShards        int    `json:"active_shards"`
	InitializingShards  int    `json:"initializing_shards"`
	NumberOfReplicas    int    `json:"number_of_replicas"`
	NumberOfShards      int    `json:"number_of_shards"`
	RelocatingShards    int    `json:"relocating_shards"`
	Status              string `json:"status"`
	UnassignedShards    int    `json:"unassigned_shards"`
}

type clusterStats struct {
	NodeName    string      `json:"node_name"`
	ClusterName string      `json:"cluster_name"`
	Status      string      `json:"status"`
	Indices     interface{} `json:"indices"`
	Nodes       interface{} `json:"nodes"`
}

type indexStat struct {
	Primaries interface{}              `json:"primaries"`
	Total     interface{}              `json:"total"`
	Shards    map[string][]interface{} `json:"shards"`
}

//nolint:lll
const sampleConfig = `
[[inputs.elasticsearch]]
  ## Elasticsearch server url
  # Basic Authentication is allowed
  # servers = ["http://user:pass@localhost:9200"]
  servers = ["http://localhost:9200"]

  ## Collect interval
  # Time unit: "ns", "us" (or "µs"), "ms", "s", "m", "h"
  interval = "10s"

  ## HTTP timeout
  http_timeout = "5s"

  ## Distribution: elasticsearch, opendistro, opensearch
  distribution = "elasticsearch"

  ## Set local true to collect the metrics of the current node only.
  # Or you can set local false to collect the metrics of all nodes in the cluster.
  local = true

  ## Set true to collect the health metric of the cluster.
  cluster_health = false

  ## Set cluster health level, either indices or cluster.
  # cluster_health_level = "indices"

  ## Whether to collect the stats of the cluster.
  cluster_stats = false

  ## Set true to collect cluster stats only from the master node.
  cluster_stats_only_from_master = true

  ## Indices to be collected, such as _all.
  indices_include = ["_all"]

  ## Indices level, may be one of "shards", "cluster", "indices".
  # Currently only "shards" is implemented.
  indices_level = "shards"

  ## Specify the metrics to be collected for the node stats, such as "indices", "os", "process", "jvm", "thread_pool", "fs", "transport", "http", "breaker".
  # node_stats = ["jvm", "http"]

  ## HTTP Basic Authentication
  # username = ""
  # password = ""

  ## TLS Config
  tls_open = false
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Set true to enable election
  election = true

  # [inputs.elasticsearch.log]
  # files = []
  # #grok pipeline script path
  # pipeline = "elasticsearch.p"

  [inputs.elasticsearch.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
`

//nolint:lll
const pipelineCfg = `
# Elasticsearch_search_query
grok(_, "^\\[%{TIMESTAMP_ISO8601:time}\\]\\[%{LOGLEVEL:status}%{SPACE}\\]\\[i.s.s.(query|fetch)%{SPACE}\\] (\\[%{HOSTNAME:nodeId}\\] )?\\[%{NOTSPACE:index}\\]\\[%{INT}\\] took\\[.*\\], took_millis\\[%{INT:duration}\\].*")

# Elasticsearch_slow_indexing
grok(_, "^\\[%{TIMESTAMP_ISO8601:time}\\]\\[%{LOGLEVEL:status}%{SPACE}\\]\\[i.i.s.index%{SPACE}\\] (\\[%{HOSTNAME:nodeId}\\] )?\\[%{NOTSPACE:index}/%{NOTSPACE}\\] took\\[.*\\], took_millis\\[%{INT:duration}\\].*")

# Elasticsearch_default
grok(_, "^\\[%{TIMESTAMP_ISO8601:time}\\]\\[%{LOGLEVEL:status}%{SPACE}\\]\\[%{NOTSPACE:name}%{SPACE}\\]%{SPACE}(\\[%{HOSTNAME:nodeId}\\])?.*")

cast(shard, "int")
cast(duration, "int")

duration_precision(duration, "ms", "ns")

nullif(nodeId, "")
default_time(time)
`

type Input struct {
	Interval                   string   `toml:"interval"`
	Local                      bool     `toml:"local"`
	Distribution               string   `toml:"distribution"`
	Servers                    []string `toml:"servers"`
	HTTPTimeout                string   `toml:"http_timeout"`
	ClusterHealth              bool     `toml:"cluster_health"`
	ClusterHealthLevel         string   `toml:"cluster_health_level"`
	ClusterStats               bool     `toml:"cluster_stats"`
	ClusterStatsOnlyFromMaster bool     `toml:"cluster_stats_only_from_master"`
	IndicesInclude             []string `toml:"indices_include"`
	IndicesLevel               string   `toml:"indices_level"`
	NodeStats                  []string `toml:"node_stats"`
	Username                   string   `toml:"username"`
	Password                   string   `toml:"password"`
	Log                        *struct {
		Files             []string `toml:"files"`
		Pipeline          string   `toml:"pipeline"`
		IgnoreStatus      []string `toml:"ignore"`
		CharacterEncoding string   `toml:"character_encoding"`
		MultilineMatch    string   `toml:"multiline_match"`
	} `toml:"log"`

	Tags map[string]string `toml:"tags"`

	TLSOpen            bool   `toml:"tls_open"`
	CacertFile         string `toml:"tls_ca"`
	CertFile           string `toml:"tls_cert"`
	KeyFile            string `toml:"tls_key"`
	InsecureSkipVerify bool   `toml:"insecure_skip_verify"`

	httpTimeout     Duration
	client          *http.Client
	serverInfo      map[string]serverInfo
	serverInfoMutex sync.Mutex
	duration        time.Duration
	tail            *tailer.Tailer

	collectCache []*point.Point

	Election bool `toml:"election"`
	pause    bool
	pauseCh  chan bool

	feeder dkio.Feeder
	tagger datakit.GlobalTagger

	semStop *cliutils.Sem // start stop signal
}

func (ipt *Input) ElectionEnabled() bool {
	return ipt.Election
}

//nolint:lll
func (ipt *Input) LogExamples() map[string]map[string]string {
	return map[string]map[string]string{
		inputName: {
			"ElasticSearch log":             `[2021-06-01T11:45:15,927][WARN ][o.e.c.r.a.DiskThresholdMonitor] [master] high disk watermark [90%] exceeded on [A2kEFgMLQ1-vhMdZMJV3Iw][master][/tmp/elasticsearch-cluster/nodes/0] free: 17.1gb[7.3%], shards will be relocated away from this node; currently relocating away shards totalling [0] bytes; the node is expected to continue to exceed the high disk watermark when these relocations are complete`,
			"ElasticSearch search slow log": `[2021-06-01T11:56:06,712][WARN ][i.s.s.query              ] [master] [shopping][0] took[36.3ms], took_millis[36], total_hits[5 hits], types[], stats[], search_type[QUERY_THEN_FETCH], total_shards[1], source[{"query":{"match":{"name":{"query":"Nariko","operator":"OR","prefix_length":0,"max_expansions":50,"fuzzy_transpositions":true,"lenient":false,"zero_terms_query":"NONE","auto_generate_synonyms_phrase_query":true,"boost":1.0}}},"sort":[{"price":{"order":"desc"}}]}], id[],`,
			"ElasticSearch index slow log":  `[2021-06-01T11:56:19,084][WARN ][i.i.s.index              ] [master] [shopping/X17jbNZ4SoS65zKTU9ZAJg] took[34.1ms], took_millis[34], type[_doc], id[LgC3xXkBLT9WrDT1Dovp], routing[], source[{"price":222,"name":"hello"}]`,
		},
	}
}

type userPrivilege struct {
	Cluster struct {
		Monitor bool `json:"monitor"`
	} `json:"cluster"`
	Index map[string]struct {
		Monitor bool `json:"monitor"`
		Ilm     bool `json:"manage_ilm"`
	} `json:"index"`
}

type serverInfo struct {
	nodeID        string
	masterID      string
	version       string
	userPrivilege *userPrivilege
}

func (i serverInfo) isMaster() bool {
	return i.nodeID == i.masterID
}

var maxPauseCh = inputs.ElectionPauseChannelLength

func defaultInput() *Input {
	return &Input{
		httpTimeout:                Duration{Duration: time.Second * 5},
		ClusterStatsOnlyFromMaster: true,
		ClusterHealthLevel:         "indices",
		pauseCh:                    make(chan bool, maxPauseCh),
		Election:                   true,
		semStop:                    cliutils.NewSem(),
		feeder:                     dkio.DefaultFeeder(),
		tagger:                     datakit.DefaultGlobalTagger(),
	}
}

// perform status mapping.
func mapHealthStatusToCode(s string) int {
	switch strings.ToLower(s) {
	case "green":
		return 1
	case "yellow":
		return 2
	case "red":
		return 3
	}
	return 0
}

var (
	inputName   = "elasticsearch"
	catalogName = "db"
	l           = logger.DefaultSLogger("elasticsearch")
)

func (*Input) Catalog() string {
	return catalogName
}

func (*Input) SampleConfig() string {
	return sampleConfig
}

func (*Input) PipelineConfig() map[string]string {
	pipelineMap := map[string]string{
		"elasticsearch": pipelineCfg,
	}
	return pipelineMap
}

func (ipt *Input) GetPipeline() []*tailer.Option {
	return []*tailer.Option{
		{
			Source:  inputName,
			Service: inputName,
			Pipeline: func() string {
				if ipt.Log != nil {
					return ipt.Log.Pipeline
				}
				return ""
			}(),
		},
	}
}

func (ipt *Input) extendSelfTag(tags map[string]string) {
	if ipt.Tags != nil {
		for k, v := range ipt.Tags {
			tags[k] = v
		}
	}
}

func (ipt *Input) AvailableArchs() []string {
	return datakit.AllOSWithElection
}

func (ipt *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&nodeStatsMeasurement{},
		&indicesStatsMeasurement{},
		&clusterStatsMeasurement{},
		&clusterHealthMeasurement{},
	}
}

func (ipt *Input) setServerInfo() error {
	if len(ipt.Distribution) == 0 {
		ipt.Distribution = "elasticsearch"
	}
	ipt.serverInfo = make(map[string]serverInfo)

	g := goroutine.NewGroup(goroutine.Option{Name: goroutine.GetInputName(inputName)})
	for _, serv := range ipt.Servers {
		func(s string) {
			g.Go(func(ctx context.Context) error {
				var err error
				info := serverInfo{}

				// get nodeID和masterID
				if ipt.ClusterStats || len(ipt.IndicesInclude) > 0 || len(ipt.IndicesLevel) > 0 {
					// Gather node ID
					if info.nodeID, err = ipt.gatherNodeID(s + "/_nodes/_local/name"); err != nil {
						return fmt.Errorf(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@"))
					}

					// get cat/master information here so NodeStats can determine
					// whether this node is the Master
					if info.masterID, err = ipt.getCatMaster(s + "/_cat/master"); err != nil {
						return fmt.Errorf(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@"))
					}
				}

				if info.version, err = ipt.getVersion(s); err != nil {
					return fmt.Errorf(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@"))
				}

				if mask.MatchString(s) {
					info.userPrivilege = ipt.getUserPrivilege(s)
				}

				ipt.serverInfoMutex.Lock()
				ipt.serverInfo[s] = info
				ipt.serverInfoMutex.Unlock()

				return nil
			})
		}(serv)
	}
	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

func (ipt *Input) Collect() error {
	if err := ipt.setServerInfo(); err != nil {
		return err
	}

	g := goroutine.NewGroup(goroutine.Option{Name: goroutine.GetInputName(inputName)})
	for _, serv := range ipt.Servers {
		func(s string) {
			g.Go(func(ctx context.Context) error {
				var clusterName string
				var err error
				url := ipt.nodeStatsURL(s)

				// Always gather node stats
				if clusterName, err = ipt.gatherNodeStats(url); err != nil {
					l.Warn(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@"))
				}

				if ipt.ClusterHealth {
					url = s + "/_cluster/health"
					if ipt.ClusterHealthLevel != "" {
						url = url + "?level=" + ipt.ClusterHealthLevel
					}
					if err := ipt.gatherClusterHealth(url, s); err != nil {
						l.Warn(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@"))
					}
				}

				if ipt.ClusterStats && (ipt.serverInfo[s].isMaster() || !ipt.ClusterStatsOnlyFromMaster || !ipt.Local) {
					if err := ipt.gatherClusterStats(s + "/_cluster/stats"); err != nil {
						l.Warn(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@"))
					}
				}

				if len(ipt.IndicesInclude) > 0 &&
					(ipt.serverInfo[s].isMaster() ||
						!ipt.ClusterStatsOnlyFromMaster ||
						!ipt.Local) {
					// get indices stats
					if ipt.IndicesLevel != "shards" {
						if err := ipt.gatherIndicesStats(s+
							"/"+
							strings.Join(ipt.IndicesInclude, ",")+
							"/_stats?ignore_unavailable=true", clusterName); err != nil {
							l.Warn(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@"))
						}
					} else {
						if err := ipt.gatherIndicesStats(s+
							"/"+
							strings.Join(ipt.IndicesInclude, ",")+
							"/_stats?level=shards&ignore_unavailable=true", clusterName); err != nil {
							l.Warn(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@"))
						}
					}

					// get settings
					if err := ipt.gatherIndicesSettings(s, clusterName); err != nil {
						l.Warn(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@"))
					}
				}

				return nil
			})
		}(serv)
	}

	return g.Wait()
}

const (
	maxInterval = 1 * time.Minute
	minInterval = 1 * time.Second
)

func (ipt *Input) RunPipeline() {
	if ipt.Log == nil || len(ipt.Log.Files) == 0 {
		return
	}

	opt := &tailer.Option{
		Source:            inputName,
		Service:           inputName,
		Pipeline:          ipt.Log.Pipeline,
		IgnoreStatus:      ipt.Log.IgnoreStatus,
		CharacterEncoding: ipt.Log.CharacterEncoding,
		MultilinePatterns: []string{ipt.Log.MultilineMatch},
		GlobalTags:        inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, ""),
		Done:              ipt.semStop.Wait(),
	}

	var err error
	ipt.tail, err = tailer.NewTailer(ipt.Log.Files, opt)
	if err != nil {
		l.Error(err)
		ipt.feeder.FeedLastError(err.Error(),
			dkio.WithLastErrorInput(inputName),
		)
		return
	}
	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_elasticsearch"})
	g.Go(func(ctx context.Context) error {
		ipt.tail.Start()
		return nil
	})
}

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)

	duration, err := time.ParseDuration(ipt.Interval)
	if err != nil {
		l.Error("invalid interval, %s", err.Error())
		return
	} else if duration <= 0 {
		l.Error("invalid interval, cannot be less than zero")
		return
	}

	ipt.duration = config.ProtectedInterval(minInterval, maxInterval, duration)

	ipt.httpTimeout = Duration{}
	if len(ipt.HTTPTimeout) > 0 {
		err := ipt.httpTimeout.UnmarshalTOML([]byte(ipt.HTTPTimeout))
		if err != nil {
			l.Warnf("invalid http timeout, %s", ipt.HTTPTimeout)
		}
	}

	client, err := ipt.createHTTPClient()
	if err != nil {
		l.Error(err)
		ipt.feeder.FeedLastError(err.Error(),
			dkio.WithLastErrorInput(inputName),
		)
		return
	}
	ipt.client = client

	defer ipt.stop()

	tick := time.NewTicker(ipt.duration)
	defer tick.Stop()

	for {
		if ipt.pause {
			l.Debugf("not leader, skipped")
		} else {
			start := time.Now()
			if err := ipt.Collect(); err != nil {
				ipt.feeder.FeedLastError(err.Error(),
					dkio.WithLastErrorInput(inputName),
				)
				l.Error(err)
			} else if len(ipt.collectCache) > 0 {
				err := ipt.feeder.Feed(inputName, point.Metric, ipt.collectCache, &dkio.Option{CollectCost: time.Since(start)})
				if err != nil {
					ipt.feeder.FeedLastError(err.Error(),
						dkio.WithLastErrorInput(inputName),
					)
					l.Errorf(err.Error())
				}
				ipt.collectCache = ipt.collectCache[:0]
			}
		}
		select {
		case <-datakit.Exit.Wait():
			ipt.exit()
			l.Info("elasticsearch exit")
			return

		case <-ipt.semStop.Wait():
			ipt.exit()
			l.Info("elasticsearch return")
			return

		case <-tick.C:

		case ipt.pause = <-ipt.pauseCh:
			// nil
		}
	}
}

func (ipt *Input) exit() {
	if ipt.tail != nil {
		ipt.tail.Close()
		l.Info("elasticsearch log exit")
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (ipt *Input) gatherIndicesSettings(url string, clusterName string) error {
	settingResp := map[string]interface{}{}
	if err := ipt.gatherJSONData(url+"/"+strings.Join(ipt.IndicesInclude, ",")+"/_settings", &settingResp); err != nil {
		return err
	}

	for m, s := range settingResp {
		jsonParser := JSONFlattener{}
		if err := jsonParser.FullFlattenJSON("", s, true, true); err != nil {
			return err
		}
		allFields := make(map[string]interface{})
		for k, v := range jsonParser.Fields {
			fieldName := strings.TrimPrefix(k, "settings_")
			if _, ok := indicesStatsFields[fieldName]; ok {
				if vStr, ok := v.(string); ok {
					if vFloat, err := strconv.ParseFloat(vStr, 64); err == nil {
						allFields[fieldName] = vFloat
					} else {
						l.Warnf("get indices settings error, invalid float value: %s", vStr)
					}
				} else {
					allFields[fieldName] = v
				}
			}
		}

		tags := map[string]string{"index_name": m, "cluster_name": clusterName}
		setHostTagIfNotLoopback(tags, url)
		ipt.extendSelfTag(tags)

		metric := &indicesStatsMeasurement{
			elasticsearchMeasurement: elasticsearchMeasurement{
				name:     "elasticsearch_indices_stats",
				tags:     tags,
				fields:   allFields,
				ts:       time.Now(),
				election: ipt.Election,
			},
		}

		if len(metric.fields) > 0 {
			ipt.collectCache = append(ipt.collectCache, metric.Point())
		}
	}
	return nil
}

func (ipt *Input) gatherIndicesStats(url string, clusterName string) error {
	indicesStats := &struct {
		Shards  map[string]interface{} `json:"_shards"`
		All     map[string]interface{} `json:"_all"`
		Indices map[string]indexStat   `json:"indices"`
	}{}

	if err := ipt.gatherJSONData(url, indicesStats); err != nil {
		return err
	}
	now := time.Now()

	allFields := make(map[string]interface{})
	// All Stats
	for m, s := range indicesStats.All {
		// parse Json, ignoring strings and bools
		jsonParser := JSONFlattener{}
		err := jsonParser.FullFlattenJSON(m, s, true, true)
		if err != nil {
			return err
		}

		for k, v := range jsonParser.Fields {
			_, ok := indicesStatsFields[k]
			if ok {
				allFields[k] = v
			}
		}
	}

	tags := map[string]string{"index_name": "_all", "cluster_name": clusterName}
	setHostTagIfNotLoopback(tags, url)
	ipt.extendSelfTag(tags)

	metric := &indicesStatsMeasurement{
		elasticsearchMeasurement: elasticsearchMeasurement{
			name:     "elasticsearch_indices_stats",
			tags:     tags,
			fields:   allFields,
			ts:       now,
			election: ipt.Election,
		},
	}

	if len(metric.fields) > 0 {
		ipt.collectCache = append(ipt.collectCache, metric.Point())
	}

	// Individual Indices stats
	for id, index := range indicesStats.Indices {
		allFields := make(map[string]interface{})

		stats := map[string]interface{}{
			"primaries": index.Primaries,
			"total":     index.Total,
		}
		for m, s := range stats {
			f := JSONFlattener{}
			// parse Json, getting strings and bools
			err := f.FullFlattenJSON(m, s, true, true)
			if err != nil {
				return err
			}

			for k, v := range f.Fields {
				_, ok := indicesStatsFields[k]
				if ok {
					allFields[k] = v
				}
			}
		}

		indexTag := map[string]string{"index_name": id, "cluster_name": clusterName}
		setHostTagIfNotLoopback(indexTag, url)
		ipt.extendSelfTag(indexTag)

		metric := &indicesStatsMeasurement{
			elasticsearchMeasurement: elasticsearchMeasurement{
				name:     "elasticsearch_indices_stats",
				tags:     indexTag,
				fields:   allFields,
				ts:       now,
				election: ipt.Election,
			},
		}

		if len(metric.fields) > 0 {
			ipt.collectCache = append(ipt.collectCache, metric.Point())
		}
	}

	return nil
}

func (ipt *Input) gatherNodeStats(url string) (string, error) {
	nodeStats := &struct {
		ClusterName string               `json:"cluster_name"`
		Nodes       map[string]*nodeStat `json:"nodes"`
	}{}

	if err := ipt.gatherJSONData(url, nodeStats); err != nil {
		return "", err
	}

	for id, n := range nodeStats.Nodes {
		sort.Strings(n.Roles)
		tags := map[string]string{
			"node_id":      id,
			"node_host":    n.Host,
			"node_name":    n.Name,
			"cluster_name": nodeStats.ClusterName,
			"node_roles":   strings.Join(n.Roles, ","),
		}

		for k, v := range n.Attributes {
			tags["node_attribute_"+k] = v
		}

		stats := map[string]interface{}{
			"indices":     n.Indices,
			"os":          n.OS,
			"process":     n.Process,
			"jvm":         n.JVM,
			"thread_pool": n.ThreadPool,
			"fs":          n.FS,
			"transport":   n.Transport,
			"http":        n.HTTP,
			"breakers":    n.Breakers,
		}

		//nolint:lll
		const cols = `fs_total_available_in_bytes,fs_total_free_in_bytes,fs_total_total_in_bytes,fs_data_0_available_in_bytes,fs_data_0_free_in_bytes,fs_data_0_total_in_bytes`

		now := time.Now()
		allFields := make(map[string]interface{})
		for p, s := range stats {
			// if one of the individual node stats is not even in the
			// original result
			if s == nil {
				continue
			}
			f := JSONFlattener{}
			// parse Json, ignoring strings and bools
			err := f.FlattenJSON(p, s)
			if err != nil {
				return "", err
			}
			for k, v := range f.Fields {
				filedName := k
				val := v
				// transform bytes to gigabytes
				if p == "fs" {
					if strings.Contains(cols, filedName) {
						if value, ok := v.(float64); ok {
							val = value / (1024 * 1024 * 1024)
							filedName = strings.ReplaceAll(filedName, "in_bytes", "in_gigabytes")
						}
					}
				}
				_, ok := nodeStatsFields[filedName]
				if ok {
					allFields[filedName] = val
				}
			}
		}

		setHostTagIfNotLoopback(tags, url)
		ipt.extendSelfTag(tags)
		metric := &nodeStatsMeasurement{
			elasticsearchMeasurement: elasticsearchMeasurement{
				name:     "elasticsearch_node_stats",
				tags:     tags,
				fields:   allFields,
				ts:       now,
				election: ipt.Election,
			},
		}
		if len(metric.fields) > 0 {
			ipt.collectCache = append(ipt.collectCache, metric.Point())
		}
	}

	return nodeStats.ClusterName, nil
}

func (ipt *Input) gatherClusterStats(url string) error {
	clusterStats := &clusterStats{}
	if err := ipt.gatherJSONData(url, clusterStats); err != nil {
		return err
	}
	now := time.Now()
	tags := map[string]string{
		"node_name":    clusterStats.NodeName,
		"cluster_name": clusterStats.ClusterName,
		"status":       clusterStats.Status,
	}

	stats := map[string]interface{}{
		"nodes":   clusterStats.Nodes,
		"indices": clusterStats.Indices,
	}

	allFields := make(map[string]interface{})
	for p, s := range stats {
		f := JSONFlattener{}
		// parse json, including bools and strings
		err := f.FullFlattenJSON(p, s, true, true)
		if err != nil {
			return err
		}
		for k, v := range f.Fields {
			_, ok := clusterStatsFields[k]
			if ok {
				allFields[k] = v
			}
		}
	}

	setHostTagIfNotLoopback(tags, url)
	ipt.extendSelfTag(tags)
	metric := &clusterStatsMeasurement{
		elasticsearchMeasurement: elasticsearchMeasurement{
			name:     "elasticsearch_cluster_stats",
			tags:     tags,
			fields:   allFields,
			ts:       now,
			election: ipt.Election,
		},
	}

	if len(metric.fields) > 0 {
		ipt.collectCache = append(ipt.collectCache, metric.Point())
	}
	return nil
}

func (ipt *Input) isVersion6(url string) bool {
	serverInfo, ok := ipt.serverInfo[url]
	if !ok { // default
		return false
	} else {
		parts := strings.Split(serverInfo.version, ".")
		if len(parts) >= 2 {
			return parts[0] == "6"
		}
	}

	return false
}

func (ipt *Input) getLifeCycleErrorCount(url string) (errCount int) {
	errCount = 0
	// default elasticsearch
	if ipt.Distribution == "elasticsearch" || (len(ipt.Distribution) == 0) {
		// check privilege
		privilege := ipt.serverInfo[url].userPrivilege
		if privilege != nil {
			indexPrivilege, ok := privilege.Index["all"]
			if ok {
				if !indexPrivilege.Ilm {
					l.Warn("user has no ilm privilege, ingore collect indices_lifecycle_error_count")
					return 0
				}
			}
		}

		indicesRes := &indexState{}
		if ipt.isVersion6(url) { // 6.x
			if err := ipt.gatherJSONData(url+"/*/_ilm/explain", indicesRes); err != nil {
				l.Warn(err)
			} else {
				for _, index := range indicesRes.Indices {
					if index.Managed && index.Step == "ERROR" {
						errCount += 1
					}
				}
			}
		} else {
			if err := ipt.gatherJSONData(url+"/*/_ilm/explain?only_errors", indicesRes); err != nil {
				l.Warn(err)
			} else {
				errCount = len(indicesRes.Indices)
			}
		}
	}

	// opendistro or opensearch
	if ipt.Distribution == "opendistro" || ipt.Distribution == "opensearch" {
		res := map[string]interface{}{}
		pluginName := "_opendistro"

		if ipt.Distribution == "opensearch" {
			pluginName = "_plugins"
		}

		if err := ipt.gatherJSONData(url+"/"+pluginName+"/_ism/explain/*", &res); err != nil {
			l.Warn(err)
		} else {
			for _, index := range res {
				indexVal, ok := index.(map[string]interface{})
				if ok {
					if step, ok := indexVal["step"]; ok {
						if stepVal, ok := step.(map[string]interface{}); ok {
							if status, ok := stepVal["step_status"]; ok {
								if statusVal, ok := status.(string); ok {
									if statusVal == "failed" {
										errCount += 1
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return errCount
}

func (ipt *Input) gatherClusterHealth(url string, serverURL string) error {
	healthStats := &clusterHealth{}
	if err := ipt.gatherJSONData(url, healthStats); err != nil {
		return err
	}
	indicesErrorCount := ipt.getLifeCycleErrorCount(serverURL)
	now := time.Now()
	clusterFields := map[string]interface{}{
		"active_primary_shards":            healthStats.ActivePrimaryShards,
		"active_shards":                    healthStats.ActiveShards,
		"active_shards_percent_as_number":  healthStats.ActiveShardsPercentAsNumber,
		"delayed_unassigned_shards":        healthStats.DelayedUnassignedShards,
		"initializing_shards":              healthStats.InitializingShards,
		"number_of_data_nodes":             healthStats.NumberOfDataNodes,
		"number_of_in_flight_fetch":        healthStats.NumberOfInFlightFetch,
		"number_of_nodes":                  healthStats.NumberOfNodes,
		"number_of_pending_tasks":          healthStats.NumberOfPendingTasks,
		"relocating_shards":                healthStats.RelocatingShards,
		"status_code":                      mapHealthStatusToCode(healthStats.Status),
		"task_max_waiting_in_queue_millis": healthStats.TaskMaxWaitingInQueueMillis,
		"timed_out":                        healthStats.TimedOut,
		"unassigned_shards":                healthStats.UnassignedShards,
		"indices_lifecycle_error_count":    indicesErrorCount,
	}

	allFields := make(map[string]interface{})

	for k, v := range clusterFields {
		_, ok := clusterHealthFields[k]
		if ok {
			allFields[k] = v
		}
	}

	tags := map[string]string{
		"name":           healthStats.ClusterName, // deprecated, may be discarded in future
		"cluster_name":   healthStats.ClusterName,
		"cluster_status": healthStats.Status,
	}

	setHostTagIfNotLoopback(tags, url)
	ipt.extendSelfTag(tags)
	metric := &clusterHealthMeasurement{
		elasticsearchMeasurement: elasticsearchMeasurement{
			name:     "elasticsearch_cluster_health",
			tags:     tags,
			fields:   allFields,
			ts:       now,
			election: ipt.Election,
		},
	}

	if len(metric.fields) > 0 {
		ipt.collectCache = append(ipt.collectCache, metric.Point())
	}

	return nil
}

func (ipt *Input) gatherNodeID(url string) (string, error) {
	nodeStats := &struct {
		ClusterName string               `json:"cluster_name"`
		Nodes       map[string]*nodeStat `json:"nodes"`
	}{}
	if err := ipt.gatherJSONData(url, nodeStats); err != nil {
		return "", err
	}

	// Only 1 should be returned
	for id := range nodeStats.Nodes {
		return id, nil
	}
	return "", nil
}

func (ipt *Input) getVersion(url string) (string, error) {
	clusterInfo := &struct {
		Version struct {
			Number string `json:"number"`
		} `json:"version"`
	}{}
	if err := ipt.gatherJSONData(url, clusterInfo); err != nil {
		return "", err
	}

	return clusterInfo.Version.Number, nil
}

func (ipt *Input) getUserPrivilege(url string) *userPrivilege {
	privilege := &userPrivilege{}
	if ipt.Distribution == "elasticsearch" || len(ipt.Distribution) == 0 {
		body := strings.NewReader(`{"cluster": ["monitor"],"index":[{"names":["all"], "privileges":["monitor","manage_ilm"]}]}`)
		header := map[string]string{"Content-Type": "application/json"}
		if err := ipt.requestData("GET", url+"/_security/user/_has_privileges", header, body, privilege); err != nil {
			l.Warnf("get user privilege error: %s", err.Error())
			return nil
		}
	}

	return privilege
}

func (ipt *Input) nodeStatsURL(baseURL string) string {
	var url string

	if ipt.Local {
		url = baseURL + statsPathLocal
	} else {
		url = baseURL + statsPath
	}

	if len(ipt.NodeStats) == 0 {
		return url
	}

	return fmt.Sprintf("%s/%s", url, strings.Join(ipt.NodeStats, ","))
}

func (ipt *Input) stop() {
	ipt.client.CloseIdleConnections()
}

func (ipt *Input) createHTTPClient() (*http.Client, error) {
	timeout := 10 * time.Second
	if ipt.httpTimeout.Duration > 0 {
		timeout = ipt.httpTimeout.Duration
	}
	client := &http.Client{
		Timeout: timeout,
	}

	if ipt.TLSOpen {
		if ipt.InsecureSkipVerify {
			client.Transport = &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // nolint:gosec
			}
		} else {
			tc, err := TLSConfig(ipt.CacertFile, ipt.CertFile, ipt.KeyFile)
			if err != nil {
				return nil, err
			} else {
				client.Transport = &http.Transport{
					TLSClientConfig: tc,
				}
			}
		}
	} else {
		if len(ipt.Servers) > 0 {
			server := ipt.Servers[0]
			if strings.HasPrefix(server, "https://") {
				client.Transport = &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // nolint:gosec
				}
			}
		}
	}

	return client, nil
}

func (ipt *Input) requestData(method string, url string, header map[string]string, body io.Reader, v interface{}) error {
	m := "GET"
	if len(method) > 0 {
		m = method
	}
	req, err := http.NewRequest(m, url, body)
	for k, v := range header {
		req.Header.Add(k, v)
	}
	if err != nil {
		return err
	}

	if ipt.Username != "" || ipt.Password != "" {
		req.SetBasicAuth(ipt.Username, ipt.Password)
	}

	r, err := ipt.client.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close() //nolint:errcheck
	if r.StatusCode != http.StatusOK {
		// NOTE: we are not going to read/discard r.Body under the assumption we'd prefer
		// to let the underlying transport close the connection and re-establish a new one for
		// future calls.
		resBodyBytes, err := io.ReadAll(r.Body)
		resBody := ""
		if err != nil {
			l.Debugf("get response body err: %s", err.Error())
		} else {
			resBody = string(resBodyBytes)
		}

		l.Debugf("response body: %s", resBody)
		return fmt.Errorf("elasticsearch: API responded with status-code %d, expected %d, url: %s",
			r.StatusCode, http.StatusOK, mask.ReplaceAllString(url, "http(s)://XXX:XXX@"))
	}

	if err = json.NewDecoder(r.Body).Decode(v); err != nil {
		return err
	}

	return nil
}

func (ipt *Input) gatherJSONData(url string, v interface{}) error {
	return ipt.requestData("GET", url, nil, nil, v)
}

func (ipt *Input) getCatMaster(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	if ipt.Username != "" || ipt.Password != "" {
		req.SetBasicAuth(ipt.Username, ipt.Password)
	}

	r, err := ipt.client.Do(req)
	if err != nil {
		return "", err
	}
	defer r.Body.Close() //nolint:errcheck
	if r.StatusCode != http.StatusOK {
		// NOTE: we are not going to read/discard r.Body under the assumption we'd prefer
		// to let the underlying transport close the connection and re-establish a new one for
		// future calls.
		//nolint:lll
		return "", fmt.Errorf("elasticsearch: Unable to retrieve master node information. API responded with status-code %d, expected %d", r.StatusCode, http.StatusOK)
	}
	response, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}

	masterID := strings.Split(string(response), " ")[0]

	return masterID, nil
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

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
