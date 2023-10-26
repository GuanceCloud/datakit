// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package socket

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
)

type tcpTask struct {
	reqCost    time.Duration
	reqDNSCost time.Duration
	timeout    time.Duration

	Host string
	Port string
}

func (t *tcpTask) getResults() *point.Point {
	var kvs point.KVs

	kvs = kvs.MustAddTag("dest_host", t.Host)
	kvs = kvs.MustAddTag("dest_port", t.Port)
	kvs = kvs.MustAddTag("proto", "tcp")

	responseTime := int64(t.reqCost) / 1000                     // us
	responseTimeWithDNS := int64(t.reqCost+t.reqDNSCost) / 1000 // us

	// fields
	kvs = kvs.Add("response_time", responseTime, false, true)
	kvs = kvs.Add("response_time_with_dns", responseTimeWithDNS, false, true)
	kvs = kvs.Add("success", int64(1), false, true) // default set ok

	return point.NewPointV2("tcp", kvs, point.DefaultMetricOptions()...)
}

func (t *tcpTask) run() error {
	var d net.Dialer
	ctx, cancel := context.WithTimeout(context.Background(), t.timeout)
	defer cancel()

	hostIP := net.ParseIP(t.Host)

	if hostIP == nil { // host name
		start := time.Now()
		if ips, err := net.LookupIP(t.Host); err != nil {
			l.Errorf("LookupIP: %s", err)
			return err
		} else {
			if len(ips) == 0 {
				return fmt.Errorf("invalid host: %s, found no ip record", t.Host)
			} else {
				t.reqDNSCost = time.Since(start)
				hostIP = ips[0]
			}
		}
	}

	tcpIPAddr := net.JoinHostPort(hostIP.String(), t.Port)

	start := time.Now()
	if conn, err := d.DialContext(ctx, "tcp", tcpIPAddr); err != nil {
		l.Errorf("DialContext: %s", err)
		return err
	} else {
		t.reqCost = time.Since(start)
		return conn.Close()
	}
}

// nolint
func (i *input) runTCP(t *tcpTask) *point.Point {
	err := t.run()

	pt := t.getResults()

	if err != nil {
		l.Warnf("TCP run: %s, ignored", err)

		pt.MustAdd("success", int64(-1))

		i.feeder.FeedLastError(err.Error(),
			io.WithLastErrorInput(inputName),
			io.WithLastErrorCategory(point.Metric))
	}

	return pt
}

func (i *input) collectTCP(destHost string, destPort string) *point.Point {
	t := &tcpTask{
		Host:    destHost,
		Port:    destPort,
		timeout: i.TCPTimeOut.Duration,
	}

	return i.runTCP(t)
}
