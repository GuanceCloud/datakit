package aliyuncms

import (
	"encoding/json"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/cms"
)

const (
	aliyuncmsConfigSample = `
#[[cms]]
# ## Aliyun Credentials (required)
#access_key_id = ''
#access_key_secret = ''

# ## Aliyun Region (required)
# ## See: https://www.alibabacloud.com/help/zh/doc-detail/40654.htm
#region_id = 'cn-hangzhou'

#[[cms.project]]
#	## Metric Statistic Project (required)
#	name='acs_ecs_dashboard'

#	## Optional instances from which you want to pull metrics, empty means to pull all instances 
#	#instanceIds=["xxx","yyy"]

#	## Metrics to Pull (Required)
#	## See: https://help.aliyun.com/document_detail/28619.html?spm=a2c4g.11186623.2.11.6ac47694AjhHt4
#	[cms.project.metrics]
#	names = [
#		'CPUUtilization',
#	]

#	## dimensions can be used to query the specified resource, which is a collection of key-value forms. 
#	## each metric may have its own dimensions, See: https://help.aliyun.com/document_detail/28619.html?spm=a2c4g.11186623.2.11.6ac47694AjhHt4 
#	## name is metric name, value is a json string
#	[[cms.project.metrics.dimensions]]
#		name = "CPUUtilization"
#		value = '''
#		[
#		{"instanceId":"i-bp15wj5w33t8vfxi****"},
#		{"instanceId":"i-bp1bq3x84ko4ct6x****"}
#		]
#		'''
`
)

var (
	MetricsRequests = []*MetricsRequest{}
)

type (
	Dimension struct {
		Name  string `toml:"name"`
		Value string `toml:"value"`
	}

	Metric struct {
		MetricNames []string     `toml:"names"`
		Dimensions  []*Dimension `toml:"dimensions"`
	}

	Project struct {
		Name        string   `toml:"name"`
		InstanceIDs []string `toml:"instanceIds"`
		Metrics     *Metric  `toml:"metrics"`
	}

	CMS struct {
		RegionID        string     `toml:"region_id"`
		AccessKeyID     string     `toml:"access_key_id"`
		AccessKeySecret string     `toml:"access_key_secret"`
		Project         []*Project `toml:"project"`
	}

	MetricInfo struct {
		Periods    []int64
		Statistics []string
		Dimensions string
	}

	MetricsRequest struct {
		q           *cms.DescribeMetricListRequest
		info        *MetricInfo
		checkPeriod bool
	}
)

func (p *Project) GenDimension(mestric string) (string, error) {

	var dimension *Dimension
	for _, d := range p.Metrics.Dimensions {
		if d.Name == mestric {
			dimension = d
			break
		}
	}

	if dimension == nil && len(p.InstanceIDs) == 0 {
		return "", nil
	}

	vals := []map[string]string{}

	if dimension != nil && dimension.Value != "" {
		if err := json.Unmarshal([]byte(dimension.Value), &vals); err != nil {
			return "", err
		}
	}

	setId := false
	for _, m := range vals {
		if _, ok := m["instanceId"]; ok {
			setId = true
			break
		}
	}

	if !setId && len(p.InstanceIDs) > 0 {
		for _, id := range p.InstanceIDs {
			vals = append(vals, map[string]string{
				"instanceId": id})
		}
	}

	js, err := json.Marshal(&vals)
	if err != nil {
		return "", err
	}

	return string(js), nil

}

func getPeriod(namespace, metricname string) string {
	//TODO: 有些指标可能最低周期不是60s
	return "60"
}
