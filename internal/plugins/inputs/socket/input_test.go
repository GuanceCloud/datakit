// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package socket

import (
	"fmt"
	"net"
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

func TestInput_Run(t *testing.T) {
	go startTCP()
	i := &Input{
		DestURL:    []string{"tcp://127.0.0.1:22", "tcp://127.0.0.1:20000", "udp://1.1.1.1:5555"},
		UDPTimeOut: datakit.Duration{Duration: time.Second * 10},
		TCPTimeOut: datakit.Duration{Duration: time.Second * 10},
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
