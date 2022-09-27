// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/engine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/runtime"
)

func NewTestingRunner(script string) (*runtime.Script, error) {
	name := "default.p"
	ret1, ret2 := engine.ParseScript(map[string]string{
		"default.p": script,
	}, map[string]string{
		"default.p": "",
	},
		FuncsMap, FuncsCheckMap,
	)
	if len(ret1) > 0 {
		return ret1[name], nil
	}
	if len(ret2) > 0 {
		return nil, ret2[name]
	}
	return nil, fmt.Errorf("parser func error")
}

func NewTestingRunner2(scripts map[string]string) (map[string]*runtime.Script, map[string]error) {
	return engine.ParseScript(scripts, nil, FuncsMap, FuncsCheckMap)
}

func runScript(proc *runtime.Script, measurement string,
	tags map[string]string, fields map[string]any, tn time.Time) (
	string, map[string]string, map[string]any, time.Time, bool, error,
) {
	return engine.RunScript(proc, measurement, tags, fields, tn, nil)
}
