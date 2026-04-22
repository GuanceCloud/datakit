//go:build linux
// +build linux

package dnsflow

import (
	"testing"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/stretchr/testify/require"
)

func TestReadPacketInfoFromDNSParser_QueryMeta(t *testing.T) {
	cases := []struct {
		name      string
		queryType layers.DNSType
		wantType  string
	}{
		{name: "a", queryType: layers.DNSTypeA, wantType: "A"},
		{name: "ns", queryType: layers.DNSTypeNS, wantType: "NS"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			parser := NewDNSParse()
			parser.layers = []gopacket.LayerType{
				layers.LayerTypeIPv4,
				layers.LayerTypeUDP,
				layers.LayerTypeDNS,
			}

			parser.ipv4.SrcIP = []byte{10, 0, 0, 10}
			parser.ipv4.DstIP = []byte{8, 8, 8, 8}
			parser.udp.SrcPort = 43210
			parser.udp.DstPort = 53
			parser.dns.ID = 123
			parser.dns.QR = false
			parser.dns.Questions = []layers.DNSQuestion{
				{
					Name:  []byte("Example.COM."),
					Type:  tc.queryType,
					Class: layers.DNSClassIN,
				},
			}

			info, err := ReadPacketInfoFromDNSParser(time.Now(), &parser)
			require.NoError(t, err)
			require.Equal(t, "example.com", info.QueryDomain)
			require.Equal(t, tc.wantType, info.QueryType)
			require.Equal(t, uint16(43210), info.Key.ClientPort)
			require.Equal(t, uint16(53), info.Key.ServerPort)
		})
	}
}
