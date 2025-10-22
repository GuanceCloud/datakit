// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package kingbase collect Kingbase metrics
package kingbase

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

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/kingbase/driver"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
)

var _ inputs.ElectionInput = (*Input)(nil)

const (
	maxInterval = 15 * time.Minute
	minInterval = 10 * time.Second

	metricNameConnection        = "kingbase_connections"
	metricNameTransactions      = "kingbase_transactions"
	metricNameQueryPerformance  = "kingbase_query_performance"
	metricNameLocks             = "kingbase_locks"
	metricNameQueryStats        = "kingbase_query_stats"
	metricNameBufferCache       = "kingbase_buffer_cache"
	metricNameDatabaseStatus    = "kingbase_database_status"
	metricNameTablespace        = "kingbase_tablespace"
	metricNameLockDetails       = "kingbase_lock_details"
	metricNameIndexUsage        = "kingbase_index_usage"
	metricNameBackgroundWriter  = "kingbase_background_writer"
	metricNameSessionActivity   = "kingbase_session_activity"
	metricNameQueryCancellation = "kingbase_query_cancellation"
	metricNameFunctionStats     = "kingbase_function_stats"
	metricNameSlowQueries       = "kingbase_slow_query"
)

var (
	inputName           = "kingbase"
	metricName          = inputName
	catalogName         = "db"
	customQueryFeedName = dkio.FeedSource(inputName, "custom_query")
	l                   = logger.DefaultSLogger(inputName)
)

type Input struct {
	Host               string           `toml:"host"`
	Port               int              `toml:"port"`
	User               string           `toml:"user"`
	Password           string           `toml:"password"`
	Database           string           `toml:"database"`
	Interval           datakit.Duration `toml:"interval"`
	MetricExcludeList  []string         `toml:"metric_exclude_list"`
	Timeout            string           `toml:"connect_timeout"`
	timeoutDuration    time.Duration
	SlowQueryThreshold int64             `toml:"slow_query_threshold"`
	Query              []*customQuery    `toml:"custom_queries"`
	Log                *kingbaseLog      `toml:"log"`
	Election           bool              `toml:"election"`
	Tags               map[string]string `toml:"tags"`

	collectFuncs map[string]func() error
	collectCache []*point.Point
	Version      string

	semStop    *cliutils.Sem
	pause      bool
	pauseCh    chan bool
	tail       *tailer.Tailer
	feeder     dkio.Feeder
	tagger     datakit.GlobalTagger
	mergedTags map[string]string
	db         *sqlx.DB
	ptsTime    time.Time
}

func (ipt *Input) checkSysStatStatements() error {
	// Check if sys_stat_statements extension is enabled
	var count int
	err := ipt.db.Get(&count, "SELECT count(*) FROM sys_catalog.sys_extension WHERE extname = 'sys_stat_statements'")
	if err != nil {
		return fmt.Errorf("failed to check sys_stat_statements extension: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("sys_stat_statements extension is not enabled; please run 'CREATE EXTENSION sys_stat_statements'")
	}

	// Check sys_stat_statements.track setting
	var trackSetting string
	err = ipt.db.Get(&trackSetting, "SHOW sys_stat_statements.track")
	if err != nil {
		return fmt.Errorf("failed to check sys_stat_statements.track: %w", err)
	}
	if trackSetting == "none" {
		return fmt.Errorf("sys_stat_statements.track is set to 'none', no data will be collected from this view")
	}

	return nil
}

func (ipt *Input) setKBVersion() error {
	var versionStr string
	if err := ipt.db.Get(&versionStr, "SELECT version()"); err != nil {
		return fmt.Errorf("failed to get version: %w", err)
	}
	version, err := extractVersion(versionStr)
	if err != nil {
		return fmt.Errorf("failed to extract version: %w", err)
	}
	ipt.Version = version
	return nil
}

// init db connect.
func (ipt *Input) initDBConnect() error {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		ipt.Host, ipt.Port, ipt.User, ipt.Password, ipt.Database)
	db, err := sqlx.Connect("kingbase", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to Kingbase: %w", err)
	}
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping Kingbase: %w", err)
	}
	ipt.db = db
	return nil
}

