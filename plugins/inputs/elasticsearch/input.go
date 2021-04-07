package elasticsearch

import (
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
	## log file path
	#logFiles = ["/path/to/your/file.log"]

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

type Input struct {
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
	LogFiles           				 []string  `toml:"logfiles"`

	TLSOpen    bool   `toml:"tls_open"`
	CacertFile string `toml:"tls_ca"`
	CertFile   string `toml:"tls_cert"`
	KeyFile    string `toml:"tls_key"`

	client          *http.Client
	serverInfo      map[string]serverInfo
	serverInfoMutex sync.Mutex
	log             *logger.Logger
	duration        time.Duration

	collectCache []inputs.Measurement
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

func (i serverInfo) isMaster() bool {
	return i.nodeID == i.masterID
}

func NewElasticsearch() *Input {
	return &Input{
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
	l         = logger.DefaultSLogger("demo")
)

func (_ *Input) Catalog() string {
	return inputName
}

func (_ *Input) SampleConfig() string {
	return sampleConfig
}

func (_ *Input) Test() {
	return
}

type elasticsearchMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *elasticsearchMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *elasticsearchMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "elasticsearch",
		Fields: map[string]*inputs.FieldInfo{
			"active_shards_percent_as_number": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "active shards percent"},
			"active_primary_shards":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "active primary shards"},
			"status":                          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "status"},
			"timed_out":                       &inputs.FieldInfo{DataType: inputs.Bool, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "timed_out"},
		},
	}
}

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&elasticsearchMeasurement{},
	}
}

func (i *Input) addFields(measurement string, fields map[string]interface{}, tags map[string]string, t ...time.Time) {
	var tm time.Time
	if len(t) > 0 {
		tm = t[0]
	} else {
		tm = time.Now()
	}

	i.collectCache = append(i.collectCache, &elasticsearchMeasurement{
		name:   measurement,
		tags:   tags,
		fields: fields,
		ts:     tm,
	})
}

