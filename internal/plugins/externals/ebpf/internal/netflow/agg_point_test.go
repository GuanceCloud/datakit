//go:build linux
// +build linux

package netflow

import (
	"strconv"
	"testing"
	"time"
)

func TestKV2PointAddsTCPHealthFields(t *testing.T) {
	key := &aggKey{
		sAddr:       [4]uint32{0, 0, 0, 0x0100007f},
		dAddr:       [4]uint32{0, 0, 0, 0x0200007f},
		sPort:       43210,
		dPort:       443,
		transport:   transportTCP,
		netns:       42,
		family:      "IPv4",
		direction:   DirectionOutgoing,
		processName: "curl",
		sType:       "loopback",
		dType:       "loopback",
	}
	value := &aggValue{
		bytesRead:      100,
		bytesWritten:   200,
		packetsRead:    3,
		packetsWrite:   5,
		retransmits:    2,
		rtt:            1200,
		rttVar:         300,
		tcpClosed:      1,
		tcpEstablished: 4,
		tcpConnects:    2,
		tcpFailures:    1,
		tcpCloseWait:   1,
		tcpLastAck:     1,
		tcpTimeWait:    1,
		count:          2,
	}

	pt, err := kv2point(key, value, time.Unix(1, 0), nil, nil)
	if err != nil {
		t.Fatalf("kv2point failed: %v", err)
	}

	if got := pt.Get("packets_read"); got != int64(3) {
		t.Fatalf("unexpected packets_read %v", got)
	}
	if got := pt.Get("packets_written"); got != int64(5) {
		t.Fatalf("unexpected packets_written %v", got)
	}
	if got := pt.Get("tcp_closed"); got != int64(1) {
		t.Fatalf("unexpected tcp_closed %v", got)
	}
	if got := pt.Get("tcp_established"); got != int64(4) {
		t.Fatalf("unexpected tcp_established %v", got)
	}
	if got := pt.Get("tcp_connect_attempts"); got != int64(2) {
		t.Fatalf("unexpected tcp_connect_attempts %v", got)
	}
	if got := pt.Get("tcp_connect_failures"); got != int64(1) {
		t.Fatalf("unexpected tcp_connect_failures %v", got)
	}
	if got := pt.Get("tcp_close_wait"); got != int64(1) {
		t.Fatalf("unexpected tcp_close_wait %v", got)
	}
	if got := pt.Get("tcp_last_ack"); got != int64(1) {
		t.Fatalf("unexpected tcp_last_ack %v", got)
	}
	if got := pt.Get("tcp_time_wait"); got != int64(1) {
		t.Fatalf("unexpected tcp_time_wait %v", got)
	}
}

func TestKV2PointAddsDstDomainFromSharedRecord(t *testing.T) {
	RecordPeerDomain("127.0.0.2", 443, transportTCP, "42", "api.example.com")

	key := &aggKey{
		sAddr:       [4]uint32{0, 0, 0, 0x0100007f},
		dAddr:       [4]uint32{0, 0, 0, 0x0200007f},
		sPort:       43210,
		dPort:       443,
		transport:   transportTCP,
		netns:       42,
		family:      "IPv4",
		direction:   DirectionOutgoing,
		processName: "curl",
		sType:       "loopback",
		dType:       "loopback",
	}
	value := &aggValue{
		bytesRead:    100,
		bytesWritten: 200,
		count:        1,
	}

	pt, err := kv2point(key, value, time.Unix(1, 0), nil, nil)
	if err != nil {
		t.Fatalf("kv2point failed: %v", err)
	}
	if got := pt.GetTag("dst_domain"); got != "api.example.com" {
		t.Fatalf("unexpected dst_domain %q", got)
	}
}

func TestAddrDomainRecordCleansExpiredPeerRecords(t *testing.T) {
	record := &addrDomainRecord{
		ipRecord:   map[string]addrDomainEntry{},
		peerRecord: map[peerDomainKey]addrDomainEntry{},
	}
	record.peerRecord[peerDomainKey{
		ip:        "10.1.2.3",
		port:      443,
		transport: transportTCP,
		netns:     "42",
	}] = addrDomainEntry{
		domain: "old.example.com",
		ts:     time.Now().Add(-addrDomainTTL - time.Second),
	}
	record.lastCleanup = time.Now().Add(-addrDomainCleanupInterval - time.Second)

	record.RecordPeerDomain("10.1.2.4", 443, transportTCP, "42", "new.example.com")

	if len(record.peerRecord) != 1 {
		t.Fatalf("peerRecord len = %d, want 1", len(record.peerRecord))
	}
	if got := record.LookupPeerDomain("10.1.2.4", 443, transportTCP, "42"); got != "new.example.com" {
		t.Fatalf("LookupPeerDomain = %q, want new.example.com", got)
	}
}

