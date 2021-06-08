package container

import (
	"net"
	"net/url"
	"time"

	"github.com/docker/docker/api/types"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

// ## To use environment variables (ie, docker-machine), set endpoint = "ENV"

const (
	inputName = "container"

	sampleCfg = `
[inputs.container]
  endpoint = "unix:///var/run/docker.sock"
  
  enable_metric = false  
  enable_object = true   
  enable_logging = true  
  
  metric_interval = "10s"

  drop_tags = ["contaienr_id"]
  pod_name_rewrite = []
  
  ## TLS Config
  # tls_ca = "/path/to/ca.pem"
  # tls_cert = "/path/to/cert.pem"
  # tls_key = "/path/to/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
  
  [inputs.container.kubelet]
    kubelet_url = "http://127.0.0.1:10255"

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
  
  #[[inputs.container.logfilter]]
  #  filter_message = [
  #    '''<this-is-message-regexp''',
  #    '''<this-is-another-message-regexp''',
  #  ]
  #  source = "<your-source-name>"
  #  service = "<your-service-name>"
  #  pipeline = "<pipeline.p>"
  
  [inputs.container.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
`
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
)

var (
	l = logger.DefaultSLogger(inputName)

	// 容器日志的连接参数
	containerLogsOptions = types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Tail:       "0", // 默认关闭FromBeginning，避免数据量巨大。开启为 'all'
	}

	// 容器获取列表的连接参数
	containerListOptions = types.ContainerListOptions{}
)

func (this *Input) loadCfg() (err error) {
	// tlsConfig 可以为空指针，即没有配置tls
	tlsConfig, _err := this.ClientConfig.TLSConfig()
	if _err != nil {
		return _err
	}

	this.client, err = this.newClient(this.Endpoint, tlsConfig)
	if err != nil {
		return
	}

	if err = this.initLoggingConf(); err != nil {
		return
	}

	if this.Kubernetes != nil {
		l.Debugf("use kubelet_url %s", this.Kubernetes.URL)

		u, err := url.Parse(this.Kubernetes.URL)
		if err != nil {
			return err
		}

		func() {
			// kubelet API 没有提供 ping 功能
			// 此处手动检查该端口是否可以连接
			if err = rawConnect(u.Hostname(), u.Port()); err != nil {
				l.Errorf("kubelet_url connecting error(not collect kubelet): %s", err)
				// 如果检测到该端口尚未监听，或无法连接，则将 this.Kubernetes 置为 空指针
				// 后续将不再采用 kubelet 相关数据
				this.Kubernetes = nil
				return
			}

			if err := this.Kubernetes.Init(); err != nil {
				l.Debugf("read kubelet token error (use empty tokne): %s", err)
				// use empty token
				this.Kubernetes.BearerTokenString = ""
			}
		}()
	}

	return
}

func (this *Input) initLoggingConf() error {
	for _, lf := range this.LogFilters {
		if err := lf.Init(); err != nil {
			return err
		}
	}
	return nil
}

func rawConnect(host string, port string) error {
	timeout := time.Second
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
	if err != nil {
		return err
	}
	if conn != nil {
		defer conn.Close()
		return nil
	}
	return nil
}
