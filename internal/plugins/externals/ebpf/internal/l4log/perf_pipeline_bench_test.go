//go:build linux
// +build linux

package l4log

import (
	"bufio"
	"bytes"
	"net"
	"net/http"
	"testing"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func httpReqOrRespLegacy(cnt []byte) int8 {
	if s := bytes.Index(cnt, []byte{'\r', '\n'}); s > 0 {
		buf := cnt[:s]
		hinfo := bytes.Split(buf, []byte{' '})
		if len(hinfo) < 2 {
			return 0
		}

		switch string(hinfo[0]) {
		case "GET":
		case "HEAD":
		case "POST":
		case "PUT":
		case "DELETE":
		case "CONNECT":
		case "OPTIONS":
		case "PATCH":
		case "TRACE":
		default:
			if len(hinfo[0]) > 5 && string(hinfo[0][:5]) == "HTTP/" {
				return 2
			}
			return 0
		}

		return 1
	}

	return 0
}

func parseHTTPRequestMetaLegacy(cnt []byte) (method, path, traceID, parentID string, ok bool) {
	reader := bufio.NewReader(bytes.NewReader(cnt))
	req, err := http.ReadRequest(reader)
	if err != nil {
		return "", "", "", "", false
	}
	defer req.Body.Close() //nolint:errcheck

	method = req.Method
	path = req.URL.Path
	for k, v := range req.Header {
		if k == "Traceparent" && len(v) > 0 {
			traceID, parentID = parseTraceparentHeader([]byte(v[0]))
			break
		}
	}
	return method, path, traceID, parentID, true
}

func parseHTTPResponseStatusLegacy(cnt []byte) (int, bool) {
	reader := bufio.NewReader(bytes.NewReader(cnt))
	resp, err := http.ReadResponse(reader, nil)
	if err != nil {
		return 0, false
	}
	defer resp.Body.Close() //nolint:errcheck
	return resp.StatusCode, true
}

func benchPacketBytes(b *testing.B) []byte {
	b.Helper()

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{ComputeChecksums: true, FixLengths: true}
	eth := &layers.Ethernet{
		SrcMAC:       net.HardwareAddr{0, 1, 2, 3, 4, 5},
		DstMAC:       net.HardwareAddr{6, 7, 8, 9, 10, 11},
		EthernetType: layers.EthernetTypeIPv4,
	}
	ip := &layers.IPv4{
		Version:  4,
		TTL:      64,
		SrcIP:    net.IPv4(10, 0, 0, 1),
		DstIP:    net.IPv4(10, 0, 0, 2),
		Protocol: layers.IPProtocolTCP,
	}
	tcp := &layers.TCP{
		SrcPort: 12345,
		DstPort: 8080,
		Seq:     1,
		Ack:     1,
		ACK:     true,
		PSH:     true,
		Window:  4096,
	}
	if err := tcp.SetNetworkLayerForChecksum(ip); err != nil {
		b.Fatal(err)
	}

	payload := gopacket.Payload([]byte(req))
	if err := gopacket.SerializeLayers(buf, opts, eth, ip, tcp, payload); err != nil {
		b.Fatal(err)
	}

	return append([]byte(nil), buf.Bytes()...)
}

func BenchmarkHTTPReqOrRespLegacy(b *testing.B) {
	payload := []byte(req)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if httpReqOrRespLegacy(payload) == 0 {
			b.Fatal("expected http request")
		}
	}
}

func BenchmarkHTTPReqOrRespFast(b *testing.B) {
	payload := []byte(req)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if httpReqOrResp(payload) == 0 {
			b.Fatal("expected http request")
		}
	}
}

func BenchmarkParseHTTPRequestMetaLegacy(b *testing.B) {
	payload := []byte(req)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, _, _, _, ok := parseHTTPRequestMetaLegacy(payload); !ok {
			b.Fatal("expected parsed request")
		}
	}
}

func BenchmarkParseHTTPRequestMetaFast(b *testing.B) {
	payload := []byte(req)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, _, _, _, _, ok := parseHTTPRequestMeta(payload); !ok {
			b.Fatal("expected parsed request")
		}
	}
}

func BenchmarkParseHTTPResponseStatusLegacy(b *testing.B) {
	payload := []byte(resp)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, ok := parseHTTPResponseStatusLegacy(payload); !ok {
			b.Fatal("expected parsed response")
		}
	}
}

func BenchmarkParseHTTPResponseStatusFast(b *testing.B) {
	payload := []byte(resp)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, ok := parseHTTPResponseStatus(payload); !ok {
			b.Fatal("expected parsed response")
		}
	}
}

func BenchmarkDecodePacketLegacy(b *testing.B) {
	packet := benchPacketBytes(b)
	layerLi := make([]gopacket.LayerType, 0, 10)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		decoder := NewPktDecoder()
		layerLi = layerLi[:0]
		_ = decoder.pktDecode.DecodeLayers(packet, &layerLi)
	}
}

func BenchmarkDecodePacketReuseDecoder(b *testing.B) {
	packet := benchPacketBytes(b)
	layerLi := make([]gopacket.LayerType, 0, 10)
	decoder := NewPktDecoder()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		layerLi = layerLi[:0]
		_ = decoder.pktDecode.DecodeLayers(packet, &layerLi)
	}
}

func BenchmarkDecodePacketExtractReuseDecoder(b *testing.B) {
	packet := benchPacketBytes(b)
	layerLi := make([]gopacket.LayerType, 0, 10)
	decoder := NewPktDecoder()
	cache := newPacketStringCache()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		layerLi = layerLi[:0]
		_ = decoder.pktDecode.DecodeLayers(packet, &layerLi)

		var key PMeta
		var hdr PktTCPHdr
		hdr.SrcMAC = cache.macString(decoder.eth.SrcMAC)
		hdr.DstMAC = cache.macString(decoder.eth.DstMAC)
		hdr.AckSeq = decoder.tcp.Ack
		hdr.Seq = decoder.tcp.Seq
		hdr.Win = uint32(decoder.tcp.Window)
		if len(decoder.tcp.Contents) >= 14 {
			hdr.Flags = TCPFlag(decoder.tcp.Contents[13])
		}

		key.SrcPort = uint16(decoder.tcp.SrcPort)
		key.DstPort = uint16(decoder.tcp.DstPort)
		key.SrcIP = cache.ipString(decoder.ipv4.SrcIP)
		key.DstIP = cache.ipString(decoder.ipv4.DstIP)
		hdr.TCPPayloadSize = int(decoder.ipv4.Length) -
			len(decoder.ipv4.BaseLayer.Contents) -
			len(decoder.tcp.BaseLayer.Contents)

		if key.SrcPort == 0 || hdr.Seq == 0 {
			b.Fatal("unexpected decode result")
		}
	}
}

func BenchmarkDecodePacketFastPath(b *testing.B) {
	packet := benchPacketBytes(b)
	cache := newPacketStringCache()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, ok := parseFastTCPPacket(packet, 1, cache); !ok {
			b.Fatal("expected fast path packet decode")
		}
	}
}
