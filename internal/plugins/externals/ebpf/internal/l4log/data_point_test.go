//go:build linux
// +build linux

package l4log

import (
	"strings"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/netflow"
)

func TestBuildTCPLogSetsRetransmitFields(t *testing.T) {
	chunk := &PktChunk{
		ChunkID:       1,
		RetransmitsTx: 3,
		RetransmitsRx: 5,
		TCPSreries: []PktTCPHdr{
			{TS: 10},
		},
	}
	value := &PValue{
		tcpInfo: TCPLog{
			l7proto:        L7ProtoHTTP,
			RetransmitsSYN: 2,
		},
	}

	kvs, _, ok, err := buildTCPLog(chunk, time.Now().UnixNano(), point.NewTags(map[string]string{
		"src_ip": "10.0.0.1",
	}), value)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected buildTCPLog to produce a point")
	}
	if got := kvs.Get("tx_retrans"); got == nil || got.GetI() != 3 {
		t.Fatalf("unexpected tx_retrans field: %+v", got)
	}
	if got := kvs.Get("rx_retrans"); got == nil || got.GetI() != 5 {
		t.Fatalf("unexpected rx_retrans field: %+v", got)
	}
}

func TestBuildTCPLogPreservesMACMapInMessage(t *testing.T) {
	chunk := &PktChunk{
		ChunkID: 1,
		TCPSreries: []PktTCPHdr{
			{TS: 10},
		},
	}
	if got := chunk.GetMacID("aa:bb:cc:dd:ee:ff"); got != "1" {
		t.Fatalf("unexpected first mac id: %s", got)
	}
	if got := chunk.GetMacID("11:22:33:44:55:66"); got != "2" {
		t.Fatalf("unexpected second mac id: %s", got)
	}

	value := &PValue{
		tcpInfo: TCPLog{
			l7proto: L7ProtoHTTP,
		},
	}
	kvs, _, ok, err := buildTCPLog(chunk, time.Now().UnixNano(), point.NewTags(map[string]string{
		"src_ip": "10.0.0.1",
	}), value)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected buildTCPLog to produce a point")
	}
	msg := kvs.Get("message")
	if msg == nil || !strings.Contains(msg.GetS(), "\"mac_map\":{\"11:22:33:44:55:66\":\"2\",\"aa:bb:cc:dd:ee:ff\":\"1\"}") && !strings.Contains(msg.GetS(), "\"mac_map\":{\"aa:bb:cc:dd:ee:ff\":\"1\",\"11:22:33:44:55:66\":\"2\"}") {
		t.Fatalf("unexpected message field: %+v", msg)
	}
	if chunk.MACMap != nil {
		t.Fatalf("expected temporary MACMap to be released after serialization")
	}
}

func TestFeedNetworkLogDirtyOnly(t *testing.T) {
	oldEnableNetlog := enableNetlog
	enableNetlog = false
	defer func() {
		enableNetlog = oldEnableNetlog
	}()

	keyDirty := PMeta{SrcIP: "10.0.0.1", DstIP: "10.0.0.2", SrcPort: 1000, DstPort: 2000}
	keyIdle := PMeta{SrcIP: "10.0.0.3", DstIP: "10.0.0.4", SrcPort: 3000, DstPort: 4000}

	valDirty := &PValue{
		sMACEQ: true,
		tcpInfo: TCPLog{
			direction: directionOutgoing,
			metric: TCPMetrics{
				BytesWritten: 11,
			},
		},
	}
	valIdle := &PValue{
		sMACEQ: true,
		tcpInfo: TCPLog{
			direction: directionOutgoing,
			metric: TCPMetrics{
				BytesWritten: 22,
			},
		},
	}

	pool := newConnMap()
	pool.insert(keyDirty, valDirty)
	pool.insert(keyIdle, valIdle)
	pool.markDirty(keyDirty)

	conns := &TCPConns{nsUID: "42"}
	conns.feedNetworkLog(pool, false, false, false, []string{"10.0.0.1", "10.0.0.3"})

	if got := valDirty.tcpInfo.metric.BytesWritten; got != 0 {
		t.Fatalf("dirty flow bytes_written = %d, want 0", got)
	}
	if got := valIdle.tcpInfo.metric.BytesWritten; got != 22 {
		t.Fatalf("idle flow bytes_written = %d, want 22", got)
	}
	if got := conns.agg.Len(); got != 1 {
		t.Fatalf("agg len = %d, want 1", got)
	}
}

