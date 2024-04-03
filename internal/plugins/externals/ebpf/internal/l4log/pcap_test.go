//go:build linux && with_pcap
// +build linux,with_pcap

package l4log

import (
	"errors"
	"io"
	"net"
	"testing"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

func TestHTTP2(t *testing.T) {
	cases := []struct {
		name           string
		fp             string
		txIP, rxIP     net.IP
		txPort, rxPort uint16
	}{
		{
			name:   "http2",
			fp:     "./pcapdata/http2.pcapng",
			txIP:   net.ParseIP("127.0.0.1"),
			rxIP:   net.ParseIP("127.0.0.1"),
			txPort: 52980,
			rxPort: 9998,
		},
		{
			name:   "grpc",
			fp:     "./pcapdata/grpc.pcapng",
			txIP:   net.ParseIP("127.0.0.1"),
			rxIP:   net.ParseIP("127.0.0.1"),
			txPort: 55488,
			rxPort: 50051,
		},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			h2log := NewH2Log()

			t.Logf("filepath: %s", tCase.fp)

			h, err := pcap.OpenOffline(tCase.fp)
			if err != nil {
				t.Error(err)
			}
			decoder := NewPktDecoder()

			h2dec := NewH2Dec()

			// Only for unit testing
			h2dec.enableH2BodyData = true

			cnt := 0
			for {
				cnt++
				l2Buf, _, err := h.ZeroCopyReadPacketData()
				if err != nil {
					if errors.Is(err, io.EOF) {
						t.Log(err)
					} else {
						t.Fatal(err)
					}
					break
				}

				layerLi := make([]gopacket.LayerType, 0, 10)
				_ = decoder.pktDecode.DecodeLayers(l2Buf, &layerLi)

				sip := decoder.ipv4.SrcIP
				sport := decoder.tcp.SrcPort

				var txrx int8
				if sip.Equal(tCase.txIP) && sport == layers.TCPPort(tCase.txPort) {
					txrx = directionTX
				} else {
					txrx = directionRX
				}

				tcpHdr := PktTCPHdr{
					Seq:    decoder.tcp.Seq,
					AckSeq: decoder.tcp.Ack,
				}
				l4payload := decoder.tcp.Payload
				h2log.Handle(txrx, l4payload, int64(len(l4payload)), &tcpHdr, nil, 0, 0)
			}

			for _, v := range h2log.elems {
				if h2log.isGRPC {
					t.Logf("stream id %v, grpc method %v, path %v, h2 status %v, grpc status %v", v.streamid, v.Method, v.Path, v.StatusCode, v.grpcStatus)
				} else {
					t.Logf("stream id %v, http2 method %v, path %v, h2 status %v", v.streamid, v.Method, v.Path, v.StatusCode)
				}
				t.Logf("\t tx %v Bytes, rx %v bytes", v.txBytes, v.rxBytes)
			}
		})
	}
}
