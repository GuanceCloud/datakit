// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build windows
// +build windows

// Package cgroup wraps Linux cgroup or windws job object functions.
package resourcelimit

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/resourcelimit/job"

func run(opt *ResourceLimitOptions) error {
	return job.Run(&job.JobOptions{
		CPUMax: opt.CPUMax,
		MemMax: opt.MemMax,
	})
}

func info() string {
	return job.Info()
}
