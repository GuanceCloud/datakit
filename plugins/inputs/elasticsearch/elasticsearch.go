package elasticsearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var mask = regexp.MustCompile(`https?:\/\/\S+:\S+@`)

const statsPath = "/_nodes/stats"
const statsPathLocal = "/_nodes/_local/stats"

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

const sampleConfig = `
	[[inputs.elasticsearch]]
  ## specify a list of one or more Elasticsearch servers
  # you can add username and password to your url to use basic authentication:
  # servers = ["http://user:pass@localhost:9200"]
  servers = ["http://localhost:9200"]

	# valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"
	# required
	interval = "10s"

  ## Timeout for HTTP requests to the elastic search server(s)
  http_timeout = "5s"

  ## When local is true (the default), the node will read only its own stats.
  ## Set local to false when you want to read the node stats from all nodes
  ## of the cluster.
  local = true

  ## Set cluster_health to true when you want to also obtain cluster health stats
  cluster_health = false

  ## Adjust cluster_health_level when you want to also obtain detailed health stats
  ## The options are
  ##  - indices (default)
  ##  - cluster
  # cluster_health_level = "indices"

  ## Set cluster_stats to true when you want to also obtain cluster stats.
  cluster_stats = false

  ## Only gather cluster_stats from the master node. To work this require local = true
  cluster_stats_only_from_master = true

  ## Indices to collect; can be one or more indices names or _all
  indices_include = ["_all"]

  ## One of "shards", "cluster", "indices"
  indices_level = "shards"

  ## node_stats is a list of sub-stats that you want to have gathered. Valid options
  ## are "indices", "os", "process", "jvm", "thread_pool", "fs", "transport", "http",
  ## "breaker". Per default, all stats are gathered.
  # node_stats = ["jvm", "http"]

  ## HTTP Basic Authentication username and password.
  # username = ""
  # password = ""

  ## Optional TLS Config
	tls_open = false
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
`

type Elasticsearch struct {
	Interval                   string   `toml:"interval"`
	Local                      bool     `toml:"local"`
	Servers                    []string `toml:"servers"`
	HTTPTimeout                Duration `toml:"http_timeout"`
	ClusterHealth              bool     `toml:"cluster_health"`
	ClusterHealthLevel         string   `toml:"cluster_health_level"`
	ClusterStats               bool     `toml:"cluster_stats"`
	ClusterStatsOnlyFromMaster bool     `toml:"cluster_stats_only_from_master"`
	IndicesInclude             []string `toml:"indices_include"`
	IndicesLevel               string   `toml:"indices_level"`
	NodeStats                  []string `toml:"node_stats"`
	Username                   string   `toml:"username"`
	Password                   string   `toml:"password"`

	TLSOpen    bool   `toml:"tls_open"`
	CacertFile string `toml:"tls_ca"`
	CertFile   string `toml:"tls_cert"`
	KeyFile    string `toml:"tls_key"`

	client          *http.Client
	serverInfo      map[string]serverInfo
	serverInfoMutex sync.Mutex
	log             *logger.Logger
	duration        time.Duration
}

type serverInfo struct {
	nodeID   string
	masterID string
}

type metric struct {
	Measurement string
	Fields      map[string]interface{}
	Tags        map[string]string
	Time        time.Time
}

var metrics []metric

func addFields(measurement string, fields map[string]interface{}, tags map[string]string, t ...time.Time) {
	var tm time.Time
	if len(t) > 0 {
		tm = t[0]
	} else {
		tm = time.Now().UTC()
	}

	var metricNew = metric{
		Measurement: measurement,
		Fields:      fields,
		Tags:        tags,
		Time:        tm,
	}
	metrics = append(metrics, metricNew)
}

func (i serverInfo) isMaster() bool {
	return i.nodeID == i.masterID
}

func NewElasticsearch() *Elasticsearch {
	return &Elasticsearch{
		HTTPTimeout:                Duration{Duration: time.Second * 5},
		ClusterStatsOnlyFromMaster: true,
		ClusterHealthLevel:         "indices",
	}
}

// perform status mapping
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

// perform shard status mapping
func mapShardStatusToCode(s string) int {
	switch strings.ToUpper(s) {
	case "UNASSIGNED":
		return 1
	case "INITIALIZING":
		return 2
	case "STARTED":
		return 3
	case "RELOCATING":
		return 4
	}
	return 0
}

