// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package postgresql collects PostgreSQL metrics.
package postgresql

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/coreos/go-semver/semver"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cast"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
)

var (
	inputName            = "postgresql"
	customObjectFeedName = dkio.FeedSource(inputName, "CO")
	objectFeedName       = dkio.FeedSource(inputName, "O")
	loggingFeedName      = dkio.FeedSource(inputName, "L")
	customQueryFeedName  = dkio.FeedSource(inputName, "custom_query")

	catalogName                      = "db"
	l                                = logger.DefaultSLogger(inputName)
	_           inputs.ElectionInput = (*Input)(nil)
	kvMatcher                        = regexp.MustCompile(`(password|sslcert|sslkey|sslmode|sslrootcert)=\S+ ?`)
)

const (
	DBMetric          = "db"
	ReplicationMetric = "replication"
	BgwriterMetric    = "bgwriter"
	ConnectionMetric  = "connection"
	SlruMetric        = "slru"
	RelationMetric    = "relation"
	DynamicMetric     = "dynamic"
	ArchiverMetric    = "archiver"
	ReplicationSlot   = "replication_slot"
	IOMetric          = "io"
	DBMMetric         = "dbm_metric"
)

const sampleConfig = `
[[inputs.postgresql]]
  ## Server address
  # URI format
  # postgres://[datakit[:PASSWORD]]@localhost[/dbname]?sslmode=[disable|verify-ca|verify-full]
  # or simple string
  # host=localhost user=pqgotest password=... sslmode=... dbname=app_production

  address = "postgres://datakit:PASSWORD@localhost/postgres?sslmode=disable"

  ## Ignore databases which are gathered. Do not use with 'databases' option.
  #
  # ignored_databases = ["db1"]

  ## Specify the list of the databases to be gathered. Do not use with the 'ignored_databases' option.
  #
  # databases = ["db1"]

  ## Specify the name used as the "server" tag.
  #
  # outputaddress = "db01"

  ## Collect interval
  # Time unit: "ns", "us", "ms", "s", "m", "h"
  #
  interval = "10s"

  ## Set true to enable election
  #
  election = true

  ## Metric name in metric_exclude_list will not be collected.
  #
  metric_exclude_list = [""]

  ## collect object
  [inputs.postgresql.object]
    # Set true to enable collecting objects
    enabled = true

    # interval to collect postgresql object which will be greater than collection interval
    interval = "600s"

    [inputs.postgresql.object.collect_schemas]
      # Set true to enable collecting schemas
      enabled = true

      # Maximum number of tables to collect
      max_tables = 300

      # Set true to enable auto discovery database
      auto_discovery_database = false

     # Maximum number of databases to collect
     max_database = 100

  ## Relations config
  #
  # The list of relations/tables can be specified to track per-relation metrics. To collect relation
  # relation_name refer to the name of a relation, either relation_name or relation_regex must be set.
  # relation_regex is a regex rule, only takes effect when relation_name is not set.
  # schemas used for filtering, ignore this field when it is empty
  # relkind can be a list of the following options:
  #   r(ordinary table), i(index), S(sequence), t(TOAST table), p(partitioned table),
  #   m(materialized view), c(composite type), f(foreign table)
  #
  # [[inputs.postgresql.relations]]
  # relation_name = "<TABLE_NAME>"
  # relation_regex = "<TABLE_PATTERN>"
  # schemas = ["public"]
  # relkind = ["r", "p"]

  ## Run a custom SQL query and collect corresponding metrics.
  #
  # [[inputs.postgresql.custom_queries]]
  #   sql = '''
  #     select datname,numbackends,blks_read
  #     from pg_stat_database
  #     limit 10
  #   '''
  #   metric = "postgresql_custom_stat"
  #   tags = ["datname" ]
  #   fields = ["numbackends", "blks_read"]
  #   interval = "10s"

  ## Log collection
  #
  # [inputs.postgresql.log]
  # files = []
  # pipeline = "postgresql.p"

  ## Custom tags
  #
  [inputs.postgresql.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
`

//nolint:lll
const pipelineCfg = `
add_pattern("log_date", "%{YEAR}-%{MONTHNUM}-%{MONTHDAY}%{SPACE}%{HOUR}:%{MINUTE}:%{SECOND}%{SPACE}(?:CST|UTC)")
add_pattern("status", "(LOG|ERROR|FATAL|PANIC|WARNING|NOTICE|INFO)")
add_pattern("session_id", "([.0-9a-z]*)")
add_pattern("application_name", "(\\[%{GREEDYDATA:application_name}?\\])")
add_pattern("remote_host", "(\\[\\[?%{HOST:remote_host}?\\]?\\])")
grok(_, "%{log_date:time}%{SPACE}\\[%{INT:process_id}\\]%{SPACE}(%{WORD:db_name}?%{SPACE}%{application_name}%{SPACE}%{USER:user}?%{SPACE}%{remote_host}%{SPACE})?%{session_id:session_id}%{SPACE}(%{status:status}:)?")

# default
grok(_, "%{log_date:time}%{SPACE}\\[%{INT:process_id}\\]%{SPACE}%{status:status}")

nullif(remote_host, "")
nullif(session_id, "")
nullif(application_name, "")
nullif(user, "")
nullif(db_name, "")

group_in(status, [""], "INFO")

default_time(time)
`

