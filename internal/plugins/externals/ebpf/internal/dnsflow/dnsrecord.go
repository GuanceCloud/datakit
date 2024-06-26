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
	var cnameDomain string
	for _, answer := range packetInfo.Answers {
		switch answer.Type { //nolint:exhaustive
		case layers.DNSTypeA, layers.DNSTypeAAAA:
			if answer.IP == nil || answer.Name == nil {
				continue
			}
			if cnameDomain != "" {
				c.record[answer.IP.String()] = [2]interface{}{
					cnameDomain,
					packetInfo.TS,
				}
			} else {
				c.record[answer.IP.String()] = [2]interface{}{
					string(answer.Name),
					packetInfo.TS,
				}
			}

		case layers.DNSTypeCNAME:
			if cnameDomain == "" {
				cnameDomain = string(answer.Name)
			}
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
