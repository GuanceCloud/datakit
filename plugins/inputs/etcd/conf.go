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
#       ## 是否开启 HTTPS TLS，如果开启则需要同时配置下面的3个路径
#       tls_open = false
#
#       ## CA 证书路径
#       tls_cacert_file = "ca.crt"
#
#       ## 客户端证书文件路径
#	tls_cert_file = "peer.crt"
#
#	## 私钥文件路径
#	tls_key_file = "peer.key"
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
	TLSOpen     bool          `toml:"tls_open"`
	CacertFile  string        `toml:"tls_cacert_file"`
	CertFile    string        `toml:"tls_cert_file"`
	KeyFile     string        `toml:"tls_key_file"`
	Cycle       time.Duration `toml:"collect_cycle"`
	Measurement string        `toml:"measurement"`
}

type Config struct {
	Subscribes []Subscribe `toml:"subscribes"`
}
