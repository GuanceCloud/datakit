//go:build linux && cgo
// +build linux,cgo

package l7flow

import (
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/l7flow/comm"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/l7flow/protodec"
)

func BenchmarkTaskCommString(b *testing.B) {
	comm := [KernelTaskCommLen]byte{' ', 'n', 'g', 'i', 'n', 'x', '-', 'w', 'o', 'r', 'k', 'e', 'r', 0, 0, ' '}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if got := taskCommString(&comm); got != "nginx-worker" {
			b.Fatalf("unexpected task comm %q", got)
		}
	}
}

func BenchmarkGenPts(b *testing.B) {
	conn := &comm.ConnectionInfo{
		Saddr:       [4]uint32{0, 0, 0, 0x0100000A},
		Daddr:       [4]uint32{0, 0, 0, 0x0200000A},
		Sport:       43210,
		Dport:       80,
		Pid:         123,
		ProcessName: "nginx",
		TaskName:    "worker",
		ServiceName: "svc",
	}

	base := &protodec.ProtoData{
		KVs: point.KVs{
			point.NewKV(comm.FieldBytesRead, int64(100)),
			point.NewKV(comm.FieldBytesWritten, int64(200)),
			point.NewKV(comm.FieldHTTPMethod, "GET"),
			point.NewKV(comm.FieldHTTPRoute, "/bench"),
			point.NewKV(comm.FieldHTTPStatusCode, "200"),
			point.NewKV(comm.FieldStatus, "ok"),
			point.NewKV(comm.FieldOperation, "HTTP"),
			point.NewKV(comm.FieldResource, "GET /bench"),
		},
		Direction: comm.DOut,
		L7Proto:   protodec.ProtoHTTP,
		Time:      1_000_000,
		Duration:  3_000,
		KTime:     2,
		Meta: protodec.ProtoMeta{
			ReqTCPSeq:  1,
			RespTCPSeq: 2,
			Threads:    [2][2]int32{{11, 22}, {33, 0}},
		},
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v := *base
		v.KVs = append(point.KVs(nil), base.KVs...)
		pts := genPts([]*protodec.ProtoData{&v}, conn)
		if len(pts) != 1 {
			b.Fatalf("unexpected point count %d", len(pts))
		}
	}
}