func (ipt *Input) Init() {
	ipt.collectFuncs = map[string]func() error{
		metricNameConnection:        ipt.collectConnections,
		metricNameTransactions:      ipt.collectTransactions,
		metricNameQueryPerformance:  ipt.collectQueryPerformance,
		metricNameLocks:             ipt.collectLocks,
		metricNameQueryStats:        ipt.collectQueryStats,
		metricNameBufferCache:       ipt.collectBufferCache,
		metricNameDatabaseStatus:    ipt.collectDatabaseStatus,
		metricNameTablespace:        ipt.collectTablespace,
		metricNameLockDetails:       ipt.collectLockDetails,
		metricNameIndexUsage:        ipt.collectIndexUsage,
		metricNameBackgroundWriter:  ipt.collectBackgroundWriter,
		metricNameSessionActivity:   ipt.collectSessionActivity,
		metricNameQueryCancellation: ipt.collectQueryCancellation,
		metricNameFunctionStats:     ipt.collectFunctionStats,
		metricNameSlowQueries:       ipt.collectSlowQueries,
	}

	for _, metric := range ipt.MetricExcludeList {
		delete(ipt.collectFuncs, metric)
	}

	if ipt.Timeout != "" {
		dur, err := time.ParseDuration(ipt.Timeout)
		if err != nil {
			l.Warnf("Invalid timeout %s, using default 10s: %v", ipt.Timeout, err)
			ipt.timeoutDuration = 10 * time.Second
		} else {
			ipt.timeoutDuration = dur
		}
	} else {
		ipt.timeoutDuration = 10 * time.Second
	}
}

func (ipt *Input) setup() error {
	var err error
	l = logger.SLogger(inputName)

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

	l.Infof("%s input started", inputName)
	ipt.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval.Duration)
	ipt.mergedTags = inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, host)
	l.Debugf("merged tags: %+#v", ipt.mergedTags)

	if err := ipt.initDBConnect(); err != nil {
		// l.Errorf("Failed to initialize DB connection: %v", err)
		return fmt.Errorf("failed to initialize DB connection: %w", err)
	}

	// set version
	if err := ipt.setKBVersion(); err != nil {
		l.Warnf("kingbase version set error: %w", err)
	}

	// 检查 sys_stat_statements 扩展
	if err := ipt.checkSysStatStatements(); err != nil {
		l.Warnf("sys_stat_statements check failed: %w", err)
	}
	return nil
}

func (ipt *Input) Run() {
	if err := ipt.setup(); err != nil {
		l.Errorf("Setup failed: %v", err)
		ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(inputName),
			metrics.WithLastErrorCategory(point.Metric),
		)
		return
	}

	tick := time.NewTicker(ipt.Interval.Duration)
	defer tick.Stop()

	ipt.Init()

	ipt.ptsTime = ntp.Now()

	// run custom queries
	ipt.runCustomQueries()

	for {
		start := time.Now()
		if ipt.pause {
			l.Debugf("not leader, skipped")
		} else {
			if err := ipt.Collect(); err != nil {
				l.Errorf("collect: %s", err)
				ipt.feeder.FeedLastError(err.Error(),
					metrics.WithLastErrorInput(inputName),
					metrics.WithLastErrorCategory(point.Metric),
				)
			}

			if len(ipt.collectCache) > 0 {
				if err := ipt.feeder.Feed(point.Metric, ipt.collectCache,
					dkio.WithCollectCost(time.Since(start)),
					dkio.WithElection(ipt.Election),
					dkio.WithSource(metricName)); err != nil {
					ipt.feeder.FeedLastError(err.Error(),
						metrics.WithLastErrorInput(inputName),
						metrics.WithLastErrorCategory(point.Metric),
					)
					l.Errorf("feed measurement: %s", err)
				}
			}
		}
		select {
		case tt := <-tick.C:
			ipt.ptsTime = inputs.AlignTime(tt, ipt.ptsTime, ipt.Interval.Duration)
		case <-datakit.Exit.Wait():
			ipt.exit()
			l.Infof("%s input exit", inputName)
			return
		case <-ipt.semStop.Wait():
			ipt.exit()
			l.Infof("%s input return", inputName)
			return
		case ipt.pause = <-ipt.pauseCh:
			// nil
		}
	}
}

