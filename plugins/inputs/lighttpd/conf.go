package lighttpd

import "time"

const (
	pluginName = "lighttpd"

	lighttpdConfigSample = `
# [lighttpd]
# [[lighttpd.subscribes]]
#       ## lighttpd status url
#	url = "http://127.0.0.1:8080/server-status"
#
#       ## 指定 lighttpd 版本为 "v1" 或 "v2"
#       version = "v1"
#
#	## 采集周期，时间单位是秒
#	collect_cycle = 60
#
#       ## measurement，不可重复
#       measurement = "lighttpd"
`
)

type Subscribe struct {
	LighttpdURL     string        `toml:"url"`
	LighttpdVersion string        `toml:"version"`
	Cycle           time.Duration `toml:"collect_cycle"`
	Measurement     string        `toml:"measurement"`
}

type Config struct {
	Subscribes []Subscribe `toml:"subscribes"`
}
