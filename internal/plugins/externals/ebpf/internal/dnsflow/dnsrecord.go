//go:build linux
// +build linux

package dnsflow

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/gopacket/layers"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/exporter"
	dknetflow "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/netflow"
)

type DNSAnswerRecord struct {
	sync.RWMutex
	record      map[string]dnsAnswerEntry
	lastCleanup time.Time
	limit       int
}

type dnsAnswerEntry struct {
	domain string
	ts     time.Time
}

const (
	dnsAnswerRecordTTL             = 10 * time.Minute
	dnsAnswerRecordCleanupInterval = time.Minute
	dnsAnswerRecordLimitEnv        = "DK_EBPF_DNSFLOW_ANSWER_RECORD_LIMIT"
	defaultDNSAnswerRecordLimit    = 65_536
	maxDNSAnswerRecordLimit        = 1_000_000
)

func (c *DNSAnswerRecord) LookupAddr(ip string) string {
	if c == nil || ip == "" {
		return ""
	}
	now := time.Now()

	c.RLock()
	defer c.RUnlock()

	v, ok := c.record[ip]
	if !ok {
		return ""
	}
	if now.Sub(v.ts) > dnsAnswerRecordTTL {
		return ""
	}
	return v.domain
}

func (c *DNSAnswerRecord) addRecord(packetInfo *DNSPacketInfo) {
	if c == nil || packetInfo == nil {
		return
	}
	now := time.Now()

	c.Lock()
	defer c.Unlock()

	if c.record == nil {
		c.record = map[string]dnsAnswerEntry{}
	}
	c.cleanupLocked(now, false)

	var cnameDomain string
	for _, answer := range packetInfo.Answers {
		switch answer.recordType { //nolint:exhaustive
		case layers.DNSTypeA, layers.DNSTypeAAAA:
			if answer.ip == "" || answer.name == "" {
				continue
			}
			ip := answer.ip
			domain := normalizeDNSDomain(answer.name)
			if _, exists := c.record[ip]; !exists && !c.allowInsertLocked(now) {
				exporter.IncBPFEventDrop("dnsflow", "answer_record", "limit")
				continue
			}
			if cnameDomain != "" {
				domain = cnameDomain
				c.record[ip] = dnsAnswerEntry{
					domain: cnameDomain,
					ts:     now,
				}
			} else {
				c.record[ip] = dnsAnswerEntry{
					domain: domain,
					ts:     now,
				}
			}
			dknetflow.RecordAddrDomain(ip, domain)

		case layers.DNSTypeCNAME:
			if cnameDomain == "" {
				cnameDomain = normalizeDNSDomain(answer.name)
			}
		default:
		}
	}
}

func (c *DNSAnswerRecord) allowInsertLocked(now time.Time) bool {
	limit := c.entryLimit()
	if limit <= 0 || len(c.record) < limit {
		return true
	}
	c.cleanupLocked(now, true)
	return len(c.record) < limit
}

func (c *DNSAnswerRecord) entryLimit() int {
	if c == nil {
		return defaultDNSAnswerRecordLimit
	}
	if c.limit <= 0 {
		c.limit = dnsAnswerRecordLimit()
	}
	return c.limit
}

func dnsAnswerRecordLimit() int {
	raw := strings.TrimSpace(os.Getenv(dnsAnswerRecordLimitEnv))
	if raw == "" {
		return defaultDNSAnswerRecordLimit
	}
	limit, err := strconv.Atoi(raw)
	if err != nil || limit <= 0 {
		l.Warnf("invalid %s=%q, use default %d", dnsAnswerRecordLimitEnv, raw, defaultDNSAnswerRecordLimit)
		return defaultDNSAnswerRecordLimit
	}
	if limit > maxDNSAnswerRecordLimit {
		return maxDNSAnswerRecordLimit
	}
	return limit
}

func normalizeDNSDomain(domain string) string {
	domain = strings.TrimSpace(strings.ToLower(domain))
	domain = strings.TrimSuffix(domain, ".")
	return domain
}

func (c *DNSAnswerRecord) Cleanup() {
	if c == nil {
		return
	}
	c.Lock()
	defer c.Unlock()
	c.cleanupLocked(time.Now(), true)
}

func (c *DNSAnswerRecord) cleanupLocked(now time.Time, force bool) {
	if c == nil {
		return
	}
	if !force && !c.lastCleanup.IsZero() && now.Sub(c.lastCleanup) < dnsAnswerRecordCleanupInterval {
		return
	}
	c.lastCleanup = now

	for k, v := range c.record {
		if now.Sub(v.ts) > dnsAnswerRecordTTL {
			delete(c.record, k)
		}
	}
}

func NewDNSRecord() *DNSAnswerRecord {
	return &DNSAnswerRecord{
		record: map[string]dnsAnswerEntry{},
		limit:  dnsAnswerRecordLimit(),
	}
}
