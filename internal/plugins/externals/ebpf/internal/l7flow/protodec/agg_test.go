//go:build linux
// +build linux

package protodec

import (
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/l7flow/comm"
)

func TestHTTPAggNoPidTag(t *testing.T) {
	agg := &HTTPAggP{}

	conn := &comm.ConnectionInfo{
		Saddr:       [4]uint32{0, 0, 0, 0x0a000001},
		Daddr:       [4]uint32{0, 0, 0, 0x0a000002},
		Sport:       12345,
		Dport:       80,
		Pid:         5678,
		Netns:       1,
		Meta:        0, // IPv4 + TCP
		ProcessName: "http-proc",
	}

	data := &ProtoData{
		KVs: point.NewKVs(map[string]interface{}{
			comm.FieldHTTPRoute:      "/demo",
			comm.FieldHTTPMethod:     "GET",
			comm.FieldHTTPStatusCode: "200",
			comm.FieldHTTPVersion:    "1.1",
			comm.FieldBytesRead:      int64(50),
			comm.FieldBytesWritten:   int64(30),
		}),
		Cost:      1000,
		Duration:  1000,
		Direction: comm.DOut,
		L7Proto:   ProtoHTTP,
	}

	agg.Obs(conn, data)

	pts := agg.Export(nil, nil)
	if len(pts) != 1 {
		t.Fatalf("expected 1 point, got %d", len(pts))
	}
	if pid := pts[0].GetTag("pid"); pid != "" {
		t.Fatalf("unexpected pid tag: %s", pid)
	}
}

func TestHTTPAggNormalizesDirectionAndKeepsFirstBytes(t *testing.T) {
	agg := &HTTPAggP{}

	conn := &comm.ConnectionInfo{
		Saddr:       [4]uint32{0, 0, 0, 0x0100000A},
		Daddr:       [4]uint32{0, 0, 0, 0x0200000A},
		Sport:       80,
		Dport:       52345,
		Pid:         5678,
		Netns:       1,
		Meta:        0,
		ProcessName: "http-proc",
	}

	data := &ProtoData{
		KVs: point.NewKVs(map[string]interface{}{
			comm.FieldHTTPRoute:      "/demo",
			comm.FieldHTTPMethod:     "GET",
			comm.FieldHTTPStatusCode: "200",
			comm.FieldHTTPVersion:    "1.1",
			comm.FieldBytesRead:      int64(50),
			comm.FieldBytesWritten:   int64(30),
		}),
		Cost:      1000,
		Duration:  1000,
		Direction: comm.DOut,
		L7Proto:   ProtoHTTP,
	}

	agg.Obs(conn, data)

	pts := agg.Export(nil, nil)
	if len(pts) != 1 {
		t.Fatalf("expected 1 point, got %d", len(pts))
	}

	if got := pts[0].GetTag("direction"); got != DirectionIncoming {
		t.Fatalf("unexpected direction %q", got)
	}
	if got := pts[0].GetTag("server_port"); got != "80" {
		t.Fatalf("unexpected server_port %q", got)
	}
	if got := pts[0].GetTag("client_port"); got != "*" {
		t.Fatalf("unexpected client_port %q", got)
	}
	if got := pts[0].Get("bytes_read"); got != int64(50) {
		t.Fatalf("unexpected bytes_read %v", got)
	}
	if got := pts[0].Get("bytes_written"); got != int64(30) {
		t.Fatalf("unexpected bytes_written %v", got)
	}
}

func TestHTTPAggAddsDstDomainFromHTTPHost(t *testing.T) {
	agg := &HTTPAggP{}

	conn := &comm.ConnectionInfo{
		Saddr:       [4]uint32{0, 0, 0, 0x0100000A},
		Daddr:       [4]uint32{0, 0, 0, 0x0200000A},
		Sport:       52345,
		Dport:       443,
		Pid:         5678,
		Netns:       1,
		Meta:        0,
		ProcessName: "http-proc",
	}

	data := &ProtoData{
		KVs: point.NewKVs(map[string]interface{}{
			comm.FieldHTTPRoute:      "/demo",
			comm.FieldHTTPMethod:     "GET",
			comm.FieldHTTPHost:       "api.example.com",
			comm.FieldHTTPStatusCode: "200",
			comm.FieldHTTPVersion:    "1.1",
			comm.FieldBytesRead:      int64(50),
			comm.FieldBytesWritten:   int64(30),
		}),
		Cost:      1000,
		Duration:  1000,
		Direction: comm.DOut,
		L7Proto:   ProtoHTTP,
	}

	agg.Obs(conn, data)

	pts := agg.Export(nil, nil)
	if len(pts) != 1 {
		t.Fatalf("expected 1 point, got %d", len(pts))
	}
	if got := pts[0].GetTag("dst_domain"); got != "api.example.com" {
		t.Fatalf("unexpected dst_domain %q", got)
	}
	if got := pts[0].GetTag("server_domain"); got != "api.example.com" {
		t.Fatalf("unexpected server_domain %q", got)
	}
}
