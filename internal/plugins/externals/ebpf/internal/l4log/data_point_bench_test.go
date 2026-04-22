//go:build linux
// +build linux

package l4log

import (
	"testing"
	"time"
)

func benchmarkConnContext() (*PMeta, *PValue, *TCPConns, map[string]string, *HTTPLogElem, *PktChunk) {
	key := &PMeta{
		SrcIP:   "10.0.0.1",
		DstIP:   "10.0.0.2",
		SrcPort: 43210,
		DstPort: 8080,
		VNIID:   7,
	}
	value := &PValue{
		sMACEQ: true,
		tcpInfo: TCPLog{
			synSeq:         100,
			synAckSeq:      200,
			direction:      directionOutgoing,
			l7proto:        L7ProtoHTTP,
			RetransmitsSYN: 1,
		},
	}
	conns := &TCPConns{
		ifaceNameMAC: [2]string{"eth0", "aa:bb:cc:dd:ee:ff"},
		nsUID:        "42",
		hostNetwork:  true,
		virtualNIC:   false,
	}
	tags := buildCommTags(key, value, conns)
	httpElem := &HTTPLogElem{
		Direction:     DOutging,
		TraceID:       "trace",
		ParentID:      "parent",
		Path:          "/bench",
		Method:        "GET",
		StatusCode:    200,
		reqSeq:        100,
		respSeq:       200,
		txFirstByteTS: 1,
		txLastByteTS:  int64(3 * time.Millisecond),
		rxFirstByteTS: int64(5 * time.Millisecond),
		rxLastByteTS:  int64(8 * time.Millisecond),
		txPkts:        4,
		rxPkts:        5,
		txBytes:       512,
		rxBytes:       1024,
		txRetransmits: 1,
		rxRetransmits: 2,
	}
	chunk := &PktChunk{
		ChunkID:       1,
		chunkKind:     chunkKindSYN | chunkKindFINRST,
		txSeq:         [2]uint32{100, 300},
		rxSeq:         [2]uint32{200, 400},
		TimePos:       50,
		TXPacket:      4,
		RXPacket:      5,
		TxBytes:       512,
		RxBytes:       1024,
		RetransmitsTx: 1,
		RetransmitsRx: 2,
		TCPSreries: []PktTCPHdr{
			{TS: 100},
		},
	}
	return key, value, conns, tags, httpElem, chunk
}

func BenchmarkBuildCommTags(b *testing.B) {
	key, value, conns, tags, httpElem, chunk := benchmarkConnContext()
	_ = tags
	_ = httpElem
	_ = chunk

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = buildCommTags(key, value, conns)
	}
}

func BenchmarkBuildHTTPLog(b *testing.B) {
	key, value, conns, _, httpElem, _ := benchmarkConnContext()
	baseKVs := buildCommKVs(key, value, conns)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, _, ok, err := buildHTTPLog(key, value, httpElem, baseKVs, nil, "42", []string{"10.0.0.1"}); err != nil {
			b.Fatal(err)
		} else if !ok {
			b.Fatal("expected http log point")
		}
	}
}

func BenchmarkBuildTCPLog(b *testing.B) {
	key, value, conns, tags, httpElem, chunk := benchmarkConnContext()
	_ = tags
	_ = httpElem
	_ = chunk
	baseKVs := buildCommKVs(key, value, conns)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, _, _, _, chunk := benchmarkConnContext()
		if _, _, ok, err := buildTCPLog(chunk, time.Now().UnixNano(), baseKVs, value); err != nil {
			b.Fatal(err)
		} else if !ok {
			b.Fatal("expected tcp log point")
		}
	}
}
