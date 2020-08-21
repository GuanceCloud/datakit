package aliyunfc

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName = "aliyunfc"

	sampleCfg = `
[[inputs.aliyuncms]]

##(required)
access_key_id = 'xxx'
access_key_secret = 'xxxxx'
region_id = 'cn-hangzhou'

# ##(optional)（Security Token Service，STS）
#security_token = ''

# ##(optional) global collect interval, default is 5min.
#interval = '5m'

# ##(optional) delay collect duration
#delay = '5m'

#[inputs.aliyuncms.tags]
#key1 = "val1"
#key2 = "val2"

# ##(required)
[[inputs.aliyuncms.project]]
# ##(required) product namespace
name='acs_fc'
# ##(required)
[inputs.aliyuncms.project.metrics]

##(required)
## names of metrics
names =[
	'FuntionTotalInvocations',
	'ServiceTotalInvocations',
	'FunctionAvgDuration',
	'FunctionBillableInvocations',
	'FunctionBillableInvocationsRate',
	'FunctionBillableInvocationsRate',
	'FunctionClientErrors',
	'FunctionClientErrorsRate',
	'FunctionFunctionErrors',
	'FunctionFunctionErrorsRate',
	'FunctionMaxMemoryUsage',
	'FunctionServerErrors',
	'FunctionServerErrorsRate',
	'FuntionThrottles',
	'FuntionThrottlesRate',
	'RegionBillableInvocations',
	'RegionbillableInvocationsRate',
	'RegionClientErrors',
	'RegionClientErrorsRate',
	'RegionServerErrors',
	'RegionThrottles',
	'RegionThrttlesRate',
	'RegionTotalInvocations',
	'ServiceBillableInvocations',
	'ServiceBillableInvocationsRate',
	'ServiceClientErrors',
	'ServiceClientErrorsRate',
	'ServiceClientErrorsRate',
	'ServiceClientErrorsRate',
	'ServiceThrottles',
	'ServiceThrottles']

# ##(optional)
#[[inputs.aliyuncms.project.metrics.property]]
# ##(optional) you may specify period of this metric
#period = 60

# ##(optional) collect interval of thie metric
#interval = '5m'

# ##(optional) collect filter, a json string
#dimensions = '''
#  [
#   {"userId":"******"}
#   ]
#   '''

# ##(optional) custom tags
#[inputs.aliyuncms.project.metrics.property.tags]
#key1 = "val1"
#key2 = "val2"
	`
)

type FC struct {
	Interval string `toml:"interval"`
	Metric   string `toml:"metric"`
}

func (m *FC) SampleConfig() string {
	return sampleCfg
}

func (m *FC) Description() string {
	return ""
}

func (m *FC) Catalog() string {
	return "aliyun"
}

func (m *FC) Run() {
	return
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &FC{}
	})
}
