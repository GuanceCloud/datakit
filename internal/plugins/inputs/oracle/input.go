// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package oracle collect Oracle metrics
package oracle

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/jmoiron/sqlx"
	go_ora "github.com/sijms/go-ora/v2"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
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
	customQueryFeedName  = inputName + "-custom_query"
	loggingFeedName      = inputName + "-L"
	catalogName          = "db"
)

var l = logger.DefaultSLogger(inputName)

type Input struct {
	Host              string           `toml:"host"`
	Port              int              `toml:"port"`
	User              string           `toml:"user"`
	Password          string           `toml:"password"`
	Interval          datakit.Duration `toml:"interval"`
	Timeout           string           `toml:"connect_timeout"`
	Service           string           `toml:"service"`
	MetricExcludeList []string         `toml:"metric_exclude_list"`
	timeoutDuration   time.Duration
	Query             []*customQuery    `toml:"custom_queries"`
	SlowQueryTime     string            `toml:"slow_query_time"`
	Election          bool              `toml:"election"`
	Tags              map[string]string `toml:"tags"`

	mainVersion, // simple version like 11
	fullVersion string // full version like 'Oracle Database 11g Express Edition Release 11.2.0.2.0 - 64bit Production'

	Uptime             int
	CollectCoStatus    string
	CollectCoErrMsg    string
	LastCustomerObject *customerObjectMeasurement

	semStop        *cliutils.Sem // start stop signal
	pauseCh        chan bool
	feeder         dkio.Feeder
	tagger         datakit.GlobalTagger
	mergedTags     map[string]string
	db             *sqlx.DB
	pause          bool
	ptsTime        time.Time
	slowQueryTime  time.Duration
	lastActiveTime string
	cacheSQL       map[string]string

	UpState int
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

	db.SetConnMaxLifetime(ipt.Interval.Duration) // avoid max cursor problem

	ctx, cancel := context.WithTimeout(context.Background(), ipt.timeoutDuration)
	defer cancel()

	if err := ipt.db.PingContext(ctx); err != nil {
		l.Errorf("init config connect error %v", err)
		ipt.db.Close() //nolint:errcheck,gosec
		return err
	}

	ipt.getOracleVersion()

	return nil
}

func (ipt *Input) getConnString() string {
	opt := map[string]string{
		"timeout": fmt.Sprintf("%d", ipt.timeoutDuration/time.Second),
	}

	connStr := go_ora.BuildUrl(ipt.Host, ipt.Port, ipt.Service, ipt.User, ipt.Password, opt)

	return connStr
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

	ipt.collectOracleProcess()
	ipt.collectOracleTableSpace()
	ipt.collectOracleSystem()
	ipt.collectSlowQuery()
	ipt.collectWaitingEvent()
	ipt.collectLockedSession()

	ipt.getOracleUptime()
}

func (ipt *Input) Init() {
	var err error

	l = logger.SLogger(inputName)
	ipt.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval.Duration)
	tick := time.NewTicker(ipt.Interval.Duration)
	defer tick.Stop()

	setHost := false
	host := strings.ToLower(ipt.Host)
	switch host {
	case "", "localhost":
		setHost = true
	default:
		if net.ParseIP(host).IsLoopback() {
			setHost = true
		}
	}

	if setHost {
		host, err = os.Hostname()
		if err != nil {
			l.Errorf("os.Hostname failed: %v", err)
		}
	}

	ipt.mergedTags = inputs.MergeTags(ipt.tagger.ElectionTags(), ipt.Tags, host)
	ipt.mergedTags["oracle_service"] = ipt.Service
	ipt.mergedTags["oracle_server"] = fmt.Sprintf("%s:%d", ipt.Host, ipt.Port)

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
			return

		case <-ipt.semStop.Wait():
			return

		case <-tick.C:
		}

		// on init failing, we still upload up metric to show that the oracle input not working.
		ipt.FeedUpMetric()
	}
}

func (ipt *Input) Run() {
	tick := time.NewTicker(ipt.Interval.Duration)
	defer tick.Stop()
	defer func() {
		l.Info("oracle exit")
	}()

	ipt.Init()

	l.Infof("collecting each %v", ipt.Interval.Duration)

	// run custom queries
	ipt.runCustomQueries()

	ipt.ptsTime = ntp.Now()
	for {
		if ipt.pause {
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

		case ipt.pause = <-ipt.pauseCh:
			// nil
		}
	}
}

func (ipt *Input) Catalog() string { return catalogName }

func (ipt *Input) SampleConfig() string { return configSample }

func (ipt *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&processMeasurement{},
		&tablespaceMeasurement{},
		&systemMeasurement{},
		&customerObjectMeasurement{},
		&slowQueryMeasurement{},
		&waitingEventMeasurement{},
		&lockMeasurement{},

		&inputs.UpMeasurement{},
	}
}

func (ipt *Input) AvailableArchs() []string {
	return datakit.AllOSWithElection
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

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
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

func defaultInput() *Input {
	return &Input{
		Tags:     make(map[string]string),
		Timeout:  "10s",
		pauseCh:  make(chan bool, inputs.ElectionPauseChannelLength),
		Election: true,
		feeder:   dkio.DefaultFeeder(),
		tagger:   datakit.DefaultGlobalTagger(),
		semStop:  cliutils.NewSem(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
