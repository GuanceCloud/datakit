package aliyuncms

import (
	"encoding/json"
	"strconv"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/cms"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
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
#	#instanceIds=['','']

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
#       period = 60
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
		Name   string `toml:"name"`
		Value  string `toml:"value"`
		Period int    `toml:"period"`
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

	MetricMeta struct {
		Periods     []int64
		Statistics  []string
		Dimensions  []string
		Description string
		Unit        string
	}

	MetricsRequest struct {
		q             *cms.DescribeMetricListRequest
		meta          *MetricMeta
		haveGetMeta   bool
		tunePeriod    bool
		tuneDimension bool
	}
)

func (p *Project) genDimension(metric string, logger *models.Logger) (string, error) {

	var dimension *Dimension
	for _, d := range p.Metrics.Dimensions {
		if d.Name == metric {
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
			logger.Errorf("invalid dimension(%s): %s, %s", metric, dimension.Value, err)
			return "", err
		}
	}

	bHaveSetID := false
	for _, m := range vals {
		if _, ok := m["instanceId"]; ok {
			bHaveSetID = true
			break
		}
	}

	if !bHaveSetID && len(p.InstanceIDs) > 0 {
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

func (p *Project) getPeriod(metricname string) string {
	for _, d := range p.Metrics.Dimensions {
		if d.Name == metricname {
			if d.Period >= 60 {
				return strconv.FormatInt(int64(d.Period), 10)
			}
			break
		}
	}
	return "60" //默认60s
}
