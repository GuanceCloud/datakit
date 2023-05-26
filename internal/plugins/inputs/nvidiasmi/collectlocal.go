// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2022-present Guance, Inc.

// Package nvidiasmi collects host nvidiasmi metrics.
package nvidiasmi

import (
	"bytes"
	"os/exec"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

func (ipt *Input) collectLocal() error {
	// cycle exec all binPath (if have multiple "nvidia-smi")
	for _, binPath := range ipt.BinPaths {
		// get data （[]byte）
		data, err := ipt.getBytes(binPath)
		if err != nil {
			l.Errorf("get bytes by binPath log: %s .", err)
			// 这里，出错，data如果!=nil，可能有有价值的信息，需要继续处理。
			if data == nil {
				continue
			}
		}

		// convert and calculate GPU metrics
		_ = ipt.convert(data, "")
		/*
			// convert xml -> SMI{} struct
			smi := &SMI{}
			err = xml.Unmarshal(data, smi)
			if err != nil {
				l.Errorf("Unmarshal xml data log: %s .", err)
				continue // 不能return，否则，后面的msi运行不了
			}

			// convert to tags + fields
			metrics, metricsLog := smi.genTagsFields(ipt)

			// Append to the cache, the Run() function will handle it
			for _, metric := range metrics {
				ipt.collectCache = append(ipt.collectCache, &nvidiaSmiMeasurement{
					name:   metricName,
					tags:   metric.tags,
					fields: metric.fields,
				})
			}
			for i := 0; i < len(metricsLog); i++ {
				pt, err := point.NewPoint(
					metricName,
					metricsLog[i].tags,
					metricsLog[i].fields,
					point.LOpt(),
				)
				if err != nil {
					l.Errorf("collect gpu_smi process log: %s .", err)
				} else {
					ipt.pts = append(ipt.pts, pt)
				}
			} */
	}

	return nil
}

// Get the result of binPath execution
// @binPath One of run bin files.
func (ipt *Input) getBytes(binPath string) ([]byte, error) {
	c := exec.Command(binPath, "-q", "-x")
	// dd exec.Command ENV
	if len(ipt.Envs) != 0 {
		// in windows here will broken old PATH
		c.Env = ipt.Envs
	}

	var b bytes.Buffer
	c.Stdout = &b
	c.Stderr = &b
	if err := c.Start(); err != nil {
		return nil, err
	}
	err := datakit.WaitTimeout(c, ipt.Timeout.Duration)
	return b.Bytes(), err
}
