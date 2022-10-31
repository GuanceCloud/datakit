// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ebpf

//nolint:lll
const configSample = `
[[inputs.ebpf]]
  daemon = true
  name = 'ebpf'
  cmd = "/usr/local/datakit/externals/datakit-ebpf"
  args = [
    "--datakit-apiserver", "0.0.0.0:9529",
  ]
  envs = []

  ## automatically takes effect when running DataKit in 
  ## Kubernetes daemonset mode
  ##
  # kubernetes_url = "https://kubernetes.default:443"
  # bearer_token = "/run/secrets/kubernetes.io/serviceaccount/token"
  # or # bearer_token_string = "<your-token-string>"
  
  ## all supported plugins:
  ## - "ebpf-net"  :
  ##     contains L4-network(netflow), L7-network(httpflow, dnsflow) collection
  ## - "ebpf-bash" :
  ##     log bash
  ##
  enabled_plugins = [
    "ebpf-net",
  ]

  ## 若开启 ebpf-net 插件，需选配: 
  ##  - "httpflow" (* 默认开启)
  ##  - "httpflow-tls"
  ##
  l7net_enabled = [
    "httpflow",
    # "httpflow-tls"
  ]

  ## if the system does not enable ipv6, it needs to be changed to true
  ##
  ipv6_disabled = false

  ## ephemeral port strart from <ephemeral_port>
  ##
  # ephemeral_port = 10001

  # interval = "60s"

  [inputs.ebpf.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"

#############################
# 参数说明(若标 * 为必选项)
#############################
#  --hostname               : 主机名，此参数可改变该采集器上传数据时 host tag 的值, 优先级为: 指定该参数 > datakit.conf 中的 ENV_HOSTNAME 值(若非空，启动时自动添加该参数) > 采集器自行获取(默认值)
#  --datakit-apiserver      : DataKit API Server 地址, 默认值 0.0.0.0:9529
#  --log                    : 日志输出路径, 默认值 DataKitInstallDir/externals/datakit-ebpf.log
#  --log-level              : 日志级别，默认 info
#  --service                : 默认值 ebpf
`
