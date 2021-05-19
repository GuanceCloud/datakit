package redis

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/go-redis/redis"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	maxInterval = 30 * time.Minute
	minInterval = 15 * time.Second
)

var (
	inputName   = "redis"
	catalogName = "db"
	l           = logger.DefaultSLogger("redis")
)

type Input struct {
	Host              string `toml:"host"`
	Port              int    `toml:"port"`
	UnixSocketPath    string `toml:"unix_socket_path"`
	DB                int    `toml:"db"`
	Password          string `toml:"password"`
	Service           string `toml:"service"`
	SocketTimeout     int    `toml:"socket_timeout"`
	Interval          datakit.Duration
	Keys              []string             `toml:"keys"`
	WarnOnMissingKeys bool                 `toml:"warn_on_missing_keys"`
	CommandStats      bool                 `toml:"command_stats"`
	Slowlog           bool                 `toml:"slow_log"`
	SlowlogMaxLen     int                  `toml:"slowlog-max-len"`
	Tags              map[string]string    `toml:"tags"`
	client            *redis.Client        `toml:"-"`
	collectCache      []inputs.Measurement `toml:"-"`
	Addr              string               `toml:"-"`
	Log               *inputs.TailerOption `toml:"log"`
	tailer            *inputs.Tailer       `toml:"-"`
	resKeys           []string             `toml:"-"`
	err               error
}

func (i *Input) initCfg() {
	i.Addr = fmt.Sprintf("%s:%d", i.Host, i.Port)
	client := redis.NewClient(&redis.Options{
		Addr:     i.Addr,
		Password: i.Password, // no password set
		DB:       i.DB,       // use default DB
	})

	i.client = client

	i.globalTag()
}

func (i *Input) globalTag() {
	i.Tags["server"] = i.Addr
	i.Tags["service_name"] = i.Service
}

func (i *Input) PipelineConfig() map[string]string {
	pipelineMap := map[string]string{
		"redis": pipelineCfg,
	}
	return pipelineMap
}

func (i *Input) Collect() error {
	i.collectCache = []inputs.Measurement{}

	i.collectInfoMeasurement()

	// 获取客户端信息
	i.collectClientMeasurement()

	// db command
	if i.CommandStats {
		i.collectCommandMeasurement()
	}

	// slowlog
	if i.Slowlog {
		slowlogMeasurement := CollectSlowlogMeasurement(i)
		i.collectCache = append(i.collectCache, slowlogMeasurement)
	}

	// bigkey
	if len(i.Keys) > 0 {
		i.collectBigKeyMeasurement()
	}

	if i.err != nil {
		io.FeedLastError(inputName, i.err.Error())
		i.err = nil
	}

	return nil
}

func (i *Input) collectInfoMeasurement() {
	m := &infoMeasurement{
		i:       i,
		resData: make(map[string]interface{}),
		tags:    make(map[string]string),
		fields:  make(map[string]interface{}),
	}

	m.name = "redis_info"
	for key, value := range i.Tags {
		m.tags[key] = value
	}

	if err := m.getData(); err != nil {
		m.i.err = err
	}
	m.submit()

	i.collectCache = append(i.collectCache, m)
}

func (i *Input) collectClientMeasurement() {
	if err := i.getClientData(); err != nil {
		i.err = err
	}
}

func (i *Input) collectBigKeyMeasurement() {
	i.getKeys()
	i.getData()
}

func (i *Input) collectCommandMeasurement() {
	if err := i.getCommandData(); err != nil {
		i.err = err
	}
}

func (i *Input) runLog(defaultPile string) {
	if i.Log != nil {
		go func() {
			pfile := defaultPile
			if i.Log.Pipeline != "" {
				pfile = i.Log.Pipeline
			}

			i.Log.Service = i.Service
			i.Log.Pipeline = filepath.Join(datakit.PipelineDir, pfile)

			i.Log.Source = inputName
			i.Log.Tags = make(map[string]string)
			for k, v := range i.Tags {
				i.Log.Tags[k] = v
			}
			tailer, err := inputs.NewTailer(i.Log)
			if err != nil {
				i.err = err
				l.Errorf("init tailf err:%s", err.Error())
				return
			}
			i.tailer = tailer
			tailer.Run()
		}()
	}
}

func (i *Input) Run() {
	l = logger.SLogger("redis")

	i.Interval.Duration = datakit.ProtectedInterval(minInterval, maxInterval, i.Interval.Duration)

	i.initCfg()

	i.runLog("redis.p")

	tick := time.NewTicker(i.Interval.Duration)
	defer tick.Stop()

	n := 0

	for {
		n++
		select {
		case <-tick.C:
			l.Debugf("redis input gathering...")
			start := time.Now()
			if err := i.Collect(); err != nil {
				io.FeedLastError(inputName, err.Error())
			} else {
				if err := inputs.FeedMeasurement(inputName, datakit.Metric, i.collectCache,
					&io.Option{CollectCost: time.Since(start), HighFreq: (n%2 == 0)}); err != nil {
					io.FeedLastError(inputName, err.Error())
				}

				i.collectCache = i.collectCache[:] // NOTE: do not forget to clean cache
			}

		case <-datakit.Exit.Wait():
			if i.tailer != nil {
				i.tailer.Close()
				l.Info("redis log exit")
			}
			l.Info("redis exit")
			return
		}
	}
}

func (i *Input) Catalog() string { return catalogName }

func (i *Input) SampleConfig() string { return configSample }

func (i *Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&infoMeasurement{},
		&clientMeasurement{},
		&commandMeasurement{},
		&slowlogMeasurement{},
		&bigKeyMeasurement{},
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
