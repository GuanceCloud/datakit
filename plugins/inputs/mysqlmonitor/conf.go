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

type MysqlMonitor struct {
	Servers                             []string          `toml:"servers"`
	MetricName                          string            `toml:"metricName"`
	Interval                            string            `toml:"interval"`
	IntervalDuration                    time.Duration     `json:"-" toml:"-"`
	PerfEventsStatementsDigestTextLimit int64             `toml:"perf_events_statements_digest_text_limit"`
	PerfEventsStatementsLimit           int64             `toml:"perf_events_statements_limit"`
	PerfEventsStatementsTimeLimit       int64             `toml:"perf_events_statements_time_limit"`
	TableSchemaDatabases                []string          `toml:"table_schema_databases"`
	GatherProcessList                   bool              `toml:"gather_process_list"`
	GatherUserStatistics                bool              `toml:"gather_user_statistics"`
	GatherInfoSchemaAutoInc             bool              `toml:"gather_info_schema_auto_inc"`
	GatherInnoDBMetrics                 bool              `toml:"gather_innodb_metrics"`
	GatherSlaveStatus                   bool              `toml:"gather_slave_status"`
	GatherBinaryLogs                    bool              `toml:"gather_binary_logs"`
	GatherTableIOWaits                  bool              `toml:"gather_table_io_waits"`
	GatherTableLockWaits                bool              `toml:"gather_table_lock_waits"`
	GatherIndexIOWaits                  bool              `toml:"gather_index_io_waits"`
	GatherEventWaits                    bool              `toml:"gather_event_waits"`
	GatherTableSchema                   bool              `toml:"gather_table_schema"`
	GatherFileEventsStats               bool              `toml:"gather_file_events_stats"`
	GatherPerfEventsStatements          bool              `toml:"gather_perf_events_statements"`
	GatherGlobalVars                    bool              `toml:"gather_global_variables"`
	GatherGlobalStatus                  bool              `toml:"gather_global_status"`
	IntervalSlow                        string            `toml:"interval_slow"`
	Tags                                map[string]string `toml:"tags"`

	lastT            time.Time
	initDone         bool
	scanIntervalSlow uint32
	db               *sql.DB
}
