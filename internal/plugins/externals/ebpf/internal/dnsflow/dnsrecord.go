//go:build linux
// +build linux

package dnsflow

import (
	"sync"
	"time"

	"github.com/google/gopacket/layers"
)

type DNSAnswerRecord struct {
	sync.RWMutex
	record map[string][2]interface{}
}

func (c *DNSAnswerRecord) LookupAddr(ip string) string {
	c.RLock()
	defer c.RUnlock()

	v, ok := c.record[ip]
	if !ok {
		return ""
	}

	if domian, ok := v[0].(string); ok {
		return domian
	} else {
		return ""
	}
}

func (c *DNSAnswerRecord) addRecord(packetInfo *DNSPacketInfo) {
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
				packetInfo.TS,
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

func (c *DNSAnswerRecord) Cleanup() {
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

func NewDNSRecord() *DNSAnswerRecord {
	return &DNSAnswerRecord{
		record: map[string][2]interface{}{},
	}
}
