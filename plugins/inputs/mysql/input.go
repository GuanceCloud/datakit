// Package mysql collect MySQL metrics
package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var _ inputs.ElectionInput = (*Input)(nil)

const (
	maxInterval   = 15 * time.Minute
	minInterval   = 10 * time.Second
	dbmMetricName = "database_performance"
	strMariaDB    = "MariaDB"
)

var (
	inputName   = "mysql"
	catalogName = "db"
	l           = logger.DefaultSLogger("mysql")
)

type tls struct {
	TLSKey  string `toml:"tls_key"`
	TLSCert string `toml:"tls_cert"`
	TLSCA   string `toml:"tls_ca"`
}

type customQuery struct {
	sql    string   `toml:"sql"`
	metric string   `toml:"metric"`
	tags   []string `toml:"tags"`
	fields []string `toml:"fields"`

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
	Host      string    `toml:"host"`
	Port      int       `toml:"port"`
	User      string    `toml:"user"`
	Pass      string    `toml:"pass"`
	Sock      string    `toml:"sock"`
	Tables    []string  `toml:"tables"`
	Users     []string  `toml:"users"`
	Dbm       bool      `toml:"dbm"`
	DbmMetric dbmMetric `toml:"dbm_metric"`
	DbmSample dbmSample `toml:"dbm_sample"`

	Charset string `toml:"charset"`

	Timeout         string `toml:"connect_timeout"`
	timeoutDuration time.Duration

	TLS *tls `toml:"tls"`

	Service  string `toml:"service"`
	Interval datakit.Duration

	Tags map[string]string `toml:"tags"`

	Query  []*customQuery `toml:"custom_queries"`
	Addr   string         `toml:"-"`
	InnoDB bool           `toml:"innodb"`
	Log    *mysqllog      `toml:"log"`

	MatchDeprecated string `toml:"match,omitempty"`

	start time.Time
	db    *sql.DB
	// response   []map[string]interface{}
	tail *tailer.Tailer
	// collectors []func() ([]inputs.Measurement, error)
	collectors []func() ([]*io.Point, error)

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

func (i *Input) getDsnString() string {
	cfg := mysql.Config{
		AllowNativePasswords: true,
		CheckConnLiveness:    true,
		User:                 i.User,
		Passwd:               i.Pass,
	}

	// set addr
	if i.Sock != "" {
		cfg.Net = "unix"
		cfg.Addr = i.Sock
	} else {
		addr := fmt.Sprintf("%s:%d", i.Host, i.Port)
		cfg.Net = "tcp"
		cfg.Addr = addr
	}
	i.Addr = cfg.Addr

	// set timeout
	if i.timeoutDuration != 0 {
		cfg.Timeout = i.timeoutDuration
	}

	// set Charset
	if i.Charset != "" {
		cfg.Params["charset"] = i.Charset
	}

	// tls (todo)
	return cfg.FormatDSN()
}

func (*Input) PipelineConfig() map[string]string {
	pipelineMap := map[string]string{
		inputName: pipelineCfg,
	}
	return pipelineMap
}

func (i *Input) GetPipeline() []*tailer.Option {
	return []*tailer.Option{
		{
			Source:  inputName,
			Service: inputName,
			Pipeline: func() string {
				if i.Log != nil {
					return i.Log.Pipeline
				}
				return ""
			}(),
		},
	}
}

func (i *Input) initCfg() error {
	var err error
	i.timeoutDuration, err = time.ParseDuration(i.Timeout)
	if err != nil {
		i.timeoutDuration = 10 * time.Second
	}

	dsnStr := i.getDsnString()

	db, err := sql.Open("mysql", dsnStr)
	if err != nil {
		l.Errorf("sql.Open(): %s", err.Error())
		return err
	} else {
		i.db = db
	}

	ctx, cancel := context.WithTimeout(context.Background(), i.timeoutDuration)
	defer cancel()

	if err := i.db.PingContext(ctx); err != nil {
		l.Errorf("init config connect error %v", err)
		return err
	}

	i.globalTag()
	if i.Dbm {
		i.initDbm()
	}
	return nil
}

func (i *Input) initDbm() {
	i.dbmSampleCache.explainCache.Size = 1000 // max size
	i.dbmSampleCache.explainCache.TTL = 60    // 60 second to live
}

func (i *Input) globalTag() {
	i.Tags["server"] = i.Addr
	i.Tags["service_name"] = i.Service
}

func (i *Input) q(s string) rows {
	rows, err := i.db.Query(s)
	if err != nil {
		l.Errorf("query %s failed: %s, ignored", s, err.Error())
		return nil
	}

	if err := rows.Err(); err != nil {
		closeRows(rows)
		l.Errorf("query %s failed: %s, ignored", s, err.Error())
		return nil
	}

	return rows
}

// init db connect.
func (i *Input) initDBConnect() error {
	isNeedConnect := false

	if i.db == nil {
		isNeedConnect = true
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer func() {
			cancel()
		}()

		if err := i.db.PingContext(ctx); err != nil {
			isNeedConnect = true
		}
	}

	if isNeedConnect {
		if err := i.initCfg(); err != nil {
			return err
		}
	}

	return nil
}

// mysql.
func (i *Input) metricCollectMysql() ([]*io.Point, error) {
	if err := i.collectMysql(); err != nil {
		return []*io.Point{}, err
	}

	pts, err := i.buildMysql()
	if err != nil {
		return []*io.Point{}, err
	}
	return pts, nil
}

// mysql_schema.
func (i *Input) metricCollectMysqlSchema() ([]*io.Point, error) {
	if err := i.collectMysqlSchema(); err != nil {
		return []*io.Point{}, err
	}

	pts, err := i.buildMysqlSchema()
	if err != nil {
		return []*io.Point{}, err
	}
	return pts, nil
}

// mysql_table_schema.
func (i *Input) metricCollectMysqlTableSschema() ([]*io.Point, error) {
	if err := i.collectMysqlTableSchema(); err != nil {
		return []*io.Point{}, err
	}

	pts, err := i.buildMysqlTableSchema()
	if err != nil {
		return []*io.Point{}, err
	}
	return pts, nil
}

// mysql_user_status.
func (i *Input) metricCollectMysqlUserStatus() ([]*io.Point, error) {
	if err := i.collectMysqlUserStatus(); err != nil {
		return []*io.Point{}, err
	}

	pts, err := i.buildMysqlUserStatus()
	if err != nil {
		return []*io.Point{}, err
	}
	return pts, nil
}

// mysql_custom_queries.
func (i *Input) metricCollectMysqlCustomQueries() ([]*io.Point, error) {
	if err := i.collectMysqlCustomQueries(); err != nil {
		return []*io.Point{}, err
	}

	pts, err := i.buildMysqlCustomQueries()
	if err != nil {
		return []*io.Point{}, err
	}
	return pts, nil
}

// mysql_innodb.
func (i *Input) metricCollectMysqlInnodb() ([]*io.Point, error) {
	if err := i.collectMysqlInnodb(); err != nil {
		return []*io.Point{}, err
	}

	pts, err := i.buildMysqlInnodb()
	if err != nil {
		return []*io.Point{}, err
	}
	return pts, nil
}

// mysql_dbm_metric.
func (i *Input) metricCollectMysqlDbmMetric() ([]*io.Point, error) {
	if err := i.collectMysqlDbmMetric(); err != nil {
		return []*io.Point{}, err
	}

	pts, err := i.buildMysqlDbmMetric()
	if err != nil {
		return []*io.Point{}, err
	}
	return pts, nil
}

// mysql_dbm_sample.
func (i *Input) metricCollectMysqlDbmSample() ([]*io.Point, error) {
	if err := i.collectMysqlDbmSample(); err != nil {
		return []*io.Point{}, err
	}

	pts, err := i.buildMysqlDbmSample()
	if err != nil {
		return []*io.Point{}, err
	}
	return pts, nil
}

func (i *Input) resetLastError() {
	i.lastErrors = []string{}
}

func (i *Input) handleLastError() {
	if len(i.lastErrors) > 0 {
		io.FeedLastError(inputName, strings.Join(i.lastErrors, "; "))
		i.resetLastError()
	}
}

func (i *Input) appendLastError(err error) {
	if err != nil {
		i.lastErrors = append(i.lastErrors, err.Error())
	}
}

func (i *Input) Collect() (map[string][]*io.Point, error) {
	if err := i.initDBConnect(); err != nil {
		return map[string][]*io.Point{}, err
	}

	if len(i.collectors) == 0 {
		i.collectors = []func() ([]*io.Point, error){
			i.metricCollectMysql,              // mysql
			i.metricCollectMysqlSchema,        // mysql_schema
			i.metricCollectMysqlTableSschema,  // mysql_table_schema
			i.metricCollectMysqlUserStatus,    // mysql_user_status
			i.metricCollectMysqlCustomQueries, // mysql_custom_queries
		}
	}

	i.start = time.Now()

	var ptsMetric, ptsLoggingMetric, ptsLoggingSample []*io.Point

	for idx, f := range i.collectors {
		l.Debugf("collecting %d(%v)...", idx, f)

		pts, err := f()
		if err != nil {
			l.Errorf("collectors %v failed: %s", f, err.Error())
			i.appendLastError(err)
		}

		if len(pts) > 0 {
			ptsMetric = append(ptsMetric, pts...)
		}
	}

	if i.InnoDB {
		// mysql_innodb
		pts, err := i.metricCollectMysqlInnodb()
		if err != nil {
			l.Errorf("metricCollectMysqlInnodb failed: %s", err.Error())
			i.appendLastError(err)
		}

		if len(pts) > 0 {
			ptsMetric = append(ptsMetric, pts...)
		}
	}

	if i.Dbm && (i.DbmMetric.Enabled || i.DbmSample.Enabled) {
		g := goroutine.NewGroup(goroutine.Option{Name: goroutine.GetInputName("mysql_dbm")})
		if i.DbmMetric.Enabled {
			g.Go(func(ctx context.Context) error {
				// mysql_dbm_metric
				pts, err := i.metricCollectMysqlDbmMetric()
				if err != nil {
					l.Errorf("metricCollectMysqlDbmMetric failed: %s", err.Error())
					i.appendLastError(err)
				}

				if len(pts) > 0 {
					ptsLoggingMetric = append(ptsLoggingMetric, pts...)
				}
				return nil
			})
		}
		if i.DbmSample.Enabled {
			g.Go(func(ctx context.Context) error {
				// mysql_dbm_sample
				pts, err := i.metricCollectMysqlDbmSample()
				if err != nil {
					l.Errorf("metricCollectMysqlDbmSample failed: %s", err.Error())
					i.appendLastError(err)
				}

				if len(pts) > 0 {
					ptsLoggingSample = append(ptsLoggingSample, pts...)
				}
				return nil
			})
		}

		err := g.Wait()
		if err != nil {
			l.Errorf("mysql dmb collect error: %v", err)
			io.FeedLastError(inputName, err.Error())
		}
	} // if

	mpts := make(map[string][]*io.Point)
	mpts[datakit.Metric] = ptsMetric

	ptsLoggingMetric = append(ptsLoggingMetric, ptsLoggingSample...) // two combine in one
	mpts[datakit.Logging] = ptsLoggingMetric

	return mpts, nil
}

func (i *Input) RunPipeline() {
	if i.Log == nil || len(i.Log.Files) == 0 {
		return
	}

	if i.Log.Pipeline == "" {
		i.Log.Pipeline = "mysql.p" // use default
	}

	opt := &tailer.Option{
		Source:            "mysql",
		Service:           "mysql",
		Pipeline:          i.Log.Pipeline,
		GlobalTags:        i.Tags,
		CharacterEncoding: i.Log.CharacterEncoding,
		MultilineMatch:    i.Log.MultilineMatch,
	}
	var err error
	i.tail, err = tailer.NewTailer(i.Log.Files, opt, i.Log.IgnoreStatus)
	if err != nil {
		l.Error(err)
		io.FeedLastError(inputName, err.Error())
		return
	}

	go i.tail.Start()
}

func (i *Input) Run() {
	l = logger.SLogger("mysql")
	i.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, i.Interval.Duration)

	tick := time.NewTicker(i.Interval.Duration)
	defer tick.Stop()

	// Try until init OK.
	for {
		if err := i.initCfg(); err != nil {
			io.FeedLastError(inputName, err.Error())
		} else {
			break
		}

		select {
		case <-datakit.Exit.Wait():

			if i.tail != nil {
				i.tail.Close() //nolint:errcheck
			}
			l.Info("mysql exit")

			return

		case <-i.semStop.Wait():
			return

		case <-tick.C:
		}
	}

	l.Infof("collecting each %v", i.Interval.Duration)

	for {
		if i.pause {
			l.Debugf("not leader, skipped")
		} else {
			l.Debugf("mysql input gathering...")
		}

		l.Debugf("mysql input gathering...")

		i.resetLastError()

		mpts, err := i.Collect()
		if err != nil {
			l.Warnf("i.Collect failed: %v", err)
			io.FeedLastError(inputName, err.Error())
		}

		for category, pts := range mpts {
			if len(pts) > 0 {
				if err := io.Feed(inputName, category, pts,
					&io.Option{CollectCost: time.Since(i.start)}); err != nil {
					l.Warnf("io.Feed failed: %v", err)
					io.FeedLastError(inputName, err.Error())
				} // if err
			}
		} // for

		i.handleLastError()

		select {
		case <-datakit.Exit.Wait():
			i.exit()
			l.Info("mysql exit")
			return

		case <-i.semStop.Wait():
			i.exit()
			l.Info("mysql return")
			return

		case <-tick.C:

		case i.pause = <-i.pauseCh:
			// nil
		}
	}
}

func (i *Input) exit() {
	if i.tail != nil {
		i.tail.Close()
		l.Info("mysql log exit")
	}
}

func (i *Input) Terminate() {
	if i.semStop != nil {
		i.semStop.Close()
	}
}

func (*Input) Catalog() string { return catalogName }

func (*Input) SampleConfig() string { return configSample }

func (*Input) AvailableArchs() []string { return datakit.AllArch }

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&baseMeasurement{},
		&schemaMeasurement{},
		&innodbMeasurement{},
		&tbMeasurement{},
		&userMeasurement{},
		&dbmStateMeasurement{},
		&dbmSampleMeasurement{},
	}
}

func (i *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	defer tick.Stop()
	select {
	case i.pauseCh <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (i *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	defer tick.Stop()
	select {
	case i.pauseCh <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			Tags:    make(map[string]string),
			Timeout: "10s",
			pauseCh: make(chan bool, inputs.ElectionPauseChannelLength),

			semStop: cliutils.NewSem(),
		}
	})
}
