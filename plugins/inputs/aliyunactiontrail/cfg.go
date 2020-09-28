package aliyunactiontrail

import (
	"context"

	"golang.org/x/time/rate"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/actiontrail"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

const (
	configSample = `
#[[inputs.aliyunactiontrail]]

# ##(required) 
#region = 'cn-hangzhou'
#access_id = ''
#access_key = ''

# ##(optional) ISO8601 unix time format: 2020-02-01T06:00:00Z
# ## the earliest is 90 days from now.
# ## if empty, from now on.
#from = ''

# ##(optional) default is 10m, must not be less than 10m
#interval = '10m'
`
)

type (
	AliyunActiontrail struct {
		Region     string
		AccessKey  string
		AccessID   string
		MetricName string
		From       string
		Interval   datakit.Duration //至少10分钟

		client *actiontrail.Client

		regions []string

		rateLimiter *rate.Limiter

		ctx       context.Context
		cancelFun context.CancelFunc

		historyFlag int32

		debugMode bool
	}
)
