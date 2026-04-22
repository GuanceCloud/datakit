//go:build linux
// +build linux

package netflow

import (
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
