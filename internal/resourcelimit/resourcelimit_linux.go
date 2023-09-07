// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package resourcelimit

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/resourcelimit/cgroup"

func run(opt *ResourceLimitOptions) error {
	return cgroup.Run(&cgroup.CgroupOptions{
		Path:       opt.Path,
		CPUMax:     opt.CPUMax,
		MemMax:     opt.MemMax,
		DisableOOM: opt.DisableOOM,
		Enable:     opt.Enable,
	})
}

func info() string {
	return cgroup.Info()
}