func (ipt *Input) Collect() error {
	ipt.collectCache = make([]*point.Point, 0)
	for name, fn := range ipt.collectFuncs {
		if err := fn(); err != nil {
			l.Errorf("Failed to collect %s: %v", name, err)
			ipt.feeder.FeedLastError(err.Error(),
				metrics.WithLastErrorInput(inputName),
				metrics.WithLastErrorCategory(point.Metric),
			)
			continue
		}
	}
	return nil
}

type kingbaseLog struct {
	Files             []string `toml:"files"`
	Pipeline          string   `toml:"pipeline"`
	IgnoreStatus      []string `toml:"ignore"`
	CharacterEncoding string   `toml:"character_encoding"`
	MultilineMatch    string   `toml:"multiline_match"`
}

func (*Input) PipelineConfig() map[string]string {
	return map[string]string{
		"kingbase": pipelineCfg,
	}
}

//nolint:lll
func (ipt *Input) LogExamples() map[string]map[string]string {
	return map[string]map[string]string{
		"kingbase": {
			"Kingbase log": `2025-06-17 13:07:10.952 UTC [999] ERROR:  relation "sys_stat_activity" does not exist at character 240
			`,
		},
	}
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

func (ipt *Input) RunPipeline() {
	if ipt.Log == nil || len(ipt.Log.Files) == 0 {
		return
	}

	if ipt.Log.MultilineMatch == "" {
		ipt.Log.MultilineMatch = `^\d{4}-\d{2}-\d{2}`
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

	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_kingbase"})
	g.Go(func(ctx context.Context) error {
		ipt.tail.Start()
		return nil
	})
}

func (ipt *Input) exit() {
	if ipt.tail != nil {
		ipt.tail.Close()
		l.Info("kingbase log exit")
	}
}

func (*Input) Singleton() {}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
	if ipt.db != nil {
		if err := ipt.db.Close(); err != nil {
			l.Error(err)
		}
	}
}
func (*Input) Catalog() string { return catalogName }

func (*Input) SampleConfig() string { return sampleCfg }

func (*Input) AvailableArchs() []string {
	return datakit.AllOSWithElection
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&ConnectionsMeasurement{},
		&TransactionsMeasurement{},
		&QueryPerformanceMeasurement{},
		&LocksMeasurement{},
		&QueryStatsMeasurement{},
		&BufferCacheMeasurement{},
		&DatabaseStatusMeasurement{},
		&TablespaceMeasurement{},
		&LockDetailsMeasurement{},
		&IndexUsageMeasurement{},
		&BackgroundWriterMeasurement{},
		&SessionActivityMeasurement{},
		&QueryCancellationMeasurement{},
		&FunctionStatsMeasurement{},
		&SlowQueriesMeasurement{},
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

func (ipt *Input) ElectionEnabled() bool {
	return ipt.Election
}

func defaultInput() *Input {
	ipt := &Input{
		Interval:   datakit.Duration{Duration: time.Second * 10},
		pauseCh:    make(chan bool, inputs.ElectionPauseChannelLength),
		Election:   true,
		Tags:       make(map[string]string),
		feeder:     dkio.DefaultFeeder(),
		semStop:    cliutils.NewSem(),
		tagger:     datakit.DefaultGlobalTagger(),
		mergedTags: make(map[string]string),
	}
	return ipt
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
