// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package redis collects redis metrics.
package redis

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/go-redis/redis/v8"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	dknet "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
)

const (
	maxInterval         = 60 * time.Minute
	minInterval         = 15 * time.Second
	defaultKeyInterval  = 5 * time.Minute
	defaultKeyScanSleep = "0.1"
)

var (
	inputName                                 = "redis"
	customObjectFeedName                      = inputName + "/CO"
	catalogName                               = "db"
	l                                         = logger.DefaultSLogger("redis")
	_                    inputs.ElectionInput = (*Input)(nil)
)

type redislog struct {
	Files             []string `toml:"files"`
	Pipeline          string   `toml:"pipeline"`
	IgnoreStatus      []string `toml:"ignore"`
	CharacterEncoding string   `toml:"character_encoding"`
	MultilineMatch    string   `toml:"multiline_match"`

	MatchDeprecated string `toml:"match,omitempty"`
}

type Input struct {
	Username       string `toml:"username"`
	Host           string `toml:"host"`
	UnixSocketPath string `toml:"unix_socket_path"`
	Password       string `toml:"password"`
	Timeout        string `toml:"connect_timeout"`
	Service        string `toml:"service"`
	Addr           string `toml:"-"`
	Port           int    `toml:"port"`
	*dknet.TLSClientConfig
	TLSOpen            bool   `toml:"tls_open"`             // Deprecated
	CacertFile         string `toml:"tls_ca"`               // Deprecated (use TLSConf.CaCerts)
	CertFile           string `toml:"tls_cert"`             // Deprecated (use TLSConf.Cert)
	KeyFile            string `toml:"tls_key"`              // Deprecated (use TLSConf.CertKey)
	InsecureSkipVerify bool   `toml:"insecure_skip_verify"` // Deprecated (use TLSConf.InsecureSkipVerify)

	DB                 int           `toml:"db"`
	SocketTimeout      int           `toml:"socket_timeout"`
	SlowlogMaxLen      int           `toml:"slowlog-max-len"`
	Interval           time.Duration `toml:"interval"`
	WarnOnMissingKeys  bool          `toml:"warn_on_missing_keys"`
	CommandStats       bool          `toml:"command_stats"`
	LatencyPercentiles bool          `toml:"latency_percentiles"`
	Slowlog            bool          `toml:"slow_log"`
	AllSlowLog         bool          `toml:"all_slow_log"`
	RedisCliPath       string        `toml:"redis_cli_path"`
	Hotkey             bool          `toml:"hotkey"`
	BigKey             bool          `toml:"bigkey"`
	KeyInterval        time.Duration `toml:"key_interval"`
	KeyTimeout         time.Duration `toml:"key_timeout"`
	KeyScanSleep       string        `toml:"key_scan_sleep"`

	Version            string
	Uptime             int
	CollectCoStatus    string
	CollectCoErrMsg    string
	LastCustomerObject *customerObjectMeasurement

	Tags map[string]string `toml:"tags"`
	Keys []string          `toml:"keys"`
	DBS  []int             `toml:"dbs"`
	Log  *redislog         `toml:"log"`

	MatchDeprecated   string   `toml:"match,omitempty"`
	ServersDeprecated []string `toml:"servers,omitempty"`

	UpState int

	timeoutDuration time.Duration
	keyDBS          []int

	tail       *tailer.Tailer
	start      time.Time
	collectors []func() ([]*point.Point, error)

	client        *redis.Client
	isInitialized bool

	Election        bool `toml:"election"`
	pause           bool
	pauseCh         chan bool
	hashMap         [][16]byte
	latencyLastTime map[string]time.Time
	// a pointer, set value in m.lastCollect.*=...
	cpuUsage redisCPUUsage

	semStop *cliutils.Sem // start stop signal

	startUpUnix int64
	feeder      dkio.Feeder
	mergedTags  map[string]string
	tagger      datakit.GlobalTagger
	alignTS     int64
}

type redisCPUUsage struct {
	usedCPUSys    float64
	usedCPUSysTS  time.Time
	usedCPUUser   float64
	usedCPUUserTS time.Time
}

func (ipt *Input) ElectionEnabled() bool {
	return ipt.Election
}