func (i *Input) Collect() error {
	// 获取nodeID和masterID
	if i.ClusterStats || len(i.IndicesInclude) > 0 || len(i.IndicesLevel) > 0 {
		var wgC sync.WaitGroup
		wgC.Add(len(i.Servers))

		i.serverInfo = make(map[string]serverInfo)
		for _, serv := range i.Servers {
			go func(s string) {
				defer wgC.Done()
				info := serverInfo{}

				var err error

				// Gather node ID
				if info.nodeID, err = i.gatherNodeID(s + "/_nodes/_local/name"); err != nil {
					l.Error(fmt.Errorf(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@")))
					return
				}

				// get cat/master information here so NodeStats can determine
				// whether this node is the Master
				if info.masterID, err = i.getCatMaster(s + "/_cat/master"); err != nil {
					l.Error(fmt.Errorf(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@")))
					return
				}

				i.serverInfoMutex.Lock()
				i.serverInfo[s] = info
				i.serverInfoMutex.Unlock()

			}(serv)
		}
		wgC.Wait()
	}

	var wg sync.WaitGroup
	wg.Add(len(i.Servers))

	for _, serv := range i.Servers {
		go func(s string) {
			defer wg.Done()
			url := i.nodeStatsURL(s)

			// Always gather node stats
			if err := i.gatherNodeStats(url); err != nil {
				l.Error(fmt.Errorf(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@")))
				return
			}

			if i.ClusterHealth {
				url = s + "/_cluster/health"
				if i.ClusterHealthLevel != "" {
					url = url + "?level=" + i.ClusterHealthLevel
				}
				if err := i.gatherClusterHealth(url); err != nil {
					l.Error(fmt.Errorf(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@")))
					return
				}
			}

			if i.ClusterStats && (i.serverInfo[s].isMaster() || !i.ClusterStatsOnlyFromMaster || !i.Local) {
				if err := i.gatherClusterStats(s + "/_cluster/stats"); err != nil {
					l.Error(fmt.Errorf(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@")))
					return
				}
			}

			if len(i.IndicesInclude) > 0 && (i.serverInfo[s].isMaster() || !i.ClusterStatsOnlyFromMaster || !i.Local) {
				if i.IndicesLevel != "shards" {
					if err := i.gatherIndicesStats(s + "/" + strings.Join(i.IndicesInclude, ",") + "/_stats"); err != nil {
						l.Error(fmt.Errorf(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@")))
						return
					}
				} else {
					if err := i.gatherIndicesStats(s + "/" + strings.Join(i.IndicesInclude, ",") + "/_stats?level=shards"); err != nil {
						l.Error(fmt.Errorf(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@")))
						return
					}
				}
			}
		}(serv)
	}

	wg.Wait()

	return nil
}

func (i *Input) Run() {
	go func () {
		option := inputs.TailerOption {
			Source: inputName,
			Service: inputName
		}
		tailer, err := inputs.NewTailer(inputName, i.LogFiles, &option)
		if err != nil {
			l.Error(err)
			return
		}
	
		tailer.Run()
	}()


	duration, err := time.ParseDuration(i.Interval)
	if err != nil {
		l.Error(fmt.Errorf("invalid interval, %s", err.Error()))
		return
	} else if duration <= 0 {
		l.Error(fmt.Errorf("invalid interval, cannot be less than zero"))
		return
	}

	i.duration = duration
	client, err := i.createHTTPClient()
	if err != nil {
		l.Error(err)
		return
	}
	i.client = client

	defer i.stop()

	tick := time.NewTicker(i.duration)
	defer tick.Stop()

	n := 0

	for {
		n++
		select {
		case <-datakit.Exit.Wait():
			l.Info("elasticsearch exit")
			return

		case <-tick.C:
			l.Debugf("elasticsearchs input gathering...")
			start := time.Now()
			if err := i.Collect(); err != nil {
				l.Error(err)
			} else {
				inputs.FeedMeasurement("elasticsearch", io.Metric, i.collectCache, &io.Option{CollectCost: time.Since(start), HighFreq: (n%2 == 0)})
				i.collectCache = i.collectCache[:0]
			}
		}
	}
}

func (i *Input) gatherIndicesStats(url string) error {
	indicesStats := &struct {
		Shards  map[string]interface{} `json:"_shards"`
		All     map[string]interface{} `json:"_all"`
		Indices map[string]indexStat   `json:"indices"`
	}{}

	if err := i.gatherJSONData(url, indicesStats); err != nil {
		return err
	}
	now := time.Now()

	// Total Shards Stats
	shardsStats := map[string]interface{}{}
	for k, v := range indicesStats.Shards {
		shardsStats[k] = v
	}
	i.addFields("elasticsearch_indices_stats_shards_total", shardsStats, map[string]string{}, now)

	// All Stats
	for m, s := range indicesStats.All {
		// parse Json, ignoring strings and bools
		jsonParser := JSONFlattener{}
		err := jsonParser.FullFlattenJSON(m+"_", s, true, true)
		if err != nil {
			return err
		}
		i.addFields("elasticsearch_indices_stats", jsonParser.Fields, map[string]string{"index_name": "_all"}, now)
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
			err := f.FullFlattenJSON(m, s, true, true)
			if err != nil {
				return err
			}
			i.addFields("elasticsearch_indices_stats", f.Fields, indexTag, now)
		}

		if i.IndicesLevel == "shards" {
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

					i.addFields("elasticsearch_indices_stats_shards",
						flattened.Fields,
						shardTags,
						now)
				}
			}
		}
	}

	return nil
}

func (i *Input) gatherNodeStats(url string) error {
	nodeStats := &struct {
		ClusterName string               `json:"cluster_name"`
		Nodes       map[string]*nodeStat `json:"nodes"`
	}{}

	if err := i.gatherJSONData(url, nodeStats); err != nil {
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
				return err
			}
			for k, v := range f.Fields {
				allFields[k] = v
			}
		}
		i.addFields("elasticsearch_node_stats", allFields, tags, now)
	}
	return nil
}

func (i *Input) gatherClusterStats(url string) error {
	clusterStats := &clusterStats{}
	if err := i.gatherJSONData(url, clusterStats); err != nil {
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
		for k, v := range(f.Fields) {
			allFields[k] = v
		}

	}
	i.addFields("elasticsearch_cluster_stats", allFields, tags, now)

	return nil
}

func (i *Input) gatherClusterHealth(url string) error {
	healthStats := &clusterHealth{}
	if err := i.gatherJSONData(url, healthStats); err != nil {
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
	i.addFields(
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
		i.addFields(
			"elasticsearch_cluster_health_indices",
			indexFields,
			map[string]string{"index": name, "name": healthStats.ClusterName},
			measurementTime,
		)
	}
	return nil
}

func (i *Input) gatherNodeID(url string) (string, error) {
	nodeStats := &struct {
		ClusterName string               `json:"cluster_name"`
		Nodes       map[string]*nodeStat `json:"nodes"`
	}{}
	if err := i.gatherJSONData(url, nodeStats); err != nil {
		return "", err
	}

	// Only 1 should be returned
	for id := range nodeStats.Nodes {
		return id, nil
	}
	return "", nil
}

func (i *Input) nodeStatsURL(baseURL string) string {
	var url string

	if i.Local {
		url = baseURL + statsPathLocal
	} else {
		url = baseURL + statsPath
	}

	if len(i.NodeStats) == 0 {
		return url
	}

	return fmt.Sprintf("%s/%s", url, strings.Join(i.NodeStats, ","))
}

func (i *Input) stop() {
	i.client.CloseIdleConnections()
}

func (i *Input) createHTTPClient() (*http.Client, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	if i.TLSOpen {
		tc, err := TLSConfig(i.CacertFile, i.CertFile, i.KeyFile)
		if err != nil {
			return nil, err
		} else {
			i.client.Transport = &http.Transport{
				TLSClientConfig: tc,
			}
		}
	}

	return client, nil
}

func (i *Input) gatherJSONData(url string, v interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	if i.Username != "" || i.Password != "" {
		req.SetBasicAuth(i.Username, i.Password)
	}

	r, err := i.client.Do(req)
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

func (i *Input) getCatMaster(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	if i.Username != "" || i.Password != "" {
		req.SetBasicAuth(i.Username, i.Password)
	}

	r, err := i.client.Do(req)
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
