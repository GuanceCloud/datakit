package oraclemonitor

import (
	"testing"

	"github.com/influxdata/toml"
)

func TestLoadConf(t *testing.T) {
	x := `
[[inputs.oraclemonitor]]
  ## 采集的频度，最小粒度5m
	libPath = "a/b/c"
  interval = '10s'
  ## 指标集名称，默认值oracle_monitor
  metricName = ''
  ## 实例ID(非必要属性)
  instanceId = ''
  ## # 实例描述(非必要属性)
  instanceDesc = ''
  ## oracle实例地址(ip)
  host = ''
  ## oracle监听端口
  port = ''
  ## 帐号
  username = ''
  ## 密码
  password = ''
  ## oracle的服务名
  server = ''
  ## 实例类型 例如 单实例、DG、RAC 等，非必要属性
  type= 'singleInstance'
`

	tree, err := toml.Parse([]byte(x))
	if err != nil {
		t.Fatal(err)
	}

	om := &OracleMonitor{}
	if err := toml.UnmarshalTable(tree, om); err != nil {
		t.Fatal(err)
	}

	t.Logf("%+#v", om)
}