type Rows interface {
	Close()
	Columns() ([]string, error)
	Next() bool
	Scan(...interface{}) error
}

type Service interface {
	Start() error
	Stop()
	Ping() error
	Query(string) (Rows, error)
	QueryByDatabase(string, string) (Rows, error)
	SetAddress(string)
	GetColumnMap(scanner, []string) (map[string]*interface{}, error)
}

type scanner interface {
	Scan(dest ...interface{}) error
}

type Relation struct {
	RelationName  string   `toml:"relation_name"`
	RelationRegex string   `toml:"relation_regex"`
	Schemas       []string `toml:"schemas"`
	RelKind       []string `toml:"relkind"`
}

type queryCacheItem struct {
	ptsTime         time.Time
	q               string
	measurementInfo *inputs.MeasurementInfo
}

// customQuery represents configuration for executing a custom query.
type customQuery struct {
	SQL      string           `toml:"sql"`
	Metric   string           `toml:"metric"`
	Tags     []string         `toml:"tags"`
	Fields   []string         `toml:"fields"`
	Interval datakit.Duration `toml:"interval"`
}

type pgObject struct {
	Enable         bool                 `toml:"enabled"`
	Interval       datakit.Duration     `toml:"interval"`
	CollectSchemas configCollectSchemas `toml:"collect_schemas"`

	name               string
	lastCollectionTime time.Time
}

type configCollectSchemas struct {
	Enabled               bool `toml:"enabled"`
	MaxTables             int  `toml:"max_tables"`
	MaxDatabases          int  `toml:"max_databases"`
	AutoDiscoveryDatabase bool `toml:"auto_discovery_database"`
}

type Input struct {
	Address           string            `toml:"address"`
	Outputaddress     string            `toml:"outputaddress"`
	IgnoredDatabases  []string          `toml:"ignored_databases"`
	Databases         []string          `toml:"databases"`
	Interval          datakit.Duration  `toml:"interval"`
	MetricExcludeList []string          `toml:"metric_exclude_list"`
	Tags              map[string]string `toml:"tags"`
	mergedTags        map[string]string
	Relations         []Relation     `toml:"relations"`
	CustomQuery       []*customQuery `toml:"custom_queries"`
	Log               *postgresqllog `toml:"log"`
	Object            pgObject       `toml:"object"`

	Uptime             int
	CollectCoStatus    string
	CollectCoErrMsg    string
	LastCustomerObject *customerObjectMeasurement

	MaxLifetimeDeprecated string `toml:"max_lifetime,omitempty"`

	service      Service
	tail         *tailer.Tailer
	collectCache map[point.Category][]*point.Point
	host         string
	port         uint16
	dbName       string

	Election bool `toml:"election"`
	pause    bool
	pauseCh  chan bool

	feeder dkio.Feeder
	tagger datakit.GlobalTagger

	version  *semver.Version
	isAurora bool
	semStop  *cliutils.Sem // start stop signal
	ptsTime  time.Time

	collectFuncs     map[string]func() error
	relationMetrics  map[string]relationMetric
	metricQueryCache map[string]*queryCacheItem

	UpState int

	objectMetric *objectMertric
}

type postgresqllog struct {
	Files             []string `toml:"files"`
	Pipeline          string   `toml:"pipeline"`
	IgnoreStatus      []string `toml:"ignore"`
	CharacterEncoding string   `toml:"character_encoding"`
	MultilineMatch    string   `toml:"multiline_match"`
}

func (*Input) Catalog() string {
	return catalogName
}

func (*Input) SampleConfig() string {
	return sampleConfig
}

func (*Input) AvailableArchs() []string {
	return datakit.AllOSWithElection
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&inputMeasurement{},
		&lockMeasurement{},
		&indexMeasurement{},
		&replicationMeasurement{},
		&replicationSlotMeasurement{},
		&sizeMeasurement{},
		&statIOMeasurement{},
		&statMeasurement{},
		&slruMeasurement{},
		&bgwriterMeasurement{},
		&connectionMeasurement{},
		&conflictMeasurement{},
		&archiverMeasurement{},
		&inputs.UpMeasurement{},
		&postgresqlObjectMeasurement{},
	}
}

func (*Input) PipelineConfig() map[string]string {
	return map[string]string{
		"postgresql": pipelineCfg,
	}
}

