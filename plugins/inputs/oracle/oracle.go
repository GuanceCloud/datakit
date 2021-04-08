// +build linux

package oracle

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/external"
)

const (
	configSample = `
[[inputs.external]]
	daemon = true
	name = 'oraclemonitor'
	cmd  = "/usr/local/cloudcare/dataflux/datakit/externals/oraclemonitor"
	args = [
		'-data-type'      , '<metric/logging>'          ,
		'-instance-id'    , '<your-instance-id>'        ,
		'-metric-name'    , 'oracle_monitor'            ,
		'-interval'       , '1m'                        ,
		'-instance-desc'  , '<your-oracle-description>' ,
		'-host'           , '<your-oracle-host>'        ,
		'-port'           , '1521'                      ,
		'-username'       , '<oracle-user-name>'        ,
		'-password'       , '<oracle-password>'         ,
		'-service-name'   , '<oracle-service-name>'     ,
		'-oracle-version' , '11g'                       ,
	]
	envs = [
		'LD_LIBRARY_PATH=/opt/oracle/instantclient_19_8:$LD_LIBRARY_PATH',
	]

	#############################
	# 参数说明(标 * 为必选项)
	#############################
	# *-interval       : 采集的频度，最小粒度5m
	#  -data-type      : 数据类型，默认值metric
	#  -metric-name    : 指标集名称，默认值oracle_monitor
	#  -instance-id    : 实例ID
	#  -instance-desc  : 实例描述
	# *-host           : oracle实例地址(ip)
	#  -port           : oracle监听端口
	# *-username       : oracle 用户名
	# *-password       : oracle 密码
	# *-service-name   : oracle的服务名
`
)

var (
	inputName   = "oracle"
	catalogName = "db"
	l           = logger.DefaultSLogger("oracle")
)

type Input struct {
	external.ExernalInput
}

func (i *Input) initCfg() {
	// 采集频度
	i.IntervalDuration = 1 * time.Minute

	if i.Interval != "" {
		du, err := time.ParseDuration(i.Interval)
		if err != nil {
			l.Errorf("bad interval %s: %s, use default: 1m", i.Interval, err.Error())
		} else {
			i.IntervalDuration = du
		}
	}

	i.Addr = fmt.Sprintf("%s:%d", i.Host, i.Port)
	client := redis.NewClient(&redis.Options{
		Addr:     i.Addr,
		Password: i.Password, // no password set
		DB:       i.DB,       // use default DB
	})

	i.client = client

	i.globalTag()
}

func (i *Input) Run() {
	// FIXME: 如果改成松散配置读取方式（只要是 .conf，直接读取并启动之）
	// 这里得到 .Run() 方法要去掉。
	i.ExernalInput.Run()
}

func (i *Input) Catalog() string { return catalogName }

func (i *Input) SampleConfig() string { return configSample }

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&processMeasurement{},
		&tablespaceMeasurement{},
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
