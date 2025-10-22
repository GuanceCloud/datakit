// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package mysql collect MySQL metrics
package mysql

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/go-sql-driver/mysql"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
)

var _ inputs.ElectionInput = (*Input)(nil)

const (
	maxInterval = 15 * time.Minute
	minInterval = 10 * time.Second
	strMariaDB  = "MariaDB"

	metricNameMySQL               = "mysql"
	metricNameMySQLReplication    = "mysql_replication"
	metricNameMySQLSchema         = "mysql_schema"
	metricNameMySQLTableSchema    = "mysql_table_schema"
	metricNameMySQLUserStatus     = "mysql_user_status"
	metricNameMySQLInnodb         = "mysql_innodb"
	metricNameMySQLDbmMetric      = "mysql_dbm_metric"
	metricNameMySQLDbmSample      = "mysql_dbm_sample"
	metricNameMySQLDbmActivity    = "mysql_dbm_activity"
	metricNameMySQLReplicationLog = "mysql_replication_log"
)

var (
	inputName            = "mysql"
	customObjectFeedName = dkio.FeedSource(inputName, "CO")
	objectFeedName       = dkio.FeedSource(inputName, "O")
	customQueryFeedName  = dkio.FeedSource(inputName, "custom_query")
	loggingFeedName      = dkio.FeedSource(inputName, "L")
	catalogName          = "db"
	l                    = logger.DefaultSLogger("mysql")
)

type TLS struct {
	TLSKey             string `toml:"tls_key"`
	TLSCert            string `toml:"tls_cert"`
	TLSCA              string `toml:"tls_ca"`
	InsecureSkipVerify bool   `toml:"insecure_skip_verify"`
	AllowTLS10         bool   `toml:"allow_tls10,omitempty"`
}

type customQuery struct {
	SQL      string           `toml:"sql"`
	Metric   string           `toml:"metric"`
	Interval datakit.Duration `toml:"interval"`
	Tags     []string         `toml:"tags"`
	Fields   []string         `toml:"fields"`
	ptsTime  time.Time
}

type mysqllog struct {
	Files             []string `toml:"files"`
	Pipeline          string   `toml:"pipeline"`
	IgnoreStatus      []string `toml:"ignore"`
	CharacterEncoding string   `toml:"character_encoding"`
	MultilineMatch    string   `toml:"multiline_match"`

	MatchDeprecated string `toml:"match,omitempty"`
}

type Input struct {
	Host              string      `toml:"host"`
	Port              int         `toml:"port"`
	User              string      `toml:"user"`
	Pass              string      `toml:"pass"`
	Sock              string      `toml:"sock"`
	Tables            []string    `toml:"tables"`
	Users             []string    `toml:"users"`
	MetricExcludeList []string    `toml:"metric_exclude_list"`
	Dbm               bool        `toml:"dbm"`
	DbmMetric         dbmMetric   `toml:"dbm_metric"`
	DbmSample         dbmSample   `toml:"dbm_sample"`
	DbmActivity       dbmActivity `toml:"dbm_activity"`
	Object            mysqlObject `toml:"object"`

	Replica      bool `toml:"replication"`
	GroupReplica bool `toml:"group_replication"`

	Charset string `toml:"charset"`

	Timeout         string `toml:"connect_timeout"`
	timeoutDuration time.Duration

	Service  string           `toml:"service"`
	Interval datakit.Duration `toml:"interval"`

	TLS  *TLS              `toml:"tls"`
	Tags map[string]string `toml:"tags"`

	Query  []*customQuery `toml:"custom_queries"`
	Addr   string         `toml:"-"`
	InnoDB bool           `toml:"innodb"`
	Log    *mysqllog      `toml:"log"`

	MatchDeprecated string `toml:"match,omitempty"`

	UpState int

	Version            *mysqlVersion
	Uptime             int
	CollectCoStatus    string
	CollectCoErrMsg    string
	LastCustomerObject *customerObjectMeasurement

	ptsTime time.Time
	db      *sql.DB
	feeder  dkio.Feeder
	tagger  datakit.GlobalTagger

	// response   []map[string]interface{}
	tail       *tailer.Tailer
	collectors map[string]func() ([]*point.Point, error)

	Election bool `toml:"election"`
	pause    bool
	pauseCh  chan bool
	binLogOn bool

	semStop *cliutils.Sem // start stop signal

	dbmCache       map[string]dbmRow
	dbmSampleCache dbmSampleCache

	objectMetric *objectMertric

	// collected metrics - mysql
	globalStatus    map[string]interface{}
	globalVariables map[string]interface{}
	binlog          map[string]interface{}

	// collected metrics - mysql_replication
	mReplication      map[string]interface{}
	mGroupReplication map[string]interface{}

	// collected metrics - mysql_schema
	mSchemaSize          map[string]interface{}
	mSchemaQueryExecTime map[string]interface{}

	// collected metrics - mysql_innodb
	mInnodb map[string]interface{}

	// collected metrics - mysql_table_schema
	mTableSchema []map[string]interface{}

	// collected metrics - mysql_user_status
	mUserStatusName       map[string]interface{}
	mUserStatusVariable   map[string]map[string]interface{}
	mUserStatusConnection map[string]map[string]interface{}

	// collected metrics - mysql_dbm_metric
	dbmMetricRows []dbmRow

	// collected metrics - mysql_dbm_sample
	dbmSamplePlans []planObj

	lastErrors []string
	mergedTags map[string]string
}