//nolint:lll
func (ipt *Input) LogExamples() map[string]map[string]string {
	return map[string]map[string]string{
		"postgresql": {
			"PostgreSQL log": `2021-05-31 15:23:45.110 CST [74305] test [pgAdmin 4 - DB:postgres] postgres [127.0.0.1] 60b48f01.12241 LOG: statement: 		SELECT psd.*, 2^31 - age(datfrozenxid) as wraparound, pg_database_size(psd.datname) as pg_database_size 		FROM pg_stat_database psd 		JOIN pg_database pd ON psd.datname = pd.datname 		WHERE psd.datname not ilike 'template%' AND psd.datname not ilike 'rdsadmin' 		AND psd.datname not ilike 'azure_maintenance' AND psd.datname not ilike 'postgres'`,
		},
	}
}

func (ipt *Input) ElectionEnabled() bool {
	return ipt.Election
}

func (ipt *Input) GetPipeline() []tailer.Option {
	opts := []tailer.Option{
		tailer.WithSource(inputName),
		tailer.WithService(inputName),
	}
	if ipt.Log != nil {
		opts = append(opts, tailer.WithPipeline(ipt.Log.Pipeline))
	}
	return opts
}

func (ipt *Input) SanitizedAddress() (sanitizedAddress string, err error) {
	var canonicalizedAddress string

	if ipt.Outputaddress != "" {
		return ipt.Outputaddress, nil
	}

	if strings.HasPrefix(ipt.Address, "postgres://") || strings.HasPrefix(ipt.Address, "postgresql://") {
		if canonicalizedAddress, err = parseURL(ipt.Address); err != nil {
			return sanitizedAddress, err
		}
	} else {
		canonicalizedAddress = ipt.Address
	}

	sanitizedAddress = kvMatcher.ReplaceAllString(canonicalizedAddress, "")

	return sanitizedAddress, err
}

func (ipt *Input) getQueryPoints(cache *queryCacheItem, dealColumnFn ...func(map[string]*interface{})) ([]*point.Point, error) {
	var (
		columns []string
		err     error
		points  []*point.Point
	)

	if cache == nil || cache.q == "" {
		return nil, fmt.Errorf("query cache is empty")
	}

	measurementInfo := cache.measurementInfo
	if measurementInfo == nil {
		measurementInfo = inputMeasurement{}.Info()
	}

	start := time.Now()
	rows, err := ipt.service.Query(cache.q)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	sqlQueryCostSummary.WithLabelValues(measurementInfo.Name,
		measurementInfo.Name).Observe(time.Since(start).Seconds())
	if columns, err = rows.Columns(); err != nil {
		return nil, err
	}

	for rows.Next() {
		columnMap, err := ipt.service.GetColumnMap(rows, columns)
		if err != nil {
			return nil, err
		}
		if len(dealColumnFn) > 0 {
			for _, fn := range dealColumnFn {
				fn(columnMap)
			}
		}
		if point := ipt.makePoints(columnMap, measurementInfo); point != nil {
			point.SetTime(cache.ptsTime)
			points = append(points, point)
		}
	}

	return points, nil
}

func (ipt *Input) executeQuery(cache *queryCacheItem) error {
	// set point's time on non-custome query
	cache.ptsTime = ipt.ptsTime

	if points, err := ipt.getQueryPoints(cache); err != nil {
		return fmt.Errorf("getQueryPoints error: %w", err)
	} else {
		ipt.collectCache[point.Metric] = append(ipt.collectCache[point.Metric], points...)
	}

	return nil
}

func (ipt *Input) getDBMetrics() error {
	cache, ok := ipt.metricQueryCache[DBMetric]
	if !ok {
		query := `
		SELECT psd.*,
			2^31 - age(datfrozenxid) as wraparound,
			psd.datname as db,
			pg_database_size(psd.datname) as database_size
		FROM pg_stat_database psd
		JOIN pg_database pd ON psd.datname = pd.datname
		WHERE psd.datname not ilike 'template%'   AND psd.datname not ilike 'rdsadmin'
		AND psd.datname not ilike 'azure_maintenance'   AND psd.datname not ilike 'postgres'
		`
		if len(ipt.IgnoredDatabases) != 0 {
			query += fmt.Sprintf(` AND psd.datname NOT IN ('%s')`, strings.Join(ipt.IgnoredDatabases, "','"))
		} else if len(ipt.Databases) != 0 {
			query += fmt.Sprintf(` AND psd.datname IN ('%s')`, strings.Join(ipt.Databases, "','"))
		}

		cache = &queryCacheItem{
			q:               query,
			measurementInfo: inputMeasurement{}.Info(),
		}
		ipt.metricQueryCache[DBMetric] = cache
		l.Infof("Query for metric [%s]: %s", cache.measurementInfo.Name, query)
	}

	return ipt.executeQuery(cache)
}

