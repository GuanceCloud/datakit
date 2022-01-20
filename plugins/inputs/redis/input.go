// Package redis collects redis metrics.
package redis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var _ inputs.ElectionInput = (*Input)(nil)

const (
	maxInterval = 30 * time.Minute
	minInterval = 15 * time.Second
)

var (
	inputName   = "redis"
	catalogName = "db"
	l           = logger.DefaultSLogger("redis")
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
	Username          string `toml:"username"`
	Host              string `toml:"host"`
	UnixSocketPath    string `toml:"unix_socket_path"`
	Password          string `toml:"password"`
	Timeout           string `toml:"connect_timeout"`
	Service           string `toml:"service"`
	Addr              string `toml:"-"`
	Port              int    `toml:"port"`
	DB                int    `toml:"db"`
	SocketTimeout     int    `toml:"socket_timeout"`
	SlowlogMaxLen     int    `toml:"slowlog-max-len"`
	Interval          datakit.Duration
	WarnOnMissingKeys bool              `toml:"warn_on_missing_keys"`
	CommandStats      bool              `toml:"command_stats"`
	Slowlog           bool              `toml:"slow_log"`
	Tags              map[string]string `toml:"tags"`
	Keys              []string          `toml:"keys"`
	DBS               []int             `toml:"dbs"`
	Log               *redislog         `toml:"log"`

	MatchDeprecated   string   `toml:"match,omitempty"`
	ServersDeprecated []string `toml:"servers,omitempty"`

	timeoutDuration time.Duration

	tail       *tailer.Tailer
	start      time.Time
	collectors []func() ([]inputs.Measurement, error)

	client *redis.Client

	pause   bool
	pauseCh chan bool

	semStop *cliutils.Sem // start stop signal
}

func (i *Input) initCfg() error {
	var err error
	i.timeoutDuration, err = time.ParseDuration(i.Timeout)
	if err != nil {
		i.timeoutDuration = 10 * time.Second
	}

	i.Addr = fmt.Sprintf("%s:%d", i.Host, i.Port)

	client := redis.NewClient(&redis.Options{
		Addr:     i.Addr,
		Username: i.Username,
		Password: i.Password, // no password set
		DB:       i.DB,       // use default DB
	})

	if i.SlowlogMaxLen == 0 {
		i.SlowlogMaxLen = 128
	}

	i.client = client

	// ping (todo)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err = client.Ping(ctx).Result()

	if err != nil {
		return err
	}

	i.Tags["server"] = i.Addr
	i.Tags["service_name"] = i.Service

	return nil
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

func (i *Input) Collect() error {
	for idx, f := range i.collectors {
		ms, err := f()
		if err != nil {
			l.Errorf("collector %v[%d]: %s", f, idx, err)
			io.FeedLastError(inputName, err.Error())
		}

		if len(ms) > 0 {
			if err := inputs.FeedMeasurement(inputName,
				datakit.Metric,
				ms,
				&io.Option{CollectCost: time.Since(i.start)}); err != nil {
				l.Errorf("FeedMeasurement: %s", err)
			}
		}
	}
	return nil
}

func (i *Input) collectInfoMeasurement() ([]inputs.Measurement, error) {
	var collectCache []inputs.Measurement

	m := &infoMeasurement{
		cli:     i.client,
		resData: make(map[string]interface{}),
		tags:    make(map[string]string),
		fields:  make(map[string]interface{}),
	}

	m.name = "redis_info"

	for key, value := range i.Tags {
		m.tags[key] = value
	}

	// get data
	if err := m.getData(); err != nil {
		return nil, err
	}

	// build line data
	if err := m.submit(); err != nil {
		return nil, err
	}

	if len(m.fields) > 0 {
		collectCache = append(collectCache, m)
	}

	return collectCache, nil
}

func (i *Input) collectBigKeyMeasurement() ([]inputs.Measurement, error) {
	keys, err := i.getKeys()
	if err != nil {
		return nil, err
	}

	return i.getData(keys)
}

// 数据源获取数据.
func (i *Input) collectClientMeasurement() ([]inputs.Measurement, error) {
	ctx := context.Background()
	list, err := i.client.ClientList(ctx).Result()
	if err != nil {
		l.Error("client list get error,", err)
		return nil, err
	}

	return i.parseClientData(list)
}

// 数据源获取数据.
func (i *Input) collectCommandMeasurement() ([]inputs.Measurement, error) {
	ctx := context.Background()
	list, err := i.client.Info(ctx, "commandstats").Result()
	if err != nil {
		l.Error("command stats error,", err)
		return nil, err
	}

	return i.parseCommandData(list)
}

func (i *Input) collectSlowlogMeasurement() ([]inputs.Measurement, error) {
	return i.getSlowData()
}

func (i *Input) RunPipeline() {
	if i.Log == nil || len(i.Log.Files) == 0 {
		return
	}

	if i.Log.Pipeline == "" {
		i.Log.Pipeline = "redis.p" // use default
	}

	opt := &tailer.Option{
		Source:            inputName,
		Service:           inputName,
		Pipeline:          i.Log.Pipeline,
		GlobalTags:        i.Tags,
		IgnoreStatus:      i.Log.IgnoreStatus,
		CharacterEncoding: i.Log.CharacterEncoding,
		MultilineMatch:    i.Log.MultilineMatch,
	}

	var err error
	i.tail, err = tailer.NewTailer(i.Log.Files, opt)
	if err != nil {
		l.Error("NewTailer: %s", err)

		io.FeedLastError(inputName, err.Error())
		return
	}

	go i.tail.Start()
}

func (i *Input) Run() {
	l = logger.SLogger("redis")
	io.FeedEventLog(&io.Reporter{Message: inputName + " start ok, ready for collecting metrics.", Logtype: "event"})

	i.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, i.Interval.Duration)

	tick := time.NewTicker(i.Interval.Duration)
	defer tick.Stop()

	// Try init until ok.
	for {
		select {
		case <-datakit.Exit.Wait():
			return
		case <-i.semStop.Wait():
			return
		case <-tick.C:
		}

		if err := i.initCfg(); err != nil {
			io.FeedLastError(inputName, err.Error())
		} else {
			break
		}
	}

	i.collectors = []func() ([]inputs.Measurement, error){
		i.collectInfoMeasurement,
		i.collectClientMeasurement,
		i.collectCommandMeasurement,
		i.collectSlowlogMeasurement,
		i.collectDBMeasurement,
		i.CollectLatencyMeasurement,
		i.collectReplicaMeasurement,
	}

	// 判断是否采集集群
	ctx := context.Background()
	list1 := i.client.Do(ctx, "info", "cluster").String()
	part := strings.Split(list1, ":")
	if len(part) >= 3 {
		if strings.Compare(part[2], "1") == 1 {
			i.collectors = append(i.collectors, i.CollectClusterMeasurement)
		}
	}

	if len(i.Keys) > 0 {
		i.collectors = append(i.collectors, i.collectBigKeyMeasurement)
	}

	for {
		if !i.pause {
			l.Debugf("redis input gathering...")
			i.start = time.Now()
			if err := i.Collect(); err != nil {
				l.Errorf("Collect: %s", err)
			}
		} else {
			l.Debugf("not leader, skipped")
		}

		select {
		case <-datakit.Exit.Wait():
			i.exit()
			l.Info("redis exit")
			return

		case <-i.semStop.Wait():
			i.exit()
			l.Info("redis return")
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
		l.Info("redis log exit")
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
			Timeout: "10s",
			pauseCh: make(chan bool, inputs.ElectionPauseChannelLength),
			DB:      -1,

			semStop: cliutils.NewSem(),
		}
	})
}
