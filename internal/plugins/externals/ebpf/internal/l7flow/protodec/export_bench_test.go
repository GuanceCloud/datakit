//go:build linux
// +build linux

package protodec

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/l7flow/comm"
)

func BenchmarkHTTPDecPipeExport(b *testing.B) {
	dec := &httpDecPipe{direction: comm.DOut}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		dec.infCache = []*httpInfo{{
			method:      "GET",
			path:        "/bench",
			httpVersion: "1.1",
			statusCode:  200,
			reqBytes:    120,
			respBytes:   240,
			ts:          1,
			ktime:       [4]uint64{1, 2, 3, 4},
		}}
		out := dec.Export(false)
		if len(out) != 1 {
			b.Fatalf("unexpected export size %d", len(out))
		}
	}
}

func BenchmarkH2DecPipeExport(b *testing.B) {
	dec := &h2DecPipe{direction: comm.DOut, proto: ProtoHTTP2}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		dec.elems = []*h2Info{{
			method:     "GET",
			path:       "/bench",
			statusCode: 200,
			reqBytes:   120,
			respBytes:  240,
			ts:         1,
			ktime:      [4]uint64{1, 2, 3, 4},
			hFinished:  true,
		}}
		out := dec.Export(false)
		if len(out) != 1 {
			b.Fatalf("unexpected export size %d", len(out))
		}
	}
}
