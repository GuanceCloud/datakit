package huaweiyunces

import (
	"context"
	"strings"
	"time"

	"golang.org/x/time/rate"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"

	ces "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1"
	iammodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/model"
)

const (
	sampleConfig = `
#[[inputs.huaweiyunces]]
# ##(required)
#access_key_id = ''
#access_key_secret = ''

# ##(optional) default is 5min
#interval = '5m'

# ##(optional) default is 1min, should not more then interval
#delay = '1m'

# ##(optional) control the frequency of huaweiyun api call in every second, the minimum is 1 and the maximum is 1000. default is 20
#api_frequency = 20

# ##(optional) specify the project ids to collect. default will apply to all projects
#projectids = [
#	'614439cb10ad4bdc9f3b0bc8xxx',
#	'214439cb10ad4bdc9f3b0bc8xxx'
#]

# ##(optional) defaultly will collect all available metrics, you can specify the metrics of namespaces
# ##each string specify the metric names of one namespace, separate by ':', if no metric name, collect all metrics of this namespace
# metrics = [
#'SYS.ECS',
#'SYS.OBS:download_bytes,upload_bytes',
# ]

# ##(optional) exclude some metrics that you not want
# exclude_metrics = [
#'SYS.ECS',
#'SYS.OBS:download_bytes,upload_bytes',
# ]

# ##(optional) mapping projectId to regionId, eg. mapping 0747621b760026a52f02c009e91xxxx to cn-north-1
# ##supported regionIds:
# ##"af-south-1"
# ##"cn-north-4"
# ##"cn-north-1"
# ##"cn-east-2"
# ##"cn-east-3"
# ##"cn-south-1"
# ##"cn-southwest-2"
# ##"ap-southeast-2"
# ##"ap-southeast-1"
# ##"ap-southeast-3"
#[inputs.huaweiyunces.project_regions]
#projectId1 = 'regionId1'

# ##(optional) custom tags
#[inputs.huaweiyunces.tags]
#key1 = 'val1'

# ##(optional) specify the collect option for some metrics
#[[inputs.huaweiyunces.property]]
#metric = 'SYS.ECS'
#interval = '10m'
#delay = '5m'
#period = 300
#filter = 'max'
#dimension = [
#     'instance_id,694244a4-659e-4931-8e72-9e90993xxxx'
#]
`
)

type (
	Property struct {
		Metric string `toml:"metric"`

		Name string `toml:"name"` //deprated

		Period   int              `toml:"period"`
		Filter   string           `toml:"filter"`
		Interval datakit.Duration `toml:"interval"`
		Delay    datakit.Duration `toml:"delay"`

		Dimension []string `toml:"dimension,omitempty"`

		Dimensions map[string]string `toml:"dimensions,omitempty"` //deprated

		Tags map[string]string `toml:"tags,omitempty"`

		namespace   string
		metricNames []string
		dimensions  []MetricsDimension
	}

	agent struct {
		AccessKeyID     string `toml:"access_key_id"`
		AccessKeySecret string `toml:"access_key_secret"`

		Interval datakit.Duration `toml:"interval"`
		Delay    datakit.Duration `toml:"delay"`

		ApiFrequency int `toml:"api_frequency"`

		IncludeProjectIDs []string `toml:"projectids,omitempty"`
		ExcludeProjectIDs []string `toml:"exclude_projectids,omitempty"`

		IncludeMetrics []string `toml:"metrics,omitempty"`
		ExcludeMetrics []string `toml:"exclude_metrics,omitempty"`

		Properties []*Property `toml:"property,omitempty"`

		EndPoint  string `toml:"endpoint"`  //deprated
		RegionID  string `toml:"region"`    //deprated
		ProjectID string `toml:"projectid"` //deprated

		Namespace []*Namespace `toml:"namespace"` //deprated

		ProjectRegions map[string]string `toml:"project_regions,omitempty"`

		Tags map[string]string `toml:"tags,omitempty"`

		ecsInstanceIDs []string          //deprated
		cesClient      *ces.CesClient    //deprated
		reqs           []*metricsRequest //deprated

		includeMetrics map[string][]string
		excludeMetrics map[string][]string

		clients []*cesCli

		limiter *rate.Limiter

		ctx       context.Context
		cancelFun context.CancelFunc

		mode string

		testError error

		reloadCloudTime time.Time
	}

	cesCli struct {
		proj iammodel.AuthProjectResult
		cli  *ces.CesClient

		requests map[string]*requestsOfNamespace
	}

	metricsRequest struct {
		tags map[string]string

		namespace  string
		metricname string

		originalMetricName string
		prefix             string

		period int
		filter string
		unit   string

		from int64
		to   int64

		//每个指标可单独配置interval，默认使用全局的配置
		interval time.Duration
		delay    time.Duration

		lastTime time.Time

		dimsOld []*Dimension //deprated

		dimensoions   []MetricsDimension
		fixDimensions bool //配置文件中显式指定了dimension
	}

	requestsOfNamespace struct {
		namespace string
		requests  map[string]*metricsRequest
	}
)

