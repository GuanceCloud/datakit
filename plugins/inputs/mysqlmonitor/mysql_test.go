package mysqlmonitor

import (
	"path/filepath"
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

func TestRun(t *testing.T) {
	// t.Run("case-run", func(t *testing.T) {
	// 	m := MysqlMonitor{}
	// 	m.Servers = []string{"root:root@tcp(10.200.6.53:3306)/"}
	// 	m.GatherProcessList = true
	// 	m.GatherUserStatistics = false
	// 	m.GatherInfoSchemaAutoInc = false
	// 	m.GatherInnoDBMetrics = false
	// 	m.GatherSlaveStatus = false
	// 	m.GatherBinaryLogs = false
	// 	m.GatherTableIOWaits = false
	// 	m.GatherTableLockWaits = false
	// 	m.GatherIndexIOWaits = false
	// 	m.GatherEventWaits = false
	// 	m.GatherTableSchema = false
	// 	m.GatherFileEventsStats = false
	// 	m.GatherPerfEventsStatements = false
	// 	m.GatherGlobalVars = false
	// 	m.GatherGlobalStatus = false
	// 	m.Interval = "10s"
	// 	m.MetricName = name

	// 	m.Run()

	// 	t.Log("ok")
	// })

	t.Run("case-push-data", func(t *testing.T) {
		datakit.InstallDir = "."
		datakit.DataDir = "."
		// datakit.OutputFile = "metrics.txt"
		datakit.GRPCDomainSock = filepath.Join(datakit.InstallDir, "datakit.sock")
		datakit.Exit = cliutils.NewSem()

		config.Cfg.MainCfg = &config.MainConfig{}
		config.Cfg.MainCfg.DataWay = &config.DataWayCfg{}

		config.Cfg.MainCfg.DataWay.Host = "preprod-openway.cloudcare.cn"
		config.Cfg.MainCfg.DataWay.Token = "tkn_f299ad7b7c0d4acdb8657be8d086f13a"
		config.Cfg.MainCfg.DataWay.Scheme = "https"
		datakit.IntervalDuration = time.Second * 10

		io.Start()

		m := MysqlMonitor{}
		m.Servers = []string{"root:root@tcp(10.200.6.53:3306)/"}
		// m.GatherProcessList = true
		// // 无测试数 ok
		// m.GatherUserStatistics = true
		// // int64 ok
		// m.GatherInfoSchemaAutoInc = true
		// // ok
		// m.GatherInnoDBMetrics = true
		// // 无测试数 ok
		// m.GatherSlaveStatus = true
		// // ok
		// m.GatherBinaryLogs = true
		// //  ok
		// m.GatherTableIOWaits = true
		// // 无测试数 ok
		// m.GatherTableLockWaits = true
		// // 数据切分(todo) ok
		// m.GatherIndexIOWaits = true
		// // ok
		// m.GatherEventWaits = true
		// // ok
		// m.GatherTableSchema = true
		// // ok
		// m.GatherFileEventsStats = true
		// // 无测试数据 ok
		// m.GatherPerfEventsStatements = true
		// // fail
		m.GatherGlobalVars = true
		// fail
		// m.GatherGlobalStatus = true
		m.Interval = "10s"
		m.MetricName = name

		m.Run()

		t.Log("ok")
	})
}
