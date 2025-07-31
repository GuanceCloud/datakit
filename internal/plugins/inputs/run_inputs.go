// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !datakit_aws_lambda
// +build !datakit_aws_lambda

package inputs

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/usagetrace"
)

func RunInputs(iptInfos map[string][]*InputInfo) error {
	mtx.RLock()
	defer mtx.RUnlock()
	usagetrace.ClearInputNames()
	var utOpts []usagetrace.UsageTraceOption

	for name, arr := range iptInfos {
		if len(arr) > 1 {
			if _, ok := arr[0].Input.(Singleton); ok {
				arr = arr[:1]
			}
		}

		inputInstanceVec.WithLabelValues(name).Set(float64(len(arr)))

		// For each input, only add once
		utOpts = append(utOpts, usagetrace.WithInputNames(name))

		for _, ii := range arr {
			RunInput(name, ii)
		}
	}

	// Notify datakit usage all the started inputs.
	usagetrace.UpdateTraceOptions(utOpts...)
	return nil
}
