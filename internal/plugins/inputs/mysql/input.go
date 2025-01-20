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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
)

var _ inputs.ElectionInput = (*Input)(nil)

const (
	maxInterval = 15 * time.Minute
	minInterval = 10 * time.Second
	strMariaDB  = "MariaDB"
)

var (
	inputName            = "mysql"
	customObjectFeedName = inputName + "/CO"
	loggingFeedName      = inputName + "/L"
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
	SQL    string   `toml:"sql"`
	Metric string   `toml:"metric"`
	Tags   []string `toml:"tags"`
	Fields []string `toml:"fields"`

	md5Hash string
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
	Host        string      `toml:"host"`
	Port        int         `toml:"port"`
	User        string      `toml:"user"`
	Pass        string      `toml:"pass"`
	Sock        string      `toml:"sock"`
	Tables      []string    `toml:"tables"`
	Users       []string    `toml:"users"`
	Dbm         bool        `toml:"dbm"`
	DbmMetric   dbmMetric   `toml:"dbm_metric"`
	DbmSample   dbmSample   `toml:"dbm_sample"`
	DbmActivity dbmActivity `toml:"dbm_activity"`

	Replica      bool `toml:"replication"`
	GroupReplica bool `toml:"group_replication"`

	Charset string `toml:"charset"`

	Timeout         string `toml:"connect_timeout"`
	timeoutDuration time.Duration

	Service  string `toml:"service"`
	Interval datakit.Duration

	TLS  *TLS              `toml:"tls"`
	Tags map[string]string `toml:"tags"`

	Query  []*customQuery `toml:"custom_queries"`
	Addr   string         `toml:"-"`
	InnoDB bool           `toml:"innodb"`
	Log    *mysqllog      `toml:"log"`

	MatchDeprecated string `toml:"match,omitempty"`

	UpState int

	Version            string
	Uptime             int
	CollectCoStatus    string
	CollectCoErrMsg    string
	LastCustomerObject *customerObjectMeasurement

	start  time.Time
	db     *sql.DB
	feeder dkio.Feeder
	tagger datakit.GlobalTagger

	// response   []map[string]interface{}
	tail       *tailer.Tailer
	collectors []func() ([]*point.Point, error)

	Election bool `toml:"election"`
	pause    bool
	pauseCh  chan bool
	binLogOn bool

	semStop *cliutils.Sem // start stop signal

	dbmCache       map[string]dbmRow
	dbmSampleCache dbmSampleCache

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

	// collected metrics - mysql custom queries
	mCustomQueries map[string][]map[string]interface{}

	lastErrors []string
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
	return nil
}

func (ipt *Input) initDbm() {
	ipt.dbmSampleCache.explainCache.Size = 1000 // max size
	ipt.dbmSampleCache.explainCache.TTL = 60    // 60 second to live
}

func (ipt *Input) globalTag() {
	ipt.Tags["server"] = ipt.Addr
	if len(ipt.Service) > 0 {
		ipt.Tags["service_name"] = ipt.Service
	}
}

func (ipt *Input) q(s string) rows {
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

	return rows
}

