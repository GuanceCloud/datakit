package yarn

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/antchfx/jsonquery"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type IoFeed func(data []byte, category, name string) error

type Metrcis struct {
	ClusterMetrics ClusterMetrics `json:"clusterMetrics"`
}
type ClusterMetrics struct {
	AppsSubmitted int `json:"appsSubmitted"`
	AppsCompleted int `json:"appsCompleted"`
	AppsPending   int `json:"appsPending"`
	AppsRunning   int `json:"appsRunning"`
	AppsFailed    int `json:"appsFailed"`
	AppsKilled    int `json:"appsKilled"`

	ReservedMB  int `json:"reservedMB"`
	AvailableMB int `json:"availableMB"`
	AllocatedMB int `json:"allocatedMB"`
	TotalMB     int `json:"totalMB"`

	ReservedVirtualCores  int `json:"reservedVirtualCores"`
	AvailableVirtualCores int `json:"availableVirtualCores"`
	AllocatedVirtualCores int `json:"allocatedVirtualCores"`
	TotalVirtualCores     int `json:"totalVirtualCores"`

	ContainersAllocated int `json:"containersAllocated"`
	ContainersReserved  int `json:"containersReserved"`
	ContainersPending   int `json:"containersPending"`

	TotalNodes          int `json:"totalNodes"`
	ActiveNodes         int `json:"activeNodes"`
	LostNodes           int `json:"lostNodes"`
	UnhealthyNodes      int `json:"unhealthyNodes"`
	DecommissionedNodes int `json:"decommissionedNodes"`
	RebootedNodes       int `json:"rebootedNodes"`
	ShutdownNodes       int `json:"shutdownNodes"`
}

type APP struct {
	Apps Apps `json:"apps"`
}

type Apps struct {
	App []AppItem `json:"app"`
}

type AppItem struct {
	Id   string `json:"id"`
	User string `json:"user"`
	Name string `json:"name"`

	Progress          int `json:"progress"`
	StartedTime       int `json:"startedTime"`
	FinishedTime      int `json:"finishedTime"`
	ElapsedTime       int `json:"elapsedTime"`
	AllocatedMB       int `json:"allocatedMB"`
	AllocatedVCores   int `json:"allocatedVCores"`
	RunningContainers int `json:"runningContainers"`
	MemorySeconds     int `json:"memorySeconds"`
	VcoreSeconds      int `json:"vcoreSeconds"`
}

type NODE struct {
	Nodes Nodes `json:"nodes"`
}

type Nodes struct {
	Node []NodeItem `json:"node"`
}

type NodeItem struct {
	Id string `json:"id"`

	LastHealthUpdate      int `json:"lastHealthUpdate"`
	UsedMemoryMB          int `json:"usedMemoryMB"`
	AvailMemoryMB         int `json:"availMemoryMB"`
	UsedVirtualCores      int `json:"usedVirtualCores"`
	AvailableVirtualCores int `json:"availableVirtualCores"`
	NumContainers         int `json:"numContainers"`
}

type Yarn struct {
	Interval    interface{}
	Active      bool
	Host        string
	MetricsName string
	Tags        map[string]string
	hostPath    string
}

type YarnInput struct {
	Yarn
}

type YarnOutput struct {
	IoFeed
}

type YarnParam struct {
	input  YarnInput
	output YarnOutput
	log    *logger.Logger
}

const (
	yarnConfigSample = `### You need to configure an [[inputs.yarn]] for each yarn to be monitored.
### interval: monitor interval, the default value is "60s".
### host: yarn service WebUI host, such as http://ip:port.
### metricsName: the name of metric, default is "yarn"

#[[inputs.yarn]]
#	interval    = "60s"
#	host        = "http://127.0.0.1:8088"
#	metricsName = "yarn"
#	[inputs.yarn.tags]
#		tag1 = "tag1"
#		tag2 = "tag2"
#		tagn = "tagn"

#[[inputs.yarn]]
#	interval    = "60s"
#	host        = "http://127.0.0.1:8088"
#	metricsName = "yarn"
#	[inputs.yarn.tags]
#		tag1 = "tag1"
#		tag2 = "tag2"
#		tagn = "tagn"
`

	inputName         = "yarn"
	defaultMetricName = inputName
	defaultInterval   = "60s"
	urlPrefix         = "/ws/v1/cluster/"
	host              = "host"
	canConect         = "can_connect"
	section           = "section"
	sectionMain       = "MAIN"
	sectionAPP        = "APP."
	sectionNode       = "NODE."
	sectionQueue      = "QUEUE."
)

