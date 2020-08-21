package mysqlmonitor

import (
	"fmt"
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

func (_ *Mysql) Catalog() string {
	return "db"
}

func (_ *Mysql) SampleConfig() string {
	return configSample
}

func (mysql *Mysql) Run() {
	l = logger.SLogger("mysqlMonitor")
	l.Info("mysqlMonitor input started...")

	connStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", mysql.Username, mysql.Password, mysql.Host, mysql.Port, mysql.Database)
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		l.Errorf("mysql connect faild %v", err)
	}

	mysql.db = db

	interval, err := time.ParseDuration(mysql.Interval)
	if err != nil {
		l.Error(err)
	}

	tick := time.NewTicker(interval)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			// handle
			mysql.handle()
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return
		}
	}
}

func init() {
	inputs.Add(name, func() inputs.Input {
		return &Mysql{}
	})
}
