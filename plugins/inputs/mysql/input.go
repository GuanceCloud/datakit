// Package mysql collect MySQL metrics
package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-sql-driver/mysql"
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
	tail       *tailer.Tailer
	collectors []func() ([]inputs.Measurement, error)

	err error

	pause   bool
	pauseCh chan bool

	dbmCache       map[string]dbmRow
	dbmSampleCache dbmSampleCache
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

func (i *Input) PipelineConfig() map[string]string {
	pipelineMap := map[string]string{
		"mysql": pipelineCfg,
	}
	return pipelineMap
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

func (i *Input) Collect() {
	ctx, cancel := context.WithTimeout(context.Background(), i.timeoutDuration)
	defer cancel()

	if err := i.db.PingContext(ctx); err != nil {
		l.Errorf("connect error %v", err)
		io.FeedLastError(inputName, err.Error())
		return
	}

	for idx, f := range i.collectors {
		l.Debugf("collecting %d(%v)...", idx, f)

		if ms, err := f(); err != nil {
			io.FeedLastError(inputName, err.Error())
		} else {
			if len(ms) == 0 {
				continue
			}

			if err := inputs.FeedMeasurement(inputName,
				datakit.Metric,
				ms,
				&io.Option{CollectCost: time.Since(i.start)}); err != nil {
				l.Error(err)
			}
		}
	}

	if i.Dbm && (i.DbmMetric.Enabled || i.DbmSample.Enabled) {
		g := goroutine.NewGroup(goroutine.Option{Name: goroutine.GetInputName("mysql_dbm")})
		if i.DbmMetric.Enabled {
			g.Go(func(ctx context.Context) error {
				ms, err := i.collectStatementMetrics()
				if err != nil {
					return err
				}
				if len(ms) > 0 {
					if err := inputs.FeedMeasurement(dbmMetricName,
						datakit.Logging,
						ms,
						&io.Option{CollectCost: time.Since(i.start)}); err != nil {
						l.Error(err)
					}
				}
				return nil
			})
		}
		if i.DbmSample.Enabled {
			g.Go(func(ctx context.Context) error {
				ms, err := i.collectStatementSamples()
				if err != nil {
					return err
				}
				if len(ms) > 0 {
					if err := inputs.FeedMeasurement(dbmMetricName,
						datakit.Logging,
						ms,
						&io.Option{CollectCost: time.Since(i.start)}); err != nil {
						l.Error(err)
					}
				}
				return nil
			})
		}

		err := g.Wait()
		if err != nil {
			l.Errorf("mysql dmb collect error: %v", err)
			io.FeedLastError(inputName, err.Error())
		}
	}
}

// 获取base指标.
func (i *Input) collectBaseMeasurement() ([]inputs.Measurement, error) {
	m := &baseMeasurement{
		i:       i,
		resData: make(map[string]interface{}),
		tags:    make(map[string]string),
		fields:  make(map[string]interface{}),
	}

	m.name = "mysql"
	for key, value := range i.Tags {
		m.tags[key] = value
	}

	if err := m.getStatus(); err != nil {
		return nil, err
	}

	if err := m.getVariables(); err != nil {
		return nil, err
	}

	// 如果没有打开 bin-log，这里可能报错：Error 1381: You are not using binary logging
	// 不过忽略这一错误
	// TODO: if-bin-log-enabled
	if m.resData["log_bin"] == "ON" || m.resData["log_bin"] == "on" {
		_ = m.getLogStats()
	}

	if err := m.submit(); err == nil {
		if len(m.fields) > 0 {
			return []inputs.Measurement{m}, nil
		}
	}

	return nil, nil
}

// 获取innodb指标.
func (i *Input) collectInnodbMeasurement() ([]inputs.Measurement, error) {
	return i.getInnodb()
}

// 获取tableSchema指标.
func (i *Input) collectTableSchemaMeasurement() ([]inputs.Measurement, error) {
	return i.getTableSchema()
}

// 获取用户指标.
func (i *Input) collectUserMeasurement() ([]inputs.Measurement, error) {
	return i.getUserData()
}

// 获取schema指标.
func (i *Input) collectSchemaMeasurement() ([]inputs.Measurement, error) {
	x, err := i.getSchemaSize()
	if err != nil {
		return nil, err
	}

	y, err := i.getQueryExecTimePerSchema()
	if err != nil {
		return nil, err
	}

	return append(x, y...), nil
}

// dbm metric.
func (i *Input) collectStatementMetrics() ([]inputs.Measurement, error) {
	return i.getDbmMetric()
}

// dbm sample.
func (i *Input) collectStatementSamples() ([]inputs.Measurement, error) {
	return i.getDbmSample()
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
		GlobalTags:        i.Tags,
		CharacterEncoding: i.Log.CharacterEncoding,
		MultilineMatch:    i.Log.MultilineMatch,
	}

	pl := filepath.Join(datakit.PipelineDir, i.Log.Pipeline)
	if _, err := os.Stat(pl); err != nil {
		l.Warn("%s missing: %s", pl, err.Error())
	} else {
		opt.Pipeline = pl
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
		case <-tick.C:
		}
	}

	l.Infof("collecting each %v", i.Interval.Duration)

	i.collectors = []func() ([]inputs.Measurement, error){
		i.collectBaseMeasurement,
		i.collectSchemaMeasurement,
		i.customSchemaMeasurement,
		i.collectTableSchemaMeasurement,
		i.collectUserMeasurement,
	}

	if i.InnoDB {
		i.collectors = append(i.collectors, i.collectInnodbMeasurement)
	}

	for {
		if i.pause {
			l.Debugf("not leader, skipped")
			continue
		}
		l.Debugf("mysql input gathering...")
		i.start = time.Now()
		i.Collect()

		select {
		case <-datakit.Exit.Wait():
			if i.tail != nil {
				i.tail.Close()
			}
			l.Info("mysql exit")
			return
		case <-tick.C:

		case i.pause = <-i.pauseCh:
			// nil
		}
	}
}

func (i *Input) Catalog() string { return catalogName }

func (i *Input) SampleConfig() string { return configSample }

func (i *Input) AvailableArchs() []string { return datakit.AllArch }

func (i *Input) SampleMeasurement() []inputs.Measurement {
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
		}
	})
}
