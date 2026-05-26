//go:build linux
// +build linux

package l4log

import (
	"encoding/binary"
	"net/netip"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/netflow"
)

const (
	tlsRecordTypeHandshake      = 22
	tlsHandshakeTypeClientHello = 1
	tlsExtensionServerName      = 0

	tlsSNIProbePacketBudget = 6
	tlsSNIProbeByteBudget   = 16 * 1024
)

type tlsClientHelloState struct {
	buf       []byte
	packets   uint8
	totalSize int
	found     bool
	exhausted bool
}

func (v *PValue) observeTLSSNI(txRx int8, payload []byte, k *PMeta, nsUID string) {
	if v == nil || k == nil || txRx != directionTX || len(payload) == 0 {
		return
	}
	if v.tlsSNI.found || v.tlsSNI.exhausted {
		return
	}

	v.tlsSNI.packets++
	v.tlsSNI.totalSize += len(payload)
	if v.tlsSNI.totalSize > tlsSNIProbeByteBudget {
		remain := tlsSNIProbeByteBudget - (v.tlsSNI.totalSize - len(payload))
		if remain > 0 {
			v.tlsSNI.buf = append(v.tlsSNI.buf, payload[:remain]...)
		}
		v.tlsSNI.exhausted = true
	} else {
		v.tlsSNI.buf = append(v.tlsSNI.buf, payload...)
	}

	if domain, ok, incomplete := parseTLSClientHelloServerName(v.tlsSNI.buf); ok {
		netflow.RecordPeerDomain(k.DstIP, uint32(k.DstPort), "tcp", nsUID, domain)
		v.tlsSNI.found = true
		v.tlsSNI.buf = nil
		return
	} else if !incomplete {
		v.tlsSNI.exhausted = true
		v.tlsSNI.buf = nil
		return
	}

	if v.tlsSNI.packets >= tlsSNIProbePacketBudget || v.tlsSNI.totalSize >= tlsSNIProbeByteBudget {
		v.tlsSNI.exhausted = true
		v.tlsSNI.buf = nil
	}
}

func parseTLSClientHelloServerName(buf []byte) (domain string, ok bool, incomplete bool) {
	if len(buf) < 5 {
		return "", false, true
	}

	handshakePayload := make([]byte, 0, len(buf))
	sawHandshake := false
	for off := 0; off+5 <= len(buf); {
		recType := buf[off]
		recLen := int(binary.BigEndian.Uint16(buf[off+3 : off+5]))
		off += 5
		if off+recLen > len(buf) {
			if sawHandshake || recType == tlsRecordTypeHandshake {
				return "", false, true
			}
			return "", false, false
		}

		if recType == tlsRecordTypeHandshake {
			sawHandshake = true
			handshakePayload = append(handshakePayload, buf[off:off+recLen]...)
		}
		off += recLen
	}

	if !sawHandshake {
		return "", false, false
	}

	for off := 0; off+4 <= len(handshakePayload); {
		msgType := handshakePayload[off]
		msgLen := int(handshakePayload[off+1])<<16 | int(handshakePayload[off+2])<<8 | int(handshakePayload[off+3])
		off += 4
		if off+msgLen > len(handshakePayload) {
			return "", false, true
		}

		if msgType == tlsHandshakeTypeClientHello {
			domain, ok, incomplete = parseClientHelloServerName(handshakePayload[off : off+msgLen])
			return domain, ok, incomplete
		}

		off += msgLen
	}

	return "", false, false
}

func parseClientHelloServerName(body []byte) (domain string, ok bool, incomplete bool) {
	if len(body) < 34 {
		return "", false, true
	}

	off := 34 // legacy_version + random
	if off >= len(body) {
		return "", false, true
	}

	sessionIDLen := int(body[off])
	off++
	if off+sessionIDLen > len(body) {
		return "", false, true
	}
	off += sessionIDLen

	if off+2 > len(body) {
		return "", false, true
	}
	cipherSuiteLen := int(binary.BigEndian.Uint16(body[off : off+2]))
	off += 2
	if off+cipherSuiteLen > len(body) {
		return "", false, true
	}
	off += cipherSuiteLen

	if off >= len(body) {
		return "", false, true
	}
	compressionMethodsLen := int(body[off])
	off++
	if off+compressionMethodsLen > len(body) {
		return "", false, true
	}
	off += compressionMethodsLen

	if off == len(body) {
		return "", false, false
	}
	if off+2 > len(body) {
		return "", false, true
	}
	extensionsLen := int(binary.BigEndian.Uint16(body[off : off+2]))
	off += 2
	if off+extensionsLen > len(body) {
		return "", false, true
	}

	extensions := body[off : off+extensionsLen]
	for off = 0; off+4 <= len(extensions); {
		extType := binary.BigEndian.Uint16(extensions[off : off+2])
		extLen := int(binary.BigEndian.Uint16(extensions[off+2 : off+4]))
		off += 4
		if off+extLen > len(extensions) {
			return "", false, true
		}
		if extType == tlsExtensionServerName {
			return parseServerNameExtension(extensions[off : off+extLen])
		}
		off += extLen
	}

	return "", false, false
}

func parseServerNameExtension(data []byte) (domain string, ok bool, incomplete bool) {
	if len(data) < 2 {
		return "", false, true
	}
	listLen := int(binary.BigEndian.Uint16(data[:2]))
	if 2+listLen > len(data) {
		return "", false, true
	}

	list := data[2 : 2+listLen]
	for off := 0; off+3 <= len(list); {
		nameType := list[off]
		nameLen := int(binary.BigEndian.Uint16(list[off+1 : off+3]))
		off += 3
		if off+nameLen > len(list) {
			return "", false, true
		}
		if nameType == 0 {
			domain = normalizeTLSServerName(string(list[off : off+nameLen]))
			if domain == "" {
				return "", false, false
			}
			return domain, true, false
		}
		off += nameLen
	}

	return "", false, false
}

func normalizeTLSServerName(domain string) string {
	domain = strings.TrimSpace(strings.ToLower(domain))
	domain = strings.TrimSuffix(domain, ".")
	if domain == "" {
		return ""
	}
	if _, err := netip.ParseAddr(strings.Trim(domain, "[]")); err == nil {
		return ""
	}
	return domain
}
