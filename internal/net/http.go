// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package net

import (
	"net"
	"net/http"
	"strings"
)

func RemoteAddr(req *http.Request) (ip, port string) {
BREAKPOINT:
	for _, h := range []string{
		"x-forwarded-for",
		"X-FORWARDED-FOR",
		"X-Forwarded-For",
		"x-real-ip",
		"X-REAL-IP",
		"X-Real-Ip",
		"proxy-client-ip",
		"PROXY-CLIENT-IP",
		"Proxy-Client-Ip",
	} {
		addrs := strings.Split(req.Header.Get(h), ",")
		for _, addr := range addrs {
			if ip, port, _ = net.SplitHostPort(addr); ip == "" {
				continue
			}
			break BREAKPOINT
		}
	}
	if ip == "" {
		ip, port, _ = net.SplitHostPort(req.RemoteAddr)
	}

	return
}
