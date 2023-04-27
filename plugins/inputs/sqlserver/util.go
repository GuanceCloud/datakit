// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package sqlserver collects SQL Server metrics.
package sqlserver

import "net"

func setHostTagIfNotLoopback(tags map[string]string, ipAndPort string) {
	host, _, err := net.SplitHostPort(ipAndPort)
	if err != nil {
		l.Errorf("split host and port: %v", err)
		return
	}
	if host != "localhost" && !net.ParseIP(host).IsLoopback() {
		tags["host"] = host
	}
}