func (ag *agent) isTestOnce() bool {
	return ag.mode == "test"
}

func (ag *agent) isDebug() bool {
	return ag.mode == "debug"
}

func (ag *agent) parseConfig() {
	if len(ag.IncludeMetrics) > 0 {
		ag.includeMetrics = map[string][]string{}
		for _, item := range ag.IncludeMetrics {
			//每项格式为 namespace:metrcinames, AGT.ECS:cpu_util,disk_util...
			parts := strings.Split(item, ":")
			if len(parts) == 0 {
				continue
			}
			namespace := strings.TrimSpace(parts[0])
			if ag.includeMetrics[namespace] == nil {
				ag.includeMetrics[namespace] = []string{}
			}
			if len(parts) > 1 {
				metrics := strings.Split(parts[1], ",")
				for _, name := range metrics {
					ag.includeMetrics[namespace] = append(ag.includeMetrics[namespace], strings.TrimSpace(name))
				}
			}
		}
	}

	if len(ag.ExcludeMetrics) > 0 {
		ag.excludeMetrics = map[string][]string{}
		for _, item := range ag.ExcludeMetrics {
			//每项格式为 namespace:metrcinames, AGT.ECS:cpu_util,disk_util...
			parts := strings.Split(item, ":")
			if len(parts) == 0 {
				continue
			}
			namespace := strings.TrimSpace(parts[0])
			if ag.excludeMetrics[namespace] == nil {
				ag.excludeMetrics[namespace] = []string{}
			}
			if len(parts) > 1 {
				metrics := strings.Split(parts[1], ",")
				for _, name := range metrics {
					ag.excludeMetrics[namespace] = append(ag.excludeMetrics[namespace], strings.TrimSpace(name))
				}
			}
		}
	}

	for _, p := range ag.Properties {
		if p.Metric == "" {
			continue
		}
		parts := strings.Split(p.Metric, ":")
		p.namespace = strings.TrimSpace(parts[0])
		if len(parts) > 1 {
			p.metricNames = strings.Split(parts[1], ",")
		}

		for _, dm := range p.Dimension {
			projID := ""
			kv := dm
			parts := strings.Split(dm, ":")
			if len(parts) > 1 {
				projID = strings.TrimSpace(parts[1])
				kv = strings.TrimSpace(parts[0])
			}
			parts = strings.Split(kv, ",")
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if key == "" || value == "" {
				continue
			}
			p.dimensions = append(p.dimensions, MetricsDimension{
				Name:      &key,
				Value:     &value,
				projectID: projID,
			})
		}
	}
}

func (ag *agent) checkProjectIgnore(projectid string) bool {
	for _, pid := range ag.ExcludeProjectIDs {
		if projectid == pid {
			return true
		}
	}
	if len(ag.IncludeProjectIDs) > 0 {
		for _, pid := range ag.IncludeProjectIDs {
			if pid == projectid {
				return false
			}
		}
		return true
	}
	return false
}

func (ag *agent) checkMetricIgnore(namespace, metricname string) bool {

	if ag.excludeMetrics != nil {
		if names, ok := ag.excludeMetrics[namespace]; ok {
			if len(names) == 0 {
				return true
			}
			for _, name := range names {
				if name == metricname {
					return true
				}
			}
		}
	}

	if ag.includeMetrics != nil {
		if names, ok := ag.includeMetrics[namespace]; ok {
			if len(names) == 0 {
				return false
			}
			for _, name := range names {
				if name == metricname {
					return false
				}
			}
		}
		return true
	}

	return false
}

func (ag *agent) applyProperty(req *metricsRequest, projectID string) (fixDimensions bool) {
	fixDimensions = false
	var prop *Property
	for _, p := range ag.Properties {
		if p.namespace != req.namespace {
			continue
		}
		if len(p.metricNames) == 0 {
			prop = p
			break
		} else {
			for _, name := range p.metricNames {
				if name == req.metricname {
					prop = p
					break
				}
			}
		}
	}
	if prop == nil {
		return
	}
	if prop.Filter != "" {
		req.filter = prop.Filter
	}
	if prop.Interval.Duration > 0 {
		req.interval = prop.Interval.Duration
	}
	if prop.Delay.Duration > 0 {
		req.delay = prop.Delay.Duration
	}
	if prop.Period > 0 {
		req.period = prop.Period
	}
	if len(prop.dimensions) > 0 {
		for _, d := range prop.dimensions {
			if d.projectID != "" {
				if d.projectID != projectID {
					continue
				}
			}
			req.dimensoions = append(req.dimensoions, d)
		}
		fixDimensions = len(req.dimensoions) > 0
	}
	req.tags = prop.Tags

	return
}

func newMetricReq(namespace, metricname, unit string) *metricsRequest {

	req := &metricsRequest{
		namespace:  namespace,
		metricname: metricname,
		filter:     "average",
		period:     300,
	}

	return req
}
