// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package sqlserver

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/obfuscate"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
)

var (
	sample = `
[[inputs.sqlserver]]
  ## your sqlserver host ,example ip:port
  host = ""

  ## your sqlserver user,password
  user = ""
  password = ""

  ## Instance name. If not specified, a connection to the default instance is made.
  instance_name = ""

  ## Database name to query. Default is master.
  database = "master"

  ## by default, support TLS 1.2 and above.
  ## set to true if server side uses TLS 1.0 or TLS 1.1
  allow_tls10 = false

  ## connection timeout default: 30s
  connect_timeout = "30s"

  ## Metric name in metric_exclude_list will not be collected.
  metric_exclude_list = [""]

  ## parameters to be added to the connection string
  ## Examples:
  ##   "encrypt=disable"
  ##   "certificate=/path/to/cert.pem"
  ## reference: https://github.com/microsoft/go-mssqldb?tab=readme-ov-file#connection-parameters-and-dsn 
  #
  # connection_parameters = "encrypt=disable"

  ## (optional) collection interval, default is 10s
  interval = "10s"


  ## Set true to enable election
  election = true

  ## configure db_filter to filter out metrics from certain databases according to their database_name tag.
  ## If leave blank, no metric from any database is filtered out.
  # db_filter = ["some_db_instance_name", "other_db_instance_name"]


  ## Run a custom SQL query and collect corresponding metrics.
  #
  # [[inputs.sqlserver.custom_queries]]
  #   sql = '''
  #     select counter_name,cntr_type,cntr_value
  #     from sys.dm_os_performance_counters
  #   '''
  #   metric = "sqlserver_custom_stat"
  #   tags = ["counter_name","cntr_type"]
  #   fields = ["cntr_value"]

  # [inputs.sqlserver.log]
  # files = []
  # #grok pipeline script path
  # pipeline = "sqlserver.p"

  [inputs.sqlserver.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`

	pScrpit = `
grok(_,"%{TIMESTAMP_ISO8601:time} %{NOTSPACE:origin}\\s+%{GREEDYDATA:msg}")
default_time(time, "+0")
`

	inputName            = `sqlserver`
	customObjectFeedName = inputName + "/CO"
	loggingFeedName      = inputName + "/L"
	catalogName          = "db"
	l                    = logger.DefaultSLogger(inputName)

	collectCache        []*point.Point
	loggingCollectCache []*point.Point

	minInterval = time.Second * 5
	maxInterval = time.Second * 30
	query       = map[string]string{
		"sqlserver_waitstats":       sqlServerWaitStatsCategorized,
		"sqlserver_database_io":     sqlServerDatabaseIO,
		"sqlserver":                 sqlServerProperties,
		"sqlserver_schedulers":      sqlServerSchedulers,
		"sqlserver_volumespace":     sqlServerVolumeSpace,
		"sqlserver_database_size":   sqlServerDatabaseSize,
		"sqlserver_database_backup": sqlServerDatabaseBackup,
	}

	loggingQuery = map[string]string{
		"sqlserver_lock_table":  sqlServerLockTable,
		"sqlserver_lock_row":    sqlServerLockRow,
		"sqlserver_lock_dead":   sqlServerLockDead,
		"sqlserver_logical_io":  sqlServerLogicIO,
		"sqlserver_worker_time": sqlServerWorkerTime,
	}
)

type customQuery struct {
	SQL    string   `toml:"sql"`
	Metric string   `toml:"metric"`
	Tags   []string `toml:"tags"`
	Fields []string `toml:"fields"`
}

type Input struct {
	Host                 string            `toml:"host"`
	User                 string            `toml:"user"`
	Password             string            `toml:"password"`
	Interval             datakit.Duration  `toml:"interval"`
	InstanceName         string            `toml:"instance_name"`
	MetricExcludeList    []string          `toml:"metric_exclude_list"`
	ConnectionParameters string            `toml:"connection_parameters,omitempty"`
	Tags                 map[string]string `toml:"tags"`
	Log                  *sqlserverlog     `toml:"log"`
	Database             string            `toml:"database,omitempty"`
	CustomQuery          []*customQuery    `toml:"custom_queries"`
	AllowTLS10           bool              `toml:"allow_tls10,omitempty"`

	Timeout         string `toml:"connect_timeout"`
	timeoutDuration time.Duration

	QueryVersionDeprecated int      `toml:"query_version,omitempty"`
	ExcludeQuery           []string `toml:"exclude_query,omitempty"`

	DBFilter    []string `toml:"db_filter,omitempty"`
	dbFilterMap map[string]struct{}

	Version            string
	Uptime             int
	CollectCoStatus    string
	CollectCoErrMsg    string
	LastCustomerObject *customerObjectMeasurement

	lastErr error
	tail    *tailer.Tailer
	start   time.Time
	db      *sql.DB

	Election bool `toml:"election"`
	pauseCh  chan bool
	pause    bool

	semStop *cliutils.Sem // start stop signal
	feeder  dkio.Feeder
	tagger  datakit.GlobalTagger
	opt     point.Option

	collectFuncs        map[string]func() error
	collectQuery        map[string]string
	collectLoggingQuery map[string]string

	UpState int
}

