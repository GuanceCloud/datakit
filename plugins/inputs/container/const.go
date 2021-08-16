package container

import (
	"time"
)

const (
	inputName = "container"
	catelog   = "container"

	// docker endpoint
	dockerEndpoint = "unix:///var/run/docker.sock"

	// docker sock 文件路径，用以判断主机是否已安装 docker 服务
	dockerEndpointPath = "/var/run/docker.sock"

	// Docker API 超时时间
	apiTimeoutDuration = time.Second * 5

	// 最小指标采集间隔
	minMetricDuration = time.Second * 10

	// 最大指标采集间隔
	maxMetricDuration = time.Second * 60

	// 对象采集间隔
	objectDuration = time.Minute * 5

	// 定时发现新日志源
	loggingHitDuration = time.Second * 5

	// 是否采集全部容器，包括未在运行的
	containerAllForMetric  = false
	containerAllForObject  = true
	containerAllForLogging = false
)

const sampleCfg = `
[inputs.container]
  endpoint = "unix:///var/run/docker.sock"
  
  enable_metric = false  
  enable_object = true   
  enable_logging = true  
  
  metric_interval = "10s"

  drop_tags = ["contaienr_id"]

  ## Examples:
  ##    '''nginx*'''
  ignore_image_name = []
  ignore_container_name = []
  
  ## TLS Config
  # tls_ca = "/path/to/ca.pem"
  # tls_cert = "/path/to/cert.pem"
  # tls_key = "/path/to/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
  
  [inputs.container.kubelet]
    kubelet_url = "http://127.0.0.1:10255"
    ignore_pod_name = []

    ## Use bearer token for authorization. ('bearer_token' takes priority)
    ## If both of these are empty, we'll use the default serviceaccount:
    ## at: /run/secrets/kubernetes.io/serviceaccount/token
    # bearer_token = "/path/to/bearer/token"
    ## OR
    # bearer_token_string = "abc_123"

    ## Optional TLS Config
    # tls_ca = /path/to/ca.pem
    # tls_cert = /path/to/cert.pem
    # tls_key = /path/to/key.pem
    ## Use TLS but skip chain & host verification
    # insecure_skip_verify = false
  
  #[[inputs.container.log]]
  #  match_by = "container-name"
  #  match = [
  #    '''<this-is-regexp''',
  #  ]
  #  source = "<your-source-name>"
  #  service = "<your-service-name>"
  #  pipeline = "<pipeline.p>"
  
  [inputs.container.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
`