func (ipt *Input) getDynamicQueryMetrics() error {
	cache, ok := ipt.metricQueryCache[DynamicMetric]
	if !ok {
		query := `
		SELECT
		datname AS db,
		confl_tablespace,
		confl_lock,
		confl_snapshot,
		confl_bufferpin,
		confl_deadlock
	FROM pg_stat_database_conflicts
	`
		if len(ipt.IgnoredDatabases) != 0 {
			query += fmt.Sprintf(` WHERE datname NOT IN ('%s')`, strings.Join(ipt.IgnoredDatabases, "','"))
		} else if len(ipt.Databases) != 0 {
			query += fmt.Sprintf(` WHERE datname IN ('%s')`, strings.Join(ipt.Databases, "','"))
		}

		cache = &queryCacheItem{
			q:               query,
			measurementInfo: conflictMeasurement{}.Info(),
		}
		ipt.metricQueryCache[DynamicMetric] = cache
		l.Infof("Query for metric [%s], %s", cache.measurementInfo.Name, query)
	}

	return ipt.executeQuery(cache)
}

func (ipt *Input) getBgwMetrics() error {
	cache, ok := ipt.metricQueryCache[BgwriterMetric]

	if !ok {
		query := `
		select * FROM pg_stat_bgwriter
	`
		cache = &queryCacheItem{
			q:               query,
			measurementInfo: bgwriterMeasurement{}.Info(),
		}
		ipt.metricQueryCache[BgwriterMetric] = cache
		l.Infof("Query for metric [%s]: %s", cache.measurementInfo.Name, query)
	}
	err := ipt.executeQuery(cache)
	return err
}

func (ipt *Input) getRelationMetrics() error {
	if len(ipt.Relations) == 0 {
		return fmt.Errorf("no relations set")
	}

	for _, relationInfo := range ipt.relationMetrics {
		cacheName := fmt.Sprintf("%s_%s", RelationMetric, relationInfo.name)
		cache, ok := ipt.metricQueryCache[cacheName]

		if !ok {
			query := ipt.getRelationQuery(relationInfo.query, relationInfo.schemaField)
			if len(query) == 0 {
				l.Warnf("relation query is empty, ignore %s", relationInfo.name)
				continue
			}

			cache = &queryCacheItem{
				q:               query,
				measurementInfo: relationInfo.measurementInfo,
			}
			ipt.metricQueryCache[cacheName] = cache
			l.Infof("Query for metric [%s], %s", cache.measurementInfo.Name, query)
		}

		if err := ipt.executeQuery(cache); err != nil {
			l.Warnf("collect %s error: %s", relationInfo.name, err.Error())
		}
	}

	return nil
}

func (ipt *Input) getArchiverMetrics() error {
	cache, ok := ipt.metricQueryCache[ArchiverMetric]
	if !ok {
		query := "select archived_count, failed_count as archived_failed_count FROM pg_stat_archiver"
		cache = &queryCacheItem{
			q:               query,
			measurementInfo: archiverMeasurement{}.Info(),
		}
		ipt.metricQueryCache[ArchiverMetric] = cache
		l.Infof("Query for metric [%s]: %s", cache.measurementInfo.Name, query)
	}

	return ipt.executeQuery(cache)
}

func (ipt *Input) getReplicationMetrics() error {
	if ipt.isAurora {
		l.Debugf("ignore replication check in aurora")
		return nil
	}

	cache, ok := ipt.metricQueryCache[ReplicationMetric]
	if !ok {
		query := ""
		if V100.LessThan(*ipt.version) || V100.Equal(*ipt.version) {
			query = `
SELECT CASE WHEN pg_last_wal_receive_lsn() IS NULL OR pg_last_wal_receive_lsn() = pg_last_wal_replay_lsn()
	THEN 0 ELSE GREATEST (0, EXTRACT (EPOCH FROM now() - pg_last_xact_replay_timestamp())) END AS replication_delay,
	abs(pg_wal_lsn_diff(pg_last_wal_receive_lsn(), pg_last_wal_replay_lsn())) AS replication_delay_bytes
         WHERE (SELECT pg_is_in_recovery())
`
		} else if V91.LessThan(*ipt.version) || V91.Equal(*ipt.version) {
			query = `
SELECT CASE WHEN pg_last_xlog_receive_location() IS NULL OR pg_last_xlog_receive_location() = pg_last_xlog_replay_location()
	THEN 0 ELSE GREATEST (0, EXTRACT (EPOCH FROM now() - pg_last_xact_replay_timestamp())) END AS replication_delay
`
			if V92.LessThan(*ipt.version) || V92.Equal(*ipt.version) {
				query += `,
	abs(pg_xlog_location_diff(pg_last_xlog_receive_location(), pg_last_xlog_replay_location())) AS replication_delay_bytes
`
			}

			query += " WHERE (SELECT pg_is_in_recovery())"
		}
		cache = &queryCacheItem{
			q:               query,
			measurementInfo: replicationMeasurement{}.Info(),
		}
		ipt.metricQueryCache[ReplicationMetric] = cache
		l.Infof("Query for metric [%s]: %s", cache.measurementInfo.Name, query)
	}

	return ipt.executeQuery(cache)
}