type mysqlObject struct {
	Enable   bool             `toml:"enabled"`
	Interval datakit.Duration `toml:"interval"`

	name               string
	lastCollectionTime time.Time
}

func (ipt *Input) ElectionEnabled() bool {
	return ipt.Election
}

func (ipt *Input) getDsnString() (string, error) {
	cfg := mysql.Config{
		AllowNativePasswords: true,
		CheckConnLiveness:    true,
		User:                 ipt.User,
		Passwd:               ipt.Pass,
		Params:               map[string]string{},
	}

	if ipt.Port == 0 {
		ipt.Port = 3306
	}

	// set addr
	if ipt.Sock != "" {
		cfg.Net = "unix"
		cfg.Addr = ipt.Sock
	} else {
		addr := fmt.Sprintf("%s:%d", ipt.Host, ipt.Port)
		cfg.Net = "tcp"
		cfg.Addr = addr
	}
	ipt.Addr = cfg.Addr

	// set timeout
	if ipt.timeoutDuration != 0 {
		cfg.Timeout = ipt.timeoutDuration
	}

	// set Charset
	if ipt.Charset != "" {
		cfg.Params["charset"] = ipt.Charset
	}

	if ipt.TLS != nil {
		tlsConfig, err := createTLSConf(ipt.TLS.TLSCA, ipt.TLS.TLSCert, ipt.TLS.TLSKey)
		if err != nil {
			return "", err
		}

		tlsConfig.InsecureSkipVerify = ipt.TLS.InsecureSkipVerify
		if ipt.TLS.AllowTLS10 {
			tlsConfig.MinVersion = tls.VersionTLS10
		}

		if err := mysql.RegisterTLSConfig("custom", tlsConfig); err != nil {
			return "", fmt.Errorf("register tls config failed: %w", err)
		} else {
			cfg.Params["tls"] = "custom"
		}
	}
	return cfg.FormatDSN(), nil
}

func createTLSConf(caFile, certFile, keyFile string) (*tls.Config, error) {
	if caFile == "" {
		return &tls.Config{
			MinVersion: tls.VersionTLS12,
		}, nil
	}

	var certs []tls.Certificate
	if certFile != "" && keyFile != "" {
		// Load client cert
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, err
		} else {
			certs = append(certs, cert)
		}
	}
	// Load CA cert
	caCert, err := os.ReadFile(filepath.Clean(caFile))
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, errors.New("failed to append certs from PEM")
	}

	tlsConfig := &tls.Config{ //nolint:gosec
		Certificates: certs,
		RootCAs:      caCertPool,
	}

	return tlsConfig, nil
}

//nolint:lll
func (*Input) LogExamples() map[string]map[string]string {
	return map[string]map[string]string{
		inputName: {
			"MySQL log": `2017-12-29T12:33:33.095243Z         2 Query     SELECT TABLE_SCHEMA, TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE CREATE_OPTIONS LIKE '%partitioned%';`,
			"MySQL slow log": `# Time: 2019-11-27T10:43:13.460744Z
# User@Host: root[root] @ localhost [1.2.3.4]  Id:    35
# Query_time: 0.214922  Lock_time: 0.000184 Rows_sent: 248832  Rows_examined: 72
# Thread_id: 55   Killed: 0  Errno: 0
# Bytes_sent: 123456   Bytes_received: 0
SET timestamp=1574851393;
SELECT * FROM fruit f1, fruit f2, fruit f3, fruit f4, fruit f5`,
		},
	}
}

