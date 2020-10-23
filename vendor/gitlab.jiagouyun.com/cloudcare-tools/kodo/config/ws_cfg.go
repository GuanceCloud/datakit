package config

import (
	"io/ioutil"

	"github.com/go-redis/redis"
	"gopkg.in/yaml.v2"
)

var (
	Wscfg = WsConfig{

	RdbConfig: WsRedis{
		Db: 0,
		RedisChann: "datakit-chann",
	},
	MysqlConfig: WsMysql{
		Dialect: "mysql",

	},
	Server: WsServer{
		Path: "/v1/datakit/ws",
	},
	Log: LogConfig{
		LogPath: "/logdata/kodows.log",
		Level: "info",
	},

}
	RedisCli *redis.Client

)


type WsConfig struct {
	RdbConfig WsRedis `yaml:"redis"`
	MysqlConfig WsMysql `yaml:"database"`
	Server WsServer `yaml:"server"`
	Log LogConfig `yaml:"log"`

}

type WsRedis struct {
	Host string `yaml:"host"`
	Password string `yaml:"password"`
	Db   int    `yaml:"db"`
	RedisChann string `yaml:"chan"`
}

type WsMysql struct {
	Dialect    string `yaml:"db_dialect"`
	Connection string `yaml:"connection"`

}

type WsServer struct {
	Bind string `yaml:"bind"`
	Path string `yaml:"path"`
}

type LogConfig struct {
	LogPath string `yaml:"path"`
	Level string `yaml:"level"`
}

func LoadWsCfg(f string) error  {
	data, err := ioutil.ReadFile(f)
	if err != nil {
	return err
	}

	if err := yaml.Unmarshal(data, &Wscfg); err != nil {
	return err
	}
	return nil
}
