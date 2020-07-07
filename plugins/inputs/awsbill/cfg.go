package awsbill

import (
	"context"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"golang.org/x/time/rate"
)

const (
	sampleConfig = `
#[[inputs.aws_billing]]
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

	ctx       context.Context
	cancelFun context.CancelFunc

	cloudwatchClient *cloudwatch.CloudWatch

	rateLimiter *rate.Limiter

	billingMetrics map[string]*cloudwatch.Metric
}