func (*Input) PipelineConfig() map[string]string {
	pipelineMap := map[string]string{
		inputName: pipelineCfg,
	}
	return pipelineMap
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

func (ipt *Input) initCfg() error {
	var err error
	ipt.timeoutDuration, err = time.ParseDuration(ipt.Timeout)
	if err != nil {
		ipt.timeoutDuration = 10 * time.Second
	}

	dsnStr, err := ipt.getDsnString()
	if err != nil {
		return err
	}

	db, err := sql.Open("mysql", dsnStr)
	if err != nil {
		l.Errorf("sql.Open(): %s ,and db data source=%s", err.Error(), dsnStr)
		return err
	} else {
		ipt.db = db
	}

	// TODO: These settings are hardcoded for now, but may be made configurable in the future.
	db.SetConnMaxLifetime(10 * time.Minute)
	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(10)

	ctx, cancel := context.WithTimeout(context.Background(), ipt.timeoutDuration)
	defer cancel()

	if err := ipt.db.PingContext(ctx); err != nil {
		l.Errorf("init config connect error %v db data source=%s", err, dsnStr)
		ipt.db.Close() //nolint:errcheck,gosec
		return err
	}

	ipt.globalTag()
	if ipt.Dbm {
		ipt.initDbm()
	}

	const sqlSelect = "SELECT VERSION();"
	version := getCleanMysqlVersion(ipt.q(sqlSelect, getMetricName(metricNameMySQL, "select_version")))
	if version == nil || version.version == "" {
		return fmt.Errorf("failed to get mysql empty version")
	} else {
		ipt.Version = version
	}

	ipt.Object.name = fmt.Sprintf("%s:%d", ipt.Host, ipt.Port)
	if ipt.Object.Enable {
		ipt.objectMetric = &objectMertric{}
	}

	if _, ok := ipt.mergedTags["server"]; !ok {
		ipt.mergedTags["server"] = ipt.Object.name
	}

	return nil
}

func (ipt *Input) initDbm() {
	ipt.dbmSampleCache.explainCache.Size = 1000 // max size
	ipt.dbmSampleCache.explainCache.TTL = 60    // 60 second to live
}

func (ipt *Input) globalTag() {
	if len(ipt.Service) > 0 {
		ipt.Tags["service_name"] = ipt.Service
	}

	hostTag := getHostTag(ipt.Host)
	if len(hostTag) > 0 {
		ipt.Tags["host"] = hostTag
	}
}

func (ipt *Input) q(s string, names ...string) rows {
	var name string
	if len(names) == 1 {
		name = names[0]
	}

	start := time.Now()
	rows, err := ipt.db.Query(s)
	if err != nil {
		l.Errorf(`query failed, sql (%q), error: %s, ignored`, s, err.Error())
		return nil
	}

	if err := rows.Err(); err != nil {
		closeRows(rows)
		l.Errorf(`query row failed, sql (%q), error: %s, ignored`, s, err.Error())
		return nil
	}

	metricName, sqlName := getMetricNames(name)

	if len(metricName) > 0 {
		sqlQueryCostSummary.WithLabelValues(metricName, sqlName).Observe(float64(time.Since(start)) / float64(time.Second))
	}

	return rows
}

// mysql.
func (ipt *Input) metricCollectMysql() ([]*point.Point, error) {
	if err := ipt.collectMysql(); err != nil {
		return []*point.Point{}, err
	}

	pts, err := ipt.buildMysql()
	if err != nil {
		return []*point.Point{}, err
	}
	return pts, nil
}

// mysql_replication.
func (ipt *Input) metricCollectMysqlReplication() ([]*point.Point, error) {
	if err := ipt.collectMysqlReplication(); err != nil {
		return []*point.Point{}, err
	}

	pts, err := ipt.buildMysqlReplication()
	if err != nil {
		return []*point.Point{}, err
	}
	return pts, nil
}

// mysql_schema.
func (ipt *Input) metricCollectMysqlSchema() ([]*point.Point, error) {
	if err := ipt.collectMysqlSchema(); err != nil {
		return []*point.Point{}, err
	}

	pts, err := ipt.buildMysqlSchema()
	if err != nil {
		return []*point.Point{}, err
	}
	return pts, nil
}

// mysql_table_schema.
func (ipt *Input) metricCollectMysqlTableSschema() ([]*point.Point, error) {
	if err := ipt.collectMysqlTableSchema(); err != nil {
		return []*point.Point{}, err
	}

	pts, err := ipt.buildMysqlTableSchema()
	if err != nil {
		return []*point.Point{}, err
	}
	return pts, nil
}

// mysql_user_status.
func (ipt *Input) metricCollectMysqlUserStatus() ([]*point.Point, error) {
	if err := ipt.collectMysqlUserStatus(); err != nil {
		return []*point.Point{}, err
	}

	pts, err := ipt.buildMysqlUserStatus()
	if err != nil {
		return []*point.Point{}, err
	}
	return pts, nil
}

// mysql_innodb.
func (ipt *Input) metricCollectMysqlInnodb() ([]*point.Point, error) {
	if err := ipt.collectMysqlInnodb(); err != nil {
		return []*point.Point{}, err
	}
	pts, err := ipt.buildMysqlInnodb()
	if err != nil {
		return []*point.Point{}, err
	}
	return pts, nil
}

// mysql_dbm_metric.
func (ipt *Input) metricCollectMysqlDbmMetric() ([]*point.Point, error) {
	if err := ipt.collectMysqlDbmMetric(); err != nil {
		return []*point.Point{}, err
	}

	pts, err := ipt.buildMysqlDbmMetric()
	if err != nil {
		return []*point.Point{}, err
	}
	return pts, nil
}

// mysql_dbm_sample.
func (ipt *Input) metricCollectMysqlDbmSample() ([]*point.Point, error) {
	if err := ipt.collectMysqlDbmSample(); err != nil {
		return []*point.Point{}, err
	}

	pts, err := ipt.buildMysqlDbmSample()
	if err != nil {
		return []*point.Point{}, err
	}
	return pts, nil
}

func (ipt *Input) metricCollectMysqlCustomerObject() ([]*point.Point, error) {
	ipt.setIptCOStatus()
	pts, err := ipt.buildMysqlCustomerObject()
	if err != nil {
		return []*point.Point{}, err
	}
	return pts, nil
}

func (ipt *Input) resetLastError() {
	ipt.lastErrors = []string{}
}

func (ipt *Input) handleLastError() {
	if len(ipt.lastErrors) > 0 {
		ipt.feeder.FeedLastError(strings.Join(ipt.lastErrors, "; "),
			metrics.WithLastErrorInput(inputName),
		)
	}
}

func (ipt *Input) Collect() (map[point.Category][]*point.Point, error) {
	var ptsMetric,
		ptsLoggingMetric,
		ptsLoggingSample,
		ptsCustomerObject []*point.Point

	mpts := make(map[point.Category][]*point.Point)

	// collect basic metrics
	for _, metricName := range []string{
		metricNameMySQL,
		metricNameMySQLReplication,
		metricNameMySQLSchema,
		metricNameMySQLTableSchema,
		metricNameMySQLUserStatus,
	} {
		f, ok := ipt.collectors[metricName]
		if !ok {
			l.Debugf("collector %s not found", metricName)
			continue
		}

		l.Debugf("collecting %s...", metricName, f)

		pts, err := f()
		if err != nil {
			l.Errorf("collect failed: %s", err.Error())
		}

		if len(pts) > 0 {
			ptsMetric = append(ptsMetric, pts...)
		}
	}

	if ipt.InnoDB {
		f, ok := ipt.collectors[metricNameMySQLInnodb]
		if ok {
			// mysql_innodb
			pts, err := f()
			if err != nil {
				l.Errorf("metricCollectMysqlInnodb failed: %s", err.Error())
			}

			if len(pts) > 0 {
				ptsMetric = append(ptsMetric, pts...)
			}
		}
	}

	if ipt.Replica {
		f, ok := ipt.collectors[metricNameMySQLReplicationLog]
		if ok {
			// mysql_replication_log
			pts, err := f()
			if err != nil {
				l.Errorf("metricCollectMysqlReplicationLog failed: %s", err.Error())
			}

			if len(pts) > 0 {
				ptsLoggingMetric = append(ptsLoggingMetric, pts...)
			}
		}
	}

	if ipt.Dbm && (ipt.DbmMetric.Enabled || ipt.DbmSample.Enabled || ipt.DbmActivity.Enabled) {
		g := goroutine.NewGroup(goroutine.Option{Name: goroutine.GetInputName("mysql")})
		if ipt.DbmMetric.Enabled {
			f, ok := ipt.collectors[metricNameMySQLDbmMetric]
			if ok {
				g.Go(func(ctx context.Context) error {
					// mysql_dbm_metric
					pts, err := f()
					if err != nil {
						l.Errorf("metricCollectMysqlDbmMetric failed: %s", err.Error())
					}

					if len(pts) > 0 {
						ptsLoggingMetric = append(ptsLoggingMetric, pts...)
					}
					return nil
				})
			}
		}

		if ipt.DbmSample.Enabled {
			f, ok := ipt.collectors[metricNameMySQLDbmSample]
			if ok {
				g.Go(func(ctx context.Context) error {
					// mysql_dbm_sample
					pts, err := f()
					if err != nil {
						l.Errorf("metricCollectMysqlDbmSample failed: %s", err.Error())
					}

					if len(pts) > 0 {
						ptsLoggingSample = append(ptsLoggingSample, pts...)
					}
					return nil
				})
			}
		}

		if ipt.DbmActivity.Enabled {
			f, ok := ipt.collectors[metricNameMySQLDbmActivity]
			if ok {
				g.Go(func(ctx context.Context) error {
					// mysql_dbm_activity
					if pts, err := f(); err != nil {
						l.Errorf("Collect mysql dbm activity failed: %s", err.Error())
					} else if len(pts) > 0 {
						ptsLoggingSample = append(ptsLoggingSample, pts...)
					}
					return nil
				})
			}
		}

		err := g.Wait()
		if err != nil {
			l.Errorf("mysql dbm collect error: %v", err)
			ipt.feeder.FeedLastError(err.Error(),
				metrics.WithLastErrorInput(inputName),
				metrics.WithLastErrorCategory(point.Metric),
			)
		}
	}

	if err := ipt.collectMysqlBasicInfo(); err != nil {
		l.Errorf("collectMysqlBasicInfo failed: %s", err.Error())
	}

	// deprecated, may be removed in the future
	pts, err := ipt.metricCollectMysqlCustomerObject()
	if err != nil {
		l.Errorf("metricCollectMysqlCustomerObject failed: %s", err.Error())
	}
	ptsCustomerObject = append(ptsCustomerObject, pts...)

	if ipt.Object.Enable {
		if pts, err := ipt.metricCollectMysqlObject(); err != nil {
			l.Errorf("metricCollectMysqlObject failed: %s", err.Error())
		} else {
			mpts[point.Object] = pts
		}
	}

	mpts[point.Metric] = ptsMetric

	ptsLoggingMetric = append(ptsLoggingMetric, ptsLoggingSample...) // two combine in one
	mpts[point.Logging] = ptsLoggingMetric

	mpts[point.CustomObject] = ptsCustomerObject
	return mpts, nil
}

func (ipt *Input) RunPipeline() {
	if ipt.Log == nil || len(ipt.Log.Files) == 0 {
		return
	}

	opts := []tailer.Option{
		tailer.WithSource("mysql"),
		tailer.WithService("mysql"),
		tailer.WithPipeline(ipt.Log.Pipeline),
		tailer.WithIgnoredStatuses(ipt.Log.IgnoreStatus),
		tailer.WithCharacterEncoding(ipt.Log.CharacterEncoding),
		tailer.EnableMultiline(true),
		tailer.WithMaxMultilineLength(int64(float64(config.Cfg.Dataway.MaxRawBodySize) * 0.8)),
		tailer.WithMultilinePatterns([]string{ipt.Log.MultilineMatch}),
		tailer.WithExtraTags(inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, "")),
		tailer.EnableDebugFields(config.Cfg.EnableDebugFields),
	}
	var err error
	ipt.tail, err = tailer.NewTailer(ipt.Log.Files, opts...)
	if err != nil {
		l.Error(err)
		ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(inputName),
			metrics.WithLastErrorCategory(point.Metric),
		)
		return
	}

	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_mysql"})
	g.Go(func(ctx context.Context) error {
		ipt.tail.Start()
		return nil
	})
}