func (ipt *Input) initCfg() error {
	var err error
	ipt.timeoutDuration, err = time.ParseDuration(ipt.Timeout)
	if err != nil {
		ipt.timeoutDuration = 10 * time.Second
	}

	ipt.Addr = fmt.Sprintf("%s:%d", ipt.Host, ipt.Port)

	ipt.TLSClientConfig = dknet.MergeTLSConfig(
		ipt.TLSClientConfig,
		[]string{ipt.CacertFile},
		ipt.CertFile,
		ipt.KeyFile,
		ipt.TLSOpen,
		ipt.InsecureSkipVerify,
	)

	if ipt.RedisCliPath == "" {
		ipt.RedisCliPath = "redis-cli"
	}

	tlsCfg, err := ipt.TLSClientConfig.TLSConfigWithBase64()
	if err != nil {
		return err
	}

	client := redis.NewClient(&redis.Options{
		Addr:      ipt.Addr,
		TLSConfig: tlsCfg,
		Username:  ipt.Username,
		Password:  ipt.Password, // no password set
		DB:        ipt.DB,       // use default DB
	})

	if ipt.SlowlogMaxLen == 0 {
		ipt.SlowlogMaxLen = 128
	}

	// ping (todo)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err = client.Ping(ctx).Result()
	if err != nil {
		_ = client.Close()
		return err
	}

	ipt.client = client

	ipt.mergedTags["server"] = ipt.Addr
	ipt.mergedTags["service_name"] = ipt.Service

	return nil
}

func (ipt *Input) tryInit() {
	if ipt.isInitialized {
		return
	}
	if err := ipt.initCfg(); err != nil {
		ipt.FeedCoErr(err)
		l.Errorf("initCfg error: %v", err)
		ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(inputName),
		)
		return
	}
	ipt.isInitialized = true
	ipt.hashMap = make([][16]byte, ipt.SlowlogMaxLen)
	ipt.collectors = []func() ([]*point.Point, error){
		ipt.collectInfoMeasurement,
		ipt.collectClientMeasurement,
		ipt.collectCommandMeasurement,
		ipt.collectDBMeasurement,
		ipt.collectReplicaMeasurement,
		ipt.collectCustomerObjectMeasurement,
	}
	// check if cluster
	ctx := context.Background()
	list1 := ipt.client.Do(ctx, "info", "cluster").String()
	part := strings.Split(list1, ":")
	if len(part) >= 3 {
		if strings.Compare(part[2], "1") == 1 {
			ipt.collectors = append(ipt.collectors, ipt.CollectClusterMeasurement)
		}
	}
}

func (*Input) PipelineConfig() map[string]string {
	pipelineMap := map[string]string{
		inputName: pipelineCfg,
	}
	return pipelineMap
}

