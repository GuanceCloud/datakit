// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package healthcheck

import (
	"net"
	"strings"
	"time"

	dt "github.com/GuanceCloud/cliutils/dialtesting"
	"github.com/GuanceCloud/cliutils/point"
)

const tcpMetricName = "host_tcp_exception"

func (ipt *Input) collectTCP() error {
	ts := time.Now()

	for _, tcp := range ipt.tcp {
		for _, ip := range tcp.HostPorts {
			host, port, err := net.SplitHostPort(ip)
			if err != nil {
				l.Warnf("parse ip port(%s) failed: %s", ip, err.Error())
				continue
			}
			task := dt.TCPTask{
				Host:    host,
				Port:    port,
				Timeout: tcp.ConnectionTimeOut,
				SuccessWhen: []*dt.TCPSuccess{
					{
						ResponseTime: []*dt.TCPResponseTime{
							{
								IsContainDNS: false,
								Target:       "60m",
							},
						},
					},
				},
				ExternalID: "-",
			}
			if err := task.InitDebug(); err != nil {
				l.Warnf("init tcp task failed: %s", err.Error())
				continue
			}

			if err := task.Run(); err != nil {
				l.Warnf("run tcp task failed: %s", err.Error())
			}

			tags, fields := task.GetResults()

			if tags["status"] == "FAIL" {
				if reason, ok := fields["fail_reason"].(string); ok {
					failType := "unknown-type"
					if strings.Contains(reason, "i/o timeout") {
						failType = "connection-timeout"
					} else if strings.Contains(reason, "connection refused") {
						failType = "connection-refused"
					}
					var kvs point.KVs
					kvs = kvs.Add("type", failType, true, true)
					kvs = kvs.Add("port", ip, true, true)
					kvs = kvs.Add("exception", true, false, true)

					for k, v := range ipt.mergedTags {
						kvs = kvs.AddTag(k, v)
					}

					opts := point.DefaultMetricOptions()
					opts = append(opts, point.WithTime(ts))

					ipt.collectCache = append(ipt.collectCache, point.NewPointV2(tcpMetricName, kvs, opts...))
				}
			}
		}
	}

	return nil
}
