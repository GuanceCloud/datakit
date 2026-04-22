//go:build linux
// +build linux

package l7flow

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/l7flow/comm"
)

func TestPutNetwrkDataResetsState(t *testing.T) {
	data := getNetwrkData(128)
	if data == nil {
		t.Fatal("expected pooled network data")
	}

	data.Conn = comm.ConnectionInfo{Pid: 123}
	data.CaptureSize = 64
	data.FnCallSize = 32
	data.TCPSeq = 100
	data.Thread = [2]int32{1, 2}
	data.TS = 1
	data.TSTail = 2
	data.Index = 3
	data.Fn = comm.FnSSLRead
	data.Payload = append(data.Payload, []byte("hello")...)

	putNetwrkData(data)

	reused := getNetwrkData(128)
	if reused == nil {
		t.Fatal("expected reused network data")
	}
	defer putNetwrkData(reused)

	if reused.Conn != (comm.ConnectionInfo{}) {
		t.Fatalf("expected reset connection info, got %+v", reused.Conn)
	}
	if reused.CaptureSize != 0 || reused.FnCallSize != 0 || reused.TCPSeq != 0 {
		t.Fatalf("expected counters reset, got %+v", reused)
	}
	if len(reused.Payload) != 0 {
		t.Fatalf("expected empty payload, got len=%d", len(reused.Payload))
	}
}

func BenchmarkNetwrkDataPool(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		data := getNetwrkData(256)
		if data == nil {
			b.Fatal("expected pooled object")
		}
		data.Payload = append(data.Payload[:0], []byte("payload")...)
		putNetwrkData(data)
	}
}
