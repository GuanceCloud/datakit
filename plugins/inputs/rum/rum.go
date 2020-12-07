package rum

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName    = `rum`
	moduleLogger *logger.Logger
)

func (_ *Rum) Catalog() string {
	return "rum"
}

func (_ *Rum) SampleConfig() string {
	return configSample
}

func (r *Rum) Run() {
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Rum{}
	})
}
