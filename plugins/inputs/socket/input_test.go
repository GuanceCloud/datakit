// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package socket

import (
	"sync"
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

func TestInput_Run(t *testing.T) {
	i := &Input{
		DestURL:    []string{"tcp://47.110.144.10:443", "tcp://497.110.144.10:443", "udp://1.1.1.1:5555", "udp://1.1.1.6541:5555"},
		curTasks:   map[string]*dialer{},
		wg:         sync.WaitGroup{},
		UDPTimeOut: datakit.Duration{Duration: time.Second * 20},
		TCPTimeOut: datakit.Duration{Duration: time.Second * 20},
	}

	if err := i.Collect(); err != nil {
		l.Warnf(err.Error())
	}
}
