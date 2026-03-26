// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package oracle collect Oracle metrics
package oracle

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	go_version "github.com/hashicorp/go-version"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/jmoiron/sqlx"
	go_ora "github.com/sijms/go-ora/v2"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var _ inputs.ElectionInput = (*Input)(nil)

const (
	maxInterval          = 15 * time.Minute
	minInterval          = 10 * time.Second
	inputName            = "oracle"
	customObjectFeedName = inputName + "-CO"
	objectFeedName       = inputName + "-O"
	customQueryFeedName  = inputName + "-custom_query"
	loggingFeedName      = inputName + "-L"
	catalogName          = "db"
	measurementOracle    = "oracle"
)

var (
	dbmFeedName         = dkio.FeedSource(inputName, "DBM")
	dbmMetricInterval   = datakit.Duration{Duration: time.Second * 60}
	dbmActivityInterval = datakit.Duration{Duration: time.Second * 10}
	dbmPlanCacheTTL     = datakit.Duration{Duration: time.Hour} // Default plan cache TTL: 1 hour
)

var l = logger.DefaultSLogger(inputName)

type oracleObject struct {
	Enable   bool             `toml:"enabled"`
	Interval datakit.Duration `toml:"interval"`

	name               string
	lastCollectionTime time.Time
}

// tablespaceConfig holds tablespace collection config (runs in separate goroutine with own interval).
type tablespaceConfig struct {
	Enable   bool             `toml:"enabled"`
	Interval datakit.Duration `toml:"interval"`
}

// slowQueryConfig holds slow query collection config (runs in separate goroutine with own interval).
type slowQueryConfig struct {
	Enable   bool             `toml:"enabled"`
	Interval datakit.Duration `toml:"interval"`
}

// processConfig holds process collection config (runs in separate goroutine with own interval).
type processConfig struct {
	Enable   bool             `toml:"enabled"`
	Interval datakit.Duration `toml:"interval"`
}

// systemConfig holds system metrics collection config (runs in separate goroutine with own interval).
type systemConfig struct {
	Enable   bool             `toml:"enabled"`
	Interval datakit.Duration `toml:"interval"`
}

type Input struct {
	Host               string           `toml:"host"`
	Port               int              `toml:"port"`
	User               string           `toml:"user"`
	Password           string           `toml:"password"`
	Interval           datakit.Duration `toml:"interval"`
	Timeout            string           `toml:"connect_timeout"`
	Service            string           `toml:"service"`
	MetricExcludeList  []string         `toml:"metric_exclude_list"`
	timeoutDuration    time.Duration
	Query              []*customQuery    `toml:"custom_queries"`
	SlowQueryTime      string            `toml:"slow_query_time"`
	Election           bool              `toml:"election"`
	Tags               map[string]string `toml:"tags"`
	MeasurementVersion string            `toml:"measurement_version"` // v1 or v2, default: v2

	Object oracleObject `toml:"object"`

	Tablespace *tablespaceConfig `toml:"tablespace"`
	SlowQuery  *slowQueryConfig  `toml:"slow_query"`
	Process    *processConfig    `toml:"process"`
	System     *systemConfig     `toml:"system"`

	// DBM configuration
	Dbm *dbmConfig `toml:"dbm"`

	mainVersion, // simple version like 11
	fullVersion string // full version like 'Oracle Database 11g Express Edition Release 11.2.0.2.0 - 64bit Production'
	dbVersion        string // database version like '11.2.0.2.0' (extracted from version_full or fullVersion)
	cdbName          string // CDB name from v$database
	isMultitenant    bool
	databaseInstance string

	objectMetric *objectMertric

	Uptime             int
	CollectCoStatus    string
	CollectCoErrMsg    string
	LastCustomerObject *customerObjectMeasurement

	semStop         *cliutils.Sem // start stop signal
	feeder          dkio.Feeder
	tagger          datakit.GlobalTagger
	mergedTags      map[string]string
	db              *sqlx.DB
	goOraConnection *go_ora.Connection // Dedicated go-ora connection for handling LOBs

	pause atomic.Bool

	ptsTime        time.Time
	slowQueryTime  time.Duration
	lastActiveTime string
	cacheSQL       map[string]string

	collectorsGroup *goroutine.Group

	// DBM
	dbmGroup                                *goroutine.Group
	statementMetricsMonotonicCountsPrevious map[StatementMetricsKeyDB]StatementMetricsMonotonicCountDB
	dbmQueryObjectCache                     *expirable.LRU[string, struct{}]
	dbmPlanObjectCache                      *expirable.LRU[string, struct{}]
	sqlSubstringLength                      int

	UpState int
}

