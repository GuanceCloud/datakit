// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package netpath

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"syscall"

	"github.com/vishvananda/netlink"
)

func lookupCurrentProbeGateway(destinationIP string) (probeGateway, error) {
	destination := net.ParseIP(destinationIP)
	if destination == nil {
		return probeGateway{}, fmt.Errorf("invalid destination IP %q", destinationIP)
	}

	routes, err := netlink.RouteGet(destination)
	if err != nil {
		return probeGateway{}, fmt.Errorf("look up route to %s: %w", destinationIP, err)
	}
	if len(routes) == 0 {
		return probeGateway{}, fmt.Errorf("no route to %s", destinationIP)
	}

	route := routes[0]
	via := probeGateway{
		sourceIP:  routeIPString(route.Src),
		gatewayIP: routeIPString(route.Gw),
		netNS:     currentNetworkNamespace(),
	}
	if route.LinkIndex <= 0 {
		return via, nil
	}

	link, err := netlink.LinkByIndex(route.LinkIndex)
	if err != nil {
		return via, nil
	}
	attrs := link.Attrs()
	if attrs == nil {
		return via, nil
	}
	via.interfaceName = attrs.Name
	via.interfaceMAC = attrs.HardwareAddr.String()
	return via, nil
}

func routeIPString(ip net.IP) string {
	if ip == nil || ip.IsUnspecified() {
		return ""
	}
	return ip.String()
}

func currentNetworkNamespace() string {
	info, err := os.Stat("/proc/self/ns/net")
	if err != nil {
		return ""
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return ""
	}
	return strconv.FormatUint(stat.Ino, 10)
}
