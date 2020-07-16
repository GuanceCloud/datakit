package aliyunlog

import "sync"

const (
	aliyunlogConfigSample = `
#[[inputs.consumer]]
# ##(required)
#endpoint = ''
#access_key_id = ''
#access_key_secret = ''
	
#[[inputs.consumer.projects]]
# ##(required) project name 
#name = ''
	
#[[inputs.consumer.projects.stores]]
# ##(required) name of log store
#name = ''
	
# ##(optional) metric name, default is 'aliyunlog_+store-name' 
#metric_name = ''
	
# ##(required) 指定当前日志库的消费组名称以及消费数据客户端名称
#consumer_group_name = ''
#consumer_name = ''
	
# ##(optional) 指定哪些key作为tag, 默认都为field
# ##eg., tags=["status_code","protocol"]
# ##By default, the key used as tag cannot be field，you can still specify a key both be tag and field: tags=["status_code:*"]
# ##可指定作为tag时的别名, 例: tags=["status_code::status"]
# ##同时作为tag和field，同时设置tag别名: tags=["status_code:*:status"]
#tags = []
	
# # ##(optional) 指定fields的类型, 默认为string, 可指定为int或float
# # ##例: fields = ["int:status,request_length", "float:cpuUsage"]
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
