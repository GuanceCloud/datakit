package mysqlmonitor

import (
	"database/sql"
	"sync"
	"time"

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
	l             *logger.Logger
	name          = "mysqlMonitor"
	mariaDBMetric = "mariaDBMonitor"
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

	m.initCfg()

	tick := time.NewTicker(m.IntervalDuration)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			m.handle()
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return
		}
	}
}

func (m *MysqlMonitor) initCfg() {
	// 采集频度
	m.IntervalDuration = 10 * time.Minute

	if m.Interval != "" {
		du, err := time.ParseDuration(m.Interval)
		if err != nil {
			l.Errorf("bad interval %s: %s, use default: 10m", m.Interval, err.Error())
		} else {
			m.IntervalDuration = du
		}
	}

	// 指标集名称
	if m.MetricName == "" {
		m.MetricName = name
	}

	var timeout time.Time
	if m.Timeout != "" {
		timeout, err := time.ParseDuration(m.Timeout)
		if err != nil {
			l.Errorf("config timeout value (%v)  error %v ", m.Timeout, err)
		}
	}

	// build dsn string
	dsnStr := m.getDsnString(timeout)

	l.Infof("db build dsn connect str %s", dsnStr)

	db, err := sql.Open("mysql", dsnStr)
	if err != nil {
		l.Errorf("sql.Open(): %s", err.Error())
	}
}

func (m *MysqlMonitor) getDsnString(timeout time.Duration) string {
	cfg := mysql.Config{
	    User:                 m.User,
	    Passwd:               m.Pass,
	}

	// set addr
	if m.Sock != "" {
		cfg.Net = "unix"
		cfg.Addr = m.Sock
	} else {
		addr := fmt.Sprintf("%s:%d", m.Host, m.Port)
		cfg.Net = "tcp"
		cfg.Addr = addr
	}

	// set timeout
	if timeout != 0 {
		m.Timeout = timeout
	}

	// set Charset
	if m.Charset != "" {
		m.Params["charset"] = m.Charset
	}

	// tls (todo)
	return cfg.FormatDSN()
}

func (m *MysqlMonitor) collectMetrics() error {
	defer m.db.Close()
	// ping
	if err := m.Ping(); err != nil {
		l.Errorf("db connect error %v", err)
		return err
	}

	m.resData = make(map[string]*sql.RawBytes)

	//STATUS data
	m.getStatus()

	// VARIABLES data
	m.getVariables()

	// innodb
	if m.options.DisableInnodbMetrics && m.innodbEngineEnabled()  {
		m.getInnodbStatus()
	}

	// Binary log statistics
    if _, ok := m.resData["log_bin"]; ok {
    	metric["INNODB_VARS"].disable = true
    	m.getLogStats()
    }

    // Compute key cache utilization metric
    m.computeCacheUtilization()

    if m.options.ExtraStatusMetrics {
    	// 额外的status metric 设置标志
    	metric["OPTIONAL_STATUS_VARS"].disable = true
    	if m.versionCompatible("5.6.6") {
    		metric["OPTIONAL_STATUS_VARS_5_6_6"].disable = true
    	}
    }

    if m.options.GaleraCluster {
    	metric["GALERA_VARS"].disable = true
    }

    if m.options.ExtraPerformanceMetrics && m.versionCompatible("5.6.0") {
    	if _, ok := m.resData["performance_schema"] {
    		metric["PERFORMANCE_VARS"].disable = true
    		m.getQueryExecTime95thus()
            m.queryExecTimePerSchema()
    	}
    }

    if m.options.SchemaSizeMetrics {
    	metric["SCHEMA_VARS"].disable = true
    	m.querySizePerschema()
    }

    // replication
    if m.options.Replication {
    	metric["SCHEMA_VARS"].disable = true
    	m.collectReplication()
    }

    m.submitMetrics()

	return nil
}

func (m *MysqlMonitor) submitMetrics() error {

    var (
    	tags   = make(map[string]string)
    	fields = make(map[string]interface{})
    )

    if m.Service != "" {
    	tags = m.Service
    }

    for tag, tagV := range m.Tags {
		tags[tag] = tagV
	}

    for key, kind := range metric {
   		if !kind.disable {
   			for k, item := range kind.metric {
   				// 数据类型转化(todo)
   				fields[item.name] = m.resData[k]
   			}
   		}
   	}

   	pt, err := io.MakeMetric(m.MetricName, tags, fields, time.Now())
	if err != nil {
		l.Errorf("make metric point error %v", err)
	}

	_, err = influxm.ParsePointsWithPrecision(pt, time.Now().UTC(), "")
	if err != nil {
		l.Errorf("[error] : %s", err.Error())
		return err
	}

	err = io.NamedFeed([]byte(pt), io.Metric, name)
	if err != nil {
		l.Errorf("push metric point error %v", err)
	}
}

