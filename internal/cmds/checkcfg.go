// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
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
	cp.Infof("\n------------------------\n")
	cp.Infof("checked %d samples, %d ignored, %d passed, %d failed, %d unknown, total input instance %d\n",
		checked, len(ignored), passed, len(failed), len(unknown), inputsCnt)

	if len(ignored) > 0 {
		cp.Infof("ignored:\n")
		for _, x := range ignored {
			cp.Infof("\t%s\n", x)
		}
	}

	if len(unknown) > 0 {
		cp.Infof("unknown:\n")
		for _, x := range unknown {
			cp.Warnf("\t%s\n", x)
		}
	}

	if len(failed) > 0 {
		cp.Infof("failed:\n")
		for _, x := range failed {
			cp.Errorf("\t%s\n", x)
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
			cp.Warnf("[W] ignore self input\n")
			ignored = append(ignored, k)
			continue
		}

		if _, err := config.LoadSingleConf(i.SampleConfig(), inputs.Inputs); err != nil {
			cp.Errorf("[E] failed to parse %s: %s\n", k, err.Error())
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
	fps := config.SearchDir(dir, suffix, ".git")

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

		if v, err := config.LoadSingleConfFile(fp, inputs.Inputs, false); err != nil {
			cp.Errorf("[E] failed to parse %s: %s, %s\n", fp, err.Error(), reflect.TypeOf(err))
			failed = append(failed, fp+": "+err.Error())
		} else {
			passed++
			for k, arr := range v {
				if len(arr) == 0 {
					cp.Warnf("[W] no input enabled in %s\n", fp)
				} else {
					cp.Infof("[I] got %d %s input in %s\n", len(arr), k, fp)
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
