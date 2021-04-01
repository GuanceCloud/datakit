package awscloudtrail

import (
	"context"

	"golang.org/x/time/rate"

	"github.com/aws/aws-sdk-go/service/cloudtrail"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

const (
	sampleConfig = `
# ##(required)
#[[inputs.awscloudtrail]]
# ##(required)
#access_key = ''
#access_secret = ''
#region_id = 'us-east-1'

# ##(optional) collect interval
#interval = '5m'`
)

type AwsInstance struct {
	AccessKey    string
	AccessSecret string
	RegionID     string

	Interval datakit.Duration

	ctx       context.Context
	cancelFun context.CancelFunc

	awsClient *cloudtrail.CloudTrail

	rateLimiter *rate.Limiter

	mode string

	testError error
}

func (ag *AwsInstance) isTest() bool {
	return ag.mode == "test"
}

func (ag *AwsInstance) isDebug() bool {
	return ag.mode == "debug"
}
