package awstrail

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"

const (
	inputName = "aws_cloudtrail"

	sampleConfig = `
#[[aws_cloudtrail]]
#access_key = ''
#access_secret = ''
#access_token = ''
#region_id = 'us-east-1'
#metric_name = '' #default is aws_cloudtrail
#interval = '5m'
`
)

type (
	AwsTrailInstance struct {
		AccessKey    string
		AccessSecret string
		AccessToken  string
		RegionID     string
		MetricName   string
		Interval     internal.Duration

		// //以下字段用于过滤结果
		// UserName     string //产生该事件的用户名
		// EventName    string //事件名称，eg., ConsoleLogin
		// ResourceType string //事件相关联的资源类型
		// ResourceName string //事件相关联的资源名称
		// Readonly     bool   //该事件是写操作还是读操作
	}
)