func (ipt *Input) getReplicationSlotMetrics() error {
	cache, ok := ipt.metricQueryCache[ReplicationSlot]
	if !ok {
		query := `
SELECT
    stat.slot_name,
    slot_type,
    CASE WHEN active THEN 'active' ELSE 'inactive' END,
    spill_txns, spill_count, spill_bytes,
    stream_txns, stream_count, stream_bytes,
    total_txns, total_bytes
FROM pg_stat_replication_slots AS stat
JOIN pg_replication_slots ON pg_replication_slots.slot_name = stat.slot_name
`
		cache = &queryCacheItem{
			q:               query,
			measurementInfo: replicationSlotMeasurement{}.Info(),
		}
		ipt.metricQueryCache[ReplicationSlot] = cache
		l.Infof("Query for metric [%s]: %s", cache.measurementInfo.Name, query)
	}

	return ipt.executeQuery(cache)
}

func (ipt *Input) getIOMetrics() error {
	cache, ok := ipt.metricQueryCache[IOMetric]
	if !ok {
		// PG16+
		query := `
SELECT backend_type,
       object,
       context,
       evictions,
       extend_time,
       extends,
       fsync_time,
       fsyncs,
       hits,
       read_time,
       reads,
       write_time,
       writes
FROM pg_stat_io
LIMIT 200
		`

		cache = &queryCacheItem{
			q:               query,
			measurementInfo: ioMeasurement{}.Info(),
		}
		ipt.metricQueryCache[IOMetric] = cache
		l.Infof("Query for metric [%s]: %s", cache.measurementInfo.Name, query)
	}

	return ipt.executeQuery(cache)
}

func (ipt *Input) getSlruMetrics() error {
	cache, ok := ipt.metricQueryCache[SlruMetric]
	if !ok {
		query := `
SELECT name, blks_zeroed, blks_hit, blks_read,
	blks_written , blks_exists, flushes, truncates
FROM pg_stat_slru
`
		cache = &queryCacheItem{
			q:               query,
			measurementInfo: slruMeasurement{}.Info(),
		}
		ipt.metricQueryCache[SlruMetric] = cache
		l.Infof("Query for metric [%s]: %s", cache.measurementInfo.Name, query)
	}
	return ipt.executeQuery(cache)
}

func (ipt *Input) getConnectionMetrics() error {
	cache, ok := ipt.metricQueryCache[ConnectionMetric]
	if !ok {
		query := `
		WITH max_con AS (SELECT setting::float FROM pg_settings WHERE name = 'max_connections')
		SELECT MAX(setting) AS max_connections, SUM(numbackends)/MAX(setting) AS percent_usage_connections
		FROM pg_stat_database, max_con
	`
		cache = &queryCacheItem{
			q:               query,
			measurementInfo: connectionMeasurement{}.Info(),
		}
		ipt.metricQueryCache[ConnectionMetric] = cache
		l.Infof("Query for metric [%s]: %s", cache.measurementInfo.Name, query)
	}

	err := ipt.executeQuery(cache)
	return err
}

func (ipt *Input) getRelationQuery(query, schemaField string) string {
	relationFilters := []string{}
	for _, relation := range ipt.Relations {
		relationFilter := []string{}
		switch {
		case len(relation.RelationName) > 0:
			relationFilter = append(relationFilter, fmt.Sprintf("( relname = '%s'", relation.RelationName))
		case len(relation.RelationRegex) > 0:
			relationFilter = append(relationFilter, fmt.Sprintf("( relname ~ '%s'", relation.RelationRegex))
		default:
			l.Warnf("relation_name and relation_regex are both empty, ignore this relation config: %+#v", relation)
			continue
		}

		if len(relation.Schemas) > 0 {
			schemaFilter := ""
			comma := ""
			for _, schema := range relation.Schemas {
				schemaFilter += fmt.Sprintf("%s'%s'", comma, schema)
				comma = ","
			}
			relationFilter = append(relationFilter, fmt.Sprintf("AND %s = ANY(array[%s]::text[])", schemaField, schemaFilter))
		}

		if strings.Contains(query, "FROM pg_locks") && len(relation.RelKind) > 0 {
			relKindFilter := ""
			comma := ""
			for _, k := range relation.RelKind {
				relKindFilter += fmt.Sprintf("%s'%s'", comma, k)
				comma = ","
			}
			relationFilter = append(relationFilter, fmt.Sprintf("AND relkind = ANY(array[%s])", relKindFilter))
		}

		relationFilter = append(relationFilter, ")")
		relationFilters = append(relationFilters, strings.Join(relationFilter, " "))
	}

	if len(relationFilters) == 0 {
		return ""
	}

	relationQuery := fmt.Sprintf("(%s)", strings.Join(relationFilters, " OR "))

	return fmt.Sprintf(query, relationQuery)
}

func (ipt *Input) setAurora() {
	rows, err := ipt.service.Query("select AURORA_VERSION();")
	if err != nil {
		l.Debugf("The db is not aurora")
		return
	}

	defer rows.Close() //nolint:errcheck

	ipt.isAurora = true
}