func (*Input) LogExamples() map[string]map[string]string {
	return map[string]map[string]string{
		inputName: {
			"Redis log": `122:M 14 May 2019 19:11:40.164 * Background saving terminated with success`,
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

func (ipt *Input) Collect() error {
	if !ipt.isInitialized {
		return fmt.Errorf("un initialized")
	}

	for idx, f := range ipt.collectors {
		pts, err := f()
		if err != nil {
			l.Errorf("collector %v[%d]: %s", f, idx, err)
			ipt.feeder.FeedLastError(err.Error(),
				metrics.WithLastErrorInput(inputName),
			)
		}

		if len(pts) > 0 {
			if pts[0].Name() == "database" {
				if err := ipt.feeder.FeedV2(point.CustomObject, pts,
					dkio.WithCollectCost(time.Since(ipt.start)),
					dkio.WithElection(ipt.Election),
					dkio.WithInputName(customObjectFeedName)); err != nil {
					ipt.feeder.FeedLastError(err.Error(),
						metrics.WithLastErrorInput(inputName),
						metrics.WithLastErrorCategory(point.CustomObject),
					)
					l.Errorf("feed measurement: %s", err)
				}
			} else {
				if err := ipt.feeder.FeedV2(point.Metric, pts,
					dkio.WithCollectCost(time.Since(ipt.start)),
					dkio.WithElection(ipt.Election),
					dkio.WithInputName(inputName)); err != nil {
					ipt.feeder.FeedLastError(err.Error(),
						metrics.WithLastErrorInput(inputName),
						metrics.WithLastErrorCategory(point.Metric),
					)
					l.Errorf("feed measurement: %s", err)
				}
			}
		}
	}
	if ipt.Slowlog {
		if err := ipt.getSlowData(); err != nil {
			return err
		}
	}

	// Old way get big key
	if len(ipt.Keys) > 0 {
		err := ipt.collectBigKey()
		if err != nil {
			return err
		}
	}

	if err := ipt.getLatencyData(); err != nil {
		return err
	}

	return nil
}

func (ipt *Input) collectInfoMeasurement() ([]*point.Point, error) {
	m := &infoMeasurement{
		cli:                ipt.client,
		tags:               make(map[string]string),
		lastCollect:        &ipt.cpuUsage,
		latencyPercentiles: ipt.LatencyPercentiles,
	}

	m.name = redisInfoM

	m.tags = map[string]string{}
	for k, v := range ipt.mergedTags {
		m.tags[k] = v
	}

	return m.getData(ipt.alignTS)
}

func (ipt *Input) collectClientMeasurement() ([]*point.Point, error) {
	ctx := context.Background()
	list, err := ipt.client.ClientList(ctx).Result()
	if err != nil {
		l.Error("client list get error,", err)
		return nil, err
	}

	return ipt.parseClientData(list)
}

func (ipt *Input) collectCommandMeasurement() ([]*point.Point, error) {
	ctx := context.Background()
	list, err := ipt.client.Info(ctx, "commandstats").Result()
	if err != nil {
		l.Error("command stats error,", err)
		return nil, err
	}

	return ipt.parseCommandData(list)
}

func (ipt *Input) collectCustomerObjectMeasurement() ([]*point.Point, error) {
	err := ipt.getVersionAndUptime()
	if err != nil {
		l.Errorf("getVersionAndUptime err: %s", err)
	}
	ipt.setIptCOStatus()
	ms := []inputs.MeasurementV2{}
	fields := map[string]interface{}{
		"display_name": fmt.Sprintf("%s:%d", ipt.Host, ipt.Port),
		"uptime":       ipt.Uptime,
		"version":      ipt.Version,
	}
	tags := map[string]string{
		"name":          fmt.Sprintf("redis-%s:%d", ipt.Host, ipt.Port),
		"host":          ipt.Host,
		"ip":            fmt.Sprintf("%s:%d", ipt.Host, ipt.Port),
		"col_co_status": ipt.CollectCoStatus,
	}
	m := &customerObjectMeasurement{
		name:     "database",
		tags:     tags,
		fields:   fields,
		election: ipt.Election,
	}
	ipt.setIptLastCOInfo(m)
	ms = append(ms, m)
	if len(ms) > 0 {
		pts := getPointsFromMeasurement(ms)
		return pts, nil
	}
	return []*point.Point{}, nil
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
		tailer.WithMultilinePatterns([]string{ipt.Log.MultilineMatch}),
		tailer.WithGlobalTags(inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, "")),
		tailer.EnableDebugFields(config.Cfg.EnableDebugFields),
	}

	var err error
	ipt.tail, err = tailer.NewTailer(ipt.Log.Files, opts...)
	if err != nil {
		l.Error("NewTailer: %s", err)

		ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(inputName),
		)
		return
	}

	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_redis"})
	g.Go(func(ctx context.Context) error {
		ipt.tail.Start()
		return nil
	})
}

func (ipt *Input) Run() {
	ipt.setup()

	if ipt.TLSClientConfig != nil && (ipt.TLSClientConfig.CertBase64 != "" || ipt.TLSClientConfig.CertKeyBase64 != "") {
		caCerts, cert, certKey, err := ipt.TLSClientConfig.Base64ToTLSFiles()
		if err != nil {
			l.Errorf("Collect: %s", err)
			return
		}

		ipt.TLSClientConfig.CaCerts = caCerts
		ipt.TLSClientConfig.Cert = cert
		ipt.TLSClientConfig.CertKey = certKey

		for _, caCert := range caCerts {
			defer os.Remove(caCert) // nolint:errcheck
		}
		defer os.Remove(cert)    // nolint:errcheck
		defer os.Remove(certKey) // nolint:errcheck
	}

	tick := time.NewTicker(ipt.Interval)
	defer tick.Stop()

	ctxKey, cancelKey := context.WithCancel(context.Background())
	defer cancelKey() // To kill all in goroutineHotkey & goroutineBigKey
	if ipt.Hotkey {
		ipt.goroutineHotkey(ctxKey)
	}
	if ipt.BigKey {
		ipt.goroutineBigKey(ctxKey)
	}

	lastTS := time.Now()
	for {
		ipt.alignTS = lastTS.UnixNano()

		if !ipt.pause {
			ipt.tryInit()

			ipt.setUpState()

			l.Debugf("redis input gathering...")
			ipt.start = time.Now()
			if err := ipt.Collect(); err != nil {
				l.Errorf("Collect: %s", err)
				ipt.setErrUpState()
			}
			ipt.FeedUpMetric()
		} else {
			l.Debugf("not leader, skipped")
		}

		select {
		case <-datakit.Exit.Wait():
			ipt.exit()
			l.Info("redis exit")
			return

		case <-ipt.semStop.Wait():
			ipt.exit()
			l.Info("redis return")
			return

		case tt := <-tick.C:
			nextts := inputs.AlignTimeMillSec(tt, lastTS.UnixMilli(), ipt.Interval.Milliseconds())
			lastTS = time.UnixMilli(nextts)
		case ipt.pause = <-ipt.pauseCh:
			// nil
		}
	}
}

