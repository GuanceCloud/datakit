package mysqlmonitor

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestGetDsnString(t *testing.T) {
	m := MysqlMonitor{}
	m.Host = "127.0.0.1"
	m.Port = 3306
	m.User = "root"
	m.Pass = "test"

	expected := "root:test@tcp(127.0.0.1:3306)/"
	actual := m.getDsnString()
	assert.Equal(t, expected, actual)
}

func TestGetStatus(t *testing.T) {
	m := MysqlMonitor{}
	m.Host = "127.0.0.1"
	m.Port = 3306
	m.User = "root"
	m.Pass = "test"

	initLog()

	m.initCfg()
	m.resData = make(map[string]interface{})

	m.getStatus()
	m.submitMetrics()
}

func TestGetVariables(t *testing.T) {
	m := MysqlMonitor{}
	m.Host = "127.0.0.1"
	m.Port = 3306
	m.User = "root"
	m.Pass = "test"

	initLog()

	m.initCfg()
	m.resData = make(map[string]interface{})

	m.getVariables()
	m.submitMetrics()
}

// todo
func TestGetInnodbStatus(t *testing.T) {
	m := MysqlMonitor{}
	m.Host = "127.0.0.1"
	m.Port = 3306
	m.User = "root"
	m.Pass = "test"

	initLog()

	m.initCfg()
	m.resData = make(map[string]interface{})

	m.getInnodbStatus()
	m.submitMetrics()
}

func TestGetLogStats(t *testing.T) {
	m := MysqlMonitor{}
	m.Host = "rm-bp15268nefz6870hg.mysql.rds.aliyuncs.com"
	m.Port = 3306
	m.User = "cc_monitor"
	m.Pass = "Zyadmin123"

	initLog()

	m.initCfg()
	m.resData = make(map[string]interface{})

	m.getLogStats()
	m.submitMetrics()
}

func TestQuerySizePerschema(t *testing.T) {
	m := MysqlMonitor{}
	m.Host = "rm-bp15268nefz6870hg.mysql.rds.aliyuncs.com"
	m.Port = 3306
	m.User = "cc_monitor"
	m.Pass = "Zyadmin123"

	initLog()

	m.initCfg()
	m.resData = make(map[string]interface{})

	m.querySizePerschema()
	m.submitMetrics()
}

func TestGetQueryExecTimePerSchema(t *testing.T) {
	m := MysqlMonitor{}
	m.Host = "127.0.0.1"
	m.Port = 3306
	m.User = "root"
	m.Pass = "test"

	initLog()

	m.initCfg()
	m.resData = make(map[string]interface{})

	m.getQueryExecTimePerSchema()
	m.submitMetrics()
}

func TestGetQueryExecTime95th(t *testing.T) {
	m := MysqlMonitor{}
	m.Host = "127.0.0.1"
	m.Port = 3306
	m.User = "root"
	m.Pass = "test"

	initLog()

	m.initCfg()
	m.resData = make(map[string]interface{})

	m.getQueryExecTime95th()
	m.submitMetrics()
}


