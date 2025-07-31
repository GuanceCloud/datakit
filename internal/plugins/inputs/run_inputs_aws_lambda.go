// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build datakit_aws_lambda
// +build datakit_aws_lambda

package inputs

import (
	"context"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

func RunInputs(iptInfos map[string][]*InputInfo) error {
	mtx.RLock()
	defer mtx.RUnlock()
	g := datakit.G("inputs")

	envs := getEnvs()

	for name, arr := range iptInfos {
		if len(arr) > 1 {
			if _, ok := arr[0].Input.(Singleton); ok {
				arr = arr[:1]
			}
		}

		inputInstanceVec.WithLabelValues(name).Set(float64(len(arr)))

		for _, ii := range arr {
			if ii.Input == nil {
				l.Debugf("skip non-datakit-input %s", name)
				continue
			}

			if inp, ok := ii.Input.(ReadEnv); ok && datakit.Docker {
				inp.ReadEnv(envs)
			}

			if inp, ok := ii.Input.(HTTPInput); ok {
				inp.RegHTTPHandler()
			}

			if inp, ok := ii.Input.(PipelineInput); ok {
				inp.RunPipeline()
			}

			func(name string, ii *InputInfo) {
				g.Go(func(ctx context.Context) error {
					protectRunningInput(name, ii)
					l.Infof("input %s exited, this maybe a input that only register a HTTP handle", name)
					return nil
				})
			}(name, ii)
		}
	}
	return nil
}
