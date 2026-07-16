//go:build linux
// +build linux

package dnsflow

import (
	"testing"
	"time"

	"github.com/google/gopacket/layers"
)

func TestDNSAnswerRecordLookupAndCleanup(t *testing.T) {
	record := NewDNSRecord()
	record.addRecord(&DNSPacketInfo{
		Answers: []dnsAnswer{
			{
				recordType: layers.DNSTypeA,
				name:       "Example.COM.",
				ip:         "10.1.2.3",
			},
		},
	})

	if got := record.LookupAddr("10.1.2.3"); got != "example.com" {
		t.Fatalf("LookupAddr = %q, want example.com", got)
	}

	record.record["10.9.8.7"] = dnsAnswerEntry{
		domain: "old.example.com",
		ts:     time.Now().Add(-dnsAnswerRecordTTL - time.Second),
	}
	record.Cleanup()

	if got := record.LookupAddr("10.9.8.7"); got != "" {
		t.Fatalf("expired LookupAddr = %q, want empty", got)
	}
}

func TestDNSAnswerRecordAddCleanupIsRateLimited(t *testing.T) {
	record := NewDNSRecord()
	record.record["10.9.8.7"] = dnsAnswerEntry{
		domain: "old.example.com",
		ts:     time.Now().Add(-dnsAnswerRecordTTL - time.Second),
	}
	record.lastCleanup = time.Now()

	record.addRecord(&DNSPacketInfo{
		Answers: []dnsAnswer{
			{
				recordType: layers.DNSTypeA,
				name:       "Example.COM.",
				ip:         "10.1.2.3",
			},
		},
	})

	if _, ok := record.record["10.9.8.7"]; !ok {
		t.Fatal("expected addRecord cleanup to be skipped inside cleanup interval")
	}

	record.lastCleanup = time.Now().Add(-dnsAnswerRecordCleanupInterval - time.Second)
	record.addRecord(&DNSPacketInfo{
		Answers: []dnsAnswer{
			{
				recordType: layers.DNSTypeA,
				name:       "Other.COM.",
				ip:         "10.1.2.4",
			},
		},
	})

	if _, ok := record.record["10.9.8.7"]; ok {
		t.Fatal("expected addRecord cleanup after cleanup interval")
	}
}

func TestDNSAnswerRecordLimitDropsNewIPsAndKeepsExisting(t *testing.T) {
	record := &DNSAnswerRecord{
		record: map[string]dnsAnswerEntry{},
		limit:  1,
	}

	record.addRecord(&DNSPacketInfo{
		Answers: []dnsAnswer{
			{
				recordType: layers.DNSTypeA,
				name:       "First.COM.",
				ip:         "10.1.2.3",
			},
		},
	})
	record.addRecord(&DNSPacketInfo{
		Answers: []dnsAnswer{
			{
				recordType: layers.DNSTypeA,
				name:       "Second.COM.",
				ip:         "10.1.2.4",
			},
		},
	})

	if got := record.LookupAddr("10.1.2.4"); got != "" {
		t.Fatalf("over-limit LookupAddr = %q, want empty", got)
	}

	record.addRecord(&DNSPacketInfo{
		Answers: []dnsAnswer{
			{
				recordType: layers.DNSTypeA,
				name:       "First-New.COM.",
				ip:         "10.1.2.3",
			},
		},
	})
	if got := record.LookupAddr("10.1.2.3"); got != "first-new.com" {
		t.Fatalf("existing LookupAddr = %q, want first-new.com", got)
	}
}

func TestDNSAnswerRecordZeroValueIsUsable(t *testing.T) {
	var record DNSAnswerRecord
	record.addRecord(&DNSPacketInfo{
		Answers: []dnsAnswer{
			{
				recordType: layers.DNSTypeA,
				name:       "Example.COM.",
				ip:         "10.1.2.3",
			},
		},
	})

	if got := record.LookupAddr("10.1.2.3"); got != "example.com" {
		t.Fatalf("LookupAddr = %q, want example.com", got)
	}
}

func TestDNSAnswerRecordLimitEnv(t *testing.T) {
	t.Setenv(dnsAnswerRecordLimitEnv, "bad")
	if got := dnsAnswerRecordLimit(); got != defaultDNSAnswerRecordLimit {
		t.Fatalf("invalid answer record limit = %d, want %d", got, defaultDNSAnswerRecordLimit)
	}

	t.Setenv(dnsAnswerRecordLimitEnv, "9999999")
	if got := dnsAnswerRecordLimit(); got != maxDNSAnswerRecordLimit {
		t.Fatalf("clamped answer record limit = %d, want %d", got, maxDNSAnswerRecordLimit)
	}
}