// dbmConfig represents the top-level DBM configuration.
type dbmConfig struct {
	Enabled  bool               `toml:"enabled"`
	Metric   *dbmMetricConfig   `toml:"metric"`
	Activity *dbmActivityConfig `toml:"activity"`
}

// dbmMetricConfig represents DBM metric (query metrics) configuration.
type dbmMetricConfig struct {
	Enabled            bool             `toml:"enabled"`
	CollectionInterval datakit.Duration `toml:"collection_interval"`
	DBRowsLimit        int              `toml:"db_rows_limit"`
	MaxQueries         int              `toml:"max_queries"`
	LookbackWindow     int              `toml:"lookback_window"`
	PlanEnabled        bool             `toml:"plan_enabled"`        // Enable plan collection
	PlanCacheTTL       datakit.Duration `toml:"plan_cache_ttl"`      // Plan object cache TTL
	MaxRunTime         int              `toml:"max_run_time"`        // Maximum runtime in seconds
	DisableLastActive  bool             `toml:"disable_last_active"` // Disable last active time filter
}

// dbmActivityConfig represents DBM activity (current active queries) configuration.
type dbmActivityConfig struct {
	Enabled            bool             `toml:"enabled"`
	CollectionInterval datakit.Duration `toml:"collection_interval"`
	DBRowsLimit        int              `toml:"db_rows_limit"`
	IncludeAllSessions bool             `toml:"include_all_sessions"` // Include all sessions, not just active ones
}

type vDatabase struct {
	Name string `db:"NAME"`
	Cdb  string `db:"CDB"`
}

func (ipt *Input) queryCDBInfo(ctx context.Context) error {
	if ipt.db == nil {
		return fmt.Errorf("database connection is nil")
	}

	var d vDatabase
	if isDBVersionGreaterOrEqualThan(ipt.dbVersion, "12") {
		err := ipt.db.GetContext(ctx, &d, "SELECT /* DK */ lower(name) name, cdb FROM v$database")
		if err != nil {
			return fmt.Errorf("failed to query v$database: %w", err)
		}
	} else {
		err := ipt.db.GetContext(ctx, &d, "SELECT /* DK */ lower(name) name FROM v$database")
		if err != nil {
			return fmt.Errorf("failed to query v$database: %w", err)
		}
		d.Cdb = "NO"
	}

	ipt.cdbName = d.Name
	ipt.isMultitenant = true
	if d.Cdb == "NO" {
		ipt.isMultitenant = false
	}
	return nil
}

func isDBVersionLessThan(dbVersion, v string) bool {
	vParsed, err := go_version.NewVersion(v)
	if err != nil {
		l.Errorf("Can't parse %s version string", v)
		return false
	}
	parsedDBVersion, err := go_version.NewVersion(dbVersion)
	if err != nil {
		l.Errorf("Can't parse db version string %s", dbVersion)
		return false
	}
	if parsedDBVersion.LessThan(vParsed) {
		return true
	}
	return false
}

func isDBVersionGreaterOrEqualThan(dbVersion, v string) bool {
	if dbVersion == "" {
		return false
	}
	vParsed, err := go_version.NewVersion(v)
	if err != nil {
		l.Errorf("Can't parse %s version string", v)
		return false
	}
	parsedDBVersion, err := go_version.NewVersion(dbVersion)
	if err != nil {
		l.Errorf("Can't parse db version string %s", dbVersion)
		return false
	}
	return parsedDBVersion.GreaterThanOrEqual(vParsed)
}

