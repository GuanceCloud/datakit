// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package endpoint

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/dnswatcher"
)

const (
	defaultDNSCacheFreq          = time.Minute
	defaultDNSCacheLookUpTimeout = 10 * time.Second
)

type dnsUpdateCallBackFunc func() error

type DNSCacher struct {
	domain   string
	ips      []string
	isFirst  bool
	callback dnsUpdateCallBackFunc
}

// Make sure dnsCacher implements the dnswatcher.IDNSWatcher interface.
var _ dnswatcher.IDNSWatcher = new(DNSCacher)

func (d *DNSCacher) InitDNSCache(host string, callback dnsUpdateCallBackFunc) {
	d.domain = host
	d.callback = callback

	dnswatcher.AddWatcher(d)
}

func (d *DNSCacher) GetDomain() string {
	return d.domain
}

func (d *DNSCacher) GetIPs() []string {
	return d.ips
}

func (d *DNSCacher) SetIPs(ips []string) {
	d.ips = ips
}

func (d *DNSCacher) Update() error {
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