type sqlserverlog struct {
	Files             []string `toml:"files"`
	Pipeline          string   `toml:"pipeline"`
	IgnoreStatus      []string `toml:"ignore"`
	CharacterEncoding string   `toml:"character_encoding"`
}

func newCountFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     desc,
	}
}

func newStringFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.String,
		Type:     inputs.String,
		Unit:     inputs.TODO,
		Desc:     desc,
	}
}

func newTimeFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationMS,
		Desc:     desc,
	}
}

func newByteFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     desc,
	}
}

func newKByteFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeKB,
		Desc:     desc,
	}
}

func newIntKByteFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeKB,
		Desc:     desc,
	}
}

func newBoolFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Bool,
		Type:     inputs.Gauge,
		Unit:     inputs.UnknownUnit,
		Desc:     desc,
	}
}

func newPercentFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.Percent,
		Desc:     desc,
	}
}

var reg = regexp.MustCompile(`\n|\s+`) //nolint:gocritic

func obfuscateSQL(text string) (sql string) {
	defer func() {
		sql = strings.TrimSpace(reg.ReplaceAllString(sql, " "))
	}()

	if out, err := obfuscate.NewObfuscator(nil).Obfuscate("sql", text); err != nil {
		return fmt.Sprintf("ERROR: failed to obfuscate: %s", err.Error())
	} else {
		return out.Query
	}
}

func transformData(measurement string, tags map[string]string, fields map[string]interface{}) {
	if tags == nil {
		return
	}
	switch measurement {
	case "sqlserver_lock_dead":
		if field, ok := fields["blocking_text"]; ok {
			if text, isString := field.(string); isString {
				fields["blocking_text"] = obfuscateSQL(text)
				fields["message"] = fields["blocking_text"]
			}
		}
	case "sqlserver_logical_io":
		if field, ok := fields["message"]; ok {
			if text, isString := field.(string); isString {
				fields["message"] = obfuscateSQL(text)
			}
		}
	case "sqlserver_database_size":
		if field, ok := fields["data_size"]; ok {
			if data, isUint := field.([]uint8); isUint {
				if dataSize, err := strconv.ParseFloat(string(data), 64); err == nil {
					fields["data_size"] = dataSize
				}
			}
		}
		if field, ok := fields["log_size"]; ok {
			if data, isUint := field.([]uint8); isUint {
				if dataSize, err := strconv.ParseFloat(string(data), 64); err == nil {
					fields["log_size"] = dataSize
				}
			}
		}
	default:
	}
}

var counterNameMap = map[string]string{
	"Processes blocked":                "processes_blocked",
	"Page Splits/sec":                  "page_splits",
	"Full Scans/sec":                   "full_scans",
	"Memory Grants Pending":            "memory_grants_pending",
	"Total Server Memory (KB)":         "total_server_memory",
	"SQL Cache Memory (KB)":            "sql_cache_memory",
	"Memory Grants Outstanding":        "memory_grants_outstanding",
	"Database Cache Memory (KB)":       "database_cache_memory",
	"Connection Memory (KB)":           "connection_memory",
	"Optimizer Memory (KB)":            "optimizer_memory",
	"Granted Workspace Memory (KB)":    "granted_workspace_memory",
	"Lock Memory (KB)":                 "lock_memory",
	"Stolen Server Memory (KB)":        "stolen_server_memory",
	"Log Pool Memory (KB)":             "log_pool_memory",
	"Buffer cache hit ratio":           "buffer_cache_hit_ratio",
	"Page life expectancy":             "page_life_expectancy",
	"Page reads/sec":                   "page_reads",
	"Page writes/sec":                  "page_writes",
	"Checkpoint pages/sec":             "checkpoint_pages",
	"Auto-Param Attempts/sec":          "auto_param_attempts",
	"Failed Auto-Params/sec":           "failed_auto_params",
	"Safe Auto-Params/sec":             "safe_auto_params",
	"Batch Requests/sec":               "batch_requests",
	"SQL Compilations/sec":             "sql_compilations",
	"SQL Re-Compilations/sec":          "sql_re_compilations",
	"Lock Waits/sec":                   "lock_waits",
	"Latch Waits/sec":                  "latch_waits",
	"Number of Deadlocks/sec":          "deadlocks",
	"Cache Object Counts":              "cache_object_counts",
	"Cache Pages":                      "cache_pages",
	"Transaction Delay":                "transaction_delay",
	"Flow Control/sec":                 "flow_control",
	"Version Store Size (KB)":          "version_store_size",
	"Version Cleanup rate (KB/s)":      "version_cleanup_rate",
	"Version Generation rate (KB/s)":   "version_generation_rate",
	"Longest Transaction Running Time": "longest_transaction_running_time",
	"Backup/Restore Throughput/sec":    "backup_restore_throughput",
	"Log Bytes Flushed/sec":            "log_bytes_flushed",
	"Log Flushes/sec":                  "log_flushes",
	"Log Flush Wait Time":              "log_flush_wait_time",
	"Transactions/sec":                 "transactions",
	"Write Transactions/sec":           "write_transactions",
	"Active Transactions":              "active_transactions",
	"User Connections":                 "user_connections",
}
