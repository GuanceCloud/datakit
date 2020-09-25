package aliyunlog

import "sync"

const (
	aliyunlogConfigSample = `
#[[inputs.aliyunlog]]
# ##(required)
#endpoint = ''
#access_key_id = ''
#access_key_secret = ''
	
#[[inputs.aliyunlog.projects]]
# ##(required) project name 
#name = ''
	
#[[inputs.aliyunlog.projects.stores]]
# ##(required) name of log store
#name = ''
	
# ##(optional) metric name, default is 'aliyunlog_+store-name' 
#metric_name = ''
	
# ##(required) consumer group and consumer name for this log store
#consumer_group_name = ''
#consumer_name = ''
	
# ##(optional) specify which are tags and which are fields
# ##eg., tags=["status_code","protocol"]
# ##By default, the key used as tag cannot be field，you can still specify a key both be tag and field: tags=["status_code:*"]
# ##specify tag alias, eg., tags=["status_code::status"]
# ##both as tag and field，and specify tag alias: tags=["status_code:*:status"]
#tags = []
	
# # ##(optional) the data type of fields, default is string, can be int or float
# # ##eg., fields = ["int:status,request_length", "float:cpuUsage"]
#fields = []	
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

		wg sync.WaitGroup
	}
)
