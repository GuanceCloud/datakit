package container

import (
	"context"
	"os"
	"regexp"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
	timex "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/time"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	l = logger.DefaultSLogger(inputName)
)

type Input struct {
	Endpoint string `toml:"endpoint"`

	EnableMetric  bool `toml:"enable_metric"`
	EnableObject  bool `toml:"enable_object"`
	EnableLogging bool `toml:"enable_logging"`

	MetricInterval string `toml:"metric_interval"`
	metricDuration time.Duration

	IgnoreImageName     []string          `toml:"ignore_image_name"`
	IgnoreContainerName []string          `toml:"ignore_container_name"`
	DropTags            []string          `toml:"drop_tags"`
	Tags                map[string]string `toml:"tags"`

	TLSCA              string `toml:"tls_ca"`
	TLSCert            string `toml:"tls_cert"`
	TLSKey             string `toml:"tls_key"`
	InsecureSkipVerify bool   `toml:"insecure_skip_verify"`

	Kubernetes *Kubernetes `toml:"kubelet"`
	LogFilters LogFilters  `toml:"logfilter"`

	in chan *job

	clients []collector

	wg sync.WaitGroup
	mu sync.Mutex

	DeprecatedPodNameRewrite []string `toml:"pod_name_write"`
}

func newInput() *Input {
	return &Input{
		Endpoint:       dockerEndpoint,
		Tags:           make(map[string]string),
		metricDuration: minMetricDuration,
		in:             make(chan *job, 64),
	}
}

func (*Input) SampleConfig() string {
	return sampleCfg
}

func (*Input) Catalog() string {
	return catelog
}

func (*Input) PipelineConfig() map[string]string {
	return nil
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&containerMetricMeasurement{},
		&containerObjectMeasurement{},
		&containerLogMeasurement{},
		&kubeletPodMetricMeasurement{},
		&kubeletPodObjectMeasurement{},
	}
}

func (*Input) AvailableArchs() []string {
	return []string{datakit.OSLinux}
}

// TODO
func (*Input) RunPipeline() {
}

func (this *Input) Run() {
	l = logger.SLogger(inputName)

	if this.setup() {
		return
	}
	l.Info("container input startd")

	if this.EnableObject {
		for _, c := range this.clients {
			c.Object(context.Background(), this.in)
		}
	}

	if this.EnableLogging {
		for _, c := range this.clients {
			c.Logging(context.Background())
		}
	}

	metricsTick := time.NewTicker(this.metricDuration)
	defer metricsTick.Stop()

	objectTick := time.NewTicker(objectDuration)
	defer objectTick.Stop()

	loggingTick := time.NewTicker(loggingHitDuration)
	defer loggingTick.Stop()

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("container exit success")
			return

		case <-metricsTick.C:
			if this.EnableMetric {
				for _, c := range this.clients {
					c.Metric(context.Background(), this.in)
				}
			}

		case <-objectTick.C:
			if this.EnableObject {
				for _, c := range this.clients {
					c.Object(context.Background(), this.in)
				}
			}

		case <-loggingTick.C:
			if this.EnableLogging {
				for _, c := range this.clients {
					c.Logging(context.Background())
				}
			}
		}
	}
}

func (this *Input) setup() bool {
	// 如果配置文件中使用默认 endpoint 且该文件不存在，说明其没有安装 docker（经测试，docker service 停止后，sock 文件依然存在）
	// 此行为是为了应对 default_enabled_inputs 行为，避免在没有安装 docker 的主机上开启 input，然后无限 error
	if this.Endpoint == dockerEndpoint {
		_, staterr := os.Stat(dockerEndpointPath)
		if os.IsNotExist(staterr) {
			l.Infof("check defaultEndpoint: %s is not exist, maybe docker.service is not installed, exit", dockerEndpointPath)
			return true
		}
	}

	var err error

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return true
		default:
			// nil
		}

		time.Sleep(time.Second)

		if err = this.verifyIgnoreRegexps(); err != nil {
			l.Error(err)
			continue
		}

		if err = this.buildDockerClient(); err != nil {
			l.Error(err)
			continue
		}

		if err = this.buildK8sClient(); err != nil {
			l.Error(err)
			continue
		}

		if err = this.initLogFilters(); err != nil {
			l.Error(err)
			continue
		}

		break
	}

	if dur, err := timex.ParseDuration(this.MetricInterval); err != nil {
		l.Debug(err)
	} else {
		this.metricDuration = dur
	}

	if this.EnableMetric || this.EnableObject {
		go this.doFeed()
	}

	return false
}

func (this *Input) buildDockerClient() error {
	t := net.TlsClientConfig{
		CaCerts: func() []string {
			if this.TLSCA == "" {
				return nil
			}
			return []string{this.TLSCA}
		}(),
		Cert:               this.TLSCert,
		CertKey:            this.TLSKey,
		InsecureSkipVerify: this.InsecureSkipVerify,
	}

	tlsConfig, err := t.TlsConfig()
	if err != nil {
		l.Error(err)
		return err
	}

	client, err := newDockerClient(this.Endpoint, tlsConfig)
	if err != nil {
		l.Error(err)
		return err
	}
	client.IgnoreImageName = this.IgnoreImageName
	client.IgnoreContainerName = this.IgnoreContainerName
	client.ProcessTags = this.processTags
	client.LogFilters = this.LogFilters
	client.K8s = this.Kubernetes

	this.clients = append(this.clients, client)

	return nil
}

func (this *Input) buildK8sClient() error {
	if this.Kubernetes == nil {
		return nil
	}

	err := this.Kubernetes.Init()
	if err != nil {
		return err
	}

	this.clients = append(this.clients, this.Kubernetes)

	return nil
}

func (this *Input) initLogFilters() error {
	for _, lf := range this.LogFilters {
		if err := lf.Init(); err != nil {
			return err
		}
	}
	return nil
}

func (this *Input) verifyIgnoreRegexps() error {
	for _, n := range this.IgnoreImageName {
		if _, err := regexp.Compile(n); err != nil {
			return err
		}
	}

	for _, n := range this.IgnoreContainerName {
		if _, err := regexp.Compile(n); err != nil {
			return err
		}
	}

	return nil
}

func (this *Input) doFeed() {
	for {
		select {
		case <-datakit.Exit.Wait():
			return

		case in := <-this.in:
			this.processTags(in.tags)

			pt, err := io.MakePoint(in.measurement, in.tags, in.fields, in.ts)
			if err != nil {
				l.Error(err)
				continue
			}

			opt := func() *io.Option {
				if in.cost != 0 {
					return &io.Option{CollectCost: in.cost}
				}
				return nil
			}()

			if err := io.Feed(inputName, in.category, []*io.Point{pt}, opt); err != nil {
				l.Error(err)
			}
		}
	}
}

func (this *Input) processTags(tags map[string]string) {
	for _, key := range this.DropTags {
		if _, ok := tags[key]; ok {
			delete(tags, key)
		}
	}

	for k, v := range this.Tags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		}
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return newInput()
	})
}
