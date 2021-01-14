package process

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "process"

	sampleConfig = ``

	pipelineSample = ``
)

type collector struct {
	testResult *inputs.TestResult
	testError  error

	mode string
}