type VAL_TYPE = int

const (
	INT VAL_TYPE = iota
	FLOAT
	STRING

	MaxGatherInterval = 30 * time.Minute
	MinGatherInterval = 1 * time.Second
)

func (y *Yarn) Catalog() string {
	return "yarn"
}

func (y *Yarn) SampleConfig() string {
	return yarnConfigSample
}

func (y *Yarn) Run() {
	if y.Host == "" {
		return
	}

	p := y.genParam()
	p.log.Infof("yarn input started...")
	p.gather()
}

func (y *Yarn) genParam() *YarnParam {
	if y.MetricsName != "" {
		y.MetricsName = defaultMetricName
	}

	if y.Interval == nil {
		y.Interval = defaultInterval
	}
	y.hostPath = strings.TrimRight(y.Host, "/") + urlPrefix

	input := YarnInput{*y}
	output := YarnOutput{io.NamedFeed}
	p := &YarnParam{input, output, logger.SLogger("yarn")}
	return p
}

func (p *YarnParam) gather() {
	var d time.Duration
	var err error

	switch p.input.Interval.(type) {
	case int64:
		d = time.Duration(p.input.Interval.(int64)) * time.Second
	case string:
		d, err = time.ParseDuration(p.input.Interval.(string))
		if err != nil {
			p.log.Errorf("parse interval err: %s", err.Error())
			return
		}
	default:
		p.log.Errorf("interval type unsupported")
		return
	}

	d = config.ProtectedInterval(MinGatherInterval, MaxGatherInterval, d)
	tick := time.NewTicker(d)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			_, err = p.getMetrics(false)
			if err != nil {
				io.FeedLastError(inputName, err.Error())
				p.log.Errorf("%s", err.Error())
			}
		case <-datakit.Exit.Wait():
			p.log.Info("input yarn exit")
			return
		}
	}
}

