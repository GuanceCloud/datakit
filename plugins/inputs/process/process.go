package process

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func (_ *collector) Catalog() string {
	return inputName
}

func (_ *collector) SampleConfig() string {
	return sampleConfig
}

func (c *collector) PipelineConfig() map[string]string {
	return map[string]string{
		inputName: pipelineSample,
	}
}

func (c *collector) Test() (*inputs.TestResult, error) {
	c.mode = "test"
	c.testResult = &inputs.TestResult{}
	c.Run()
	return c.testResult, c.testError
}

func (c *collector) Run() {

}
