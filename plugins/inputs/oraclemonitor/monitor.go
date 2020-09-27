// +build linux

package oraclemonitor

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	configSample = `
[[inputs.external]]
	daemon = true
	name = 'oraclemonitor'
	cmd  = "/usr/local/cloudcare/dataflux/datakit/externals/oraclemonitor"
	args = [
		'-instance-id'    , '<your-instance-id>'        ,
		'-metric-name'    , 'oracle_monitor'            ,
		'-interval'       , '1m'                        ,
		'-instance-desc'  , '<your-oracle-description>' ,
		'-host'           , '<your-oracle-host>'        ,
		'-port'           , '1521'                      ,
		'-username'       , '<oracle-user-name>'        ,
		'-password'       , '<oracle-password>'         ,
		'-service-name'   , '<oracle-service-name>'     ,
		'-cluster-type'   , 'single'                    ,
		'-oracle-version' , '11g'                       ,
	]
	envs = [
		'LD_LIBRARY_PATH=/opt/oracle/instantclient_19_8:$LD_LIBRARY_PATH',
	]

	#############################
	# 参数说明(标 * 为必选项)
	#############################
	# *-interval       : 采集的频度，最小粒度5m
	#  -metric-name    : 指标集名称，默认值oracle_monitor
	#  -instance-id    : 实例ID
	#  -instance-desc  : 实例描述
	# *-host           : oracle实例地址(ip)
	#  -port           : oracle监听端口
	# *-username       : oracle 用户名
	# *-password       : oracle 密码
	# *-service-name   : oracle的服务名
	# *-cluster-type   : 实例类型(例如 single/dg/rac)
	# *-oracle-version : 采集的oracle版本(支持10g, 11g, 12c)
`
)

var (
	inputName = "oraclemonitor"
)

type OracleMonitor struct{}

func (_ *OracleMonitor) Catalog() string { return "db" }

func (_ *OracleMonitor) SampleConfig() string { return configSample }

func (o *OracleMonitor) Run() {}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &OracleMonitor{}
	})
}
