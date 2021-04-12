package docker

import (
	"context"
	"crypto/tls"
	"sync"
	"time"

	"github.com/docker/docker/api/types"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			newEnvClient: NewEnvClient,
			newClient:    NewClient,
			Tags:         make(map[string]string),
		}
	})
}

type Input struct {
	Endpoint              string            `toml:"endpoint"`
	CollectMetric         bool              `toml:"collect_metric"`
	CollectObject         bool              `toml:"collect_object"`
	CollectLogging        bool              `toml:"collect_logging"`
	CollectMetricInterval string            `toml:"collect_metric_interval"`
	CollectObjectInterval string            `toml:"collect_object_interval"`
	IncludeExited         bool              `toml:"include_exited"`
	ClientConfig                            // tls config
	LogOption             []*LogOption      `toml:"log_option"`
	Tags                  map[string]string `toml:"tags"`

	collectMetricDuration time.Duration
	collectObjectDuration time.Duration
	timeoutDuration       time.Duration

	newEnvClient         func() (Client, error)
	newClient            func(string, *tls.Config) (Client, error)
	containerLogsOptions types.ContainerLogsOptions
	containerLogList     map[string]context.CancelFunc

	client     Client
	kubernetes *Kubernetes

	opts types.ContainerListOptions
	wg   sync.WaitGroup
	mu   sync.Mutex
}

func (*Input) SampleConfig() string {
	return sampleCfg
}

func (*Input) Catalog() string {
	return "docker"
}

func (*Input) PipelineConfig() map[string]string {
	return nil
}

func (this *Input) Run() {
	l = logger.SLogger(inputName)

	if this.initCfg() {
		return
	}
	l.Info("docker input start")

	if this.CollectMetric {
		go func() {
			gatherTick := time.NewTicker(this.collectMetricDuration)
			defer gatherTick.Stop()
			for {
				select {
				case <-datakit.Exit.Wait():
					return

				case <-gatherTick.C:
					data, err := this.gather()
					if err != nil {
					}
					if err := io.NamedFeed(data, io.Metric, inputName); err != nil {
						l.Error(err)
					}
				}
			}
		}()
	}

	if this.CollectLogging {
		go func() {
			gatherTick := time.NewTicker(this.collectMetricDuration)
			defer gatherTick.Stop()
			for {
				select {
				case <-datakit.Exit.Wait():
					this.Stop()
					return

				case <-gatherTick.C:
					this.gatherLog()
				}
			}
		}()
	}

	if this.CollectObject {
		go func() {
			for {
				select {
				case <-datakit.Exit.Wait():
					return

				case <-time.After(this.collectObjectDuration):
					data, err := this.gather()
					if err != nil {
					}
					if err := io.NamedFeed(data, io.Object, inputName); err != nil {
						l.Error(err)
					}
				}
			}
		}()
	}
}

func (this *Input) Stop() {
	this.cancelTails()
	this.wg.Wait()
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return nil
}
func (this *Input) initCfg() bool {
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
			time.Sleep(time.Second)
		} else {
			break
		}
	}
	return false
}
