// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !datakit_aws_lambda
// +build !datakit_aws_lambda

package inputs

import (
	"context"
	"math/rand"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/usagetrace"
)

func RunInputs() error {
	mtx.RLock()
	defer mtx.RUnlock()
	g := datakit.G("inputs")

	envs := getEnvs()

	usagetrace.ClearInputNames()
	var utOpts []usagetrace.UsageTraceOption

	for name, arr := range InputsInfo {
		if len(arr) > 1 {
			if _, ok := arr[0].input.(Singleton); ok {
				arr = arr[:1]
			}
		}

		inputInstanceVec.WithLabelValues(name).Set(float64(len(arr)))

		// For each input, only add once
		utOpts = append(utOpts, usagetrace.WithInputNames(name))

		for _, ii := range arr {
			if ii.input == nil {
				l.Debugf("skip non-datakit-input %s", name)
				continue
			}

			if inp, ok := ii.input.(ReadEnv); ok && datakit.Docker {
				inp.ReadEnv(envs)
			}

			if inp, ok := ii.input.(HTTPInput); ok {
				inp.RegHTTPHandler()
			}

			if inp, ok := ii.input.(PipelineInput); ok {
				inp.RunPipeline()
			}

			func(name string, ii *inputInfo) {
				g.Go(func(ctx context.Context) error {
					// NOTE: 让每个采集器间歇运行，防止每个采集器扎堆启动，导致主机资源消耗出现规律性的峰值
					tick := time.NewTicker(time.Duration(rand.Int63n(int64(10 * time.Second)))) //nolint:gosec
					defer tick.Stop()
					select {
					case <-tick.C:
						l.Infof("starting input %s ...", name)

						protectRunningInput(name, ii)

						l.Infof("input %s exited, this maybe a input that only register a HTTP handle", name)
						return nil
					case <-datakit.Exit.Wait():
						l.Infof("start input %s interrupted", name)
					}
					return nil
				})
			}(name, ii)
		}
	}

	// Notify datakit usage all the started inputs.
	usagetrace.UpdateTraceOptions(utOpts...)
	return nil
}
