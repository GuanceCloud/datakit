package cmds

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	failed    = []string{}
	unknown   = []string{}
	ignored   = []string{}
	passed    = 0
	checked   = 0
	inputsCnt = 0
)

func showCheckResult() {
	infof("\n------------------------\n")
	infof("checked %d samples, %d ignored, %d passed, %d failed, %d unknown, total input instance %d\n",
		checked, len(ignored), passed, len(failed), len(unknown), inputsCnt)

	if len(ignored) > 0 {
		infof("ignored:\n")
		for _, x := range ignored {
			infof("\t%s\n", x)
		}
	}

	if len(unknown) > 0 {
		infof("unknown:\n")
		for _, x := range unknown {
			warnf("\t%s\n", x)
		}
	}

	if len(failed) > 0 {
		infof("failed:\n")
		for _, x := range failed {
			errorf("\t%s\n", x)
		}
	}
}

// check samples of every inputs.
func checkSample() error {
	failed = []string{}
	unknown = []string{}
	passed = 0
	checked = len(inputs.Inputs)
	ignored = []string{}

	for k, c := range inputs.Inputs {
		i := c()

		if k == datakit.DatakitInputName {
			warnf("[W] ignore self input\n")
			ignored = append(ignored, k)
			continue
		}

		if _, err := config.LoadSingleConf(i.SampleConfig(), inputs.Inputs); err != nil {
			errorf("[E] failed to parse %s: %s\n", k, err.Error())
			failed = append(failed, k+": "+err.Error())
		} else {
			passed++
		}
	}

	l.Debugf("checked %d inptus samples", len(inputs.Inputs))

	showCheckResult()

	if len(failed) > 0 {
		return fmt.Errorf("load %v sample failed", failed)
	}
	return nil
}

func checkConfig(dir, suffix string) error {
	fps := config.SearchDir(dir, suffix)

	failed = []string{}
	unknown = []string{}
	passed = 0
	checked = 0
	ignored = []string{}

	for _, fp := range fps {
		// Skip hidden files.
		if strings.HasPrefix(filepath.Base(fp), ".") {
			continue
		}

		if v, err := config.LoadSingleConfFile(fp, inputs.Inputs); err != nil {
			errorf("[E] failed to parse %s: %s, %s\n", fp, err.Error(), reflect.TypeOf(err))
			failed = append(failed, fp+": "+err.Error())
		} else {
			passed++
			for k, arr := range v {
				if len(arr) == 0 {
					warnf("[W] no input enabled in %s\n", fp)
				} else {
					infof("[I] got %d %s input in %s\n", len(arr), k, fp)
					inputsCnt += len(arr)
				}
			}
		}
		checked++
	}

	showCheckResult()

	if len(failed) > 0 {
		return fmt.Errorf("load %d conf failed", len(failed))
	}

	return nil
}
