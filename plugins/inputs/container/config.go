package container

import (
	"fmt"
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

    ## TLS Config
    # tls_ca = "/path/to/ca.pem"
    # tls_cert = "/path/to/cert.pem"
    # tls_key = "/path/to/key.pem"
    ## Use TLS but skip chain & host verification
    # insecure_skip_verify = false

    #[[inputs.container.logfilter]]
        # filter_message = [
        #   '''<this-is-message-regexp''',
        #   '''<this-is-another-message-regexp''',
        # ]

        # source = "<your-source-name>"
        # service = "<your-service-name>"
        # pipeline = "<pipeline.p>"

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
	minMetricDuration = time.Second * 5

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

	// 始终认为，docker和k8s在同一台主机上
	// 避免进行冗杂的k8s连接配置
	var k8sURL = fmt.Sprintf(defaultKubernetesURL, "127.0.0.1")
	if this.Endpoint != dockerEndpoint {
		if u, err := url.Parse(this.Endpoint); err == nil {
			k8sURL = fmt.Sprintf(defaultKubernetesURL, u.Hostname())
		}
	}

	l.Debugf("use k8sURL %s", k8sURL)

	this.kubernetes = func() *Kubernetes {
		k := Kubernetes{URL: k8sURL}
		if err := k.Init(); err != nil {
			l.Debugf("read k8s token error (use empty tokne): %s", err)
			// use empty token
			k.BearerTokenString = ""
		}
		return &k
	}()

	return
}
