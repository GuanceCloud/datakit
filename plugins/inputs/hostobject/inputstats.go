package hostobject

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	tgi "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/telegraf_inputs"
)

type enabledInput struct {
	Input     string `json:"input"`
	Instances int    `json:"instances"`
	Panics    int    `json:"panic"`
}

type inputStats struct {
	EnabledInputs []*enabledInput `json:"enabled_inputs"`
}

func getInputStats() *inputStats {

	var stats inputStats

	for k := range inputs.Inputs {
		n, _ := inputs.InputEnabled(k)
		npanic := inputs.GetPanicCnt(k)
		if n > 0 {
			stats.EnabledInputs = append(stats.EnabledInputs, &enabledInput{Input: k, Instances: n, Panics: npanic})
		}
	}

	for k := range tgi.TelegrafInputs {
		n, _ := inputs.InputEnabled(k)
		if n > 0 {
			stats.EnabledInputs = append(stats.EnabledInputs, &enabledInput{Input: k, Instances: n})
		}
	}

	return &stats
}
