package coredns

import "time"

const (
	pluginName = "coredns"

	corednsConfigSample = `
# [coredns]
# [[coredns.subscribes]]
#       ## coredns 地址
#	host = "127.0.0.1"
#
#       ## coredns prometheus 监控端口
#	port = "9153"
#
#	## 采集周期，时间单位是秒
#	collect_cycle = 60
#
#       ## measurement，不可重复
#       measurement = "coredns"
`
)

type Subscribe struct {
	CorednsHost string        `toml:"host"`
	CorednsPort int           `toml:"port"`
	Cycle       time.Duration `toml:"collect_cycle"`
	Measurement string        `toml:"measurement"`
}

type Config struct {
	Subscribes []Subscribe `toml:"subscribes"`
}