func (ipt *Input) setup() {
	ipt.startUpUnix = time.Now().Unix()

	l = logger.SLogger(inputName)
	l.Infof("%s input started", inputName)
	ipt.Interval = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval)
	ipt.KeyInterval = config.ProtectedInterval(minInterval, maxInterval, ipt.KeyInterval)
	ipt.KeyTimeout = config.ProtectedInterval(minInterval, ipt.KeyInterval, ipt.KeyTimeout)
	if ipt.Election {
		ipt.mergedTags = inputs.MergeTags(ipt.tagger.ElectionTags(), ipt.Tags, ipt.Host)
	} else {
		ipt.mergedTags = inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, ipt.Host)
	}
	l.Debugf("merged tags: %+#v", ipt.mergedTags)

	// config have "DB"，join DBS
	if ipt.DB != -1 && !IsSlicesHave(ipt.DBS, ipt.DB) {
		ipt.DBS = append(ipt.DBS, ipt.DB)
	}
	if len(ipt.DBS) < 1 && (ipt.Hotkey || ipt.BigKey) {
		l.Errorf("dbs is nil in redis.conf, example: dbs=[0]")
	}

	ipt.keyDBS = append(ipt.keyDBS, ipt.DBS...)
}

func (ipt *Input) exit() {
	if ipt.tail != nil {
		ipt.tail.Close()
		l.Info("redis log exit")
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
		&bigKeyMeasurement{},
		&hotkeyMeasurement{},
		&clientMeasurement{},
		&clusterMeasurement{},
		&commandMeasurement{},
		&dbMeasurement{},
		&infoMeasurement{},
		&replicaMeasurement{},
		&latencyMeasurement{},
		&slowlogMeasurement{},
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
	getClientFieldMap()
	getInfoFieldMap()
	getReplicaFieldMap()
	getClusterFieldMap()

	return &Input{
		Timeout:         "10s",
		pauseCh:         make(chan bool, inputs.ElectionPauseChannelLength),
		DB:              -1,
		Tags:            make(map[string]string),
		KeyInterval:     defaultKeyInterval,
		KeyTimeout:      defaultKeyInterval,
		KeyScanSleep:    defaultKeyScanSleep,
		latencyLastTime: map[string]time.Time{},
		semStop:         cliutils.NewSem(),
		Election:        true,
		feeder:          dkio.DefaultFeeder(),
		tagger:          datakit.DefaultGlobalTagger(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}

func getPointsFromMeasurement(ms []inputs.MeasurementV2) []*point.Point {
	pts := []*point.Point{}
	for _, m := range ms {
		pts = append(pts, m.Point())
	}

	return pts
}

func parseVersion(info string) string {
	for _, line := range strings.Split(info, "\n") {
		if strings.HasPrefix(line, "redis_version:") {
			return strings.TrimSpace(strings.Split(line, ":")[1])
		}
	}
	return ""
}

func parseUptime(info string) int {
	for _, line := range strings.Split(info, "\n") {
		if strings.HasPrefix(line, "uptime_in_seconds:") {
			uptimeStr := strings.TrimSpace(strings.Split(line, ":")[1])
			uptime, err := strconv.Atoi(uptimeStr)
			if err != nil {
				l.Error("failed to parse uptime: %v", err)
				return 0
			}
			return uptime
		}
	}
	return 0
}
