// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package redis collects redis metrics.
package redis

import (
	"context"
	"fmt"
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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
)

const (
	maxInterval = 30 * time.Minute
	minInterval = 15 * time.Second
)

var (
	inputName                        = "redis"
	catalogName                      = "db"
	l                                = logger.DefaultSLogger("redis")
	_           inputs.ElectionInput = (*Input)(nil)
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
	Username           string `toml:"username"`
	Host               string `toml:"host"`
	UnixSocketPath     string `toml:"unix_socket_path"`
	Password           string `toml:"password"`
	Timeout            string `toml:"connect_timeout"`
	Service            string `toml:"service"`
	Addr               string `toml:"-"`
	Port               int    `toml:"port"`
	DB                 int    `toml:"db"`
	SocketTimeout      int    `toml:"socket_timeout"`
	SlowlogMaxLen      int    `toml:"slowlog-max-len"`
	Interval           time.Duration
	WarnOnMissingKeys  bool              `toml:"warn_on_missing_keys"`
	CommandStats       bool              `toml:"command_stats"`
	LatencyPercentiles bool              `toml:"latency_percentiles"`
	Slowlog            bool              `toml:"slow_log"`
	AllSlowLog         bool              `toml:"all_slow_log"`
	Tags               map[string]string `toml:"tags"`
	Keys               []string          `toml:"keys"`
	DBS                []int             `toml:"dbs"`
	Log                *redislog         `toml:"log"`

	MatchDeprecated   string   `toml:"match,omitempty"`
	ServersDeprecated []string `toml:"servers,omitempty"`

	timeoutDuration time.Duration

	tail       *tailer.Tailer
	start      time.Time
	collectors []func() ([]*point.Point, error)

	client *redis.Client

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

	client := redis.NewClient(&redis.Options{
		Addr:     ipt.Addr,
		Username: ipt.Username,
		Password: ipt.Password, // no password set
		DB:       ipt.DB,       // use default DB
	})

	if ipt.SlowlogMaxLen == 0 {
		ipt.SlowlogMaxLen = 128
	}

	ipt.client = client

	// ping (todo)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err = client.Ping(ctx).Result()

	if err != nil {
		return err
	}

	ipt.mergedTags["server"] = ipt.Addr
	ipt.mergedTags["service_name"] = ipt.Service

	return nil
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

func (ipt *Input) Collect() error {
	for idx, f := range ipt.collectors {
		pts, err := f()
		if err != nil {
			l.Errorf("collector %v[%d]: %s", f, idx, err)
			ipt.feeder.FeedLastError(err.Error(),
				dkio.WithLastErrorInput(inputName),
			)
		}

		if len(pts) > 0 {
			if err := ipt.feeder.Feed(inputName, point.Metric, pts,
				&dkio.Option{CollectCost: time.Since(ipt.start)}); err != nil {
				l.Errorf("FeedMeasurement: %s", err)
			}
		}
	}
	if ipt.Slowlog {
		if err := ipt.getSlowData(); err != nil {
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

	return m.getData()
}

func (ipt *Input) collectBigKeyMeasurement() ([]*point.Point, error) {
	keys, err := ipt.getKeys()
	if err != nil {
		return nil, err
	}

	return ipt.getData(keys)
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
		l.Error("NewTailer: %s", err)

		ipt.feeder.FeedLastError(err.Error(),
			dkio.WithLastErrorInput(inputName),
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

	tick := time.NewTicker(ipt.Interval)
	defer tick.Stop()

	// Try init until ok.
	for {
		select {
		case <-datakit.Exit.Wait():
			return
		case <-ipt.semStop.Wait():
			return
		case <-tick.C:
		}

		if err := ipt.initCfg(); err != nil {
			ipt.feeder.FeedLastError(err.Error(),
				dkio.WithLastErrorInput(inputName),
			)
		} else {
			break
		}
	}
	ipt.hashMap = make([][16]byte, ipt.SlowlogMaxLen)

	ipt.collectors = []func() ([]*point.Point, error){
		ipt.collectInfoMeasurement,
		ipt.collectClientMeasurement,
		ipt.collectCommandMeasurement,
		ipt.collectDBMeasurement,
		ipt.collectReplicaMeasurement,
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

	if len(ipt.Keys) > 0 {
		ipt.collectors = append(ipt.collectors, ipt.collectBigKeyMeasurement)
	}

	for {
		if !ipt.pause {
			l.Debugf("redis input gathering...")
			ipt.start = time.Now()
			if err := ipt.Collect(); err != nil {
				l.Errorf("Collect: %s", err)
			}
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

		case <-tick.C:
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
	if ipt.Election {
		ipt.mergedTags = inputs.MergeTags(ipt.tagger.ElectionTags(), ipt.Tags, ipt.Host)
	} else {
		ipt.mergedTags = inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, ipt.Host)
	}
	l.Debugf("merged tags: %+#v", ipt.mergedTags)
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
		&clientMeasurement{},
		&clusterMeasurement{},
		&commandMeasurement{},
		&dbMeasurement{},
		&infoMeasurement{},
		&latencyMeasurement{},
		&slowlogMeasurement{},
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
	getClusterFieldMap()

	return &Input{
		Timeout:         "10s",
		pauseCh:         make(chan bool, inputs.ElectionPauseChannelLength),
		DB:              -1,
		Tags:            make(map[string]string),
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
