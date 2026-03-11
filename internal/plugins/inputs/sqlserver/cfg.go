// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package sqlserver

import (
	"database/sql"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/util"
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

  ## v2+ override all measurement names to "sqlserver_metric", default: v2
  ## If you want to use the old metric set, you can change it to "v1"
  measurement_version = "v2"

  ## configure db_filter to filter out metrics from certain databases according to their database_name tag.
  ## If leave blank, no metric from any database is filtered out.
  # db_filter = ["some_db_instance_name", "other_db_instance_name"]

  ## collect object
  [inputs.sqlserver.object]
    # Set true to enable collecting objects
    enabled = true

    # interval to collect sqlserver object which will be greater than collection interval
    interval = "600s"

  ## Database Monitoring (DBM) configuration
  ## DBM provides deep visibility into database performance by collecting query metrics, activity, and execution plans
  [inputs.sqlserver.dbm]
    # Set true to enable DBM metrics collection
    enabled = false
    # Maximum number of characters to collect from stored procedures (default: 500)
    stored_procedure_characters_limit = 500

  ## Config DBM metric (query metrics)
  ## Collects cumulative execution statistics of SQL queries aggregated by query signature, query plan hash, and database
  [inputs.sqlserver.dbm.metric]
    # Set true to enable collecting query metrics
    enabled = true
    # Collection interval for query metrics (default: 60s)
    collection_interval = "60s"
    # Maximum number of rows to collect from sys.dm_exec_query_stats (default: 10000)
    # This limits the initial query result size before aggregation
    dm_exec_query_stats_row_limit = 10000
    # Maximum number of queries to report per collection interval (default: 500)
    # Only the top N queries (sorted by derivative elapsed time) will be reported as metrics
    max_queries = 500
    # Lookback window in seconds for filtering queries (default: 300)
    # Only queries that executed within this time window (based on last_execution_time + last_elapsed_time) will be collected
    lookback_window = 300
    # Enable plan collection (default: true)
    plan_enabled = true
    # Plan object cache TTL (default: 1h)
    plan_cache_ttl = "1h"
    # Maximum runtime in seconds for plan collection (default: 30)
    # If collection takes longer than this, plan collection will be skipped
    max_run_time = 30

  ## Config DBM activity (current active queries)
  ## Collects information about currently executing queries and active sessions
  [inputs.sqlserver.dbm.activity]
    # Set true to enable collecting active query information
    enabled = true
    # Collection interval for activity metrics (default: 10s)
    collection_interval = "10s"
    # Maximum number of rows to collect from sys.dm_exec_sessions (default: 1000)
    dm_exec_sessions_row_limit = 1000

  ## Run a custom SQL query and collect corresponding metrics.
  #
  # [[inputs.sqlserver.custom_queries]]
  #   sql = '''
  #     select counter_name,cntr_type,cntr_value
  #     from sys.dm_os_performance_counters
  #   '''
  #   metric = "sqlserver_custom_stat"
  #   tags = ["counter_name","cntr_type"]
  #   interval = "10s"
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
	measurementSQLServer = "sqlserver_metric"
	customObjectFeedName = dkio.FeedSource(inputName, "CO")
	loggingFeedName      = dkio.FeedSource(inputName, "L")
	customQueryFeedName  = dkio.FeedSource(inputName, "custom_query")
	objectFeedName       = dkio.FeedSource(inputName, "O")
	dbmFeedName          = dkio.FeedSource(inputName, "DBM")
	catalogName          = "db"
	l                    = logger.DefaultSLogger(inputName)

	minInterval         = time.Second * 5
	maxInterval         = time.Minute * 10
	dbmMetricInterval   = datakit.Duration{Duration: time.Second * 60}
	dbmActivityInterval = datakit.Duration{Duration: time.Second * 10}
	dbmPlanCacheTTL     = datakit.Duration{Duration: time.Hour} // Default plan cache TTL: 1 hour
	query               = map[string]string{
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
	SQL      string           `toml:"sql"`
	Metric   string           `toml:"metric"`
	Tags     []string         `toml:"tags"`
	Fields   []string         `toml:"fields"`
	Interval datakit.Duration `toml:"interval"`
}

type sqlserverObject struct {
	Enable   bool             `toml:"enabled"`
	Interval datakit.Duration `toml:"interval"`

	name               string
	host               string
	port               string
	lastCollectionTime time.Time
	queryCache         map[string]string
}

// dbmConfig represents the top-level DBM configuration.
type dbmConfig struct {
	Enabled                        bool               `toml:"enabled"`
	StoredProcedureCharactersLimit int                `toml:"stored_procedure_characters_limit"`
	Metric                         *dbmMetricConfig   `toml:"metric"`
	Activity                       *dbmActivityConfig `toml:"activity"`
}

// dbmMetricConfig represents DBM metric (query metrics) configuration.
type dbmMetricConfig struct {
	Enabled                  bool             `toml:"enabled"`
	CollectionInterval       datakit.Duration `toml:"collection_interval"`
	DmExecQueryStatsRowLimit int              `toml:"dm_exec_query_stats_row_limit"`
	MaxQueries               int              `toml:"max_queries"`
	LookbackWindow           int              `toml:"lookback_window"`
	PlanEnabled              bool             `toml:"plan_enabled"`   // Enable plan collection
	PlanCacheTTL             datakit.Duration `toml:"plan_cache_ttl"` // Plan object cache TTL
	MaxRunTime               int              `toml:"max_run_time"`   // Maximum runtime in seconds for plan collection
}

// dbmActivityConfig represents DBM activity (current active queries) configuration.
type dbmActivityConfig struct {
	Enabled                bool             `toml:"enabled"`
	CollectionInterval     datakit.Duration `toml:"collection_interval"`
	DmExecSessionsRowLimit int              `toml:"dm_exec_sessions_row_limit"`
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

	Object       sqlserverObject `toml:"object"`
	objectMetric *objectMertric

	// DBM configuration
	Dbm *dbmConfig `toml:"dbm"`

	Version            string
	MajorVersion       int
	Uptime             int
	CollectCoStatus    string
	CollectCoErrMsg    string
	LastCustomerObject *customerObjectMeasurement

	lastErr error
	tail    *tailer.Tailer
	start   time.Time
	db      *sql.DB

	Election           bool   `toml:"election"`
	MeasurementVersion string `toml:"measurement_version"` // v1 or v2, default: v2
	pause              atomic.Bool

	semStop *cliutils.Sem // start stop signal
	feeder  dkio.Feeder
	tagger  datakit.GlobalTagger
	opt     point.Option

	collectFuncs        map[string]func() error
	collectQuery        map[string]string
	collectLoggingQuery map[string]string

	UpState             int
	ptsTime             time.Time
	collectCache        []*point.Point
	loggingCollectCache []*point.Point

	// DBM caches
	dbmGroup            *goroutine.Group
	dbmColumnsCacheMu   sync.Mutex
	dbmColumnsCache     map[string][]string              // Caches available columns for DBM queries
	dbmMetricCache      map[string]*dbmMetricCache       // Cache for previous statement cumulative values to compute derivatives
	dbmQueryObjectCache *expirable.LRU[string, struct{}] // Cache for reported SQL DBO objects
	dbmPlanObjectCache  *expirable.LRU[string, struct{}] // Cache for reported Plan DBO objects
}

func (ipt *Input) getKVsOpts(categorys ...point.Category) []point.Option {
	var opts []point.Option

	category := point.Metric
	if len(categorys) > 0 {
		category = categorys[0]
	}

	switch category { //nolint:exhaustive
	case point.Logging:
		opts = point.DefaultLoggingOptions()
	case point.Metric:
		opts = point.DefaultMetricOptions()
	case point.Object:
		opts = point.DefaultObjectOptions()
	default:
		opts = point.DefaultMetricOptions()
	}

	if ipt.Election {
		opts = append(opts, point.WithExtraTags(datakit.GlobalElectionTags()))
	}

	opts = append(opts, point.WithTimestamp(ipt.start.UnixNano()))

	return opts
}

func (ipt *Input) getKVs() point.KVs {
	var kvs point.KVs

	// add extended tags
	for k, v := range ipt.Tags {
		kvs = kvs.AddTag(k, v)
	}

	return kvs
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

func newMByteFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeMB,
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
		Unit:     inputs.NoUnit,
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

func transformData(measurement string, kvs point.KVs) point.KVs {
	if kvs == nil {
		return nil
	}

	switch measurement {
	case "sqlserver_lock_dead":
		if field := kvs.Fields().Get("blocking_text"); field != nil && !field.IsTag {
			if text, isString := field.Raw().(string); isString {
				obfuscatedText := util.ObfuscateSQL(text)
				kvs = kvs.Set("message", obfuscatedText)
				kvs = kvs.Set("blocking_text", obfuscatedText)
			}
		}
	case "sqlserver_logical_io":
		if field := kvs.Fields().Get("message"); field != nil && !field.IsTag {
			if text, isString := field.Raw().(string); isString {
				kvs = kvs.Set("message", util.ObfuscateSQL(text))
			}
		}
	case "sqlserver_database_size":
		for _, mfield := range []string{"data_size", "log_size"} {
			if field := kvs.Fields().Get(mfield); field != nil && !field.IsTag {
				if data, isUint := field.Raw().([]uint8); isUint {
					if dataSize, err := strconv.ParseFloat(string(data), 64); err == nil {
						kvs = kvs.Set(mfield, dataSize)
					} else {
						l.Warnf("parse %s failed: %s, ignore", mfield, err.Error())
					}
				}
			}
		}
	default:
	}

	return kvs
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
