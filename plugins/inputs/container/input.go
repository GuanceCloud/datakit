// Package container collect container metrics/loggings/objects.
package container

import (
	"context"
	"os"
	"regexp"
	"strconv"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
	timex "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/time"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var _ inputs.ReadEnv = (*Input)(nil)

type Input struct {
	Endpoint string `toml:"endpoint"`

	EnableMetric  bool `toml:"enable_metric"`
	EnableObject  bool `toml:"enable_object"`
	EnableLogging bool `toml:"enable_logging"`

	MetricInterval               timex.Duration `toml:"metric_interval"`
	LoggingRemoveAnsiEscapeCodes bool           `toml:"logging_remove_ansi_escape_codes"`

	IgnoreImageName     []string          `toml:"ignore_image_name"`
	IgnoreContainerName []string          `toml:"ignore_container_name"`
	DropTags            []string          `toml:"drop_tags"`
	Tags                map[string]string `toml:"tags"`

	TLSCA              string `toml:"tls_ca"`
	TLSCert            string `toml:"tls_cert"`
	TLSKey             string `toml:"tls_key"`
	InsecureSkipVerify bool   `toml:"insecure_skip_verify"`

	Kubernetes *Kubernetes `toml:"kubelet"`
	Logs       Logs        `toml:"log"`

	in chan []*job

	clients []collector

	LogDepercated            DepercatedLog `toml:"logfilter,omitempty"`
	PodNameRewriteDeprecated []string      `toml:"pod_name_write,omitempty"`
	PodnameRewriteDeprecated []string      `toml:"pod_name_rewrite,omitempty"`
}

var l = logger.DefaultSLogger(inputName)

func newInput() *Input {
	return &Input{
		Endpoint: dockerEndpoint,
		Tags:     make(map[string]string),
		in:       make(chan []*job, 64),
	}
}

func (*Input) SampleConfig() string { return sampleCfg }

func (*Input) Catalog() string { return catelog }

func (*Input) PipelineConfig() map[string]string { return nil }

func (*Input) AvailableArchs() []string { return []string{datakit.OSLinux} }

func (*Input) RunPipeline() {
	// TODO.
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

func (i *Input) Run() {
	l = logger.SLogger(inputName)

	if i.setup() {
		return
	}
	l.Info("container input startd")

	if i.EnableObject {
		for _, c := range i.clients {
			c.Object(context.Background(), i.in)
		}
	}

	if i.EnableLogging {
		for _, c := range i.clients {
			c.Logging(context.Background())
		}
	}

	metricsTick := time.NewTicker(i.MetricInterval.Duration)
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
			if i.EnableMetric {
				for _, c := range i.clients {
					c.Metric(context.Background(), i.in)
				}
			}

		case <-objectTick.C:
			if i.EnableObject {
				for _, c := range i.clients {
					c.Object(context.Background(), i.in)
				}
			}

		case <-loggingTick.C:
			if i.EnableLogging {
				for _, c := range i.clients {
					c.Logging(context.Background())
				}
			}
		}
	}
}

// ReadEnv support envs：
//   ENV_INPUT_CONTAINER_ENABLE_METRIC : booler
//   ENV_INPUT_CONTAINER_ENABLE_OBJECT : booler
//   ENV_INPUT_CONTAINER_ENABLE_LOGGING : booler
//   ENV_INPUT_CONTAINER_LOGGING_REMOVE_ANSI_ESCAPE_CODES : booler
//   ENV_INPUT_CONTAINER_TAGS : "a=b,c=d"
func (i *Input) ReadEnv(envs map[string]string) {
	if enable, ok := envs["ENV_INPUT_CONTAINER_ENABLE_METRIC"]; ok {
		b, err := strconv.ParseBool(enable)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_ENABLE_METRIC to bool: %s, ignore", err)
		} else {
			i.EnableMetric = b
		}
	}

	if enable, ok := envs["ENV_INPUT_CONTAINER_ENABLE_OBJECT"]; ok {
		b, err := strconv.ParseBool(enable)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_ENABLE_OBJECT to bool: %s, ignore", err)
		} else {
			i.EnableObject = b
		}
	}

	if enable, ok := envs["ENV_INPUT_CONTAINER_ENABLE_LOGGING"]; ok {
		b, err := strconv.ParseBool(enable)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_ENABLE_LOGGING to bool: %s, ignore", err)
		} else {
			i.EnableLogging = b
		}
	}

	if remove, ok := envs["ENV_INPUT_CONTAINER_LOGGING_REMOVE_ANSI_ESCAPE_CODES"]; ok {
		b, err := strconv.ParseBool(remove)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_LOGGING_REMOVE_ANSI_ESCAPE_CODES to bool: %s, ignore", err)
		} else {
			i.LoggingRemoveAnsiEscapeCodes = b
		}
	}

	if tagsStr, ok := envs["ENV_INPUT_CONTAINER_TAGS"]; ok {
		tags := config.ParseGlobalTags(tagsStr)
		for k, v := range tags {
			i.Tags[k] = v
		}
	}
}

