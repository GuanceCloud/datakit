package container

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return newInput()
	})
}

type Input struct {
	Endpoint       string            `toml:"endpoint"`
	EnableMetric   bool              `toml:"enable_metric"`
	EnableObject   bool              `toml:"enable_object"`
	EnableLogging  bool              `toml:"enable_logging"`
	MetricInterval string            `toml:"metric_interval"`
	KubernetesURL  string            `toml:"kubernetes_url"`
	ClientConfig                     // tls config
	LogFilters     LogFilters        `toml:"logfilter"`
	Tags           map[string]string `toml:"tags"`

	newClient func(string, *tls.Config) (Client, error)

	metricDuration   time.Duration
	containerLogList map[string]context.CancelFunc

	client     Client
	kubernetes *Kubernetes

	wg sync.WaitGroup
	mu sync.Mutex
}

func newInput() *Input {
	return &Input{
		Endpoint:         dockerEndpoint,
		Tags:             make(map[string]string),
		containerLogList: make(map[string]context.CancelFunc),
		newClient:        NewClient,
		metricDuration:   minMetricDuration,
	}
}

func (*Input) SampleConfig() string {
	return sampleCfg
}

func (*Input) Catalog() string {
	return "container"
}

func (*Input) PipelineConfig() map[string]string {
	return nil
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&containersMeasurement{},
		&containersLogMeasurement{},
	}
}

func (*Input) AvailableArchs() []string {
	return []string{datakit.OSLinux}
}

func (this *Input) Run() {
	l = logger.SLogger(inputName)

	if this.initCfg() {
		return
	}
	l.Info("container input start")

	if this.EnableObject {
		this.tickObject()
	}

	tick := time.NewTicker(this.metricDuration)
	defer tick.Stop()
	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("docker exit success")
			return

		case <-tick.C:
			if this.kubernetes != nil {
				startTime := time.Now()
				pts, err := this.kubernetes.GatherPodMetrics()
				if err != nil {
					l.Error(err)
					io.FeedLastError(inputName, fmt.Sprintf("k8s pod metrics gather failed: %s", err.Error()))
					return
				}
				cost := time.Since(startTime)
				if err := io.Feed(inputName, datakit.Metric, pts, &io.Option{CollectCost: cost}); err != nil {
					l.Error(err)
					io.FeedLastError(inputName, fmt.Sprintf("k8s pod metrics gather failed: %s", err.Error()))
				}

			}
			if this.EnableMetric {
				startTime := time.Now()
				pts, err := this.gather(metricCategory)
				if err != nil {
					l.Error(err)
					io.FeedLastError(inputName, fmt.Sprintf("metrics gather failed: %s", err.Error()))
					continue
				}
				cost := time.Since(startTime)
				if err := io.Feed(inputName, datakit.Metric, pts, &io.Option{CollectCost: cost}); err != nil {
					l.Error(err)
					io.FeedLastError(inputName, fmt.Sprintf("metrics gather failed: %s", err.Error()))
				}
			}

		case <-time.After(objectDuration):
			if this.EnableObject {
				this.tickObject()
			}

		case <-time.After(loggingHitDuration):
			if this.EnableLogging {
				this.gatherLog()
			}
		}
	}
}

func (this *Input) tickObject() {
	if this.kubernetes != nil {
		startTime := time.Now()
		pts, err := this.kubernetes.GatherPodMetrics()
		if err != nil {
			l.Error(err)
			io.FeedLastError(inputName, fmt.Sprintf("k8s pod object gather failed: %s", err.Error()))
			return
		}
		cost := time.Since(startTime)
		if err := io.Feed(inputName, datakit.Object, pts, &io.Option{CollectCost: cost}); err != nil {
			l.Error(err)
			io.FeedLastError(inputName, fmt.Sprintf("k8s pod object gather failed: %s", err.Error()))
		}
	}

	startTime := time.Now()
	pts, err := this.gather(objectCategory)
	if err != nil {
		l.Error(err)
		io.FeedLastError(inputName, fmt.Sprintf("object gather failed: %s", err.Error()))
		return
	}
	cost := time.Since(startTime)
	if err := io.Feed(inputName, datakit.Object, pts, &io.Option{CollectCost: cost}); err != nil {
		l.Error(err)
		io.FeedLastError(inputName, fmt.Sprintf("object gather failed: %s", err.Error()))
	}
}

func (this *Input) initCfg() bool {
	// 如果配置文件中使用默认 endpoint 且该文件不存在，说明其没有安装 docker（经测试，docker service 停止后，sock 文件依然存在）
	// 此行为是为了应对 default_enabled_inputs 行为，避免在没有安装 docker 的主机上开启 docker，然后无限 error
	if this.Endpoint == dockerEndpoint {
		if _, err := os.Stat(dockerEndpointPath); os.IsNotExist(err) {
			l.Infof("check defaultEndpoint: %s is not exist, maybe docker.service is not installed, exit", this.Endpoint)
			// 预料之中的退出
			return true
		}
	}

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return true
		default:
			// nil
		}

		if err := this.loadCfg(); err != nil {
			l.Error(err)
			io.FeedLastError(inputName, fmt.Sprintf("load config: %s", err.Error()))
			time.Sleep(time.Second)
		} else {
			break
		}
	}
	return false
}
