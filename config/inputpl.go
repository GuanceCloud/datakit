package config

import (
	"io/ioutil"
	"path/filepath"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	pipelineWarning = `
#------------------------------------   警告   -------------------------------------
# 不要修改本文件，如果要更新，请拷贝至其它文件，最好以某种前缀区分，避免重启后被覆盖
#-----------------------------------------------------------------------------------
`
)

func initPluginPipeline() error {

	if err := pipeline.Init(datakit.DataDir); err != nil {
		return err
	}

	for name, c := range inputs.Inputs {
		switch v := c().(type) {
		case inputs.PipelineInput:
			ps := v.PipelineConfig()

			for k, v := range ps {
				if v == "" {
					continue // ignore empty pipeline
				}

				plpath := filepath.Join(datakit.PipelineDir, k+".p")

				if err := ioutil.WriteFile(plpath, []byte(pipelineWarning+v), 0600); err != nil {
					l.Errorf("failed to create pipeline script for %s/%s", name, k, err.Error())
					return err
				}
			}
		}
	}

	return nil
}
