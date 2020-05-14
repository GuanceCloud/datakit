// +build linux

package containerd

import "time"

const (
	pluginName = "containerd"

	containerdConfigSample = `
# [containerd]
# [[containerd.subscribes]]
#       ## host path
#       host_path = "/run/containerd/containerd.sock"
#
#       ## namespace
#       namespace = "moby"
#
#       ## collection ID list
#       ID_list = ["*"]
#
#	## 采集周期，时间单位是秒
#	collect_cycle = 60
#
#       ## measurement，不可重复
#       measurement = "containerd"
`
)

type Subscribe struct {
	HostPath    string        `toml:"host_path"`
	Namespace   string        `toml:"namespace"`
	IDList      []string      `toml:"ID_list"`
	Cycle       time.Duration `toml:"collect_cycle"`
	Measurement string        `toml:"measurement"`
}

type Config struct {
	Subscribes []Subscribe `toml:"subscribes"`
}
