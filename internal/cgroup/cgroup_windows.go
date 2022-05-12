// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cgroup

type Cgroup struct {
	opt *CgroupOptions

	err error
}

func (c *Cgroup) start() {
	l.Infof("not support windows system, exit")
}
