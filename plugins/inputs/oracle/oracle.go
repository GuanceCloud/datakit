package oracle

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/external"
)

const (
	configSample = `
[[inputs.external]]
  daemon = true
  name = 'oracle'
  cmd  = "/usr/local/datakit/externals/oracle"
  args = [
    '--interval'       , '1m'                        ,
    '--host'           , '<your-oracle-host>'        ,
    '--port'           , '1521'                      ,
    '--username'       , '<oracle-user-name>'        ,
    '--password'       , '<oracle-password>'         ,
    '--service-name'   , '<oracle-service-name>'     ,
  ]
  envs = [
    'LD_LIBRARY_PATH=/opt/oracle/instantclient_19_8:$LD_LIBRARY_PATH',
  ]

  [inputs.external.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"

  #############################
  # 参数说明(标 * 为必选项)
  #############################
  # *--interval       : 采集的频度，最小粒度5m
  # *--host           : oracle实例地址(ip)
  #  --port           : oracle监听端口
  # *--username       : oracle 用户名
  # *--password       : oracle 密码
  # *--service-name   : oracle的服务名
  # *--query          : 自定义查询语句，格式为<sql:metricName:tags>, sql为自定义采集的语句, tags填入使用tag字段
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

func (i *Input) Run() {
	// FIXME: 如果改成松散配置读取方式（只要是 .conf，直接读取并启动之）
	// 这里得到 .Run() 方法要去掉。

	l.Info("oracle started...")
	i.ExernalInput.Run()
}

func (i *Input) Catalog() string { return catalogName }

func (i *Input) SampleConfig() string { return configSample }

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&processMeasurement{},
		&tablespaceMeasurement{},
		&systemMeasurement{},
	}
}

func (i *Input) AvailableArchs() []string {
	return []string{datakit.OSArchLinuxAmd64}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
