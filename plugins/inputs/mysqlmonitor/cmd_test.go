package mysqlmonitor

import (
	"database/sql"
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

func TestHandle(t *testing.T) {
	logger.SetGlobalRootLogger("",
		"debug",
		logger.OPT_ENC_CONSOLE|logger.OPT_SHORT_CALLER)
	l := logger.SLogger("mysqlmonitor")

	l.Info("start....")

	t.Run("case-gatherGlobalStatuses", func(t *testing.T) {
		m := MysqlMonitor{}
		m.Servers = []string{"root:test@tcp(127.0.0.1:3306)/?tls=false"}
		m.MetricName = name
		for _, server := range m.Servers {
			serv, err := dsnAddTimeout(server)
			if err != nil {
				t.Errorf("error")
			}

			db, err := sql.Open("mysql", serv)
			if err != nil {
				t.Errorf("error")
			}

			defer db.Close()

			err = m.gatherGlobalStatuses(db, serv)
			if err != nil {
				t.Errorf("error")
			}

			// scan.Targets = []string{"127.0.0.1"}
			t.Log("ok")
		}
	})

	t.Run("case-gatherGlobalVariables", func(t *testing.T) {
		m := MysqlMonitor{}
		m.Servers = []string{"root:test@tcp(127.0.0.1:3306)/?tls=false"}
		m.MetricName = name
		for _, server := range m.Servers {
			serv, err := dsnAddTimeout(server)
			if err != nil {
				t.Errorf("error")
			}

			db, err := sql.Open("mysql", serv)
			if err != nil {
				t.Errorf("error")
			}

			defer db.Close()

			err = m.gatherGlobalVariables(db, serv)
			if err != nil {
				t.Errorf("error")
			}

			// scan.Targets = []string{"127.0.0.1"}
			t.Log("ok")
		}
	})

	t.Run("case-GatherProcessListStatuses", func(t *testing.T) {
		m := MysqlMonitor{}
		m.Servers = []string{"root:test@tcp(127.0.0.1:3306)/?tls=false"}
		m.MetricName = name
		for _, server := range m.Servers {
			serv, err := dsnAddTimeout(server)
			if err != nil {
				t.Errorf("error")
			}

			db, err := sql.Open("mysql", serv)
			if err != nil {
				t.Errorf("error")
			}

			defer db.Close()

			err = m.GatherProcessListStatuses(db, serv)
			if err != nil {
				t.Errorf("error")
			}

			// scan.Targets = []string{"127.0.0.1"}
			t.Log("ok")
		}
	})

	t.Run("case-GatherUserStatisticsStatuses", func(t *testing.T) {
		m := MysqlMonitor{}
		m.Servers = []string{"root:test@tcp(127.0.0.1:3306)/?tls=false"}
		m.MetricName = name
		for _, server := range m.Servers {
			serv, err := dsnAddTimeout(server)
			if err != nil {
				t.Errorf("error")
			}

			db, err := sql.Open("mysql", serv)
			if err != nil {
				t.Errorf("error")
			}

			defer db.Close()

			err = m.GatherUserStatisticsStatuses(db, serv)
			if err != nil {
				t.Errorf("error")
			}

			// scan.Targets = []string{"127.0.0.1"}
			t.Log("ok")
		}
	})

	t.Run("case-gatherInfoSchemaAutoIncStatuses", func(t *testing.T) {
		m := MysqlMonitor{}
		m.Servers = []string{"root:test@tcp(127.0.0.1:3306)/?tls=false"}
		m.MetricName = name
		for _, server := range m.Servers {
			serv, err := dsnAddTimeout(server)
			if err != nil {
				t.Errorf("error")
			}

			db, err := sql.Open("mysql", serv)
			if err != nil {
				t.Errorf("error")
			}

			defer db.Close()

			err = m.gatherInfoSchemaAutoIncStatuses(db, serv)
			if err != nil {
				t.Errorf("error")
			}

			// scan.Targets = []string{"127.0.0.1"}
			t.Log("ok")
		}
	})

	t.Run("case-gatherInnoDBMetrics", func(t *testing.T) {
		m := MysqlMonitor{}
		m.Servers = []string{"root:test@tcp(127.0.0.1:3306)/?tls=false"}
		m.MetricName = name
		for _, server := range m.Servers {
			serv, err := dsnAddTimeout(server)
			if err != nil {
				t.Errorf("error")
			}

			db, err := sql.Open("mysql", serv)
			if err != nil {
				t.Errorf("error")
			}

			defer db.Close()

			err = m.gatherInnoDBMetrics(db, serv)
			if err != nil {
				t.Errorf("error")
			}

			// scan.Targets = []string{"127.0.0.1"}
			t.Log("ok")
		}
	})

	// t.Run("case-gatherPerfTableIOWaits", func(t *testing.T) {
	// 	m := MysqlMonitor{}
	// 	m.Servers = []string{"root:test@tcp(127.0.0.1:3306)/?tls=false"}
	// 	m.MetricName = name
	// 	for _, server := range m.Servers {
	// 		serv, err := dsnAddTimeout(server)
	// 		if err != nil {
	// 			t.Errorf("error")
	// 		}

	// 		db, err := sql.Open("mysql", serv)
	// 		if err != nil {
	// 			t.Errorf("error")
	// 		}

	// 		defer db.Close()

	// 		err = m.gatherPerfTableIOWaits(db, serv)
	// 		if err != nil {
	// 			t.Errorf("error")
	// 		}

	// 		// scan.Targets = []string{"127.0.0.1"}
	// 		t.Log("ok")
	// 	}
	// })

	// t.Run("case-gatherPerfIndexIOWaits", func(t *testing.T) {
	// 	m := MysqlMonitor{}
	// 	m.Servers = []string{"root:test@tcp(127.0.0.1:3306)/?tls=false"}
	// 	m.MetricName = name
	// 	for _, server := range m.Servers {
	// 		serv, err := dsnAddTimeout(server)
	// 		if err != nil {
	// 			t.Errorf("error")
	// 		}

	// 		db, err := sql.Open("mysql", serv)
	// 		if err != nil {
	// 			t.Errorf("error")
	// 		}

	// 		defer db.Close()

	// 		err = m.gatherPerfIndexIOWaits(db, serv)
	// 		if err != nil {
	// 			t.Errorf("error")
	// 		}

	// 		// scan.Targets = []string{"127.0.0.1"}
	// 		t.Log("ok")
	// 	}
	// })

	t.Run("case-gatherPerfTableLockWaits", func(t *testing.T) {
		m := MysqlMonitor{}
		m.Servers = []string{"root:test@tcp(127.0.0.1:3306)/?tls=false"}
		m.MetricName = name
		for _, server := range m.Servers {
			serv, err := dsnAddTimeout(server)
			if err != nil {
				t.Errorf("error")
			}

			db, err := sql.Open("mysql", serv)
			if err != nil {
				t.Errorf("error")
			}

			defer db.Close()

			err = m.gatherPerfTableLockWaits(db, serv)
			if err != nil {
				t.Errorf("error")
			}

			// scan.Targets = []string{"127.0.0.1"}
			t.Log("ok")
		}
	})

	t.Run("case-gatherPerfEventWaits", func(t *testing.T) {
		m := MysqlMonitor{}
		m.Servers = []string{"root:test@tcp(127.0.0.1:3306)/?tls=false"}
		m.MetricName = name
		for _, server := range m.Servers {
			serv, err := dsnAddTimeout(server)
			if err != nil {
				t.Errorf("error")
			}

			db, err := sql.Open("mysql", serv)
			if err != nil {
				t.Errorf("error")
			}

			defer db.Close()

			err = m.gatherPerfEventWaits(db, serv)
			if err != nil {
				t.Errorf("error")
			}

			// scan.Targets = []string{"127.0.0.1"}
			t.Log("ok")
		}
	})

}
