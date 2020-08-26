package mysqlmonitor

import (
	"sync"
	"time"

	"database/sql"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"

	_ "github.com/go-sql-driver/mysql"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	defaultTimeout                             = 5 * time.Second
	defaultPerfEventsStatementsDigestTextLimit = 120
	defaultPerfEventsStatementsLimit           = 250
	defaultPerfEventsStatementsTimeLimit       = 86400
	defaultGatherGlobalVars                    = true
)

var (
	l    *logger.Logger
	name = "mysqlMonitor"
)

func (_ *MysqlMonitor) Catalog() string {
	return "db"
}

func (_ *MysqlMonitor) SampleConfig() string {
	return configSample
}

func (m *MysqlMonitor) Run() {
	l = logger.SLogger("mysqlMonitor")
	l.Info("mysqlMonitor input started...")

	interval, err := time.ParseDuration(m.Interval)
	if err != nil {
		l.Error(err)
	}

	m.MetricName = name

	tick := time.NewTicker(interval)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			// handle
			m.handle()
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return
		}
	}
}

func (m *MysqlMonitor) handle() {
	var wg sync.WaitGroup

	// Loop through each server and collect metrics
	for _, server := range m.Servers {
		wg.Add(1)
		go func(s string) {
			defer wg.Done()
			m.gatherServer(s)
		}(server)
	}

	wg.Wait()
}

func (m *MysqlMonitor) gatherServer(serv string) error {
	serv, err := dsnAddTimeout(serv)
	if err != nil {
		return err
	}

	db, err := sql.Open("mysql", serv)
	if err != nil {
		return err
	}

	defer db.Close()

	err = m.gatherGlobalStatuses(db, serv)
	if err != nil {
		return err
	}

	if m.GatherGlobalVars {
		// Global Variables may be gathered less often
		err = m.gatherGlobalVariables(db, serv)
		if err != nil {
			return err
		}
	}

	if m.GatherBinaryLogs {
		err = m.gatherBinaryLogs(db, serv)
		if err != nil {
			return err
		}
	}

	if m.GatherProcessList {
		err = m.GatherProcessListStatuses(db, serv)
		if err != nil {
			return err
		}
	}

	if m.GatherUserStatistics {
		err = m.GatherUserStatisticsStatuses(db, serv)
		if err != nil {
			return err
		}
	}

	if m.GatherSlaveStatus {
		err = m.gatherSlaveStatuses(db, serv)
		if err != nil {
			return err
		}
	}

	if m.GatherInfoSchemaAutoInc {
		err = m.gatherInfoSchemaAutoIncStatuses(db, serv)
		if err != nil {
			return err
		}
	}

	if m.GatherInnoDBMetrics {
		err = m.gatherInnoDBMetrics(db, serv)
		if err != nil {
			return err
		}
	}

	if m.GatherTableIOWaits {
		err = m.gatherPerfTableIOWaits(db, serv)
		if err != nil {
			return err
		}
	}

	if m.GatherIndexIOWaits {
		err = m.gatherPerfIndexIOWaits(db, serv)
		if err != nil {
			return err
		}
	}

	if m.GatherTableLockWaits {
		err = m.gatherPerfTableLockWaits(db, serv)
		if err != nil {
			return err
		}
	}

	if m.GatherEventWaits {
		err = m.gatherPerfEventWaits(db, serv)
		if err != nil {
			return err
		}
	}

	if m.GatherFileEventsStats {
		err = m.gatherPerfFileEventsStatuses(db, serv)
		if err != nil {
			return err
		}
	}

	if m.GatherPerfEventsStatements {
		err = m.gatherPerfEventsStatements(db, serv)
		if err != nil {
			return err
		}
	}

	if m.GatherTableSchema {
		err = m.gatherTableSchema(db, serv)
		if err != nil {
			return err
		}
	}
	return nil
}

func init() {
	inputs.Add(name, func() inputs.Input {
		return &MysqlMonitor{}
	})
}