var (
	inputName = "elasticsearch"
)

func (_ *Elasticsearch) Catalog() string {
	return inputName
}

func (_ *Elasticsearch) SampleConfig() string {
	return sampleConfig
}

func (_ *Elasticsearch) Test() (result *inputs.TestResult, err error) {
	return
}

func (e *Elasticsearch) Run() {
	e.log = logger.SLogger(inputName)
	duration, err := time.ParseDuration(e.Interval)
	if err != nil {
		e.log.Error(fmt.Errorf("invalid interval, %s", err.Error()))
		return
	} else if duration <= 0 {
		e.log.Error(fmt.Errorf("invalid interval, cannot be less than zero"))
		return
	}
	e.duration = duration
	client, err := e.createHTTPClient()
	if err != nil {
		e.log.Error(err)
		return
	}
	e.client = client

	defer e.stop()

	ticker := time.NewTicker(e.duration)
	defer ticker.Stop()

	for {
		select {
		case <-datakit.Exit.Wait():
			e.log.Info("elasticsearch exit")
			return

		case <-ticker.C:
			data, err := e.getMetrics()
			if err != nil {
				e.log.Error(err)
				continue
			}
			if err := io.NamedFeed(data, io.Metric, inputName); err != nil {
				e.log.Error(err)
				continue
			}
			e.log.Debugf("feed %d bytes to io ok", len(data))
		}
	}
}

