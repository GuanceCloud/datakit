package awsbill

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"

type AwsInstance struct {
	AccessKey    string
	AccessSecret string
	AccessToken  string
	RegionID     string
	MetricName   string
	Interval     internal.Duration
}
