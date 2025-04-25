// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package windowsremote

import (
	"fmt"
	"net"
	"time"
)

func getIPsInRange(ips, cidrs []string, protocol string, targetPorts []int) []string {
	res, err := getIPsFromCIDRs(cidrs)
	if err != nil {
		l.Warn(err)
	}

	foundIPs := unique(append(res, ips...))

	var reachableIPs []string
	for _, ip := range foundIPs {
		for _, port := range targetPorts {
			l.Debugf("is port open %s", ip)
			if isPortOpen(protocol, ip, port, time.Second*3) {
				reachableIPs = append(reachableIPs, ip)
				l.Infof("add %s", ip)
			}
		}
	}
	reachableIPCountVec.WithLabelValues().Set(float64(len(reachableIPs)))

	return reachableIPs
}

func getIPsFromCIDRs(cidrs []string) ([]string, error) {
	var ips []string
	for _, cidr := range cidrs {
		ip, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, fmt.Errorf("invalid CIDR %s: %w", cidr, err)
		}
		if ip.To4() == nil {
			return nil, fmt.Errorf("only supported IPv4: %s", cidr)
		}

		broadcast := make(net.IP, len(ipNet.IP))
		for i := range ipNet.IP {
			broadcast[i] = ipNet.IP[i] | ^ipNet.Mask[i]
		}

		for current := ip.Mask(ipNet.Mask); ipNet.Contains(current); incIP(current) {
			// Skip the network address (e.g., 192.168.1.0)
			// Skip the broadcast address (e.g., 192.168.1.255)
			if !current.Equal(ipNet.IP) && !current.Equal(broadcast) {
				ips = append(ips, current.String())
			}
		}
	}
	return ips, nil
}

func incIP(ip net.IP) {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] > 0 {
			break
		}
	}
}

func isPortOpen(protocol, ip string, port int, timeout time.Duration) bool {
	address := fmt.Sprintf("%s:%d", ip, port)
	_, err := net.DialTimeout(protocol, address, timeout)
	return err == nil
}

func unique(slice []string) []string {
	result := make([]string, 0, len(slice))
	seen := make(map[string]interface{}, len(slice))

	for i := range slice {
		if _, ok := seen[slice[i]]; ok {
			continue
		}
		seen[slice[i]] = struct{}{}
		result = append(result, slice[i])
	}
	return result
}