func (ipt *Input) setVersion() error {
	rows, err := ipt.service.Query("SHOW SERVER_VERSION;")
	if err != nil {
		return fmt.Errorf("query version failed: %w", err)
	}

	defer rows.Close() //nolint:errcheck

	var rawVersion string

	for rows.Next() {
		err := rows.Scan(&rawVersion)
		if err != nil {
			return err
		}
	}

	if len(rawVersion) > 0 {
		if ipt.version, err = semver.NewVersion(rawVersion); err != nil {
			parts := strings.Split(rawVersion, " ")
			if len(parts) == 0 {
				return fmt.Errorf("invalid postgresql raw version: %s", rawVersion)
			} else {
				verParts := strings.Split(parts[0], ".")
				verPartsInt := []int64{}
				for len(verParts) < 3 {
					verParts = append(verParts, "0")
				}

				isValid := true

				for _, v := range verParts {
					if vInt, err := strconv.ParseInt(v, 10, 64); err != nil {
						isValid = false
						break
					} else {
						verPartsInt = append(verPartsInt, vInt)
					}
				}
				if isValid {
					ipt.version = &semver.Version{
						Major: verPartsInt[0],
						Minor: verPartsInt[1],
						Patch: verPartsInt[2],
					}
				} else { // eg 11beta3
					re := regexp.MustCompile(`(\d+)([a-zA-Z]+)(\d+)`)
					result := re.FindAllStringSubmatch(parts[0], -1)

					if len(result) == 0 {
						return fmt.Errorf("invalid postgresql version: %s", rawVersion)
					}

					version := result[0]

					if len(version) != 4 {
						return fmt.Errorf("parse postgresql version error: %+#v, raw version: %s", version, rawVersion)
					}

					major, err := strconv.ParseInt(version[1], 10, 64)
					if err != nil {
						return fmt.Errorf("invalid postgresql version: %s", rawVersion)
					}

					ipt.version = &semver.Version{
						Major:      major,
						PreRelease: semver.PreRelease(fmt.Sprintf("%s.%s", version[2], version[3])),
					}
				}
			}
		}
	}

	return nil
}

func (ipt *Input) startService() error {
	if err := ipt.service.Ping(); err != nil {
		if err := ipt.service.Start(); err != nil {
			return fmt.Errorf("start service failed: %w", err)
		}
	}

	return nil
}

func (ipt *Input) Collect() error {
	var err error

	if err := ipt.startService(); err != nil {
		return fmt.Errorf("start service failed: %w", err)
	}

	if err := ipt.setVersion(); err != nil {
		return fmt.Errorf("set version failed: %w", err)
	}

	g := goroutine.NewGroup(goroutine.Option{Name: goroutine.GetInputName(inputName)})

	err = ipt.getUptime()
	if err != nil {
		l.Errorf("Failed to get uptime: %v", err)
	}

	// collect metrics
	g.Go(func(ctx context.Context) error {
		for name, collector := range ipt.collectFuncs {
			if err := collector(); err != nil {
				l.Warnf("collect %s metrics error: %s", name, err.Error())
			}
		}
		return nil
	})

	return g.Wait()
}

func (ipt *Input) makePoints(columnMap map[string]*interface{}, measurementInfo *inputs.MeasurementInfo) *point.Point {
	kvs := ipt.getKVs()
	if measurementInfo == nil {
		measurementInfo = inputMeasurement{}.Info()
	}

	for col, val := range columnMap {
		isMeasurementTag := false
		if _, ok := measurementInfo.Tags[col]; !ok {
			if _, ok := measurementInfo.Fields[col]; !ok {
				continue
			}
		} else {
			isMeasurementTag = true
		}

		stringVal := "" // unit8 to string
		isField := true // set isField to be true when the type is uint8 or string
		if *val != nil {
			switch trueVal := (*val).(type) {
			case []uint8:
				stringVal = string(trueVal)
			case string:
				stringVal = trueVal
			default:
				stringVal = cast.ToString(trueVal)
				isField = true
			}

			if isMeasurementTag {
				kvs = kvs.AddTag(col, stringVal)
			} else {
				switch {
				case len(stringVal) > 0:
					kvs = kvs.Add(col, cast.ToFloat64(stringVal), false, true)
				case isField:
					kvs = kvs.Add(col, *val, false, true)
				}
			}
		}
	}

	if kvs.FieldCount() > 0 {
		cat := point.Metric
		metricName := inputName
		if measurementInfo != nil {
			metricName = measurementInfo.Name
			cat = measurementInfo.Cat
		}

		opts := ipt.getKVsOpts(cat)
		return point.NewPointV2(metricName, kvs, opts...)
	}

	return nil
}

func (ipt *Input) getKVs() point.KVs {
	var kvs point.KVs

	// add extended tags
	for k, v := range ipt.mergedTags {
		kvs = kvs.AddTag(k, v)
	}

	return kvs
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

	return opts
}