func (m *MysqlMonitor) getQueryExecTime95thus() error {

}

func (m *MysqlMonitor) queryExecTimePerSchema() error {

}

func (m *MysqlMonitor) versionCompatible(version string) bool {

}

// status data
func (m *MysqlMonitor) getStatus() error {
	globalStatusSql := "SHOW /*!50002 GLOBAL */ STATUS;"
	rows, err := db.Query(globalStatusQuery)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var key string
		var val *sql.RawBytes

		if err = rows.Scan(&key, val); err != nil {
			return err
		}

		m.resData[key] = val
	}
}

// variables data
func (m *MysqlMonitor) getVariables() error {
	variablesSql := "SHOW GLOBAL VARIABLES;"
	rows, err := db.Query(variablesSql)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var key string
		var val *sql.RawBytes

		if err = rows.Scan(&key, val); err != nil {
			return err
		}

		m.resData[key] = val
	}
}

// innodb_engine_enabled
func (m *MysqlMonitor) innodbEngineEnabled() bool {
	innodbEnabledSql := `
	SELECT engine
	FROM information_schema.ENGINES
	WHERE engine='InnoDB' and support != 'no' and support != 'disabled';
	`
	result, err := db.Exec(innodbEnabledSql)
	if err != nil {
		l.Errorf("")
		return false
	}

	count, err := result.RowsAffected()
	if err != nil {
		l.Errorf("")
		return false
	}

	if count > 0 {
		return true
	}

	return false
}

// innodb status (todo)
func (m *MysqlMonitor) getInnodbStatus() error {
	innodbStatusSql := "SHOW /*!50000 ENGINE*/ INNODB STATUS;"
	rows, err := db.Query(innodbStatusSql)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var key string
		var val *sql.RawBytes

		if err = rows.Scan(&key, val); err != nil {
			return err
		}

		m.resData[key] = val
	}
}

// log stats
func (m *MysqlMonitor) getLogStats() error {
	logSql := "SHOW BINARY LOGS;"
	rows, err := db.Query(logSql)
	if err != nil {
		return err
	}
	defer rows.Close()

	var binaryLogSpace int
	for rows.Next() {
		var key string
		var val *sql.RawBytes

		if err = rows.Scan(&key, val); err != nil {
			return err
		}

		v, err := strconv.ParseInt(string(val), 10, 64)
		if err != nil {
			l.Errorf("func getLogStats, parse int %v", string(val))
			return err
		}

		binaryLogSpace += v

		m.resData["Binlog_space_usage_bytes"] = binaryLogSpace
	}
}

// Compute key cache utilization metric (todo)
func (m *MysqlMonitor) computeCacheUtilization() error {

}

func (m *MysqlMonitor) querySizePerschema() error {
	querySizePerschemaSql := `
	SELECT   table_schema, IFNULL(SUM(data_length+index_length)/1024/1024,0) AS total_mb
	FROM     information_schema.tables
	GROUP BY table_schema;
	`
	rows, err := m.db.Query(querySizePerschemaSql)
	if err != nil {
		return err
	}
	defer rows.Close()

	var schemaSize float64
	for rows.Next() {
		var key string
		var val *sql.RawBytes

		if err = rows.Scan(&key, val); err != nil {
			return err
		}

		size, err := strconv.ParseFloat(string(val), 64)
		if err != nil {
			l.Errorf("func getLogStats, parse int %v", string(val))
			return err
		}

		schemaKey := fmt.Sprintf("schema:%s", key)
		schemaSize[schemaKey] = size
	}
}

// replication (todo)
func (m *MysqlMonitor) collectReplication() error {

}

// "synthetic" metrics
func (m *MysqlMonitor) computeSynthetic() error {

}

func (m *MysqlMonitor) Test() (*inputs.TestResult, error) {
	m.test = true
	m.testData = nil

	m.handle()

	res := &inputs.TestResult{
		Result: m.testData,
		Desc:   "success!",
	}

	return res, nil
}

func init() {
	inputs.Add(name, func() inputs.Input {
		return &MysqlMonitor{}
	})
}
