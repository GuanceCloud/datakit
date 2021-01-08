package config

import (
	"io/ioutil"
	"path/filepath"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func initPluginPipeline() error {

	for name, c := range inputs.Inputs {
		switch v := c().(type) {
		case inputs.PipelineInput:
			ps := v.PipelineConfig()

			for k, v := range ps {
				plpath := filepath.Join(datakit.PipelineDir, k+".p")

				if err := ioutil.WriteFile(plpath, []byte(v), 0600); err != nil {
					l.Errorf("failed to create pipeline script for %s/%s", name, k, err.Error())
					return err
				}
			}
		}
	}

	return nil
}
