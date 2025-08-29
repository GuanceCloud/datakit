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

  ## Resource limits.
  ## The collector automatically exits when the limit is exceeded.
  ## Can configure the number of cpu cores, memory size and network bandwidth.
  ##
  # cpu_limit = "2.0"
  # mem_limit = "4GiB"
  # net_limit = "100MiB/s"

  ## automatically takes effect when running DataKit in 
  ## Kubernetes daemonset mode
  ##
  # kubernetes_url = "https://kubernetes.default:443"
  # bearer_token = "/run/secrets/kubernetes.io/serviceaccount/token"
  ##
  ## or 
  # bearer_token_string = "<your-token-string>"
  
  ## k8s workload labels
  ##
  # workload_labels = ["app"]  
  # workload_label_prefix = ""

  ## all supported plugins:
  ## - "ebpf-net"  :
  ##     contains L4-network(netflow), L7-network(httpflow, dnsflow) collection
  ## - "ebpf-bash" :
  ##     log bash
  ## - "ebpf-conntrack":
  ##     add two tags "dst_nat_ip" and "dst_nat_port" to the network flow data
  ## - "ebpf-trace":
  ##     param trace_server must be set simultaneously.
  ## - "bpf-netlog":
  ##     contains L4-network log (bpf_net_l4_log), L7-network log (bpf_net_l7_log), 
  ##              L4-network(netflow), L7-network(httpflow, dnsflow) collection
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


  ## datakit-ebpf pprof service
  pprof_host = "127.0.0.1"
  pprof_port = "6061"

  ## netlog blacklist
  ##
  # netlog_blacklist = "ip_saddr=='127.0.0.1' || ip_daddr=='127.0.0.1'"

  ## bpf-netlog plugin collection metric and log
  ##
  netlog_metric = true
  netlog_log = false
  
  ## eBPF trace generation server center address.
  trace_server = ""

  ## trace all processes containing any specified environment variable
  trace_env_list = [
    # "DK_BPFTRACE_SERVICE",
    # "DD_SERVICE",
    # "OTEL_SERVICE_NAME",
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

  # sampling_rate = "0.50"
  # sampling_rate_pts_per_min = "1500"

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
