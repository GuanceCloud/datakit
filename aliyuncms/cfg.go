package aliyuncms

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/cms"
	"github.com/influxdata/toml"
)

const (
	aliyuncmsConfigSample = `
disable = true
[[cms]]
## Aliyun Region (required)
## See: https://www.alibabacloud.com/help/zh/doc-detail/40654.htm
region_id = "cn-hangzhou"

## Aliyun Credentials (required)
access_key_id = ""
access_key_secret = ""

  [[cms.project]]
    ## Metric Statistic Project (required)
	name="acs_ecs_dashboard"

	## Optional instances from which you want to pull metrics, empty means to pull all instances 
	#instanceIds=["xxx","yyy"]

	## Metrics to Pull (Required)
	## See: https://help.aliyun.com/document_detail/28619.html?spm=a2c4g.11186623.2.11.6ac47694AjhHt4
	[cms.project.metrics]
	names = [
		"CPUUtilization",
	]

	## dimensions can be used to query the specified resource, which is a collection of key-value forms. 
	## each metric may have its own dimensions, See: https://help.aliyun.com/document_detail/28619.html?spm=a2c4g.11186623.2.11.6ac47694AjhHt4 
	## name is metric name, value is a json string, eg: '[{"instanceId":"xxx"},{"device":"xxx"}]'
	#[[cms.project.metrics.dimensions]]
	#  name = "diskusage_free"
	#  value = '[{"instanceId":"xxx"},{"device":"xxx"}]'
`
)

var (
	Cfg             ACSCmsConfig
	MetricsRequests = []*cms.DescribeMetricListRequest{}
)

type (
	Metric struct {
		MetricNames []string     `toml:"names"`
		Dimensions  []*Dimension `toml:"dimensions"`
	}

	Dimension struct {
		Name  string `toml:"name"`
		Value string `toml:"value"`
	}

	Project struct {
		Name        string   `toml:"name"`
		InstanceIDs []string `toml:"instanceIds"`
		Metrics     *Metric  `toml:"metrics"`
	}

	CmsCfg struct {
		RegionID        string     `toml:"region_id"`
		AccessKeyID     string     `toml:"access_key_id"`
		AccessKeySecret string     `toml:"access_key_secret"`
		Project         []*Project `toml:"project"`
	}

	ACSCmsConfig struct {
		Disable bool      `toml:"disable"`
		CmsCfg  []*CmsCfg `toml:"cms"`
	}
)

func (p *Project) MakeDimension(mestric string) (string, error) {

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

	if dimension.Value != "" {
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

func (c *ACSCmsConfig) SampleConfig() string {
	return aliyuncmsConfigSample
}

func (c *ACSCmsConfig) FilePath(root string) string {
	d := filepath.Join(root, "aliyuncms")
	return filepath.Join(d, "aliyuncms.toml")
}

func (c *ACSCmsConfig) ToTelegraf() (string, error) {
	return "", nil
}

func (c *ACSCmsConfig) Load(f string) error {
	data, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}

	if err = toml.Unmarshal(data, c); err != nil {
		return err
	}

	for _, item := range c.CmsCfg {
		for _, p := range item.Project {
			for _, m := range p.Metrics.MetricNames {
				req := cms.CreateDescribeMetricListRequest()
				req.Period = "60" // strconv.FormatInt(int64(metricPeriod/time.Second), 10)
				req.MetricName = m
				req.Length = "10000"
				req.Namespace = p.Name
				if ds, err := p.MakeDimension(m); err == nil {
					req.Dimensions = ds
				}
				MetricsRequests = append(MetricsRequests, req)
			}
		}
	}

	if len(MetricsRequests) == 0 {
		log.Println("[warn] no metric will be pulled")
	}

	return nil
}
