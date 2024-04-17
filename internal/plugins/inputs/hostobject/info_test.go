// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package hostobject

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetNetInfo(t *testing.T) {
	ifs, err := interfaces()
	if err != nil {
		l.Errorf("fail to get interfaces, %s", err)
	}
	var infos []*NetInfo

	// netVIfaces := map[string]bool{}
	netVIfaces, _ := NetIgnoreIfaces()

	for _, it := range ifs {
		if _, ok := netVIfaces[it.Name]; ok {
			continue
		}
		i := &NetInfo{
			Index:        it.Index,
			MTU:          it.MTU,
			Name:         it.Name,
			HardwareAddr: it.HardwareAddr,
			Flags:        it.Flags,
		}
		for _, ad := range it.Addrs {
			ip, _, _ := net.ParseCIDR(ad.Addr)
			if ip.IsLoopback() {
				continue
			}
			if ip.To4() != nil {
				i.IP4 = ad.Addr
				i.IP4All = append(i.IP4All, ad.Addr)
			} else if ip.To16() != nil {
				i.IP6 = ad.Addr
				i.IP6All = append(i.IP6All, ad.Addr)
			}
		}
		infos = append(infos, i)
	}
	assert.NotEmpty(t, infos, "infos should not be empty")
}
