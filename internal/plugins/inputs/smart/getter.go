// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package smart

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/command"
)

type diskAttributeGetter interface {
	Get([]string) []byte
	ExitStatus() int
}

// smartctlGetter implements diskAttributeGetter.
type smartctlGetter struct {
	exePath,
	nocheck string
	timeout    time.Duration
	sudo       bool
	exitStatus int
}

func (x *smartctlGetter) Get(devices []string) []byte {
	var (
		args = append([]string{
			"--info",
			"--health",
			"--attributes",
			"--tolerance=verypermissive",
			"-n",
			x.nocheck,
			"--format=brief",
		}, devices...)
		err    error
		output []byte
	)

	l.Debugf("run command %s %v", x.exePath, args)
	if output, err = command.RunWithTimeout(x.timeout, x.sudo, x.exePath, args...); err != nil {
		l.Errorf("command failed: %s", err.Error())
	}

	if s, xerr := command.ExitStatus(err); xerr != nil {
		// ignored
	} else {
		x.exitStatus = s
	}
	return output
}

func (x *smartctlGetter) ExitStatus() int {
	return x.exitStatus
}
