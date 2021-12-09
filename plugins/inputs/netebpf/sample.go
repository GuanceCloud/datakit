package netebpf

//nolint:lll
const configSample = `
[[inputs.net_ebpf]]
  daemon = true
  name = 'net_ebpf'
  cmd = "/usr/local/datakit/externals/net_ebpf"
  args = ["--datakit-apiserver", "0.0.0.0:9529"]
  envs = []
  [inputs.net_ebpf.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"

#############################
# 参数说明(若标 * 为必选项)
#############################
#  --hostname               : 主机名，此参数可改变该采集器上传数据时 host tag 的值, 优先级为: 指定该参数 > datakit.conf 中的 ENV_HOSTNAME 值(若非空，启动时自动添加该参数) > 采集器自行获取(默认值)
#  --datakit-apiserver      : DataKit API Server 地址, 默认值 0.0.0.0:9529
#  --log                    : 日志输出路径, 默认值 DataKitInstallDir/externals/net_ebpf.log
#  --log-level              : 日志级别，默认 info
#  --service                : 默认值 net_ebpf
`
