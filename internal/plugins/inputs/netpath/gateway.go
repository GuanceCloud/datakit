// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package netpath

import (
	"net"
	"strings"
)

type probeGateway struct {
	sourceIP      string
	gatewayIP     string
	interfaceName string
	interfaceMAC  string
	netNS         string
}

type probeGatewayLookup func(destinationIP string) (probeGateway, error)

// enrichProbeGateway annotates the result with the route selected by the
// network namespace which actually ran the probe. It deliberately does not
// use the eBPF-observed source netns or source IP.
func (ipt *Input) enrichProbeGateway(res *probeResult) {
	if ipt == nil || res == nil {
		return
	}

	destinationIP := strings.TrimSpace(res.tags["probe_dest_ip"])
	if destinationIP == "" && net.ParseIP(strings.TrimSpace(res.task.Target)) != nil {
		destinationIP = strings.TrimSpace(res.task.Target)
	}
	if strings.TrimSpace(destinationIP) == "" {
		return
	}

	lookup := ipt.gatewayLookup
	if lookup == nil {
		lookup = lookupCurrentProbeGateway
	}
	via, err := lookup(destinationIP)
	if err != nil {
		return
	}
	if res.tags == nil {
		res.tags = map[string]string{}
	}
	setProbeGatewayTag(res.tags, "probe_source_ip", via.sourceIP)
	setProbeGatewayTag(res.tags, "probe_gateway_ip", via.gatewayIP)
	setProbeGatewayTag(res.tags, "probe_interface", via.interfaceName)
	setProbeGatewayTag(res.tags, "probe_interface_mac", via.interfaceMAC)
	setProbeGatewayTag(res.tags, "probe_netns", via.netNS)
}

func setProbeGatewayTag(tags map[string]string, key, value string) {
	if value != "" {
		tags[key] = value
	}
}
