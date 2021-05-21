package smart

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type Input struct {
}

func (*Input) Catalog() string {

}

func (*Input) SampleConfig() string {

}

func (*Input) AvailabelArch() []string {
	return datakit.AllArch
}

func init() {
	inputs.Add("")
}