func (e *Elasticsearch) getMetrics() ([]byte, error) {
	var buffer = bytes.Buffer{}

	if e.ClusterStats || len(e.IndicesInclude) > 0 || len(e.IndicesLevel) > 0 {
		var wgC sync.WaitGroup
		wgC.Add(len(e.Servers))

		e.serverInfo = make(map[string]serverInfo)
		for _, serv := range e.Servers {
			go func(s string) {
				defer wgC.Done()
				info := serverInfo{}

				var err error

				// Gather node ID
				if info.nodeID, err = e.gatherNodeID(s + "/_nodes/_local/name"); err != nil {
					e.log.Error(fmt.Errorf(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@")))
					return
				}

				// get cat/master information here so NodeStats can determine
				// whether this node is the Master
				if info.masterID, err = e.getCatMaster(s + "/_cat/master"); err != nil {
					e.log.Error(fmt.Errorf(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@")))
					return
				}

				e.serverInfoMutex.Lock()
				e.serverInfo[s] = info
				e.serverInfoMutex.Unlock()

			}(serv)
		}
		wgC.Wait()
	}

	var wg sync.WaitGroup
	wg.Add(len(e.Servers))

	for _, serv := range e.Servers {
		go func(s string) {
			defer wg.Done()
			url := e.nodeStatsURL(s)

			// Always gather node stats
			if err := e.gatherNodeStats(url); err != nil {
				e.log.Error(fmt.Errorf(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@")))
				return
			}

			if e.ClusterHealth {
				url = s + "/_cluster/health"
				if e.ClusterHealthLevel != "" {
					url = url + "?level=" + e.ClusterHealthLevel
				}
				if err := e.gatherClusterHealth(url); err != nil {
					e.log.Error(fmt.Errorf(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@")))
					return
				}
			}

			if e.ClusterStats && (e.serverInfo[s].isMaster() || !e.ClusterStatsOnlyFromMaster || !e.Local) {
				if err := e.gatherClusterStats(s + "/_cluster/stats"); err != nil {
					e.log.Error(fmt.Errorf(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@")))
					return
				}
			}

			if len(e.IndicesInclude) > 0 && (e.serverInfo[s].isMaster() || !e.ClusterStatsOnlyFromMaster || !e.Local) {
				if e.IndicesLevel != "shards" {
					if err := e.gatherIndicesStats(s + "/" + strings.Join(e.IndicesInclude, ",") + "/_stats"); err != nil {
						e.log.Error(fmt.Errorf(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@")))
						return
					}
				} else {
					if err := e.gatherIndicesStats(s + "/" + strings.Join(e.IndicesInclude, ",") + "/_stats?level=shards"); err != nil {
						e.log.Error(fmt.Errorf(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@")))
						return
					}
				}
			}
		}(serv)
	}

	wg.Wait()

	for _, metricItem := range metrics {
		data, err := io.MakeMetric(metricItem.Measurement, metricItem.Tags, metricItem.Fields, metricItem.Time)
		if err != nil {
			continue
		}
		buffer.Write(data)
		buffer.WriteString("\n")
	}

	return buffer.Bytes(), nil
}

func (e *Elasticsearch) gatherIndicesStats(url string) error {
	indicesStats := &struct {
		Shards  map[string]interface{} `json:"_shards"`
		All     map[string]interface{} `json:"_all"`
		Indices map[string]indexStat   `json:"indices"`
	}{}

	if err := e.gatherJSONData(url, indicesStats); err != nil {
		return err
	}
	now := time.Now()

	// Total Shards Stats
	shardsStats := map[string]interface{}{}
	for k, v := range indicesStats.Shards {
		shardsStats[k] = v
	}
	addFields("elasticsearch_indices_stats_shards_total", shardsStats, map[string]string{}, now)

	// All Stats
	for m, s := range indicesStats.All {
		// parse Json, ignoring strings and bools
		jsonParser := JSONFlattener{}
		err := jsonParser.FullFlattenJSON("_", s, true, true)
		if err != nil {
			return err
		}
		addFields("elasticsearch_indices_stats_"+m, jsonParser.Fields, map[string]string{"index_name": "_all"}, now)
	}

	// Individual Indices stats
	for id, index := range indicesStats.Indices {
		indexTag := map[string]string{"index_name": id}
		stats := map[string]interface{}{
			"primaries": index.Primaries,
			"total":     index.Total,
		}
		for m, s := range stats {
			f := JSONFlattener{}
			// parse Json, getting strings and bools
			err := f.FullFlattenJSON("", s, true, true)
			if err != nil {
				return err
			}
			addFields("elasticsearch_indices_stats_"+m, f.Fields, indexTag, now)
		}

		if e.IndicesLevel == "shards" {
			for shardNumber, shards := range index.Shards {
				for _, shard := range shards {

					// Get Shard Stats
					flattened := JSONFlattener{}
					err := flattened.FullFlattenJSON("", shard, true, true)
					if err != nil {
						return err
					}

					// determine shard tag and primary/replica designation
					shardType := "replica"
					if flattened.Fields["routing_primary"] == true {
						shardType = "primary"
					}
					delete(flattened.Fields, "routing_primary")

					routingState, ok := flattened.Fields["routing_state"].(string)
					if ok {
						flattened.Fields["routing_state"] = mapShardStatusToCode(routingState)
					}

					routingNode, _ := flattened.Fields["routing_node"].(string)
					shardTags := map[string]string{
						"index_name": id,
						"node_id":    routingNode,
						"shard_name": string(shardNumber),
						"type":       shardType,
					}

					for key, field := range flattened.Fields {
						switch field.(type) {
						case string, bool:
							delete(flattened.Fields, key)
						}
					}

					addFields("elasticsearch_indices_stats_shards",
						flattened.Fields,
						shardTags,
						now)
				}
			}
		}
	}

	return nil
}

func (e *Elasticsearch) gatherNodeStats(url string) error {
	nodeStats := &struct {
		ClusterName string               `json:"cluster_name"`
		Nodes       map[string]*nodeStat `json:"nodes"`
	}{}

	if err := e.gatherJSONData(url, nodeStats); err != nil {
		return err
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

		now := time.Now()
		for p, s := range stats {
			// if one of the individual node stats is not even in the
			// original result
			if s == nil {
				continue
			}
			f := JSONFlattener{}
			// parse Json, ignoring strings and bools
			err := f.FlattenJSON("", s)
			if err != nil {
				return err
			}
			addFields("elasticsearch_"+p, f.Fields, tags, now)
		}
	}
	return nil
}

func (e *Elasticsearch) gatherClusterStats(url string) error {
	clusterStats := &clusterStats{}
	if err := e.gatherJSONData(url, clusterStats); err != nil {
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

	for p, s := range stats {
		f := JSONFlattener{}
		// parse json, including bools and strings
		err := f.FullFlattenJSON("", s, true, true)
		if err != nil {
			return err
		}
		addFields("elasticsearch_clusterstats_"+p, f.Fields, tags, now)
	}

	return nil
}

func (e *Elasticsearch) gatherClusterHealth(url string) error {
	healthStats := &clusterHealth{}
	if err := e.gatherJSONData(url, healthStats); err != nil {
		return err
	}
	measurementTime := time.Now()
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
		"status":                           healthStats.Status,
		"status_code":                      mapHealthStatusToCode(healthStats.Status),
		"task_max_waiting_in_queue_millis": healthStats.TaskMaxWaitingInQueueMillis,
		"timed_out":                        healthStats.TimedOut,
		"unassigned_shards":                healthStats.UnassignedShards,
	}
	addFields(
		"elasticsearch_cluster_health",
		clusterFields,
		map[string]string{"name": healthStats.ClusterName},
		measurementTime,
	)

	for name, health := range healthStats.Indices {
		indexFields := map[string]interface{}{
			"active_primary_shards": health.ActivePrimaryShards,
			"active_shards":         health.ActiveShards,
			"initializing_shards":   health.InitializingShards,
			"number_of_replicas":    health.NumberOfReplicas,
			"number_of_shards":      health.NumberOfShards,
			"relocating_shards":     health.RelocatingShards,
			"status":                health.Status,
			"status_code":           mapHealthStatusToCode(health.Status),
			"unassigned_shards":     health.UnassignedShards,
		}
		addFields(
			"elasticsearch_cluster_health_indices",
			indexFields,
			map[string]string{"index": name, "name": healthStats.ClusterName},
			measurementTime,
		)
	}
	return nil
}

func (e *Elasticsearch) gatherNodeID(url string) (string, error) {
	nodeStats := &struct {
		ClusterName string               `json:"cluster_name"`
		Nodes       map[string]*nodeStat `json:"nodes"`
	}{}
	if err := e.gatherJSONData(url, nodeStats); err != nil {
		return "", err
	}

	// Only 1 should be returned
	for id := range nodeStats.Nodes {
		return id, nil
	}
	return "", nil
}

func (e *Elasticsearch) nodeStatsURL(baseURL string) string {
	var url string

	if e.Local {
		url = baseURL + statsPathLocal
	} else {
		url = baseURL + statsPath
	}

	if len(e.NodeStats) == 0 {
		return url
	}

	return fmt.Sprintf("%s/%s", url, strings.Join(e.NodeStats, ","))
}

func (e *Elasticsearch) stop() {
	e.client.CloseIdleConnections()
}

func (e *Elasticsearch) createHTTPClient() (*http.Client, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	if e.TLSOpen {
		tc, err := TLSConfig(e.CacertFile, e.CertFile, e.KeyFile)
		if err != nil {
			return nil, err
		} else {
			e.client.Transport = &http.Transport{
				TLSClientConfig: tc,
			}
		}
	}

	return client, nil
}

func (e *Elasticsearch) gatherJSONData(url string, v interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	if e.Username != "" || e.Password != "" {
		req.SetBasicAuth(e.Username, e.Password)
	}

	r, err := e.client.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		// NOTE: we are not going to read/discard r.Body under the assumption we'd prefer
		// to let the underlying transport close the connection and re-establish a new one for
		// future calls.
		return fmt.Errorf("elasticsearch: API responded with status-code %d, expected %d",
			r.StatusCode, http.StatusOK)
	}

	if err = json.NewDecoder(r.Body).Decode(v); err != nil {
		return err
	}

	return nil
}

func (e *Elasticsearch) getCatMaster(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	if e.Username != "" || e.Password != "" {
		req.SetBasicAuth(e.Username, e.Password)
	}

	r, err := e.client.Do(req)
	if err != nil {
		return "", err
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		// NOTE: we are not going to read/discard r.Body under the assumption we'd prefer
		// to let the underlying transport close the connection and re-establish a new one for
		// future calls.
		return "", fmt.Errorf("elasticsearch: Unable to retrieve master node information. API responded with status-code %d, expected %d", r.StatusCode, http.StatusOK)
	}
	response, err := ioutil.ReadAll(r.Body)

	if err != nil {
		return "", err
	}

	masterID := strings.Split(string(response), " ")[0]

	return masterID, nil
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return NewElasticsearch()
	})
}
