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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
)

var (
	inputName                        = "postgresql"
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
)

const sampleConfig = `
[[inputs.postgresql]]
  ## Server address
  # URI format
  # postgres://[datakit[:PASSWORD]]@localhost[/dbname]?sslmode=[disable|verify-ca|verify-full]
  # or simple string
  # host=localhost user=pqgotest password=... sslmode=... dbname=app_production

  address = "postgres://datakit:PASSWORD@localhost?sslmode=disable"

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
  # Time unit: "ns", "us" (or "µs"), "ms", "s", "m", "h"
  #
  interval = "10s"

  ## Relations config
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

  ## Set true to enable election
  election = true

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
	Query(string) (Rows, error)
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
	query           string
	measurementInfo *inputs.MeasurementInfo
}

// customQuery represents configuration for executing a custom query.
type customQuery struct {
	SQL    string   `toml:"sql"`
	Metric string   `toml:"metric"`
	Tags   []string `toml:"tags"`
	Fields []string `toml:"fields"`
}

type Input struct {
	Address          string            `toml:"address"`
	Outputaddress    string            `toml:"outputaddress"`
	IgnoredDatabases []string          `toml:"ignored_databases"`
	Databases        []string          `toml:"databases"`
	Interval         string            `toml:"interval"`
	Tags             map[string]string `toml:"tags"`
	Relations        []Relation        `toml:"relations"`
	CustomQuery      []*customQuery    `toml:"custom_queries"`
	Log              *postgresqllog    `toml:"log"`

	MaxLifetimeDeprecated string `toml:"max_lifetime,omitempty"`

	service      Service
	tail         *tailer.Tailer
	duration     time.Duration
	collectCache []*point.Point
	host         string

	Election bool `toml:"election"`
	pause    bool
	pauseCh  chan bool

	feeder   io.Feeder
	version  *semver.Version
	isAurora bool
	semStop  *cliutils.Sem // start stop signal

	collectFuncs     map[string]func() error
	metricQueryCache map[string]*queryCacheItem
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
		&sizeMeasurement{},
		&statIOMeasurement{},
		&statMeasurement{},
		&slruMeasurement{},
		&bgwriterMeasurement{},
		&connectionMeasurement{},
		&conflictMeasurement{},
		&archiverMeasurement{},
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

func (ipt *Input) GetPipeline() []*tailer.Option {
	return []*tailer.Option{
		{
			Source:  inputName,
			Service: inputName,
			Pipeline: func() string {
				if ipt.Log != nil {
					return ipt.Log.Pipeline
				}
				return ""
			}(),
		},
	}
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

func (ipt *Input) executeQuery(cache *queryCacheItem) error {
	var (
		columns []string
		err     error
	)

	if cache == nil || cache.query == "" {
		return fmt.Errorf("query cache is empty")
	}

	measurementInfo := cache.measurementInfo
	if measurementInfo == nil {
		measurementInfo = inputMeasurement{}.Info()
	}

	rows, err := ipt.service.Query(cache.query)
	if err != nil {
		return err
	}
	defer rows.Close() //nolint:errcheck

	if columns, err = rows.Columns(); err != nil {
		return err
	}

	for rows.Next() {
		columnMap, err := ipt.service.GetColumnMap(rows, columns)
		if err != nil {
			return err
		}
		err = ipt.accRow(columnMap, measurementInfo)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ipt *Input) getCustomQueryMetrics() error {
	for _, customQuery := range ipt.CustomQuery {
		tags := map[string]interface{}{}
		fields := map[string]interface{}{}
		for _, tag := range customQuery.Tags {
			tags[tag] = tag
		}
		for _, field := range customQuery.Fields {
			fields[field] = field
		}

		queryItem := &queryCacheItem{
			query: customQuery.SQL,
			measurementInfo: &inputs.MeasurementInfo{
				Name:   customQuery.Metric,
				Tags:   tags,
				Fields: fields,
			},
		}

		if err := ipt.executeQuery(queryItem); err != nil {
			l.Warnf("collect custom query [%s] error: %s", customQuery.SQL, err.Error())
		}
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
			query:           query,
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
			query:           query,
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
			query:           query,
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

	for _, relationInfo := range relationMetrics {
		cacheName := fmt.Sprintf("%s_%s", RelationMetric, relationInfo.name)
		cache, ok := ipt.metricQueryCache[cacheName]

		if !ok {
			query := ipt.getRelationQuery(relationInfo.query, relationInfo.schemaField)
			if len(query) == 0 {
				l.Warnf("relation query is empty, ignore %s", relationInfo.name)
				continue
			}

			cache = &queryCacheItem{
				query:           query,
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
			query:           query,
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
			query:           query,
			measurementInfo: replicationMeasurement{}.Info(),
		}
		ipt.metricQueryCache[ReplicationMetric] = cache
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
			query:           query,
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
			query:           query,
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
		return err
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

func (ipt *Input) Collect() error {
	var err error

	defer ipt.service.Stop() //nolint:errcheck
	err = ipt.service.Start()
	if err != nil {
		return err
	}

	if err := ipt.setVersion(); err != nil {
		return err
	}

	g := goroutine.NewGroup(goroutine.Option{Name: goroutine.GetInputName(inputName)})

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

func (ipt *Input) accRow(columnMap map[string]*interface{}, measurementInfo *inputs.MeasurementInfo) error {
	tags := map[string]string{}
	if ipt.host != "" {
		tags["host"] = ipt.host
	}

	if ipt.Tags != nil {
		for k, v := range ipt.Tags {
			tags[k] = v
		}
	}

	fields := make(map[string]interface{})

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

		stringVal := ""
		if *val != nil {
			switch trueVal := (*val).(type) {
			case []uint8:
				stringVal = string(trueVal)
			case string:
				stringVal = trueVal
			default:
				fields[col] = trueVal
			}

			if len(stringVal) > 0 {
				if isMeasurementTag {
					tags[col] = stringVal
				} else if numVal, err := strconv.ParseInt(stringVal, 10, 64); err == nil {
					fields[col] = numVal
				}
			}
		}
	}

	if len(fields) > 0 {
		name := inputName
		if measurementInfo != nil {
			name = measurementInfo.Name
		}
		ms := &inputMeasurement{
			name:     name,
			fields:   fields,
			tags:     tags,
			ts:       time.Now(),
			election: ipt.Election,
		}
		ipt.collectCache = append(ipt.collectCache, ms.Point())
	}

	return nil
}

func (ipt *Input) RunPipeline() {
	if ipt.Log == nil || len(ipt.Log.Files) == 0 {
		return
	}

	opt := &tailer.Option{
		Source:            inputName,
		Service:           inputName,
		Pipeline:          ipt.Log.Pipeline,
		GlobalTags:        ipt.Tags,
		IgnoreStatus:      ipt.Log.IgnoreStatus,
		CharacterEncoding: ipt.Log.CharacterEncoding,
		MultilinePatterns: []string{ipt.Log.MultilineMatch},
		Done:              ipt.semStop.Wait(),
	}

	var err error
	ipt.tail, err = tailer.NewTailer(ipt.Log.Files, opt)
	if err != nil {
		l.Error(err)
		ipt.feeder.FeedLastError(err.Error(),
			io.WithLastErrorInput(inputName),
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
	maxInterval = 1 * time.Minute
	minInterval = 1 * time.Second
)

func (ipt *Input) init() error {
	ipt.service.SetAddress(ipt.Address)
	defer ipt.service.Stop() //nolint:errcheck
	err := ipt.service.Start()
	if err != nil {
		return err
	}

	if err := ipt.setVersion(); err != nil {
		return err
	}

	ipt.setAurora()

	// set default Tags
	tagAddress, err := ipt.SanitizedAddress()
	if err != nil {
		return err
	}
	dbName, err := getDBName(ipt.Address)
	if err != nil {
		l.Warnf("get db name error: %s, use postgres instead", err.Error())
		dbName = "postgres"
	}
	if ipt.Tags == nil {
		ipt.Tags = map[string]string{}
	}
	ipt.Tags["server"] = tagAddress
	ipt.Tags["db"] = dbName

	// init query cache
	ipt.metricQueryCache = map[string]*queryCacheItem{}

	// setup collectors
	ipt.collectFuncs = map[string]func() error{
		"db":          ipt.getDBMetrics,
		"replication": ipt.getReplicationMetrics,
		"bgwriter":    ipt.getBgwMetrics,
		"connection":  ipt.getConnectionMetrics,
		"customQuery": ipt.getCustomQueryMetrics,
	}

	if V130.LessThan(*ipt.version) || V130.Equal(*ipt.version) {
		ipt.collectFuncs["slru"] = ipt.getSlruMetrics
	}

	if len(ipt.Relations) > 0 {
		ipt.collectFuncs["relation"] = ipt.getRelationMetrics
	}

	if V92.LessThan(*ipt.version) || V92.Equal(*ipt.version) {
		ipt.collectFuncs["dynamic"] = ipt.getDynamicQueryMetrics
	}

	if V94.LessThan(*ipt.version) || V94.Equal(*ipt.version) {
		ipt.collectFuncs["archiver"] = ipt.getArchiverMetrics
	}

	return nil
}

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)

	duration, err := time.ParseDuration(ipt.Interval)
	if err != nil {
		l.Error("invalid interval, %s", err.Error())
		return
	} else if duration <= 0 {
		l.Error("invalid interval, cannot be less than zero")
		return
	}

	if err := ipt.setHostIfNotLoopback(); err != nil {
		l.Errorf("failed to set host: %v", err)
		return
	}

	ipt.duration = config.ProtectedInterval(minInterval, maxInterval, duration)

	tick := time.NewTicker(ipt.duration)

	defer tick.Stop()

	// try init
	for {
		if err := ipt.init(); err != nil {
			l.Errorf("failed to init postgresql: %s", err.Error())
			ipt.feeder.FeedLastError(err.Error(),
				io.WithLastErrorInput(inputName),
				io.WithLastErrorCategory(point.Metric),
			)
		} else {
			break
		}
		select {
		case <-datakit.Exit.Wait():
			if ipt.tail != nil {
				ipt.tail.Close() //nolint:errcheck
			}
			l.Infof(fmt.Sprintf("%s exit", inputName))

			return

		case <-ipt.semStop.Wait():
			return

		case <-tick.C:
		}
	}

	for {
		select {
		case <-datakit.Exit.Wait():
			ipt.exit()
			l.Info("postgresql exit")
			return

		case <-ipt.semStop.Wait():
			ipt.exit()
			l.Info("postgresql return")
			return

		case <-tick.C:
			if ipt.pause {
				l.Debugf("not leader, skipped")
				continue
			}

			start := time.Now()
			if err := ipt.Collect(); err != nil {
				ipt.feeder.FeedLastError(err.Error(),
					io.WithLastErrorInput(inputName),
				)
				l.Error(err)
			}

			if len(ipt.collectCache) > 0 {
				err := ipt.feeder.Feed(inputName, point.Metric, ipt.collectCache,
					&io.Option{CollectCost: time.Since(start)})
				if err != nil {
					ipt.feeder.FeedLastError(err.Error(),
						io.WithLastErrorInput(inputName),
					)
					l.Error(err.Error())
				}
				ipt.collectCache = ipt.collectCache[:0]
			}

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

func (ipt *Input) setHostIfNotLoopback() error {
	uu, err := url.Parse(ipt.Address)
	if err != nil {
		return err
	}
	var host string
	h, _, err := net.SplitHostPort(uu.Host)
	if err == nil {
		host = h
	} else {
		host = uu.Host
	}
	if host != "localhost" && !net.ParseIP(host).IsLoopback() {
		ipt.host = host
	}
	return nil
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
		Interval: "10s",
		pauseCh:  make(chan bool, maxPauseCh),
		Election: true,
		feeder:   io.DefaultFeeder(),
		semStop:  cliutils.NewSem(),
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
