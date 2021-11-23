//go:build (linux && ignore) || ebpf
// +build linux,ignore ebpf

package dnsflow

import (
	"net"
	"sync"
	"time"

	"github.com/google/gopacket/layers"
)

type DNSRecord struct {
	sync.RWMutex
	record map[string][2]interface{}
}

func (c *DNSRecord) LookupAddr(ip net.IP) string {
	c.RLock()
	defer c.RUnlock()

	if ip == nil {
		return ""
	}
	ipStr := ip.String()
	v, ok := c.record[ipStr]
	if !ok {
		return ""
	}

	if domian, ok := v[0].(string); ok {
		return domian
	} else {
		return ""
	}
}

func (c *DNSRecord) addRecord(packetInfo *DNSPacketInfo) {
	c.Lock()
	defer c.Unlock()
	for _, answer := range packetInfo.Answers {
		switch answer.Type {
		case layers.DNSTypeA, layers.DNSTypeAAAA:
			if answer.IP == nil || answer.Name == nil {
				continue
			}
			c.record[answer.IP.String()] = [2]interface{}{
				string(answer.Name),
				packetInfo.TS.Add(time.Second * time.Duration(answer.TTL)),
			}
		case layers.DNSTypeCNAME, layers.DNSTypeHINFO, layers.DNSTypeMB,
			layers.DNSTypeMD, layers.DNSTypeMF, layers.DNSTypeMG,
			layers.DNSTypeMINFO, layers.DNSTypeMR, layers.DNSTypeMX,
			layers.DNSTypeNS, layers.DNSTypeNULL, layers.DNSTypeOPT,
			layers.DNSTypePTR, layers.DNSTypeSOA, layers.DNSTypeSRV,
			layers.DNSTypeTXT, layers.DNSTypeURI, layers.DNSTypeWKS:
		default:
		}
	}
}

func (c *DNSRecord) Cleanup() {
	c.Lock()
	defer c.Unlock()
	for k, v := range c.record {
		if ts, ok := v[0].(time.Time); ok {
			if time.Until(ts) < 0 {
				delete(c.record, k)
			}
		}
	}
}

func NewDNSRecord() *DNSRecord {
	return &DNSRecord{
		record: map[string][2]interface{}{},
	}
}
