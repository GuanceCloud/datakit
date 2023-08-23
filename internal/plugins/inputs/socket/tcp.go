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
)

type TCPTask struct {
	Host string
	Port string

	reqCost    time.Duration
	reqDNSCost time.Duration
	reqError   string
	timeout    time.Duration
}

func (t *TCPTask) GetResults() (tags map[string]string, fields map[string]interface{}) {
	tags = map[string]string{
		"dest_host": t.Host,
		"dest_port": t.Port,
		"proto":     "tcp",
	}

	responseTime := int64(t.reqCost) / 1000                     // us
	responseTimeWithDNS := int64(t.reqCost+t.reqDNSCost) / 1000 // us

	fields = map[string]interface{}{
		"response_time":          responseTime,
		"response_time_with_dns": responseTimeWithDNS,
		"success":                int64(-1),
	}

	message := map[string]interface{}{}

	var reasons []string

	if t.reqError != "" {
		reasons = append(reasons, t.reqError)
	}

	if len(reasons) == 0 {
		message["response_time_in_micros"] = responseTime
	}

	if t.reqError == "" && len(reasons) == 0 {
		fields["success"] = int64(1)
	}

	return tags, fields
}

func (t *TCPTask) Clear() {
	t.reqCost = 0
	t.reqError = ""
}

func (t *TCPTask) Run() error {
	t.Clear()

	var d net.Dialer
	ctx, cancel := context.WithTimeout(context.Background(), t.timeout)
	defer cancel()

	hostIP := net.ParseIP(t.Host)

	if hostIP == nil { // host name
		start := time.Now()
		if ips, err := net.LookupIP(t.Host); err != nil {
			t.reqError = err.Error()
			return err
		} else {
			if len(ips) == 0 {
				err := fmt.Errorf("invalid host: %s, found no ip record", t.Host)
				t.reqError = err.Error()
				return err
			} else {
				t.reqDNSCost = time.Since(start)
				hostIP = ips[0]
			}
		}
	}

	tcpIPAddr := net.JoinHostPort(hostIP.String(), t.Port)

	start := time.Now()
	conn, err := d.DialContext(ctx, "tcp", tcpIPAddr)
	if err != nil {
		t.reqError = err.Error()
		t.reqDNSCost = 0
	} else {
		t.reqCost = time.Since(start)
		err := conn.Close()
		if err != nil {
			return fmt.Errorf("socket input close connection fail : %w", err)
		}
	}
	return nil
}

// nolint
func (i *Input) runTCP(t *TCPTask) error {
	err := t.Run() //nolint:errcheck
	// 无论成功或失败，都要记录测试结果
	i.feedMeasurement(t)
	if err != nil {
		return err
	}
	return nil
}

func (i *Input) feedMeasurement(t *TCPTask) {
	tags, fields := t.GetResults()
	ts := time.Now()
	i.collectCache = append(i.collectCache, &TCPMeasurement{name: "tcp", tags: tags, fields: fields, ts: ts})
}
