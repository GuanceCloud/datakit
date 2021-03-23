package mysqlmonitor

import (
	"database/sql"
	"time"
)

const (
	configSample = `
[[inputs.mysqlMonitor]]
# specify servers via a url matching:
#  [username[:password]@][protocol[(address)]]/[?tls=[true|false|skip-verify|custom]]
#  see https://github.com/go-sql-driver/mysql#dsn-data-source-name
#  e.g.
#    servers = ["user:passwd@tcp(127.0.0.1:3306)/?tls=false"]
#    servers = ["user@tcp(127.0.0.1:3306)/?tls=false"]
# product support MySQL and MariaDB, default MySQL
 product = "MySQL"
# If no servers are specified, then localhost is used as the host.
 servers = ["tcp(127.0.0.1:3306)/"]
# scan interval
 interval = "10m"
# Selects the metric output format.
# if the list is empty, then metrics are gathered from all database tables
 table_schema_databases = []
# gather metrics from INFORMATION_SCHEMA.TABLES for databases provided above list
 gather_table_schema = true
# gather thread state counts from INFORMATION_SCHEMA.PROCESSLIST
 gather_process_list = true
# gather user statistics from INFORMATION_SCHEMA.USER_STATISTICS
 gather_user_statistics = true
# gather auto_increment columns and max values from information schema
 gather_info_schema_auto_inc = true
# gather metrics from INFORMATION_SCHEMA.INNODB_METRICS
 gather_innodb_metrics = true
# gather metrics from SHOW SLAVE STATUS command output
 gather_slave_status = true
# gather metrics from SHOW BINARY LOGS command output
 gather_binary_logs = true
# gather metrics from PERFORMANCE_SCHEMA.GLOBAL_VARIABLES
 gather_global_variables = true
# gather metrics from PERFORMANCE_SCHEMA.GLOBAL_STATUS
 gather_global_status = true
# gather metrics from PERFORMANCE_SCHEMA.TABLE_IO_WAITS_SUMMARY_BY_TABLE
 gather_table_io_waits = true
# gather metrics from PERFORMANCE_SCHEMA.TABLE_LOCK_WAITS
 gather_table_lock_waits = true
# gather metrics from PERFORMANCE_SCHEMA.TABLE_IO_WAITS_SUMMARY_BY_INDEX_USAGE
 gather_index_io_waits = true
# gather metrics from PERFORMANCE_SCHEMA.EVENT_WAITS
 gather_event_waits = true
# gather metrics from PERFORMANCE_SCHEMA.FILE_SUMMARY_BY_EVENT_NAME
 gather_file_events_stats = true
# gather metrics from PERFORMANCE_SCHEMA.EVENTS_STATEMENTS_SUMMARY_BY_DIGEST
 gather_perf_events_statements = true
# the limits for metrics form perf_events_statements
 perf_events_statements_digest_text_limit = 120
 perf_events_statements_limit = 250
 perf_events_statements_time_limit = 86400
# Use TLS but skip chain & host verification
 [inputs.mysqlMonitor.tags]
 tags1 = "value1"
`
)

type tls struct {
	TlsKey                              string			  `toml:"tls_key"`
	TlsCert                             string			  `toml:"tls_cert"`
	TlsCA                               string			  `toml:"tls_ca"`
}

type options struct {
	Replication             bool     `toml:"replication"`
	GaleraCluster           bool	 `toml:"galera_cluster"`
	ExtraStatusMetrics      bool	 `toml:"extra_status_metrics"`
	ExtraInnodbMetrics      bool	 `toml:"extra_innodb_metrics"`
	DisableInnodbMetrics    bool	 `toml:"disable_innodb_metrics"`
	SchemaSizeMetrics       bool	 `toml:"schema_size_metrics"`
	ExtraPerformanceMetrics bool	 `toml:"extra_performance_metrics"`
}

type MysqlMonitor struct {
	// 新配置
	MetricName                          string            `toml:"metricName"`
	Host                                string			  `toml:"host"`
	Port                                int			  	  `toml:"port"`
	User                                string			  `toml:"user"`
	Pass                                string			  `toml:"pass"`
	Sock                                string			  `toml:"sock"`
	Charset                             string			  `toml:"charset"`
	Timeout                             string			  `toml:"connect_timeout"`
	TimeoutDuration                     time.Duration     `toml:"-"`
	Tls									*tls			  `toml:"tls"`
	Service                             string		  	  `toml:"service"`
	Interval                            string            `toml:"interval"`
	IntervalDuration                    time.Duration     `toml:"-"`
	Tags                                map[string]string `toml:"tags"`
	options                             *options		  `toml:"options"`
	db               					*sql.DB
	resData								map[string]interface{}

	// 测试相关
	lastT            time.Time
	initDone         bool
	scanIntervalSlow uint32
	test             bool   `toml:"-, omitempty"`
	testData         []byte `toml:"-, omitempty"`

	// 兼容老版本配置反序列化
	Product                             string            `toml:"product, omitempty"`
	Servers                             []string          `toml:"servers, omitempty"`
	PerfEventsStatementsDigestTextLimit int64             `toml:"perf_events_statements_digest_text_limit, omitempty"`
	PerfEventsStatementsLimit           int64             `toml:"perf_events_statements_limit, omitempty"`
	PerfEventsStatementsTimeLimit       int64             `toml:"perf_events_statements_time_limit, omitempty"`
	TableSchemaDatabases                []string          `toml:"table_schema_databases, omitempty"`
	GatherProcessList                   bool              `toml:"gather_process_list, omitempty"`
	GatherUserStatistics                bool              `toml:"gather_user_statistics, omitempty"`
	GatherInfoSchemaAutoInc             bool              `toml:"gather_info_schema_auto_inc, omitempty"`
	GatherInnoDBMetrics                 bool              `toml:"gather_innodb_metrics, omitempty"`
	GatherSlaveStatus                   bool              `toml:"gather_slave_status, omitempty"`
	GatherBinaryLogs                    bool              `toml:"gather_binary_logs, omitempty"`
	GatherTableIOWaits                  bool              `toml:"gather_table_io_waits, omitempty"`
	GatherTableLockWaits                bool              `toml:"gather_table_lock_waits, omitempty"`
	GatherIndexIOWaits                  bool              `toml:"gather_index_io_waits, omitempty"`
	GatherEventWaits                    bool              `toml:"gather_event_waits, omitempty"`
	GatherTableSchema                   bool              `toml:"gather_table_schema, omitempty"`
	GatherFileEventsStats               bool              `toml:"gather_file_events_stats, omitempty"`
	GatherPerfEventsStatements          bool              `toml:"gather_perf_events_statements, omitempty"`
	GatherGlobalVars                    bool              `toml:"gather_global_variables, omitempty"`
	GatherGlobalStatus                  bool              `toml:"gather_global_status, omitempty"`
	IntervalSlow                        string            `toml:"interval_slow, omitempty"`
}
