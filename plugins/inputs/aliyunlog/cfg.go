package aliyunlog

const (
	aliyunlogConfigSample = `
#[[consumer]]
#  endpoint = 'cn-hangzhou.log.aliyuncs.com'
#  access_key = ''
#  access_id = ''
	
#  [[consumer.projects]]
#    name = 'project-name'
	
#	 [[consumer.projects.stores]]
#	   name = 'store-name'

#      ##if empty, use 'aliyunlog_+store-name' 
#      metric_name = ''
#	   consumer_group_name = 'consumer-group'
#	   consumer_name = 'consumer-name'
`
)

type (
	LogProject struct {
		Name   string
		Stores []*LogStoreCfg //每个project可对应多个log store
	}

	LogStoreCfg struct {
		MetricName        string
		Name              string
		Tags              []string `toml:"tags,omitempty"`   //指定哪些作为tag(默认所有都作为field)
		Fields            []string `toml:"fields,omitempty"` //指定某些field的数据类型(默认都为字符串)
		ConsumerGroupName string
		ConsumerName      string
	}

	ConsumerInstance struct {
		Endpoint        string
		AccessKeyID     string
		AccessKeySecret string
		Projects        []*LogProject
	}
)
