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
  ##
  ## or 
  # bearer_token_string = "<your-token-string>"
  
  ## all supported plugins:
  ## - "ebpf-net"  :
  ##     contains L4-network(netflow), L7-network(httpflow, dnsflow) collection
  ## - "ebpf-bash" :
  ##     log bash
  ## - "ebpf-conntrack":
  ##     add two tags "dst_nat_ip" and "dst_nat_port" to the network flow data
  ## - "ebpf-trace":
  ##     param trace_server must be set simultaneously.
  enabled_plugins = [
    "ebpf-net",
  ]

  ## If you enable the ebpf-net plugin, you can configure:
  ##  - "httpflow" (* enabled by default)
  ##  - "httpflow-tls"
  ##
  l7net_enabled = [
    "httpflow",
    # "httpflow-tls"
  ]

  ## eBPF trace generation server center address.
  trace_server = ""

  ## trace all processes directly
  ##
  trace_all_process = false
  
  ## trace all processes containing any specified environment variable
  trace_env_list = [
    "DK_BPFTRACE_SERVICE",
    "DD_SERVICE",
    "OTEL_SERVICE_NAME",
  ]
  
  ## deny tracking any process containing any specified environment variable
  trace_env_blacklist = []
  
  ## trace all processes containing any specified process names,
  ## can be used with trace_namedenyset
  ##
  trace_name_list = []
  
  ## deny tracking any process containing any specified process names
  ##
  trace_name_blacklist = [
    
    ## The following two processes are hard-coded to never be traced,
    ## and do not need to be set:
    ##
    # "datakit",
    # "datakit-ebpf",
  ]

  ## conv other trace id to datadog trace id (base 10, 64-bit) 
  conv_to_ddtrace = false

  ## If the system does not enable ipv6, it needs to be changed to true
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
## Parameter description (if marked * is required)
#############################
##  --hostname               : Host name, this parameter can change the value of the host tag when the collector uploads data, the priority is: specify this parameter >
##                             ENV_HOSTNAME value in datakit.conf (if it is not empty, this parameter will be added automatically at startup) >
##                             collector Get it yourself (the default).
##  --datakit-apiserver      : DataKit API Server address, default value 0.0.0.0:9529 .
##  --log                    : Log output path, default <DataKitInstallDir>/externals/datakit-ebpf.log.
##  --log-level              : Log level, the default value is 'info'.
##  --service                : The default value is 'ebpf'.
`