func (p *YarnParam) getMetrics(isTest bool) ([]byte, error) {
	var err error
	pt, err := p.gatherMainSection(isTest)
	if isTest || err != nil {
		return pt, err
	}

	err = p.gatherAppSection()
	if err != nil {
		return nil, err
	}

	err = p.gatherNodeSection()
	if err != nil {
		return nil, err
	}

	err = p.gatherQueueSection()
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (p *YarnParam) gatherMainSection(isTest bool) ([]byte, error) {
	var metric Metrcis
	tags := make(map[string]string)
	fields := make(map[string]interface{})

	tags[section] = sectionMain
	tags[host] = p.input.Host
	for tag, tagV := range p.input.Tags {
		tags[tag] = tagV
	}
	fields[canConect] = true

	resp, err := http.Get(p.input.hostPath + "metrics")
	if err != nil || resp.StatusCode != 200 {
		fields[canConect] = false
		pt, _ := io.MakeMetric(p.input.MetricsName, tags, fields, time.Now())
		if !isTest {
			p.output.IoFeed(pt, datakit.Metric, inputName)
		}
		return pt, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&metric); err != nil {
		return nil, err
	}
	fields["apps_submitted"] = metric.ClusterMetrics.AppsSubmitted
	fields["apps_completed"] = metric.ClusterMetrics.AppsCompleted
	fields["apps_pending"] = metric.ClusterMetrics.AppsPending
	fields["apps_running"] = metric.ClusterMetrics.AppsRunning
	fields["apps_failed"] = metric.ClusterMetrics.AppsFailed
	fields["apps_killed"] = metric.ClusterMetrics.AppsKilled

	fields["reserved_mb"] = metric.ClusterMetrics.ReservedMB
	fields["available_mb"] = metric.ClusterMetrics.AvailableMB
	fields["allocated_mb"] = metric.ClusterMetrics.AllocatedMB
	fields["total_mb"] = metric.ClusterMetrics.TotalMB

	fields["reserved_virtual_cores"] = metric.ClusterMetrics.ReservedVirtualCores
	fields["available_virtual_cores"] = metric.ClusterMetrics.AvailableVirtualCores
	fields["allocated_virtual_cores"] = metric.ClusterMetrics.AllocatedVirtualCores
	fields["total_virtual_cores"] = metric.ClusterMetrics.TotalVirtualCores

	fields["containers_allocated"] = metric.ClusterMetrics.ContainersAllocated
	fields["containers_reserved"] = metric.ClusterMetrics.ContainersReserved
	fields["containers_pending"] = metric.ClusterMetrics.ContainersPending

	fields["total_nodes"] = metric.ClusterMetrics.TotalNodes
	fields["active_nodes"] = metric.ClusterMetrics.ActiveNodes
	fields["lost_nodes"] = metric.ClusterMetrics.LostNodes
	fields["unhealthy_nodes"] = metric.ClusterMetrics.UnhealthyNodes
	fields["decommissioned_nodes"] = metric.ClusterMetrics.DecommissionedNodes
	fields["rebooted_nodes"] = metric.ClusterMetrics.RebootedNodes
	fields["shutdown_nodes"] = metric.ClusterMetrics.ShutdownNodes

	pt, err := io.MakeMetric(p.input.MetricsName, tags, fields, time.Now())
	if err != nil {
		return nil, err
	}

	if !isTest {
		err = p.output.IoFeed(pt, datakit.Metric, inputName)
	}
	return pt, err
}

func (p *YarnParam) gatherAppSection() error {
	var app APP
	var tags map[string]string
	var fields map[string]interface{}

	resp, err := http.Get(p.input.hostPath + "apps")
	if err != nil || resp.StatusCode != 200 {
		return err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&app); err != nil {
		return err
	}

	for _, ap := range app.Apps.App {
		tags = make(map[string]string)
		fields = make(map[string]interface{})
		tags[section] = sectionAPP + ap.Id
		tags[host] = p.input.Host
		for tag, tagV := range p.input.Tags {
			tags[tag] = tagV
		}

		fields["progress"] = ap.Progress
		fields["started_time"] = ap.StartedTime
		fields["finished_time"] = ap.FinishedTime
		fields["elapsed_time"] = ap.ElapsedTime
		fields["allocated_mb"] = ap.AllocatedMB
		fields["allocated_vcores"] = ap.AllocatedVCores
		fields["running_containers"] = ap.RunningContainers
		fields["memory_seconds"] = ap.MemorySeconds
		fields["vcore_seconds"] = ap.VcoreSeconds

		pt, err := io.MakeMetric(p.input.MetricsName, tags, fields, time.Now())
		if err != nil {
			return err
		}
		err = p.output.IoFeed(pt, datakit.Metric, inputName)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *YarnParam) gatherNodeSection() error {
	var nodes NODE
	var tags map[string]string
	var fields map[string]interface{}

	resp, err := http.Get(p.input.hostPath + "nodes")
	if err != nil || resp.StatusCode != 200 {
		return err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&nodes); err != nil {
		return err
	}

	for _, node := range nodes.Nodes.Node {
		tags = make(map[string]string)
		fields = make(map[string]interface{})
		tags[section] = sectionNode + node.Id
		tags[host] = p.input.Host
		for tag, tagV := range p.input.Tags {
			tags[tag] = tagV
		}

		fields["last_health_update"] = node.LastHealthUpdate
		fields["used_memory"] = node.UsedMemoryMB
		fields["avail_memory"] = node.AvailMemoryMB

		fields["used_virtual_cores"] = node.UsedVirtualCores
		fields["available_virtual_cores"] = node.AvailableVirtualCores
		fields["num_containers"] = node.NumContainers

		pt, err := io.MakeMetric(p.input.MetricsName, tags, fields, time.Now())
		if err != nil {
			return err
		}
		err = p.output.IoFeed(pt, datakit.Metric, inputName)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *YarnParam) gatherQueueSection() error {
	var tags map[string]string
	var fields map[string]interface{}

	resp, err := http.Get(p.input.hostPath + "scheduler")
	if err != nil || resp.StatusCode != 200 {
		return err
	}
	defer resp.Body.Close()

	doc, err := jsonquery.Parse(resp.Body)
	if err != nil {
		return err
	}

	nodes := jsonquery.Find(doc, "//queue/*[type]")
	for _, node := range nodes {
		tags = make(map[string]string)
		fields = make(map[string]interface{})
		tags[host] = p.input.Host
		for tag, tagV := range p.input.Tags {
			tags[tag] = tagV
		}

		if val, err := getQueueNodeVal(node, "type", STRING); err == nil {
			switch val.(type) {
			case string:
				if val.(string) != "capacitySchedulerLeafQueueInfo" {
					continue
				}
			default:
				continue
			}
		}

		if val, err := getQueueNodeVal(node, "queueName", STRING); err == nil {
			switch val.(type) {
			case string:
				tags[section] = sectionQueue + val.(string)
			}
		}

		if val, err := getQueueNodeVal(node, "numPendingApplications", INT); err == nil {
			fields["num_pending_applications"] = val
		}

		if val, err := getQueueNodeVal(node, "userAMResourceLimit/memory", INT); err == nil {
			fields["user_am_resource_limit_memory"] = val
		}

		if val, err := getQueueNodeVal(node, "userAMResourceLimit/vCores", INT); err == nil {
			fields["user_am_resource_limit_vcores"] = val
		}

		if val, err := getQueueNodeVal(node, "absoluteCapacity", FLOAT); err == nil {
			fields["absolute_capacity"] = val
		}

		if val, err := getQueueNodeVal(node, "userLimitFactor", FLOAT); err == nil {
			fields["user_limit_factor"] = val
		}

		if val, err := getQueueNodeVal(node, "userLimit", INT); err == nil {
			fields["user_limit"] = val
		}

		if val, err := getQueueNodeVal(node, "numApplications", INT); err == nil {
			fields["num_applications"] = val
		}

		if val, err := getQueueNodeVal(node, "usedAMResource/memory", INT); err == nil {
			fields["used_am_resource_memory"] = val
		}

		if val, err := getQueueNodeVal(node, "usedAMResource/vCores", INT); err == nil {
			fields["used_am_resource_vcores"] = val
		}

		if val, err := getQueueNodeVal(node, "absoluteUsedCapacity", FLOAT); err == nil {
			fields["absolute_used_capacity"] = val
		}

		if val, err := getQueueNodeVal(node, "resourcesUsed/memory", INT); err == nil {
			fields["resources_used_memory"] = val
		}

		if val, err := getQueueNodeVal(node, "resourcesUsed/vCores", INT); err == nil {
			fields["resources_used_vcores"] = val
		}

		if val, err := getQueueNodeVal(node, "AMResourceLimit/memory", INT); err == nil {
			fields["am_resource_limit_vcores"] = val
		}

		if val, err := getQueueNodeVal(node, "AMResourceLimit/vCores", INT); err == nil {
			fields["am_resource_limit_memory"] = val
		}

		if val, err := getQueueNodeVal(node, "capacity", FLOAT); err == nil {
			fields["capacity"] = val
		}

		if val, err := getQueueNodeVal(node, "numActiveApplications", INT); err == nil {
			fields["num_active_applications"] = val
		}

		if val, err := getQueueNodeVal(node, "absoluteMaxCapacity", FLOAT); err == nil {
			fields["absolute_max_capacity"] = val
		}

		if val, err := getQueueNodeVal(node, "usedCapacity", FLOAT); err == nil {
			fields["used_capacity"] = val
		}

		if val, err := getQueueNodeVal(node, "numContainers", INT); err == nil {
			fields["num_containers"] = val
		}

		if val, err := getQueueNodeVal(node, "maxCapacity", FLOAT); err == nil {
			fields["max_capacity"] = val
		}

		if val, err := getQueueNodeVal(node, "maxApplications", INT); err == nil {
			fields["max_applications"] = val
		}

		if val, err := getQueueNodeVal(node, "maxApplicationsPerUser", INT); err == nil {
			fields["max_applications_per_user"] = val
		}

		pt, err := io.MakeMetric(p.input.MetricsName, tags, fields, time.Now())
		if err != nil {
			return err
		}
		err = p.output.IoFeed(pt, datakit.Metric, inputName)
		if err != nil {
			return err
		}
	}
	return nil
}

func getQueueNodeVal(top *jsonquery.Node, expr string, valType VAL_TYPE) (i interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = fmt.Errorf(x)
			case error:
				err = x
			default:
				err = fmt.Errorf("Unknown Panic")
			}
		}
	}()

	s := jsonquery.FindOne(top, expr).InnerText()
	switch valType {
	case INT:
		i, err = strconv.Atoi(s)
	case FLOAT:
		i, err = strconv.ParseFloat(s, 32)
	case STRING:
		i = s
	default:
		err = fmt.Errorf("Unknown Type: %d", valType)
	}
	return
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		p := &Yarn{}
		return p
	})
}