func TestAddrDomainRecordZeroValueIsUsable(t *testing.T) {
	var record addrDomainRecord

	record.RecordAddrDomain("10.1.2.3", "api.example.com")
	if got := record.LookupPeerDomain("10.1.2.3", 443, transportTCP, "42"); got != "api.example.com" {
		t.Fatalf("LookupPeerDomain by IP = %q, want api.example.com", got)
	}

	record.RecordPeerDomain("10.1.2.4", 443, transportTCP, "42", "peer.example.com")
	if got := record.LookupPeerDomain("10.1.2.4", 443, transportTCP, "42"); got != "peer.example.com" {
		t.Fatalf("LookupPeerDomain by peer = %q, want peer.example.com", got)
	}
}

func TestAddrDomainRecordLimit(t *testing.T) {
	record := &addrDomainRecord{
		ipRecord:   map[string]addrDomainEntry{},
		peerRecord: map[peerDomainKey]addrDomainEntry{},
		limit:      2,
	}

	record.RecordAddrDomain("10.1.2.1", "one.example.com")
	record.RecordPeerDomain("10.1.2.2", 443, transportTCP, "42", "two.example.com")
	record.RecordPeerDomain("10.1.2.3", 443, transportTCP, "42", "three.example.com")

	if got := record.LookupPeerDomain("10.1.2.3", 443, transportTCP, "42"); got != "" {
		t.Fatalf("expected over-limit peer domain to be dropped, got %q", got)
	}

	record.RecordPeerDomain("10.1.2.2", 443, transportTCP, "42", "two-new.example.com")
	if got := record.LookupPeerDomain("10.1.2.2", 443, transportTCP, "42"); got != "two-new.example.com" {
		t.Fatalf("expected existing peer domain update, got %q", got)
	}
}

func TestSrcIPPortRecorderLimitAndTTL(t *testing.T) {
	now := time.Unix(1000, 0)
	rec := &srcIPPortRecorder{
		Record: map[[4]uint32]IPPortRecord{},
		limit:  2,
	}
	ip1 := [4]uint32{1}
	ip2 := [4]uint32{2}
	ip3 := [4]uint32{3}

	if !rec.insertAndUpdateAt(ip1, now) {
		t.Fatal("insert ip1 failed")
	}
	if !rec.insertAndUpdateAt(ip2, now) {
		t.Fatal("insert ip2 failed")
	}
	if rec.insertAndUpdateAt(ip3, now.Add(time.Second)) {
		t.Fatal("insert ip3 succeeded while recorder is full")
	}
	if len(rec.Record) != 2 {
		t.Fatalf("recorder len = %d, want 2", len(rec.Record))
	}
	if _, err := rec.queryAt(ip1, now.Add(cleanIPPortDur+time.Second)); err == nil {
		t.Fatal("expected stale ip1 query to miss")
	}
	if !rec.insertAndUpdateAt(ip3, now.Add(cleanIPPortDur+time.Second)) {
		t.Fatal("insert ip3 after TTL cleanup failed")
	}
	if len(rec.Record) != 1 {
		t.Fatalf("recorder len after cleanup = %d, want 1", len(rec.Record))
	}
}

func TestSrcIPPortRecorderLimitEnv(t *testing.T) {
	t.Setenv(srcIPPortRecorderLimitEnv, "bad")
	if got := srcIPPortRecorderLimit(); got != defaultSrcIPPortRecorderLimit {
		t.Fatalf("invalid env limit = %d, want %d", got, defaultSrcIPPortRecorderLimit)
	}

	t.Setenv(srcIPPortRecorderLimitEnv, strconv.Itoa(maxSrcIPPortRecorderLimit+1))
	if got := srcIPPortRecorderLimit(); got != maxSrcIPPortRecorderLimit {
		t.Fatalf("clamped env limit = %d, want %d", got, maxSrcIPPortRecorderLimit)
	}
}
