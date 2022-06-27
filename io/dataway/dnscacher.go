// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dnswatcher"
)

type dnsUpdateCallBackFunc func() error

type dnsCacher struct {
	domain   string
	ips      []string
	isFirst  bool
	callback dnsUpdateCallBackFunc
}

var _ dnswatcher.IDNSWatcher = new(dnsCacher)

func (d *dnsCacher) initDNSCache(host string, callback dnsUpdateCallBackFunc) {
	d.domain = host
	d.callback = callback

	dnswatcher.AddWatcher(d)
}

func (d *dnsCacher) GetDomain() string {
	return d.domain
}

func (d *dnsCacher) GetIPs() []string {
	return d.ips
}

func (d *dnsCacher) SetIPs(ips []string) {
	d.ips = ips
}

func (d *dnsCacher) Update() error {
	// first time we used for initialize
	if d.isFirst {
		d.isFirst = false
		return nil
	}

	// after we callback
	if d.callback != nil {
		if err := d.callback(); err != nil {
			return err
		}
	}
	return nil
}