func (ipt *Input) RunPipeline() {
	if ipt.Log == nil || len(ipt.Log.Files) == 0 {
		return
	}

	opts := []tailer.Option{
		tailer.WithSource(inputName),
		tailer.WithService(inputName),
		tailer.WithPipeline(ipt.Log.Pipeline),
		tailer.WithIgnoreStatus(ipt.Log.IgnoreStatus),
		tailer.WithCharacterEncoding(ipt.Log.CharacterEncoding),
		tailer.EnableMultiline(true),
		tailer.WithMaxMultilineLength(int64(float64(config.Cfg.Dataway.MaxRawBodySize) * 0.8)),
		tailer.WithMultilinePatterns([]string{ipt.Log.MultilineMatch}),
		tailer.WithGlobalTags(inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, "")),
		tailer.EnableDebugFields(config.Cfg.EnableDebugFields),
	}

	var err error
	ipt.tail, err = tailer.NewTailer(ipt.Log.Files, opts...)
	if err != nil {
		l.Error(err)
		ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(inputName),
		)
		return
	}

	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_postgresql"})
	g.Go(func(ctx context.Context) error {
		ipt.tail.Start()
		return nil
	})
}

const (
	maxInterval = 10 * time.Minute
	minInterval = 10 * time.Second
)

func (ipt *Input) initDB() error {
	config, err := pgxpool.ParseConfig(ipt.Address)
	if err != nil {
		return fmt.Errorf("parse config error: %w", err)
	}
	ipt.port = config.ConnConfig.Port
	ipt.host = config.ConnConfig.Host

	ipt.service.SetAddress(ipt.Address)
	err = ipt.service.Start()
	if err != nil {
		return err
	}

	if err := ipt.setVersion(); err != nil {
		return err
	}

	ipt.setAurora()

	// set default Tags
	dbName, err := getDBName(ipt.Address)
	if err != nil {
		l.Warnf("get db name error: %s, use postgres instead", err.Error())
		dbName = "postgres"
	}
	ipt.dbName = dbName

	if ipt.Tags == nil {
		ipt.Tags = map[string]string{}
	}
	ipt.Tags["db"] = dbName

	return nil
}

func (ipt *Input) initCfg() {
	ipt.collectCache = map[point.Category][]*point.Point{}

	if ipt.Election {
		ipt.mergedTags = inputs.MergeTags(ipt.tagger.ElectionTags(), ipt.Tags, ipt.host)
	} else {
		ipt.mergedTags = inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, ipt.host)
	}

	if v, ok := ipt.mergedTags["host"]; ok {
		ipt.host = v
	}

	ipt.mergedTags["server"] = fmt.Sprintf("%s:%d", ipt.host, ipt.port)

	l.Infof("merged tags: %+#v", ipt.mergedTags)

	// init query cache
	ipt.metricQueryCache = map[string]*queryCacheItem{}

	// setup collectors
	ipt.collectFuncs = map[string]func() error{
		"postgresql":             ipt.getDBMetrics,
		"postgresql_replication": ipt.getReplicationMetrics,
		"postgresql_bgwriter":    ipt.getBgwMetrics,
		"postgresql_connection":  ipt.getConnectionMetrics,
	}

	if V140.LessThan(*ipt.version) || V140.Equal(*ipt.version) {
		ipt.collectFuncs["postgresql_replication_slot"] = ipt.getReplicationSlotMetrics
	}

	if V130.LessThan(*ipt.version) || V130.Equal(*ipt.version) {
		ipt.collectFuncs["postgresql_slru"] = ipt.getSlruMetrics
	}

	if V92.LessThan(*ipt.version) || V92.Equal(*ipt.version) {
		ipt.collectFuncs["postgresql_conflict"] = ipt.getDynamicQueryMetrics
	}

	if V94.LessThan(*ipt.version) || V94.Equal(*ipt.version) {
		ipt.collectFuncs["postgresql_archiver"] = ipt.getArchiverMetrics
	}

	if V160.LessThan(*ipt.version) || V160.Equal(*ipt.version) {
		ipt.collectFuncs["postgresql_io"] = ipt.getIOMetrics
	}

	ipt.relationMetrics = map[string]relationMetric{}
	for _, m := range relationMetrics {
		ipt.relationMetrics[m.measurementInfo.Name] = m
	}

	for _, metric := range ipt.MetricExcludeList {
		delete(ipt.collectFuncs, metric)
		delete(ipt.relationMetrics, metric)
	}

	if len(ipt.Relations) > 0 {
		ipt.collectFuncs["relation"] = ipt.getRelationMetrics
	}

	if ipt.Object.Enable {
		ipt.objectMetric = &objectMertric{}
		ipt.Object.name = fmt.Sprintf("%s:%d", ipt.host, ipt.port)
		ipt.collectFuncs["database"] = ipt.collectDatabaseObject
	}
}

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)

	ipt.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval.Duration)

	tick := time.NewTicker(ipt.Interval.Duration)

	defer tick.Stop()

	// try init
	for {
		if err := ipt.initDB(); err != nil {
			ipt.FeedCoByErr(err)
			l.Errorf("failed to init postgresql: %s", err.Error())
			ipt.feeder.FeedLastError(err.Error(),
				metrics.WithLastErrorInput(inputName),
				metrics.WithLastErrorCategory(point.Metric),
			)

			ipt.FeedUpMetric()
		} else {
			break
		}
		select {
		case <-datakit.Exit.Wait():
			ipt.exit()
			l.Infof(fmt.Sprintf("%s exit", inputName))
			return

		case <-ipt.semStop.Wait():
			ipt.exit()
			return

		case <-tick.C:
		}
	}

	ipt.initCfg()

	defer ipt.service.Stop() //nolint:errcheck

	// run custom queries
	ipt.runCustomQueries()
	ipt.ptsTime = ntp.Now()

	for {
		if ipt.pause {
			l.Debugf("not leader, skipped")
		} else {
			ipt.setUpState()
			start := time.Now()
			if err := ipt.Collect(); err != nil {
				ipt.feeder.FeedLastError(err.Error(),
					metrics.WithLastErrorInput(inputName),
				)
				l.Error(err)
				ipt.setErrUpState()
			}

			for category, points := range ipt.collectCache {
				if len(points) > 0 {
					feedName := inputName

					switch category { // nolint: exhaustive
					case point.CustomObject:
						feedName = customObjectFeedName
					case point.Logging:
						feedName = loggingFeedName
					case point.Object:
						feedName = objectFeedName
					}

					if err := ipt.feeder.Feed(category, points,
						dkio.WithCollectCost(time.Since(start)),
						dkio.WithElection(ipt.Election),
						dkio.WithSource(feedName),
					); err != nil {
						ipt.feeder.FeedLastError(err.Error(),
							metrics.WithLastErrorInput(inputName),
						)
						l.Errorf("feed : %s", err)
					}

					ipt.collectCache[category] = ipt.collectCache[category][:0]
				}
			}

			ipt.FeedUpMetric()

			ipt.FeedCoPts()
		}

		select {
		case <-datakit.Exit.Wait():
			ipt.exit()
			l.Info("postgresql exit")
			return

		case <-ipt.semStop.Wait():
			ipt.exit()
			l.Info("postgresql return")
			return

		case tt := <-tick.C:
			ipt.ptsTime = inputs.AlignTime(tt, ipt.ptsTime, ipt.Interval.Duration)

		case ipt.pause = <-ipt.pauseCh:
			// nil
		}
	}
}

