package awsbill

import (
	"context"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"golang.org/x/time/rate"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

const (
	sampleConfig = `
# ##(required)
#[[inputs.awsbill]]
# ##(required)
#access_key = ''
#access_secret = ''
#region_id = 'us-east-1'

# ##(optional) custom metric name, default is awsbill
#metric_name = ''

# ##(optional) collect interval, default is 4 hours. AWS billing metrics are available about once every 4 hours.
#interval = '4h'
`
)

type AwsInstance struct {
	AccessKey    string
	AccessSecret string
	//AccessToken  string
	RegionID   string
	MetricName string
	Interval   datakit.Duration

	ctx       context.Context
	cancelFun context.CancelFunc

	cloudwatchClient *cloudwatch.CloudWatch

	rateLimiter *rate.Limiter

	billingMetrics map[string]*cloudwatch.Metric

	debugMode bool
}
