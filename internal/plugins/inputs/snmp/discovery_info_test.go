// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package snmp

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

// go test -v -timeout 30s -run ^Test_initializeDiscovery$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp
func Test_initializeDiscovery(t *testing.T) {
	cases := []struct {
		name string
		in   map[string]*discoveryInfo
		out  map[string]*discoveryInfo
		err  error
	}{
		{
			name: "invalid_subnet",
			in: map[string]*discoveryInfo{
				"10.200.10.240": {},
			},
			out: map[string]*discoveryInfo{
				"10.200.10.240": {},
			},
			err: &net.ParseError{Type: "CIDR address", Text: "10.200.10.240"},
		},
		{
			name: "normal",
			in: map[string]*discoveryInfo{
				"10.200.10.0/24": {},
			},
			out: map[string]*discoveryInfo{
				"10.200.10.0/24": {
					Subnet:     "10.200.10.0/24",
					StartingIP: []byte{0x0a, 0xc8, 0x0a, 0x00}, // 10(0x0a).200(0xc8).10(0x0a).0(0x00)
					Network:    net.IPNet{IP: []byte{0x0a, 0xc8, 0x0a, 0x00}, Mask: []byte{0xff, 0xff, 0xff, 0x00}},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ipt := &Input{mAutoDiscovery: tc.in}
			err := ipt.initializeDiscovery()
			for k := range ipt.mAutoDiscovery {
				assert.Equal(t, tc.out[k].Subnet, ipt.mAutoDiscovery[k].Subnet)
				assert.Equal(t, tc.out[k].StartingIP, ipt.mAutoDiscovery[k].StartingIP)
				assert.Equal(t, tc.out[k].Network, ipt.mAutoDiscovery[k].Network)
			}
			assert.Equal(t, tc.err, err)
		})
	}
}

// go test -v -timeout 30s -run ^Test_addDynamicDevice$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp
func Test_addDynamicDevice(t *testing.T) {
	cases := []struct {
		name                  string
		deviceIP              string
		subnet                string
		originSpecificDevices map[string]*deviceInfo
		originDynamicDevices  map[string]*deviceInfo
		out                   map[string]bool
	}{
		{
			name:     "found specific",
			deviceIP: "10.200.10.240",
			subnet:   "10.200.10.0/24",
			originSpecificDevices: map[string]*deviceInfo{
				"10.200.10.240": {},
			},
		},
		{
			name:     "found dynamic",
			deviceIP: "10.200.10.240",
			subnet:   "10.200.10.0/24",
			originDynamicDevices: map[string]*deviceInfo{
				"10.200.10.240": {},
			},
			out: map[string]bool{
				"10.200.10.240": true,
			},
		},
		{
			name:     "init failed",
			deviceIP: "10.200.10.240",
			subnet:   "10.200.10.0/24",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ipt := &Input{mSpecificDevices: tc.originSpecificDevices}
			for k, v := range tc.originDynamicDevices {
				ipt.mDynamicDevices.Store(k, v)
			}
			ipt.addDynamicDevice(tc.deviceIP, tc.subnet)

			for k, v := range tc.out {
				_, ok := ipt.mDynamicDevices.Load(k)
				assert.Equal(t, v, ok)
			}
		})
	}
}

// go test -v -timeout 30s -run ^Test_removeDynamicDevice$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp
func Test_removeDynamicDevice(t *testing.T) {
	cases := []struct {
		name                  string
		deviceIP              string
		subnet                string
		originSpecificDevices map[string]*deviceInfo
		originDynamicDevices  map[string]*deviceInfo
		out                   map[string]bool
	}{
		{
			name:     "found specific",
			deviceIP: "10.200.10.240",
			subnet:   "10.200.10.0/24",
			originSpecificDevices: map[string]*deviceInfo{
				"10.200.10.240": {},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ipt := &Input{mSpecificDevices: tc.originSpecificDevices}
			ipt.removeDynamicDevice(tc.deviceIP)

			for k, v := range tc.originDynamicDevices {
				ipt.mDynamicDevices.Store(k, v)
			}

			for k, v := range tc.out {
				_, ok := ipt.mDynamicDevices.Load(k)
				assert.Equal(t, v, ok)
			}
		})
	}
}

// go test -v -timeout 30s -run ^Test_isIPIgnored$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp
func Test_isIPIgnored(t *testing.T) {
	cases := []struct {
		name       string
		deviceIP   string
		ignoredIPs map[string]struct{}
		out        bool
	}{
		{
			name:     "found",
			deviceIP: "10.200.10.240",
			ignoredIPs: map[string]struct{}{
				"10.200.10.240": {},
			},
			out: true,
		},
		{
			name:     "not found",
			deviceIP: "10.200.10.241",
			ignoredIPs: map[string]struct{}{
				"10.200.10.240": {},
			},
			out: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ipt := &Input{mDiscoveryIgnoredIPs: tc.ignoredIPs}
			ipt.isIPIgnored(tc.deviceIP)
		})
	}
}

// go test -v -timeout 30s -run ^Test_incrementIP$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp
func Test_incrementIP(t *testing.T) {
	cases := []struct {
		name   string
		origin string
		out    string
	}{
		{
			name:   "increase from 0",
			origin: "10.200.10.0",
			out:    "10.200.10.1",
		},
		{
			name:   "increment from 255",
			origin: "10.200.10.255",
			out:    "10.200.11.0",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			originIP := net.ParseIP(tc.origin)
			outIP := net.ParseIP(tc.out)
			incrementIP(originIP)
			assert.Equal(t, outIP, originIP)
		})
	}
}