func (ipt *Input) exit() {
	if ipt.tail != nil {
		ipt.tail.Close()
		l.Info("postgresql log exit")
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (ipt *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	defer tick.Stop()
	select {
	case ipt.pauseCh <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (ipt *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	defer tick.Stop()
	select {
	case ipt.pauseCh <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}

// getDBName parses out the DB name from an input URI.
// It returns the name of the DB or "postgres" if none is specified.
func getDBName(uri string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	if u.Path != "" {
		return u.Path[1:], nil
	}

	return "postgres", nil
}

func parseURL(uri string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	if u.Scheme != "postgres" && u.Scheme != "postgresql" {
		return "", fmt.Errorf("invalid connection protocol: %s", u.Scheme)
	}

	var kvs []string
	escaper := strings.NewReplacer(` `, `\ `, `'`, `\'`, `\`, `\\`)
	accrue := func(k, v string) {
		if v != "" {
			kvs = append(kvs, k+"="+escaper.Replace(v))
		}
	}

	if u.User != nil {
		v := u.User.Username()
		accrue("user", v)

		v, _ = u.User.Password()
		accrue("password", v)
	}

	if host, port, err := net.SplitHostPort(u.Host); err != nil {
		accrue("host", u.Host)
	} else {
		accrue("host", host)
		accrue("port", port)
	}

	if u.Path != "" {
		accrue("dbname", u.Path[1:])
	}

	q := u.Query()
	for k := range q {
		accrue(k, q.Get(k))
	}

	sort.Strings(kvs)
	return strings.Join(kvs, " "), nil
}

var maxPauseCh = inputs.ElectionPauseChannelLength

func NewInput(service Service) *Input {
	input := &Input{
		pauseCh:  make(chan bool, maxPauseCh),
		Election: true,
		feeder:   dkio.DefaultFeeder(),
		tagger:   datakit.DefaultGlobalTagger(),
		semStop:  cliutils.NewSem(),
		Object: pgObject{
			Enable:   true,
			Interval: datakit.Duration{Duration: 600 * time.Second},
			CollectSchemas: configCollectSchemas{
				Enabled:               true,
				MaxTables:             300,
				MaxDatabases:          100,
				AutoDiscoveryDatabase: false,
			},
		},
	}
	input.service = service
	return input
}

func defaultInput() *Input {
	return NewInput(&SQLService{})
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
