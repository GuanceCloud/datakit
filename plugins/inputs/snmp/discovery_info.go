// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmp

import (
	"net"
)

// discoveryInfo handles snmp discovery states.
type discoveryInfo struct {
	Ipt        *Input
	Subnet     string
	StartingIP net.IP
	Network    net.IPNet
}

func NewDiscoveryInfo(
	ipt *Input,
	subnet string,
	startingIP net.IP,
	network net.IPNet,
) *discoveryInfo {
	return &discoveryInfo{
		Ipt:        ipt,
		Subnet:     subnet,
		StartingIP: startingIP,
		Network:    network,
	}
}

func (ipt *Input) initializeDiscovery() error {
	for subnet := range ipt.mAutoDiscovery {
		ipAddr, ipNet, err := net.ParseCIDR(subnet)
		if err != nil {
			l.Errorf("subnet %s: Couldn't parse SNMP network: %s", subnet, err)
			return err
		}

		startingIP := ipAddr.Mask(ipNet.Mask)
		ipt.mAutoDiscovery[subnet] = NewDiscoveryInfo(ipt, subnet, startingIP, *ipNet)
	}

	return nil
}

func (ipt *Input) addDynamicDevice(deviceIP, subnet string) {
	l.Debugf("addDynamicDevice entry: %s", deviceIP)

	if _, ok := ipt.mSpecificDevices[deviceIP]; ok {
		l.Warnf("addDynamicDevice: device already in specific: %s", deviceIP)
		return
	}

	if _, ok := ipt.mDynamicDevices.Load(deviceIP); ok {
		l.Warnf("addDynamicDevice: device already in dynamic: %s", deviceIP)
		return
	}

	di, err := ipt.initializeDevice(deviceIP, subnet)
	if err != nil {
		l.Errorf("Input initialize failed: err = (%v), ip = (%s)", err, deviceIP)
		return
	}
	ipt.mDynamicDevices.Store(deviceIP, di)

	l.Debugf("addDynamicDevice success: %s", deviceIP)
}

func (ipt *Input) removeDynamicDevice(deviceIP string) {
	l.Debugf("removeDynamicDevice entry: %s", deviceIP)

	if _, ok := ipt.mSpecificDevices[deviceIP]; ok {
		l.Errorf("removeDynamicDevice: should not be here: %s", deviceIP)
		return
	}

	if di, ok := ipt.mDynamicDevices.Load(deviceIP); ok {
		device, ok := di.(*deviceInfo)
		if !ok {
			l.Errorf("should not be here")
			return
		}
		if err := device.Session.Close(); err != nil {
			l.Warnf("device.Session.Close failed: err = (%v), deviceIP = (%v)", err, deviceIP)
		}
		ipt.mDynamicDevices.Delete(deviceIP)
	}

	l.Debugf("removeDynamicDevice success: %s", deviceIP)
}

func (ipt *Input) isIPIgnored(ipString string) bool {
	_, present := ipt.mDiscoveryIgnoredIPs[ipString]
	return present
}

func incrementIP(ip net.IP) {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] > 0 {
			break
		}
	}
}
