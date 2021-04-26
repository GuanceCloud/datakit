package kubernetes

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/election"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "kubernetes"

	sampleCfg = `
[inputs.kubernetes]
  kube_apiserver_url = "http://127.0.0.1:8080/metrics"

  [inputs.docker.tags]
    # tags1 = "value1"
`
)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return newInput()
	})
}

type Input struct {
	KubeAPIServerURL string `toml:"kube_apiserver_url"`
}

func newInput() *Input {
	return nil
}

func (*Input) SampleConfig() string {
	return sampleCfg
}

func (*Input) Catalog() string {
	return "kubernetes"
}

func (*Input) PipelineConfig() map[string]string {
	return nil
}

func (*Input) AvailableArchs() []string {
	return []string{datakit.OSLinux}
}

func (n *Input) SampleMeasurement() []inputs.Measurement {
	return nil
}

func (this *Input) Run() {
	// TODO

	tick := time.NewTicker(time.Second * 5)
	defer tick.Stop()

	for {
		select {
		case <-datakit.Exit.Wait():
			return

		case <-tick.C:
			if election.CurrentStats() == election.Leader {
				this.gather()
			}
		}
	}
}

func (this *Input) Stop() {
}

func (this *Input) gather() {
	// TODO
}