func (ipt *Input) setupDB() error {
	var err error
	ipt.timeoutDuration, err = time.ParseDuration(ipt.Timeout)
	if err != nil {
		ipt.timeoutDuration = 30 * time.Second
	}

	connStr := ipt.getConnString()
	db, err := sqlx.Open("oracle", connStr)
	if err != nil {
		l.Errorf("sql.Open(): %s", err.Error())
		return err
	} else {
		ipt.db = db
	}

	// TODO: These settings are hardcoded for now, but may be made configurable in the future.
	db.SetConnMaxLifetime(10 * time.Minute) // avoid max cursor problem
	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(10)

	ctx, cancel := context.WithTimeout(context.Background(), ipt.timeoutDuration)
	defer cancel()

	if err := ipt.db.PingContext(ctx); err != nil {
		l.Errorf("init config connect error %v", err)
		ipt.db.Close() //nolint:errcheck,gosec
		return err
	}

	ipt.getOracleVersion()

	ipt.getDatabaseInstance(ctx)

	// Query v$database to get CDB information
	if err := ipt.queryCDBInfo(ctx); err != nil {
		l.Warnf("failed to query v$database: %v", err)
		// Don't return error, continue with setup
	}

	return nil
}

type hostNameRow struct {
	HostName string `db:"HOST_NAME"`
}

func (ipt *Input) getDatabaseInstance(ctx context.Context) {
	var hn hostNameRow
	err := ipt.db.GetContext(ctx, &hn, "SELECT host_name FROM v$instance")
	if err != nil {
		l.Warnf("failed to get oracle host name: %s", err)
		return
	}
	ipt.databaseInstance = hn.HostName
}

func (ipt *Input) getConnString() string {
	opt := map[string]string{
		"timeout": fmt.Sprintf("%d", ipt.timeoutDuration/time.Second),
	}

	connStr := go_ora.BuildUrl(ipt.Host, ipt.Port, ipt.Service, ipt.User, ipt.Password, opt)

	return connStr
}

// buildGoOraURL builds a connection URL for go-ora driver.
func (ipt *Input) buildGoOraURL() string {
	opt := map[string]string{
		"timeout": fmt.Sprintf("%d", ipt.timeoutDuration/time.Second),
	}
	return go_ora.BuildUrl(ipt.Host, ipt.Port, ipt.Service, ipt.User, ipt.Password, opt)
}

// connectGoOra creates a dedicated go-ora connection for handling LOBs.
// This is needed because sqlx doesn't handle CLOB types well with go-ora driver.
func (ipt *Input) connectGoOra() (*go_ora.Connection, error) {
	conn, err := go_ora.NewConnection(ipt.buildGoOraURL(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect with the oracle driver: %w", err)
	}
	err = conn.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open connection with the oracle driver: %w", err)
	}
	return conn, nil
}

// closeGoOraConnection closes the dedicated go-ora connection.
func (ipt *Input) closeGoOraConnection() {
	if ipt.goOraConnection == nil {
		return
	}
	err := ipt.goOraConnection.Close()
	if err != nil {
		l.Warnf("failed to close go-ora connection: %s", err.Error())
	}
	ipt.goOraConnection = nil
}

// isConnectionError checks if the error is a connection error.
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "ORA-03113") || // end-of-file on communication channel
		strings.Contains(errStr, "ORA-03114") || // not connected to ORACLE
		strings.Contains(errStr, "ORA-12571") || // TNS:packet writer failure
		strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "closed")
}

func (ipt *Input) isMetricExclude(metric string) bool {
	for _, m := range ipt.MetricExcludeList {
		if metric == m {
			return true
		}
	}

	return false
}

func (ipt *Input) Collect() {
	ipt.setUpState()
	ipt.FeedCoPts()

	ipt.collectWaitingEvent()
	ipt.collectLockedSession()

	ipt.getOracleUptime()

	if ipt.Object.Enable {
		ipt.collectDatabaseObject()
	}
}