func TestBuildCommKVsCache(t *testing.T) {
	key := &PMeta{SrcIP: "10.0.0.1", DstIP: "10.0.0.2", SrcPort: 1000, DstPort: 2000}
	val := &PValue{
		sMACEQ: true,
		tcpInfo: TCPLog{
			direction: directionOutgoing,
			synSeq:    1,
			synAckSeq: 2,
		},
	}
	conns := &TCPConns{
		nsUID:        "42",
		ifaceNameMAC: [2]string{"eth0", "aa:bb:cc:dd:ee:ff"},
		tags:         map[string]string{"service": "demo"},
	}

	kvs1 := buildCommKVs(key, val, conns)
	kvs2 := buildCommKVs(key, val, conns)
	if len(kvs1) == 0 || len(kvs2) == 0 {
		t.Fatal("expected cached kvs to be populated")
	}
	if kvs1.Get("service") == nil || kvs1.Get("service").GetS() != "demo" {
		t.Fatalf("unexpected service tag in first kvs: %+v", kvs1.Get("service"))
	}
	if kvs2.Get("service") == nil || kvs2.Get("service").GetS() != "demo" {
		t.Fatalf("unexpected service tag in cached kvs: %+v", kvs2.Get("service"))
	}

	val.tcpInfo.direction = directionIncoming
	kvs3 := buildCommKVs(key, val, conns)
	if len(kvs3) == 0 {
		t.Fatal("expected kvs after direction change")
	}
	if got := kvs3.Get("conn_side"); got == nil || got.GetS() != "server" {
		t.Fatalf("expected direction change to rebuild conn_side, got %+v", got)
	}
}

func TestBuildCommKVsRefreshesRuntimeTags(t *testing.T) {
	key := &PMeta{SrcIP: "10.0.0.1", DstIP: "10.0.0.2", SrcPort: 1000, DstPort: 2000}
	val := &PValue{
		sMACEQ: true,
		tcpInfo: TCPLog{
			direction: directionOutgoing,
			synSeq:    1,
			synAckSeq: 2,
		},
	}
	conns := &TCPConns{
		nsUID:        "42",
		ifaceNameMAC: [2]string{"eth0", "aa:bb:cc:dd:ee:ff"},
		tags:         map[string]string{"service": "demo"},
	}

	kvs1 := buildCommKVs(key, val, conns)
	if got := kvs1.Get("service"); got == nil || got.GetS() != "demo" {
		t.Fatalf("unexpected initial service tag: %+v", got)
	}

	conns.UpdateTags(map[string]string{"service": "checkout", "cluster": "prod"})

	kvs2 := buildCommKVs(key, val, conns)
	if got := kvs2.Get("service"); got == nil || got.GetS() != "checkout" {
		t.Fatalf("unexpected refreshed service tag: %+v", got)
	}
	if got := kvs2.Get("cluster"); got == nil || got.GetS() != "prod" {
		t.Fatalf("unexpected refreshed cluster tag: %+v", got)
	}
}

func TestBuildCommKVsAddsDstDomainDynamically(t *testing.T) {
	key := &PMeta{SrcIP: "10.0.0.1", DstIP: "10.0.0.9", SrcPort: 1000, DstPort: 443}
	val := &PValue{
		sMACEQ: true,
		tcpInfo: TCPLog{
			direction: directionOutgoing,
			synSeq:    1,
			synAckSeq: 2,
		},
	}
	conns := &TCPConns{
		nsUID:        "42",
		ifaceNameMAC: [2]string{"eth0", "aa:bb:cc:dd:ee:ff"},
		tags:         map[string]string{"service": "demo"},
	}

	kvs1 := buildCommKVs(key, val, conns)
	if got := kvs1.Get("dst_domain"); got != nil {
		t.Fatalf("unexpected initial dst_domain %+v", got)
	}

	netflow.RecordPeerDomain("10.0.0.9", 443, "tcp", "42", "edge.example.com")

	kvs2 := buildCommKVs(key, val, conns)
	if got := kvs2.Get("dst_domain"); got == nil || got.GetS() != "edge.example.com" {
		t.Fatalf("unexpected dynamic dst_domain %+v", got)
	}
	if got := kvs2.Get("server_domain"); got == nil || got.GetS() != "edge.example.com" {
		t.Fatalf("unexpected dynamic server_domain %+v", got)
	}
}

