// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package socket

import (
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

func TestInput_Run(t *testing.T) {
	go startTCP()
	i := &Input{
		DestURL:    []string{"tcp://47.110.144.10:443", "tcp://497.110.144.10:443", "udp://1.1.1.1:5555", "udp://1.1.1.6541:5555", "tcp://47.96.11.196:443"},
		curTasks:   map[string]*dialer{},
		wg:         sync.WaitGroup{},
		UDPTimeOut: datakit.Duration{Duration: time.Second * 20},
		TCPTimeOut: datakit.Duration{Duration: time.Second * 20},
	}

	if err := i.Collect(); err != nil {
		l.Warnf(err.Error())
	}
	for _, c := range i.collectCache {
		t.Log(c.LineProto())
	}
}

func startTCP() {
	listen, err := net.Listen("tcp", "127.0.0.1:20000")
	if err != nil {
		fmt.Println("listen failed, err:", err)
		return
	}
	conn, _ := listen.Accept()
	defer conn.Close()
}