func (i *Input) setup() bool {
	// 如果配置文件中使用默认 endpoint 且该文件不存在，说明其没有安装 docker（经测试，docker service 停止后，sock 文件依然存在）
	// 此行为是为了应对 default_enabled_inputs 行为，避免在没有安装 docker 的主机上开启 input，然后无限 error
	if i.Endpoint == dockerEndpoint {
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

		if err = i.verifyIgnoreRegexps(); err != nil {
			l.Error(err)
			continue
		}

		if err = i.buildK8sClient(); err != nil {
			l.Error(err)
			continue
		}

		if err = i.buildDockerClient(); err != nil {
			l.Error(err)
			continue
		}

		if err = i.initLogs(); err != nil {
			l.Error(err)
			continue
		}

		break
	}

	if i.EnableMetric || i.EnableObject {
		go i.doFeed()
	}

	return false
}

func (i *Input) buildDockerClient() error {
	t := net.TLSClientConfig{
		CaCerts: func() []string {
			if i.TLSCA == "" {
				return nil
			}
			return []string{i.TLSCA}
		}(),
		Cert:               i.TLSCert,
		CertKey:            i.TLSKey,
		InsecureSkipVerify: i.InsecureSkipVerify,
	}

	tlsConfig, err := t.TLSConfig()
	if err != nil {
		l.Error(err)
		return err
	}

	client, err := newDockerClient(i.Endpoint, tlsConfig)
	if err != nil {
		l.Error(err)
		return err
	}

	client.IgnoreImageName = i.IgnoreImageName
	client.IgnoreContainerName = i.IgnoreContainerName
	client.LoggingRemoveAnsiEscapeCodes = i.LoggingRemoveAnsiEscapeCodes
	client.ProcessTags = i.processTags
	client.Logs = i.Logs
	if verifyIntegrityOfK8sConnect(i.Kubernetes) {
		client.K8s = i.Kubernetes
	}

	i.clients = append(i.clients, client)

	return nil
}

func (i *Input) buildK8sClient() error {
	if i.Kubernetes == nil {
		return nil
	}

	if err := i.Kubernetes.Init(); err != nil {
		// 如果使用默认 k8s url，init() 失败将不会追究，忽略此错误避免影响到 container 采集

		if i.Kubernetes.URL == "http://127.0.0.1:10255" ||
			i.Kubernetes.URL == "http://localhost:10255" {
			// NOTE: 此处将该指针置空，以示后续将不再采集 k8s
			i.Kubernetes = nil
			return nil
		}
		// 如果该 k8s url 并非默认值，则说明该值是一个经过配置的、预期可用的 url，不可再忽略此报错
		return err
	}

	i.clients = append(i.clients, i.Kubernetes)

	return nil
}

func (i *Input) initLogs() error {
	return i.Logs.Init()
}

func (i *Input) verifyIgnoreRegexps() error {
	for _, n := range i.IgnoreImageName {
		if _, err := regexp.Compile(n); err != nil {
			return err
		}
	}

	for _, n := range i.IgnoreContainerName {
		if _, err := regexp.Compile(n); err != nil {
			return err
		}
	}

	return nil
}

func (i *Input) doFeed() {
	type data = struct {
		pts   []*io.Point
		costs []time.Duration
	}
	cache := make(map[string]*data)

	cleanTick := time.NewTicker(time.Second * 3)
	defer cleanTick.Stop()

	for {
		select {
		case <-datakit.Exit.Wait():
			return

		case jobs := <-i.in:
			for _, job := range jobs {
				i.processTags(job.tags)

				pt, err := io.MakePoint(job.measurement, job.tags, job.fields, job.ts)
				if err != nil {
					l.Error(err)
					continue
				}

				if _, ok := cache[job.category]; !ok {
					cache[job.category] = &data{}
				}

				cache[job.category].pts = append(cache[job.category].pts, pt)
				cache[job.category].costs = append(cache[job.category].costs, job.cost)
			}

		case <-cleanTick.C:
			for category, d := range cache {
				if len(d.pts) == 0 {
					d.costs = d.costs[:0]
					continue
				}

				opt := func() *io.Option {
					if len(d.costs) == 0 {
						return nil
					}

					var sum time.Duration
					for _, cost := range d.costs {
						sum += cost
					}
					if sum == 0 {
						return nil
					}

					return &io.Option{CollectCost: time.Duration(int64(sum) / int64(len(d.costs)))}
				}()

				if err := io.Feed(inputName, category, d.pts, opt); err != nil {
					l.Error(err)
				}

				d.pts = d.pts[:0]
				d.costs = d.costs[:0]
			}
		}
	}
}

func (i *Input) processTags(tags map[string]string) {
	for _, key := range i.DropTags {
		delete(tags, key)
	}

	for k, v := range i.Tags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		}
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return newInput()
	})
}