func (ipt *Input) Init() error {
	l = logger.SLogger(inputName)
	ipt.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval.Duration)
	tick := time.NewTicker(ipt.Interval.Duration)
	defer tick.Stop()

	// set object name
	ipt.Object.name = fmt.Sprintf("%s:%d", ipt.Host, ipt.Port)
	ipt.objectMetric = &objectMertric{}

	if ipt.Tags == nil {
		ipt.Tags = make(map[string]string)
	}
	if ipt.Election {
		ipt.mergedTags = inputs.MergeTags(ipt.tagger.ElectionTags(), ipt.Tags, ipt.Host)
	} else {
		ipt.mergedTags = inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, ipt.Host)
	}

	if _, ok := ipt.mergedTags["oracle_service"]; !ok {
		ipt.mergedTags["oracle_service"] = ipt.Service
	}

	if _, ok := ipt.mergedTags["oracle_server"]; !ok {
		ipt.mergedTags["oracle_server"] = ipt.Object.name
	}

	if _, ok := ipt.mergedTags["server"]; !ok {
		ipt.mergedTags["server"] = ipt.Object.name
	}

	// cache sql
	ipt.cacheSQL = make(map[string]string)
	// slow query
	if len(ipt.SlowQueryTime) > 0 {
		du, err := time.ParseDuration(ipt.SlowQueryTime)
		if err != nil {
			l.Warnf("bad slow query %s: %s, disable slow query", ipt.SlowQueryTime, err.Error())
		} else {
			if du >= time.Millisecond {
				ipt.slowQueryTime = du
			} else {
				l.Warnf("slow query time %v less than 1 millisecond, skip", du)
			}
		}
	}

	// Try until init OK.
	for {
		if err := ipt.setupDB(); err != nil {
			ipt.FeedCoByErr(err)
			l.Warnf("init config error: %s", err.Error())
			ipt.feeder.FeedLastError(err.Error(),
				metrics.WithLastErrorInput(inputName),
				metrics.WithLastErrorCategory(point.Metric),
			)
		} else {
			break
		}

		select {
		case <-datakit.Exit.Wait():
			return fmt.Errorf("datakit exit")

		case <-ipt.semStop.Wait():
			return fmt.Errorf("input exit")

		case <-tick.C:
		}

		// on init failing, we still upload up metric to show that the oracle input not working.
		ipt.FeedUpMetric()
	}

	if ipt.databaseInstance != "" {
		ipt.mergedTags["database_instance"] = ipt.databaseInstance
	}

	return nil
}

func (ipt *Input) Run() {
	tick := time.NewTicker(ipt.Interval.Duration)
	defer tick.Stop()
	defer func() {
		l.Info("oracle exit")
		ipt.exit()
	}()

	if err := ipt.Init(); err != nil {
		l.Errorf("init failed: %s", err.Error())
		return
	}

	l.Infof("collecting each %v", ipt.Interval.Duration)

	// run custom queries
	ipt.runCustomQueries()

	// run low-frequency collectors
	ipt.runLowFrequencyCollectors()

	// run DBM
	ipt.runDbmCollectors()

	ipt.ptsTime = ntp.Now()
	for {
		if ipt.pause.Load() {
			l.Info("not leader, skipped")
		} else {
			ipt.Collect()
		}

		select {
		case <-datakit.Exit.Wait():
			return

		case <-ipt.semStop.Wait():
			return

		case tt := <-tick.C:
			ipt.ptsTime = inputs.AlignTime(tt, ipt.ptsTime, ipt.Interval.Duration)
		}
	}
}

func (ipt *Input) Catalog() string { return catalogName }

func (ipt *Input) SampleConfig() string { return configSample }

func (ipt *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&oracleMeasurement{},
		&inputs.UpMeasurement{},
		// Logging measurements are not unified
		&slowQueryMeasurement{},
		&dbmActivityMeasurement{},
		// Object measurements are not unified
		&customerObjectMeasurement{},
		&oracleObjectMeasurement{},
		&dbmQueryObjectMeasurement{},
		&dbmPlanObjectMeasurement{},
	}
}

func (ipt *Input) AvailableArchs() []string {
	return datakit.AllOSWithElection
}

func (ipt *Input) Pause() error {
	ipt.pause.Store(true)
	return nil
}

func (ipt *Input) Resume() error {
	ipt.pause.Store(false)
	return nil
}

func (ipt *Input) ElectionEnabled() bool {
	return ipt.Election
}

