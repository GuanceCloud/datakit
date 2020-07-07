package aliyunactiontrail

import (
	"context"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/actiontrail"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"golang.org/x/time/rate"
)

const (
	configSample = `
#[[inputs.actiontrail]]
#    region = 'cn-hangzhou'
#    access_id = ''
#    access_key = ''

#    ##if empty, use "aliyun_actiontrail"
#    metric_name = ''

#    ## ISO8601 unix time format: 2020-02-01T06:00:00Z 
#    ## the earliest is 90 days from now.
#    ## if empty, from now on. 
#    from = ''

#    ## default is 10m, must not be less than 10m
#    interval = '10m'
`
)

type (
	AliyunActiontrail struct {
		Region     string
		AccessKey  string
		AccessID   string
		MetricName string
		From       string
		Interval   internal.Duration //至少10分钟

		client *actiontrail.Client

		metricName string

		rateLimiter *rate.Limiter

		ctx       context.Context
		cancelFun context.CancelFunc
	}
)
