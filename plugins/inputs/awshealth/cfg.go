package awshealth

import (
	"context"

	"golang.org/x/time/rate"

	"github.com/aws/aws-sdk-go/service/health"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

const (
	sampleConfig = `
#[[inputs.awshealth]]
# ##(required)
#access_key_id = ''
#access_key_secret = ''
#region = 'cn-north-1'

# ##(optional) custom metric name, default is awshealth
#metric_name = ''

# ##(optional) collect interval, default is 1min.
#interval = '1m'

# ##(optional) custom tags
#[inputs.awshealth.tags]
#key1 = "val1"
`
)

type (
	agent struct {
		AccessKeyID     string `toml:"access_key_id"`
		AccessKeySecret string `toml:"access_key_secret"`
		Region          string `toml:"region"`
		Interval        datakit.Duration
		MetricName      string `toml:"metric_name"`
		// EventStatus     []string
		// EventType       []string
		// EventCode       []string
		// Services        []string
		Tags map[string]string `toml:"tags,omitempty"`

		client *health.Health

		rateLimiter *rate.Limiter

		ctx       context.Context
		cancelFun context.CancelFunc
	}
)