func (ipt *Input) initCollectors() {
	l.Infof("init collectors, metric exclude list: %v", ipt.MetricExcludeList)

	ipt.collectors = map[string]func() ([]*point.Point, error){
		metricNameMySQL:               ipt.metricCollectMysql,
		metricNameMySQLReplication:    ipt.metricCollectMysqlReplication,
		metricNameMySQLSchema:         ipt.metricCollectMysqlSchema,
		metricNameMySQLTableSchema:    ipt.metricCollectMysqlTableSschema,
		metricNameMySQLUserStatus:     ipt.metricCollectMysqlUserStatus,
		metricNameMySQLInnodb:         ipt.metricCollectMysqlInnodb,
		metricNameMySQLDbmMetric:      ipt.metricCollectMysqlDbmMetric,
		metricNameMySQLDbmSample:      ipt.metricCollectMysqlDbmSample,
		metricNameMySQLDbmActivity:    ipt.metricCollectMysqlDbmActivity,
		metricNameMySQLReplicationLog: ipt.buildMysqlReplicationLog,
	}

	for _, metricName := range ipt.MetricExcludeList {
		delete(ipt.collectors, metricName)
	}
}

func (ipt *Input) runCustomQuery(query *customQuery) {
	if query == nil {
		return
	}

	// use input interval as default
	duration := ipt.Interval.Duration
	// use custom query interval if set
	if query.Interval.Duration > 0 {
		duration = config.ProtectedInterval(minInterval, maxInterval, query.Interval.Duration)
	}

	tick := time.NewTicker(duration)
	defer tick.Stop()

	query.ptsTime = ntp.Now()
	for {
		collectStart := time.Now()

		if ipt.pause {
			l.Debugf("not leader, custom query skipped")
		} else {
			l.Debugf("start collecting custom query, metric name: %s", query.Metric)

			arr := getCleanMysqlCustomQueries(ipt.q(query.SQL, query.Metric))
			if arr != nil {
				points := ipt.getCustomQueryPoints(query, arr)
				if len(points) > 0 {
					if err := ipt.feeder.Feed(point.Metric, points,
						dkio.WithCollectCost(time.Since(collectStart)),
						dkio.WithElection(ipt.Election),
						dkio.WithSource(customQueryFeedName),
					); err != nil {
						ipt.feeder.FeedLastError(err.Error(),
							metrics.WithLastErrorInput(customQueryFeedName),
							metrics.WithLastErrorCategory(point.Metric),
						)
						l.Errorf("feed custom query failed: %s", err.Error())
					}
				}
			}
		}

		select {
		case <-datakit.Exit.Wait():
			ipt.exit()
			l.Info("mysql custom query exit")
			return

		case <-ipt.semStop.Wait():
			ipt.exit()
			l.Info("mysql custom query return")
			return

		case tt := <-tick.C:
			query.ptsTime = inputs.AlignTime(tt, query.ptsTime, duration)
		}
	}
}

