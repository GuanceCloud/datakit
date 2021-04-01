package redis

import (
	"time"

	"github.com/go-redis/redis"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName   = "redis"
	catalogName = "db"
	l           = logger.DefaultSLogger("redis")
)

type Input struct {
	Host              string		    `toml:"host"`
	Port              int			    `toml:"port"`
	UnixSocketPath    string            `toml:"unix_socket_path"`
	DB                int			    `toml:"db"`
	Password          string		    `toml:"password"`
	Service           string		    `toml:"service"`
	SocketTimeout     int               `toml:"socket_timeout"`
	Interval          string            `toml:"interval"`
	IntervalDuration  time.Duration     `toml:"-"`
	Keys              []string		    `toml:"keys"`
	WarnOnMissingKeys bool              `toml:"warn_on_missing_keys"`
	SlowlogMaxLen     float64           `toml:"slowlog-max-len"`
	Tags              map[string]string `toml:"tags"`
	client           *redis.Client      `toml:"-"`
	collectCache []inputs.Measurement   `toml:"-"`
}

func (i *Input) Init() error {
	client := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "dev", // no password set
        DB:       0,  // use default DB
    })

    i.client = client

    return nil
}

func (i *Input) Collect() error {
	i.collectCache = []inputs.Measurement{}

	// demo
	demoMeasurement :=CollectDemoMeasurement()
	i.collectCache = append(i.collectCache, demoMeasurement)

	// 获取info指标
	infoMeasurement := CollectInfoMeasurement(i.client)
	i.collectCache = append(i.collectCache, infoMeasurement)

	return nil
}

func (i *Input) Run() {

	l = logger.SLogger("redis")
	tick := time.NewTicker(time.Second * 3)
	defer tick.Stop()

	i.Init()

	n := 0

	for {

		n++
		select {
		case <-tick.C:
			l.Debugf("redis input gathering...")
			start := time.Now()
			if err := i.Collect(); err != nil {
				l.Error(err)
			} else {
				inputs.FeedMeasurement(inputName, io.Metric, i.collectCache,
					&io.Option{CollectCost: time.Since(start), HighFreq: (n%2 == 0)})

				i.collectCache = i.collectCache[:] // NOTE: do not forget to clean cache
			}

		case <-datakit.Exit.Wait():
			return
		}
	}
}

func (i *Input) Catalog() string      { return catalogName }

func (i *Input) SampleConfig() string { return "[inputs.demo]" }

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&demoMeasurement{},
		&infoMeasurement{},
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