func TestBuildHTTPLogRecordsHostDomainImmediately(t *testing.T) {
	key := &PMeta{SrcIP: "10.1.0.1", DstIP: "10.1.0.9", SrcPort: 12345, DstPort: 8080}
	value := &PValue{}
	elem := &HTTPLogElem{
		Method:        "GET",
		Path:          "/health",
		Host:          "frontend.example.com",
		Direction:     DOutging,
		txFirstByteTS: 1,
		messageDirty:  true,
	}

	kvs, _, ok, err := buildHTTPLog(key, value, elem, point.NewTags(map[string]string{
		"src_ip": key.SrcIP,
	}), nil, "ns-http-host", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected buildHTTPLog to produce a point")
	}
	if got := kvs.Get("dst_domain"); got == nil || got.GetS() != "frontend.example.com" {
		t.Fatalf("unexpected dst_domain: %+v", got)
	}
	if got := netflow.LookupPeerDomain(key.DstIP, uint32(key.DstPort), "tcp", "ns-http-host"); got != "frontend.example.com" {
		t.Fatalf("unexpected recorded peer domain: %q", got)
	}
}

func TestBuildH2LogRecordsHostDomainImmediately(t *testing.T) {
	key := &PMeta{SrcIP: "10.2.0.1", DstIP: "10.2.0.9", SrcPort: 23456, DstPort: 8443}
	value := &PValue{}
	elem := &HTTP2LogElem{
		Method:        "GET",
		Path:          "/v1/orders",
		Host:          "api.example.com",
		Direction:     DOutging,
		txFirstByteTS: 1,
		messageDirty:  true,
	}

	kvs, _, ok, err := buildH2Log(key, value, elem, point.NewTags(map[string]string{
		"src_ip": key.SrcIP,
	}), nil, "ns-h2-host", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected buildH2Log to produce a point")
	}
	if got := kvs.Get("dst_domain"); got == nil || got.GetS() != "api.example.com" {
		t.Fatalf("unexpected dst_domain: %+v", got)
	}
	if got := netflow.LookupPeerDomain(key.DstIP, uint32(key.DstPort), "tcp", "ns-h2-host"); got != "api.example.com" {
		t.Fatalf("unexpected recorded peer domain: %q", got)
	}
}

func TestTCPLogMessageJSONDoesNotRetainCache(t *testing.T) {
	chunk := &PktChunk{
		ChunkID:       1,
		messageDirty:  true,
		TCPSreries:    []PktTCPHdr{{TS: 10}},
		RetransmitsTx: 1,
	}

	msg1, err := tcpLogMessageJSON(chunk)
	if err != nil {
		t.Fatal(err)
	}
	if chunk.messageDirty {
		t.Fatal("expected tcp message dirty flag to be clean after marshal")
	}
	if chunk.messageCache != "" {
		t.Fatal("expected tcp message not to be retained on chunk")
	}

	msg2, err := tcpLogMessageJSON(chunk)
	if err != nil {
		t.Fatal(err)
	}
	if msg1 != msg2 {
		t.Fatal("expected stable tcp message output")
	}

	chunk.TxBytes = 12
	chunk.markMessageDirty()
	msg3, err := tcpLogMessageJSON(chunk)
	if err != nil {
		t.Fatal(err)
	}
	if msg3 == msg2 {
		t.Fatal("expected dirty tcp chunk to rebuild message")
	}
}

func TestHTTPLogMessageCache(t *testing.T) {
	elem := &HTTPLogElem{
		Method:        "GET",
		Path:          "/health",
		TraceID:       "trace",
		ParentID:      "parent",
		txBytes:       10,
		rxBytes:       20,
		txPkts:        1,
		rxPkts:        2,
		txRetransmits: 1,
		messageDirty:  true,
	}

	msg1, err := httpLogMessageJSON(elem)
	if err != nil {
		t.Fatal(err)
	}
	msg2, err := httpLogMessageJSON(elem)
	if err != nil {
		t.Fatal(err)
	}
	if msg1 != msg2 {
		t.Fatal("expected cached http message to be reused")
	}

	elem.rxBytes++
	elem.markMessageDirty()
	msg3, err := httpLogMessageJSON(elem)
	if err != nil {
		t.Fatal(err)
	}
	if msg3 == msg2 {
		t.Fatal("expected dirty http elem to rebuild message")
	}
}
