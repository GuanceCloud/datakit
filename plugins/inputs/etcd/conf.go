package etcd

import "time"

const (
	pluginName = "etcd"

	etcdConfigSample = `
# [etcd]
# [[etcd.subscribes]]
#       ## etcd 地址
#	host = "127.0.0.1"
#
#       ## etcd 端口
#	port = 2379
#
#       ## CA 证书路径（例如：ca.crt）
#       # tls_cacert_file = ""
#
#       ## 客户端证书文件路径（例如：peer.crt）
#	# tls_cert_file = ""
#
#	## 私钥文件路径（例如：peer.key）
#	tls_key_file = ""
#
#	## 采集周期，时间单位是秒
#	collect_cycle = 60
#
#       ## measurement，不可重复
#       measurement = "etcd"
`
)

type Subscribe struct {
	EtcdHost    string        `toml:"host"`
	EtcdPort    int           `toml:"port"`
	CacertFile  string        `toml:"tls_cacert_file"`
	CertFile    string        `toml:"tls_cert_file"`
	KeyFile     string        `toml:"tls_key_file"`
	Cycle       time.Duration `toml:"collect_cycle"`
	Measurement string        `toml:"measurement"`
}

type Config struct {
	Subscribes []Subscribe `toml:"subscribes"`
}