func (ipt *Input) exit() {
	// Close go-ora connection if exists
	ipt.closeGoOraConnection()
	// Close sqlx database connection if exists
	if ipt.db != nil {
		if err := ipt.db.Close(); err != nil {
			l.Warnf("failed to close database connection: %s", err.Error())
		}
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (ipt *Input) runLowFrequencyCollectors() {
	ipt.collectorsGroup = goroutine.NewGroup(goroutine.Option{Name: "oracle_collectors"})

	if ipt.Tablespace != nil && ipt.Tablespace.Enable {
		ipt.collectorsGroup.Go(func(ctx context.Context) error {
			ipt.runTablespaceCollector()
			return nil
		})
	}
	if ipt.SlowQuery != nil && ipt.SlowQuery.Enable {
		ipt.collectorsGroup.Go(func(ctx context.Context) error {
			ipt.runSlowQueryCollector()
			return nil
		})
	}
	if ipt.Process != nil && ipt.Process.Enable {
		ipt.collectorsGroup.Go(func(ctx context.Context) error {
			ipt.runProcessCollector()
			return nil
		})
	}
	if ipt.System != nil && ipt.System.Enable {
		ipt.collectorsGroup.Go(func(ctx context.Context) error {
			ipt.runSystemCollector()
			return nil
		})
	}
}

// runTablespaceCollector runs tablespace collection in its own goroutine with dedicated interval.
func (ipt *Input) runTablespaceCollector() {
	duration := ipt.Tablespace.Interval.Duration
	if duration <= 0 {
		duration = 600 * time.Second
	}

	tick := time.NewTicker(duration)
	defer tick.Stop()

	ptsTime := ntp.Now()
	for {
		if ipt.pause.Load() {
			l.Debugf("not leader, tablespace collection skipped")
		} else {
			ipt.collectOracleTableSpace(ptsTime)
		}

		select {
		case <-datakit.Exit.Wait():
			l.Info("tablespace collection exit")
			return
		case <-ipt.semStop.Wait():
			l.Info("tablespace collection return")
			return
		case tt := <-tick.C:
			ptsTime = inputs.AlignTime(tt, ptsTime, duration)
		}
	}
}

// runProcessCollector runs process collection in its own goroutine with dedicated interval (default 60s).
func (ipt *Input) runProcessCollector() {
	duration := ipt.Process.Interval.Duration
	if duration <= 0 {
		duration = 60 * time.Second
	}

	tick := time.NewTicker(duration)
	defer tick.Stop()

	ptsTime := ntp.Now()
	for {
		if ipt.pause.Load() {
			l.Debugf("not leader, process collection skipped")
		} else {
			ipt.collectOracleProcess(ptsTime)
		}

		select {
		case <-datakit.Exit.Wait():
			l.Info("process collection exit")
			return
		case <-ipt.semStop.Wait():
			l.Info("process collection return")
			return
		case tt := <-tick.C:
			ptsTime = inputs.AlignTime(tt, ptsTime, duration)
		}
	}
}

// runSystemCollector runs system metrics collection in its own goroutine with dedicated interval (default 60s).
func (ipt *Input) runSystemCollector() {
	duration := ipt.System.Interval.Duration
	if duration <= 0 {
		duration = 60 * time.Second
	}

	tick := time.NewTicker(duration)
	defer tick.Stop()

	ptsTime := ntp.Now()
	for {
		if ipt.pause.Load() {
			l.Debugf("not leader, system collection skipped")
		} else {
			ipt.collectOracleSystem(ptsTime)
		}

		select {
		case <-datakit.Exit.Wait():
			l.Info("system collection exit")
			return
		case <-ipt.semStop.Wait():
			l.Info("system collection return")
			return
		case tt := <-tick.C:
			ptsTime = inputs.AlignTime(tt, ptsTime, duration)
		}
	}
}

// runSlowQueryCollector runs slow query collection in its own goroutine with dedicated interval.
func (ipt *Input) runSlowQueryCollector() {
	duration := ipt.SlowQuery.Interval.Duration
	if duration <= 0 {
		duration = 60 * time.Second
	}

	tick := time.NewTicker(duration)
	defer tick.Stop()

	ptsTime := ntp.Now()
	for {
		if ipt.pause.Load() {
			l.Debugf("not leader, slow query collection skipped")
		} else {
			ipt.collectSlowQuery(ptsTime)
		}

		select {
		case <-datakit.Exit.Wait():
			l.Info("slow query collection exit")
			return
		case <-ipt.semStop.Wait():
			l.Info("slow query collection return")
			return
		case tt := <-tick.C:
			ptsTime = inputs.AlignTime(tt, ptsTime, duration)
		}
	}
}

func (ipt *Input) runDbmCollectors() {
	// Check if DBM is enabled at the top level
	if ipt.Dbm == nil || !ipt.Dbm.Enabled {
		return
	}

	ipt.dbmGroup = goroutine.NewGroup(goroutine.Option{Name: "oracle_dbm"})

	// Start DBM metric collection (query metrics) with its own interval
	if ipt.Dbm.Metric != nil && ipt.Dbm.Metric.Enabled {
		ipt.dbmGroup.Go(func(ctx context.Context) error {
			ipt.runDbmMetric()
			return nil
		})
	}

	// Start DBM activity collection with its own interval
	if ipt.Dbm.Activity != nil && ipt.Dbm.Activity.Enabled {
		ipt.dbmGroup.Go(func(ctx context.Context) error {
			ipt.runDbmActivity()
			return nil
		})
	}
}

// runDbmMetric runs DBM metric.
func (ipt *Input) runDbmMetric() {
	duration := ipt.Dbm.Metric.CollectionInterval.Duration

	tick := time.NewTicker(duration)
	defer tick.Stop()

	ptsTime := ntp.Now()
	for {
		if ipt.pause.Load() {
			l.Debugf("not leader, DBM metric collection skipped")
		} else {
			l.Debugf("start collecting DBM metric")
			ipt.collectDbmMetricAndPlans(duration, ptsTime)
		}

		select {
		case <-datakit.Exit.Wait():
			l.Info("DBM metric collection exit")
			return
		case <-ipt.semStop.Wait():
			l.Info("DBM metric collection return")
			return
		case tt := <-tick.C:
			ptsTime = inputs.AlignTime(tt, ptsTime, duration)
		}
	}
}

// collectDbmMetricAndPlans collects DBM metric and plan objects.
func (ipt *Input) collectDbmMetricAndPlans(duration time.Duration, ptsTime time.Time) {
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	oracleRows, err := ipt.collectDbmMetric(ctx, ptsTime)
	if err != nil {
		l.Errorf("collectDbmMetric failed: %s", err.Error())
		return
	}

	if len(oracleRows) > 0 {
		// Collect query objects
		ipt.collectDbmQueries(oracleRows, ptsTime)

		// Collect plan objects (only if enabled)
		// max_run_time check is done inside collectDbmPlans loop
		if ipt.Dbm.Metric.PlanEnabled {
			ipt.collectDbmPlans(ctx, oracleRows, ptsTime)
		}
	}
}

// runDbmActivity runs DBM activity.
func (ipt *Input) runDbmActivity() {
	duration := ipt.Dbm.Activity.CollectionInterval.Duration

	tick := time.NewTicker(duration)
	defer tick.Stop()

	ptsTime := ntp.Now()
	for {
		if ipt.pause.Load() {
			l.Debugf("not leader, DBM activity collection skipped")
		} else {
			start := time.Now()
			l.Debugf("start collecting DBM activity")
			pts, err := ipt.collectDbmActivity(duration, ptsTime)
			if err != nil {
				l.Errorf("collectDbmActivity failed: %s", err.Error())
			} else if len(pts) > 0 {
				if err := ipt.feeder.Feed(point.Logging, pts,
					dkio.WithCollectCost(time.Since(start)),
					dkio.WithElection(ipt.Election),
					dkio.WithSource(dbmFeedName),
				); err != nil {
					ipt.feeder.FeedLastError(err.Error(),
						metrics.WithLastErrorInput(inputName),
						metrics.WithLastErrorCategory(point.Logging),
					)
					l.Errorf("feed dbm activity failed: %s", err.Error())
				}
			}
		}

		select {
		case <-datakit.Exit.Wait():
			l.Info("DBM activity collection exit")
			return
		case <-ipt.semStop.Wait():
			l.Info("DBM activity collection return")
			return
		case tt := <-tick.C:
			ptsTime = inputs.AlignTime(tt, ptsTime, duration)
		}
	}
}

func selectWrapper[T any](ipt *Input, s T, sql string, names ...string) error {
	now := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), ipt.timeoutDuration)
	defer cancel()

	var name string
	if len(names) == 1 {
		name = names[0]
	}

	err := ipt.db.SelectContext(ctx, s, sql)
	if err != nil && (strings.Contains(err.Error(), "ORA-01012") || strings.Contains(err.Error(), "database is closed")) {
		if err := ipt.setupDB(); err != nil {
			_ = ipt.db.Close()
		}
	}

	if err != nil {
		l.Errorf("executed sql: %s, cost: %v, err: %v\n", sql, time.Since(now), err)
	} else {
		metricName, sqlName := getMetricNames(name)
		if len(sqlName) > 0 {
			sqlQueryCostSummary.WithLabelValues(metricName, sqlName).Observe(float64(time.Since(now)) / float64(time.Second))
		}
	}

	return err
}

