package config

import (
	"io/ioutil"

	"github.com/go-redis/redis"
	"go.uber.org/zap"
	yaml "gopkg.in/yaml.v2"
)

var (
	C = Config{
		Influx: InfluxCfg{
			ReadTimeOut:  30,
			WriteTimeOut: 30,
		},

		Database: DatabaseCfg{
			Dialect: `mysql`,
		},

		Redis: RedisCfg{
			Host: `cloudcare-kodo-redis.db.svc.cluster.local:6379`,
			Db:   0,
		},

		LogConfig: LogCfg{
			LogFile:    `/logdata/log`,
			Level:      `info`,
			JSONFormat: true,
			GinLogFile: `/logdata/gin.log`,
		},

		NSQ: NSQCfg{
			Lookupd:   `http://nsqlookupd.middleware.svc.cluster.local:4161`,
			RefreshAt: 300,
		},

		Stat: StatCfg{
			SlowWrite: 3.0, // second
			SlowQuery: 1.0,
		},

		Global: GlobalCfg{
			EnableInnerApi: true,
			StatsOn:        256,
			Listen:         `:9527`,
			Workers:        8,
			LogWorkers:     8,
			Dataway:        "http://internal-dataway.utils:9528",
			EsConsumer:     false,
			RetryTimes:     3600 * 24,
		},

		Ck: CkCfg{
			ReadTimeOut:  60,
			WriteTimeOut: 30,
			Host:         ``,
			User:         ``,
			Passwd:       ``,
			ClusterName:  `default`,
		},

		Es: EsCfg{
			Host:   ``,
			User:   ``,
			Passwd: ``,
		},

		RpConfig: map[string][2]string{
			`rp6`:      [2]string{`25920h`, `720h`}, // 3 year
			`rp5`:      [2]string{`8640h`, `720h`},  // 1 year
			`rp4`:      [2]string{`4320h`, `720h`},  // 6 month
			`rp3`:      [2]string{`2160h`, `1w`},    // 3 month
			`rp2`:      [2]string{`720h`, `1w`},     // 1 month
			`rp1`:      [2]string{`168h`, `1d`},     // 1 week
			`rp0`:      [2]string{`24h`, `6h`},      // 1 day
			`rp_2160h`: [2]string{`2160h`, `1w`},
			`autogen`:  [2]string{`25920h`, `720h`},
		},
		Ws: WsConfig{
			Path: "/v1/datakit/ws",
		},
	}

	Redis *redis.Client
)

type DatabaseCfg struct {
	Dialect    string `yaml:"db_dialect"`
	Connection string `yaml:"connection"`
}

type NSQCfg struct {
	Lookupd   string `yaml:"lookupd"`
	RefreshAt int    `yaml:"refresh_at"`
}

type RedisCfg struct {
	Host string `yaml:"host"`
	Pass string `yaml:"password"`
	Db   int    `yaml:"db"`
}

type InfluxCfg struct {
	ReadTimeOut  int64  `yaml:"read_timeout"`
	WriteTimeOut int64  `yaml:"write_timeout"`
	DefaultRP    string `yaml:"default_rp"`
	UserAgent    string `yaml:"user_agent"`
	EnableGrant  bool   `yaml:"enable_grant"`
	EnableGZ     bool   `yaml:"enable_gz"`

	DisableWrite bool `yaml:"disable_write"` // for test
}

type LogCfg struct {
	LogFile    string `yaml:"log_file"`
	Level      string `yaml:"level"`
	JSONFormat bool   `yaml:"json_format"`
	ShortFile  bool   `yaml:"short_file"`
	GinLogFile string `yaml:"gin_log_file"`
	Rl         *zap.Logger
}

type StatCfg struct {
	SlowWrite float64 `yaml:"slow_write"`
	SlowQuery float64 `yaml:"slow_query"`
	RP        string  `yaml:"rp"`
}

type GlobalCfg struct {
	EnableInnerApi bool   `yaml:"enable_inner_api"`
	StatsOn        int    `yaml:"stats_on"`
	Listen         string `yaml:"listen"`
	Workers        int    `yaml:"workers"`
	LogWorkers     int    `yaml:"log_workers"`
	Dataway        string `yaml:"dataway"`
	EsConsumer     bool   `yaml:"es_consumer"`
	RetryTimes     int64  `yaml:"retry_time_seconds"`

	// each license should only used on 1 dataway, if any dataway mis-configured
	// license(legal) used on other dataway, kodo will refuse it's request.
	EnableLicenseDataWayBinding bool `yaml:"enable_license_dataWay_binding"`
}

type CkCfg struct {
	ReadTimeOut  int64 `yaml:"read_timeout"`
	WriteTimeOut int64 `yaml:"write_timeout"`

	Host        string `yaml:"host"`
	User        string `yaml:"user"`
	Passwd      string `yaml:"password"`
	ClusterName string `yaml:"cluster_name"`
}

type SecretCfg struct {
	EncryptKey string `yaml:"encrypt_key"`
}

type Config struct {
	Influx    InfluxCfg            `yaml:"influxdb"`
	Database  DatabaseCfg          `yaml:"database"`
	Redis     RedisCfg             `yaml:"redis"`
	LogConfig LogCfg               `yaml:"log"`
	RpConfig  map[string][2]string `yaml:"global_rp"`
	Func      FuncCfg              `yaml:"func"`
	NSQ       NSQCfg               `yaml:"nsq"`
	Global    GlobalCfg            `yaml:"global"`
	Ck        CkCfg                `yaml:"ck"`
	Secret    SecretCfg            `yaml:"secret"`
	Stat      StatCfg              `yaml:"stat"`
	Es        EsCfg                `yaml:"es"`
	Ws        WsConfig             `yaml:"ws_server"`
}

type EsCfg struct {
	Host   string `yaml:"host"`
	User   string `yaml:"user"`
	Passwd string `yaml:"password"`
	Enable bool   `yaml:"enable"`
}

type FuncCfg struct {
	Host   string `yaml:"host"`
	Enable bool   `yaml:"enable"`
}

type WsConfig struct {
	Bind string      `yaml:"bind"`
	Path string		 `yaml:"path"`
	TimeOut string   `yaml:"time_out"`
}

func DumpConfig(cfg *Config, f string) error {
	c, err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(f, c, 0644)
}

func LoadConfig(f string) error {
	data, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, &C); err != nil {
		return err
	}

	return nil
}
