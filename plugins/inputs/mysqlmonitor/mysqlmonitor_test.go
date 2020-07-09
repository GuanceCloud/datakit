package mysqlmonitor

import (
	"fmt"
	"testing"

	"github.com/naoina/toml"
)

func TestParse(t *testing.T) {
	x := []byte(`

#[[mysqlmonitor]]
  ## 采集的频度，最小粒度5m
  interval = '5m'
  ## 指标集名称，默认值(mysql_monitor)
  metricName = ''
  instanceId = ''
  instanceDesc = ''
  host = '10.200.6.53'
  port = '3306'
  username = 'root'
  password = 'root'
  database = ''
  product = ''
`)

	tbl, err := toml.Parse(x)
	if err != nil {
		t.Fatal(err)

	}

	//var o MysqlMonitor
	var o Mysql
	if err := toml.UnmarshalTable(tbl, &o); err != nil {
		t.Fatal(err)
	}

	fmt.Printf("%+#v", o)
}
