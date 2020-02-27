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
		Name              string
		ConsumerGroupName string
		ConsumerName      string
	}

	ConsumerInstance struct {
		Endpoint  string
		AccessKey string
		AccessID  string
		Projects  []*LogProject
	}
)
