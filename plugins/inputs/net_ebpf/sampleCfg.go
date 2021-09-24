package net_ebpf

const configSample = `
[[inputs.net_ebpf]]
  daemon = true
  name = 'net_ebpf'
  cmd = "/usr/local/datakit/externals/net_ebpf"
  args = []
  envs = []
  [inputs.net_ebpf.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"

#############################
# 参数说明(标 * 为必选项)
#############################
#  --hostname               : host name
#  --datakit-apiserver      : DataKit API Server 地址, 默认值 0.0.0.0:9529  
#  --log                    : 日志输出路径, 默认值 DataKitInstallDir/externals/net_ebpf.log
#  --service                : 默认值 net_ebpf
`