func selectWrapperWithBinds[T any](ipt *Input, ctx context.Context, s T, sql string, binds ...interface{}) error {
	now := time.Now()

	err := ipt.db.SelectContext(ctx, s, sql, binds...)
	if err != nil && (strings.Contains(err.Error(), "ORA-01012") || strings.Contains(err.Error(), "database is closed")) {
		if err := ipt.setupDB(); err != nil {
			_ = ipt.db.Close()
		}
		// Retry after reconnection
		if ipt.db != nil {
			err = ipt.db.SelectContext(ctx, s, sql, binds...)
		}
	}

	if err != nil {
		l.Errorf("executed sql: %s, cost: %v, err: %v\n", sql, time.Since(now), err)
	}

	return err
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

	opts = append(opts, point.WithTime(ipt.ptsTime))

	return opts
}

// getKVsOptsWithTime returns point options with the given ptsTime.
func (ipt *Input) getKVsOptsWithTime(ptsTime time.Time, categorys ...point.Category) []point.Option {
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

	opts = append(opts, point.WithTime(ptsTime))

	return opts
}

func defaultInput() *Input {
	return &Input{
		Tags:               make(map[string]string),
		Timeout:            "10s",
		pause:              atomic.Bool{},
		Election:           true,
		MeasurementVersion: "v2", // default: v2
		Tablespace: &tablespaceConfig{
			Enable:   true,
			Interval: datakit.Duration{Duration: 600 * time.Second},
		},
		SlowQuery: &slowQueryConfig{
			Enable:   true,
			Interval: datakit.Duration{Duration: 60 * time.Second},
		},
		Process: &processConfig{
			Enable:   true,
			Interval: datakit.Duration{Duration: 60 * time.Second},
		},
		System: &systemConfig{
			Enable:   true,
			Interval: datakit.Duration{Duration: 60 * time.Second},
		},
		Object: oracleObject{
			Enable:   true,
			Interval: datakit.Duration{Duration: 600 * time.Second},
		},
		feeder:  dkio.DefaultFeeder(),
		tagger:  datakit.DefaultGlobalTagger(),
		semStop: cliutils.NewSem(),
		Dbm: &dbmConfig{
			Enabled: false,
			Metric: &dbmMetricConfig{
				Enabled:            false,
				CollectionInterval: dbmMetricInterval,
				DBRowsLimit:        10000,
				MaxQueries:         500,
				LookbackWindow:     300,
				PlanEnabled:        true,
				PlanCacheTTL:       dbmPlanCacheTTL,
				MaxRunTime:         30,
				DisableLastActive:  false,
			},
			Activity: &dbmActivityConfig{
				Enabled:            false,
				CollectionInterval: dbmActivityInterval,
				DBRowsLimit:        1000,
				IncludeAllSessions: false,
			},
		},
		statementMetricsMonotonicCountsPrevious: make(map[StatementMetricsKeyDB]StatementMetricsMonotonicCountDB),
		sqlSubstringLength:                      4000,
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
