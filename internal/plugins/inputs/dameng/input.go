// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package dameng collect Dameng database metrics
package dameng

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/coreos/go-semver/semver"
	"github.com/jmoiron/sqlx"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/dameng/driver"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
)

var _ inputs.ElectionInput = (*Input)(nil)

const (
	maxInterval = 15 * time.Minute
	minInterval = 10 * time.Second

	metricNameConnection    = "dameng_connections"
	metricNameLocks         = "dameng_locks"
	metricNameDeadlock      = "dameng_deadlock"
	metricNameMemory        = "dameng_memory"
	metricNameMemPool       = "dameng_mem_pool"
	metricNameBufferCache   = "dameng_buffer_cache"
	metricNameBlockSessions = "dameng_block_sessions"
	metricNameRates         = "dameng_rates"
	metricNameTablespace    = "dameng_tablespace"
	metricNameSlowQueries   = "dameng_slow_query"
)

var (
	inputName           = "dameng"
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
	Log                *damengLog        `toml:"log"`
	Election           bool              `toml:"election"`
	Tags               map[string]string `toml:"tags"`

	collectFuncs map[string]func() error
	collectCache []*point.Point
	Version      *semver.Version

	LastStatValues map[string]int64
	LastStatTime   time.Time

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

// init db connect.
func (ipt *Input) initDBConnect() error {
	hostPort := net.JoinHostPort(ipt.Host, strconv.Itoa(ipt.Port))
	connStr := fmt.Sprintf("dm://%s:%s@%s", ipt.User, ipt.Password, hostPort)
	db, err := sqlx.Connect("dm", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to dm: %w", err)
	}
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping dm: %w", err)
	}
	ipt.db = db

	return nil
}

func (ipt *Input) Init() {
	ipt.collectFuncs = map[string]func() error{
		metricNameMemory:     ipt.collectMemory,
		metricNameMemPool:    ipt.collectMemPool,
		metricNameTablespace: ipt.collectTablespace,
		metricNameConnection: ipt.collectConnections,

		metricNameRates:         ipt.collectRates,
		metricNameSlowQueries:   ipt.collectSlowQueries,
		metricNameLocks:         ipt.collectLocks,
		metricNameDeadlock:      ipt.collectDeadlock,
		metricNameBufferCache:   ipt.collectBufferCache,
		metricNameBlockSessions: ipt.collectBlockSessions,
	}

	for _, metric := range ipt.MetricExcludeList {
		delete(ipt.collectFuncs, metric)
	}

	if ipt.Timeout != "" {
		dur, err := time.ParseDuration(ipt.Timeout)
		if err != nil {
			l.Warnf("Invalid timeout %s, using default 10s: %w", ipt.Timeout, err)
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
			l.Errorf("os.Hostname failed: %w", err)
		}
	}

	l.Infof("%s input started", inputName)
	ipt.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval.Duration)
	ipt.mergedTags = inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, host)
	l.Debugf("merged tags: %+#v", ipt.mergedTags)

	if err := ipt.initDBConnect(); err != nil {
		// l.Errorf("Failed to initialize DB connection: %w", err)
		return fmt.Errorf("failed to initialize DB connection: %w", err)
	}
	return nil
}

func (ipt *Input) Run() {
	if err := ipt.setup(); err != nil {
		l.Errorf("Setup failed: %w", err)
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
			l.Errorf("Failed to collect %s: %w", name, err)
			ipt.feeder.FeedLastError(err.Error(),
				metrics.WithLastErrorInput(inputName),
				metrics.WithLastErrorCategory(point.Metric),
			)
			continue
		}
	}
	return nil
}

type damengLog struct {
	Files             []string `toml:"files"`
	Pipeline          string   `toml:"pipeline"`
	IgnoreStatus      []string `toml:"ignore"`
	CharacterEncoding string   `toml:"character_encoding"`
	MultilineMatch    string   `toml:"multiline_match"`
}

func (*Input) PipelineConfig() map[string]string {
	return map[string]string{
		"dameng": pipelineCfg,
	}
}

//nolint:lll
func (ipt *Input) LogExamples() map[string]map[string]string {
	return map[string]map[string]string{
		"dameng": {
			"dameng log": `2025-06-23 10:04:27.793 [INFO] dminit P0000008230 T0000000000000008230  INI parameter RECYCLE_POOLS changed, the original value 0, new value 1

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

	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_dameng"})
	g.Go(func(ctx context.Context) error {
		ipt.tail.Start()
		return nil
	})
}

func (ipt *Input) exit() {
	if ipt.tail != nil {
		ipt.tail.Close()
		l.Info("dameng log exit")
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
	return []string{datakit.OSLabelLinux, datakit.OSLabelWindows, datakit.LabelK8s, datakit.LabelDocker, datakit.LabelElection}
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&MemoryMeasurement{},
		&MemPoolMeasurement{},
		&TablespaceMeasurement{},
		&ConnectionsMeasurement{},
		&RatesMeasurement{},
		&SlowQueriesMeasurement{},
		&LocksMeasurement{},
		&DeadlockMeasurement{},
		&BufferCacheMeasurement{},
		&BlockSessionsMeasurement{},
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
