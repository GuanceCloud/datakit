// +build linux

package containerd

import "time"

const (
	pluginName = "containerd"

	containerdConfigSample = `
# [containerd]
# [[containerd.subscribes]]
#       ## containerd 在本机的 sock 地址，一般使用默认即可
#       host_path = "/run/containerd/containerd.sock"
#
#       ## 需要采集的 containerd namespace
#       ## 可以使 'ps -ef | grep containerd | grep containerd-shim' 查看详情
#       namespace = "moby"
#
#       ## 需要采集的 containerd ID 列表，ID 是一串长度为 64 的字符串
#       ## 如果该值是 "*" ，会默认采集所有
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
