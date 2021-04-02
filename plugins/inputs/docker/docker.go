package docker

import (
	"crypto/tls"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &DockerUtil{
			newEnvClient: NewEnvClient,
			newClient:    NewClient,
			Tags:         make(map[string]string),
		}
	})
}

type DockerUtil struct {
	Endpoint              string               `toml:"endpoint"`
	CollectMetricInterval string               `toml:"collect_metric_interval"`
	CollectObjectInterval string               `toml:"collect_object_interval"`
	IncludeExited         bool                 `toml:"include_exited"`
	ClientConfig                               // tls config
	LogsPipeline          []*DockerLogPipeline `toml:"logs_pipeline"`
	Tags                  map[string]string    `toml:"tags"`

	collectMetricDuration time.Duration
	collectObjectDuration time.Duration

	newEnvClient func() (Client, error)
	newClient    func(string, *tls.Config) (Client, error)

	client Client
}

func (*DockerUtil) SampleConfig() string {
	return sampleCfg
}

func (*DockerUtil) Catalog() string {
	return "docker"
}

func (*DockerUtil) PipelineConfig() map[string]string {
	return nil
}

func (d *DockerUtil) Test() (*inputs.TestResult, error) {
	l = logger.SLogger(inputName)

	var result = inputs.TestResult{Desc: "数据指标获取失败，详情见错误信息"}
	var err error

	if err = d.loadCfg(); err != nil {
		return &result, err
	}

	var data []byte
	data, err = d.gather()
	if err != nil {
		return &result, err
	}

	result.Result = data
	result.Desc = "数据指标获取成功"

	return &result, err
}

func (d *DockerUtil) Run() {
	l = logger.SLogger(inputName)
	if d.initCfg() {
		return
	}
	l.Info("docker input start")
}

func (d *DockerUtil) initCfg() bool {
	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return true
		default:
			// nil
		}

		if err := d.loadCfg(); err != nil {
			l.Error(err)
			time.Sleep(time.Second)
		} else {
			break
		}
	}
	return false
}