func (ipt *Input) runCustomQueries() {
	if len(ipt.Query) == 0 {
		return
	}

	l.Infof("start to run custom queries, total %d queries", len(ipt.Query))

	g := goroutine.NewGroup(goroutine.Option{
		Name:         "mysql_custom_query",
		PanicTimes:   6,
		PanicTimeout: 10 * time.Second,
	})
	for _, q := range ipt.Query {
		func(q *customQuery) {
			g.Go(func(ctx context.Context) error {
				ipt.runCustomQuery(q)
				return nil
			})
		}(q)
	}
}

func (ipt *Input) Run() {
	l = logger.SLogger("mysql")
	ipt.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval.Duration)

	tick := time.NewTicker(ipt.Interval.Duration)
	defer tick.Stop()

	// init collectors
	ipt.initCollectors()

	if ipt.Election {
		ipt.mergedTags = inputs.MergeTags(ipt.tagger.ElectionTags(), ipt.Tags, ipt.Host)
	} else {
		ipt.mergedTags = inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, ipt.Host)
	}

	// Try until init OK.
	for {
		if err := ipt.initCfg(); err != nil {
			ipt.setInptErrCOMsg(err.Error())
			ipt.setIptErrCOStatus()
			pts := ipt.getCoPointByColErr()
			if err := ipt.feeder.Feed(point.CustomObject,
				pts,
				dkio.WithElection(ipt.Election),
				dkio.WithSource(customObjectFeedName),
			); err != nil {
				ipt.feeder.FeedLastError(err.Error(),
					metrics.WithLastErrorInput(inputName),
					metrics.WithLastErrorCategory(point.CustomObject),
				)
				l.Errorf("feed : %s", err)
			}

			l.Warnf("init config error: %s", err.Error())
			ipt.feeder.FeedLastError(err.Error(),
				metrics.WithLastErrorInput(inputName),
				metrics.WithLastErrorCategory(point.Metric),
			)

			// On init failing, we still upload up metric to show that the mysql input not working.
			ipt.FeedUpMetric()
		} else {
			break
		}

		select {
		case <-datakit.Exit.Wait():
			ipt.exit()
			return

		case <-ipt.semStop.Wait():
			ipt.exit()
			return

		case <-tick.C:
		}
	}

	l.Infof("collecting each %v", ipt.Interval.Duration)
	ipt.ptsTime = ntp.Now()

	// run custom queries
	ipt.runCustomQueries()

	for {
		if ipt.pause {
			l.Debugf("not leader, skipped")
		} else {
			l.Debugf("mysql input gathering...")

			collectStart := time.Now()

			ipt.setUpState()
			ipt.resetLastError()

			mpts, err := ipt.Collect()
			if err != nil {
				ipt.setErrUpState()
				l.Warnf("i.Collect failed: %v", err)
				ipt.feeder.FeedLastError(err.Error(),
					metrics.WithLastErrorInput(inputName),
					metrics.WithLastErrorCategory(point.Metric),
				)
			}

			for category, pts := range mpts {
				if len(pts) > 0 {
					feedName := inputName

					switch category { // nolint: exhaustive
					case point.CustomObject:
						feedName = customObjectFeedName // use specific CO-suffix feed name.
					case point.Logging:
						feedName = loggingFeedName // use specific L-suffix feed name.
					case point.Object:
						feedName = objectFeedName
					}

					if err := ipt.feeder.Feed(category, pts,
						dkio.WithCollectCost(time.Since(collectStart)),
						dkio.WithElection(ipt.Election),
						dkio.WithSource(feedName),
					); err != nil {
						ipt.feeder.FeedLastError(err.Error(),
							metrics.WithLastErrorInput(inputName),
							metrics.WithLastErrorCategory(point.Metric),
						)
						l.Errorf("feed : %s", err)
					}
				}
			}

			ipt.handleLastError()

			ipt.FeedUpMetric()
		}

		select {
		case <-datakit.Exit.Wait():
			ipt.exit()
			l.Info("mysql exit")
			return

		case <-ipt.semStop.Wait():
			ipt.exit()
			l.Info("mysql return")
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
		l.Info("mysql log exit")
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (*Input) Catalog() string { return catalogName }

func (*Input) SampleConfig() string { return configSample }

func (*Input) AvailableArchs() []string { return datakit.AllOSWithElection }

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&baseMeasurement{},
		&replicationMeasurement{},
		&schemaMeasurement{},
		&innodbMeasurement{},
		&tbMeasurement{},
		&userMeasurement{},
		&dbmStateMeasurement{},
		&dbmSampleMeasurement{},
		&dbmActivityMeasurement{},
		&replicationLogMeasurement{},
		&customerObjectMeasurement{},
		&inputs.UpMeasurement{},
		&mysqlObjectMeasurement{},
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

func defaultInput() *Input {
	return &Input{
		Tags:     make(map[string]string),
		Timeout:  "10s",
		pauseCh:  make(chan bool, inputs.ElectionPauseChannelLength),
		Election: true,
		feeder:   dkio.DefaultFeeder(),
		tagger:   datakit.DefaultGlobalTagger(),
		semStop:  cliutils.NewSem(),
		UpState:  0,
		Object: mysqlObject{
			Enable:   true,
			Interval: datakit.Duration{Duration: 600 * time.Second},
		},
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
