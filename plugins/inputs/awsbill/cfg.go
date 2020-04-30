package awsbill

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"

const (
	inputName = "aws_billing"

	sampleConfig = `
#[[aws_billing]]
#access_key = ''
#access_secret = ''
#access_token = ''
#region_id = 'us-east-1'
#metric_name = '' #default is aws_billing
#interval = '6h' #AWS billing metrics are available about once every 4 hours.
`
)

type AwsInstance struct {
	AccessKey    string
	AccessSecret string
	AccessToken  string
	RegionID     string
	MetricName   string
	Interval     internal.Duration
}