// init db connect.
func (ipt *Input) initDBConnect() error {
	isNeedConnect := false

	if ipt.db == nil {
		isNeedConnect = true
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer func() {
			cancel()
		}()

		if err := ipt.db.PingContext(ctx); err != nil {
			isNeedConnect = true
		}
	}

	if isNeedConnect {
		if err := ipt.initCfg(); err != nil {
			return err
		}
	}

	return nil
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

// mysql_custom_queries.
func (ipt *Input) metricCollectMysqlCustomQueries() ([]*point.Point, error) {
	if err := ipt.collectMysqlCustomQueries(); err != nil {
		return []*point.Point{}, err
	}

	pts, err := ipt.buildMysqlCustomQueries()
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
	if err := ipt.collectMysqlCustomerObject(); err != nil {
		ipt.setIptErrCOStatus()
		return []*point.Point{}, err
	}
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
	if err := ipt.initDBConnect(); err != nil {
		return map[point.Category][]*point.Point{}, err
	}

	if len(ipt.collectors) == 0 {
		ipt.collectors = []func() ([]*point.Point, error){
			ipt.metricCollectMysql,              // mysql
			ipt.metricCollectMysqlReplication,   // mysql_replication
			ipt.metricCollectMysqlSchema,        // mysql_schema
			ipt.metricCollectMysqlTableSschema,  // mysql_table_schema
			ipt.metricCollectMysqlUserStatus,    // mysql_user_status
			ipt.metricCollectMysqlCustomQueries, // mysql_custom_queries
		}
	}

	var ptsMetric,
		ptsLoggingMetric,
		ptsLoggingSample,
		ptsCustomerObject []*point.Point

	for idx, f := range ipt.collectors {
		l.Debugf("collecting %d(%v)...", idx, f)

		pts, err := f()
		if err != nil {
			l.Errorf("collect failed: %s", err.Error())
		}

		if len(pts) > 0 {
			ptsMetric = append(ptsMetric, pts...)
		}
	}

	if ipt.InnoDB {
		// mysql_innodb
		pts, err := ipt.metricCollectMysqlInnodb()
		if err != nil {
			l.Errorf("metricCollectMysqlInnodb failed: %s", err.Error())
		}

		if len(pts) > 0 {
			ptsMetric = append(ptsMetric, pts...)
		}
	}

	if ipt.Replica {
		// mysql_replication_log
		pts, err := ipt.buildMysqlReplicationLog()
		if err != nil {
			l.Errorf("metricCollectMysqlReplicationLog failed: %s", err.Error())
		}

		if len(pts) > 0 {
			ptsLoggingMetric = append(ptsLoggingMetric, pts...)
		}
	}

	if ipt.Dbm && (ipt.DbmMetric.Enabled || ipt.DbmSample.Enabled || ipt.DbmActivity.Enabled) {
		g := goroutine.NewGroup(goroutine.Option{Name: goroutine.GetInputName("mysql")})
		if ipt.DbmMetric.Enabled {
			g.Go(func(ctx context.Context) error {
				// mysql_dbm_metric
				pts, err := ipt.metricCollectMysqlDbmMetric()
				if err != nil {
					l.Errorf("metricCollectMysqlDbmMetric failed: %s", err.Error())
				}

				if len(pts) > 0 {
					ptsLoggingMetric = append(ptsLoggingMetric, pts...)
				}
				return nil
			})
		}

		if ipt.DbmSample.Enabled {
			g.Go(func(ctx context.Context) error {
				// mysql_dbm_sample
				pts, err := ipt.metricCollectMysqlDbmSample()
				if err != nil {
					l.Errorf("metricCollectMysqlDbmSample failed: %s", err.Error())
				}

				if len(pts) > 0 {
					ptsLoggingSample = append(ptsLoggingSample, pts...)
				}
				return nil
			})
		}

		if ipt.DbmActivity.Enabled {
			g.Go(func(ctx context.Context) error {
				// mysql_dbm_activity
				if pts, err := ipt.metricCollectMysqlDbmActivity(); err != nil {
					l.Errorf("Collect mysql dbm activity failed: %s", err.Error())
				} else if len(pts) > 0 {
					ptsLoggingSample = append(ptsLoggingSample, pts...)
				}
				return nil
			})
		}

		err := g.Wait()
		if err != nil {
			l.Errorf("mysql dmb collect error: %v", err)
			ipt.feeder.FeedLastError(err.Error(),
				metrics.WithLastErrorInput(inputName),
				metrics.WithLastErrorCategory(point.Metric),
			)
		}
	}

	pts, err := ipt.metricCollectMysqlCustomerObject()
	if err != nil {
		l.Errorf("metricCollectMysqlCustomerObject failed: %s", err.Error())
	}
	ptsCustomerObject = append(ptsCustomerObject, pts...)

	mpts := make(map[point.Category][]*point.Point)
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
		tailer.WithIgnoreStatus(ipt.Log.IgnoreStatus),
		tailer.WithCharacterEncoding(ipt.Log.CharacterEncoding),
		tailer.EnableMultiline(true),
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

func (ipt *Input) Run() {
	l = logger.SLogger("mysql")
	ipt.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval.Duration)

	tick := time.NewTicker(ipt.Interval.Duration)
	defer tick.Stop()

	// Try until init OK.
	for {
		if err := ipt.initCfg(); err != nil {
			ipt.setInptErrCOMsg(err.Error())
			ipt.setIptErrCOStatus()
			pts := ipt.getCoPointByColErr()
			if err := ipt.feeder.FeedV2(point.CustomObject,
				pts,
				dkio.WithElection(ipt.Election),
				dkio.WithInputName(customObjectFeedName),
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
	ipt.start = time.Now()

	for {
		if ipt.pause {
			l.Debugf("not leader, skipped")
		} else {
			l.Debugf("mysql input gathering...")

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
					}

					if err := ipt.feeder.FeedV2(category, pts,
						dkio.WithCollectCost(time.Since(ipt.start)),
						dkio.WithElection(ipt.Election),
						dkio.WithInputName(feedName),
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
			nextts := inputs.AlignTimeMillSec(tt, ipt.start.UnixMilli(), ipt.Interval.Duration.Milliseconds())
			ipt.start = time.UnixMilli(nextts)
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
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
