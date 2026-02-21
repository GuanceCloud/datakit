// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package lldp

import (
	"encoding/json"
	T "testing"

	"github.com/gosnmp/gosnmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp/snmputil"
)

func Test_doParseLLDPData(t *T.T) {
	t.Run(`doParseLLDPData`, func(t *T.T) {
		pdus := []*gosnmp.SnmpPDU{
			// chassis id
			{
				Value: []byte("DESKTOP-ONGRI03"),
				Name:  ".1.0.8802.1.1.2.1.4.1.1.5.0.32.3",
				Type:  gosnmp.OctetString,
			},

			// chassis id
			{
				Value: []byte("ZHUYUNPC"),
				Name:  ".1.0.8802.1.1.2.1.4.1.1.5.0.32.4",
				Type:  gosnmp.OctetString,
			},

			// ChassisSubType
			{
				Value: ChassisSubtypeMACAddress,
				Name:  ".1.0.8802.1.1.2.1.4.1.1.4.0.45.1",
				Type:  gosnmp.Integer,
			},

			{
				Value: ChassisSubtypeLocallyAssigned,
				Name:  ".1.0.8802.1.1.2.1.4.1.1.4.0.32.4",
				Type:  gosnmp.Integer,
			},

			// port id
			{
				Value: []byte("gi24"),
				Name:  ".1.0.8802.1.1.2.1.4.1.1.7.0.33.1",
				Type:  gosnmp.OctetString,
			},

			{
				Value: []byte("gi24"),
				Name:  ".1.0.8802.1.1.2.1.4.1.1.7.0.34.1",
				Type:  gosnmp.OctetString,
			},

			{
				Value: []byte("GigabitEthernet0/0/7"),
				Name:  ".1.0.8802.1.1.2.1.4.1.1.7.0.45.1",
				Type:  gosnmp.OctetString,
			},

			{
				Value: []byte("GigabitEthernet0/0/6"),
				Name:  ".1.0.8802.1.1.2.1.4.1.1.7.0.46.1",
				Type:  gosnmp.OctetString,
			},

			{
				Value: []byte("firewall-usg6325E, Huawei Versatile Routing Platform Software, Software Version : USG6300E V600R007C20SPC603 (VRP (R) Software, Version 5.170), Copyright (C) 2014-2023 Huawei Technologies Co., Ltd."),
				Name:  ".1.0.8802.1.1.2.1.4.1.1.10.0.45.1",
				Type:  gosnmp.OctetString,
			},

			{
				Value: []byte("firewall-usg6325E, Huawei Versatile Routing Platform Software, Software Version : USG6300E V600R007C20SPC603 (VRP (R) Software, Version 5.170), Copyright (C) 2014-2023 Huawei Technologies Co., Ltd."),
				Name:  ".1.0.8802.1.1.2.1.4.1.1.10.0.46.1",
				Type:  gosnmp.OctetString,
			},

			{
				Value: []byte("firewall-usg6325E"),
				Name:  ".1.0.8802.1.1.2.1.4.1.1.9.0.45.1",
				Type:  gosnmp.OctetString,
			},

			{
				Value: []byte("firewall-usg6325E"),
				Name:  ".1.0.8802.1.1.2.1.4.1.1.9.0.46.1",
				Type:  gosnmp.OctetString,
			},

			{
				Value: []byte{'c', 0xfb, 0xae, 0x9f, 0xea, 0x97},
				Name:  ".1.0.8802.1.1.2.1.4.1.1.5.0.45.1",
				Type:  gosnmp.OctetString,
			},

			{
				Value: []byte{'c', 0xfb, 0xae, 0x9f, 0xea, 0x97},
				Name:  ".1.0.8802.1.1.2.1.4.1.1.5.0.46.1",
				Type:  gosnmp.OctetString,
			},

			/*
				2025/09/19 17:06:34 col: 7, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.7.0.32.3|val: "\x04|\x16\xf7F\xfb"
				2025/09/19 17:06:34 col: 7, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.7.0.32.4|val: "\x88\x88\x88\x88\x87\x88"
				2025/09/19 17:06:34 col: 7, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.7.0.33.1|val: "gi24"
				2025/09/19 17:06:34 col: 7, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.7.0.34.1|val: "gi24"
				2025/09/19 17:06:34 col: 7, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.7.0.45.1|val: "GigabitEthernet0/0/7"
				2025/09/19 17:06:34 col: 7, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.7.0.46.1|val: "GigabitEthernet0/0/6"

			*/
		}

		tmp := map[string]*neighborInfo{}
		localPorts := map[string]string{
			"1": "InLoopBack0", "2": "NULL0", "3": "Console9/0/0", "4": "MEth0/0/1", "6": "Vlanif60",
			"7": "Vlanif305", "8": "Vlanif306", "9": "Vlanif999", "10": "Eth-Trunk1", "11": "Eth-Trunk2",
			"12": "Eth-Trunk5", "13": "GigabitEthernet0/0/1", "14": "GigabitEthernet0/0/2", "15": "GigabitEthernet0/0/3",
			"16": "GigabitEthernet0/0/4", "17": "GigabitEthernet0/0/5", "18": "GigabitEthernet0/0/6", "19": "GigabitEthernet0/0/7",
			"20": "GigabitEthernet0/0/8", "21": "GigabitEthernet0/0/9", "22": "GigabitEthernet0/0/10", "23": "GigabitEthernet0/0/11",
			"24": "GigabitEthernet0/0/12", "25": "GigabitEthernet0/0/13", "26": "GigabitEthernet0/0/14", "27": "GigabitEthernet0/0/15",
			"28": "GigabitEthernet0/0/16", "29": "GigabitEthernet0/0/17", "30": "GigabitEthernet0/0/18", "31": "GigabitEthernet0/0/19",
			"32": "GigabitEthernet0/0/20", "33": "GigabitEthernet0/0/21", "34": "GigabitEthernet0/0/22", "35": "GigabitEthernet0/0/23",
			"36": "GigabitEthernet0/0/24", "37": "GigabitEthernet0/0/25", "38": "GigabitEthernet0/0/26", "39": "GigabitEthernet0/0/27",
			"40": "GigabitEthernet0/0/28", "41": "GigabitEthernet0/0/29", "42": "GigabitEthernet0/0/30", "43": "GigabitEthernet0/0/31",
			"44": "GigabitEthernet0/0/32", "45": "GigabitEthernet0/0/33", "46": "GigabitEthernet0/0/34", "47": "GigabitEthernet0/0/35",
			"48": "GigabitEthernet0/0/36", "49": "GigabitEthernet0/0/37", "50": "GigabitEthernet0/0/38", "51": "GigabitEthernet0/0/39",
			"52": "GigabitEthernet0/0/40", "53": "GigabitEthernet0/0/41", "54": "GigabitEthernet0/0/42", "55": "GigabitEthernet0/0/43",
			"56": "GigabitEthernet0/0/44", "57": "GigabitEthernet0/0/45", "58": "GigabitEthernet0/0/46", "59": "GigabitEthernet0/0/47",
			"60": "GigabitEthernet0/0/48", "61": "GigabitEthernet0/0/49", "62": "GigabitEthernet0/0/50", "63": "GigabitEthernet0/0/51",
			"64": "GigabitEthernet0/0/52", "65": "Eth-Trunk100",
		}
		for _, x := range pdus {
			doParseLLDPData(x, localPorts, tmp)
		}

		for k, v := range tmp {
			t.Logf("%s: %+#v", k, v)
		}
	})
}

func TestLLDP(t *T.T) {
	t.Run("basic", func(t *T.T) {
		t.Skip("skip hard-coded target IP test")
		targetIP := "10.100.67.254"

		session, err := snmputil.NewGosnmpSession(&snmputil.SessionOpts{
			IPAddress:       targetIP,
			Port:            161,
			SnmpVersion:     3,
			CommunityString: "1234abcd",
			// SNMPv3 config (uncomment to use v3)
			// SnmpVersion:  3,
			// User:         "snmpuser",
			// AuthProtocol: "SHA",
			// AuthKey:      "AuthPassw0rd!",
			// PrivProtocol: "AES",
			// PrivKey:      "PrivPassw0rd!",
		})
		assert.NoError(t, err)

		assert.NoError(t, session.Connect())
		defer session.Close()

		lpm, err := getLocalPortMap(session, oidIFDescription)
		assert.NoError(t, err)

		t.Logf("discovered %d local interfaces on %s", len(lpm), targetIP)
		for k, v := range lpm {
			t.Logf("%s: %s", k, v)
		}

		var rawNeighborData []gosnmp.SnmpPDU
		err = session.BulkWalk(oidLLDPRemoteTable, func(pdu gosnmp.SnmpPDU) error {
			rawNeighborData = append(rawNeighborData, pdu)
			return nil
		})
		assert.NoError(t, err)

		t.Logf("got %d neighbor pdus", len(rawNeighborData))

		neighbors := parseLLDPData(rawNeighborData, lpm)

		j, err := json.MarshalIndent(neighbors, "", "  ")
		assert.NoError(t, err)
		t.Logf("LLDP neighbors: %s", string(j))
	})
}

func Test_formatColonSepBytes(t *T.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "MAC address",
			input:    []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55},
			expected: "00:11:22:33:44:55",
		},
		{
			name:     "Single byte",
			input:    []byte{0xAA},
			expected: "aa",
		},
		{
			name:     "Empty slice",
			input:    []byte{},
			expected: "",
		},
		{
			name:     "Binary data",
			input:    []byte{0x00, 0x01, 0xFF},
			expected: "00:01:ff",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *T.T) {
			result := formatColonSepBytes(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func Test_isPrintableASCII(t *T.T) {
	tests := []struct {
		name     string
		input    []byte
		expected bool
	}{
		{
			name:     "Printable ASCII",
			input:    []byte("Hello World"),
			expected: true,
		},
		{
			name:     "Non-printable control char",
			input:    []byte{0x00, 0x01},
			expected: false,
		},
		{
			name:     "Non-printable tab char",
			input:    []byte("Hello\tWorld"),
			expected: false,
		},
		{
			name:     "Empty slice",
			input:    []byte{},
			expected: false,
		},
		{
			name:     "ASCII numbers and symbols",
			input:    []byte("12345!@#$%"),
			expected: true,
		},
		{
			name:     "Binary data",
			input:    []byte{0xFF, 0xFE},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *T.T) {
			result := isPrintableASCII(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func Test_encodeChassisID(t *T.T) {
	tests := []struct {
		name         string
		subtype      int
		rawChassisID []byte
		expectedID   string
	}{
		{
			name:         "Subtype 1: Chassis Component",
			subtype:      ChassisSubtypeComponent,
			rawChassisID: []byte("SW-001"),
			expectedID:   "SW-001",
		},
		{
			name:         "Subtype 4: MAC address",
			subtype:      ChassisSubtypeMACAddress,
			rawChassisID: []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55},
			expectedID:   "00:11:22:33:44:55",
		},
		{
			name:         "Subtype 5: IPv4 address",
			subtype:      ChassisSubtypeNetworkAddress,
			rawChassisID: []byte{192, 168, 1, 1},
			expectedID:   "192.168.1.1",
		},
		{
			name:         "Subtype 5: IPv6 address",
			subtype:      ChassisSubtypeNetworkAddress,
			rawChassisID: []byte{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
			expectedID:   "2001:db8::1",
		},
		{
			name:         "Subtype 2: Interface Alias",
			subtype:      ChassisSubtypeInterfaceAlias,
			rawChassisID: []byte("Interface-Alias-01"),
			expectedID:   "Interface-Alias-01",
		},
		{
			name:         "Subtype 7: Locally Assigned",
			subtype:      ChassisSubtypeLocallyAssigned,
			rawChassisID: []byte("Local-Device"),
			expectedID:   "Local-Device",
		},
		{
			name:         "Unknown subtype with printable ASCII",
			subtype:      99,
			rawChassisID: []byte("Printable"),
			expectedID:   "Printable",
		},
		{
			name:         "Unknown subtype with binary data",
			subtype:      99,
			rawChassisID: []byte{0xFF, 0xFE, 0xFD},
			expectedID:   "ff:fe:fd",
		},
		{
			name:         "Empty rawChassisID",
			subtype:      ChassisSubtypeMACAddress,
			rawChassisID: nil,
			expectedID:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *T.T) {
			result := encodeChassisIDBySubtype(tt.subtype, tt.rawChassisID)
			assert.Equal(t, tt.expectedID, result)
		})
	}
}

func Test_encodePortID(t *T.T) {
	tests := []struct {
		name         string
		subtype      int
		rawPortID    []byte
		expectedPort string
	}{
		{
			name:         "Subtype 1: Interface Alias",
			subtype:      PortSubtypeInterfaceAlias,
			rawPortID:    []byte("GigabitEthernet0/0/1"),
			expectedPort: "GigabitEthernet0/0/1",
		},
		{
			name:         "Subtype 3: MAC address",
			subtype:      PortSubtypeMACAddress,
			rawPortID:    []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55},
			expectedPort: "00:11:22:33:44:55",
		},
		{
			name:         "Subtype 4: IPv4 address",
			subtype:      PortSubtypeNetworkAddress,
			rawPortID:    []byte{192, 168, 1, 1},
			expectedPort: "192.168.1.1",
		},
		{
			name:         "Subtype 4: IPv6 address",
			subtype:      PortSubtypeNetworkAddress,
			rawPortID:    []byte{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
			expectedPort: "2001:db8::1",
		},
		{
			name:         "Subtype 4: MAC address (6 bytes)",
			subtype:      PortSubtypeMACAddress,
			rawPortID:    []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55},
			expectedPort: "00:11:22:33:44:55",
		},
		{
			name:         "Subtype 5: Interface Name",
			subtype:      PortSubtypeInterfaceName,
			rawPortID:    []byte("eth0"),
			expectedPort: "eth0",
		},
		{
			name:         "Subtype 7: Local",
			subtype:      PortSubtypeLocallyAssigned,
			rawPortID:    []byte("LocalPort"),
			expectedPort: "LocalPort",
		},
		{
			name:         "Unknown subtype with printable ASCII",
			subtype:      99,
			rawPortID:    []byte("Printable"),
			expectedPort: "Printable",
		},
		{
			name:         "Unknown subtype with binary data",
			subtype:      99,
			rawPortID:    []byte{0xFF, 0xFE, 0xFD},
			expectedPort: "ff:fe:fd",
		},
		{
			name:         "Empty rawPortID",
			subtype:      PortSubtypeMACAddress,
			rawPortID:    nil,
			expectedPort: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *T.T) {
			result := encodePortIDBySubtype(tt.subtype, tt.rawPortID)
			assert.Equal(t, tt.expectedPort, result)
		})
	}
}

func Test_GetLocalChassisID(t *T.T) {
	tests := []struct {
		name              string
		setupMock         func(*snmputil.MockSession)
		expectedChassisID string
		expectedSubtype   int
		expectError       bool
		errorContains     string
	}{
		{
			name: "MAC address subtype 4",
			setupMock: func(sess *snmputil.MockSession) {
				sess.On("Get", []string{oidLLDPLocChassisIDSubtype, oidLLDPLocChassisID}).Return(&gosnmp.SnmpPacket{
					Variables: []gosnmp.SnmpPDU{
						{
							Name:  oidLLDPLocChassisIDSubtype,
							Type:  gosnmp.Integer,
							Value: ChassisSubtypeMACAddress,
						},
						{
							Name:  oidLLDPLocChassisID,
							Type:  gosnmp.OctetString,
							Value: []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55},
						},
					},
				}, nil)
			},
			expectedChassisID: "00:11:22:33:44:55",
			expectedSubtype:   ChassisSubtypeMACAddress,
			expectError:       false,
		},
		{
			name: "Chassis Component subtype 1",
			setupMock: func(sess *snmputil.MockSession) {
				sess.On("Get", []string{oidLLDPLocChassisIDSubtype, oidLLDPLocChassisID}).Return(&gosnmp.SnmpPacket{
					Variables: []gosnmp.SnmpPDU{
						{
							Name:  oidLLDPLocChassisIDSubtype,
							Type:  gosnmp.Integer,
							Value: ChassisSubtypeComponent,
						},
						{
							Name:  oidLLDPLocChassisID,
							Type:  gosnmp.OctetString,
							Value: []byte("SW-001"),
						},
					},
				}, nil)
			},
			expectedChassisID: "SW-001",
			expectedSubtype:   ChassisSubtypeComponent,
			expectError:       false,
		},
		{
			name: "IPv4 address subtype 5",
			setupMock: func(sess *snmputil.MockSession) {
				sess.On("Get", []string{oidLLDPLocChassisIDSubtype, oidLLDPLocChassisID}).Return(&gosnmp.SnmpPacket{
					Variables: []gosnmp.SnmpPDU{
						{
							Name:  oidLLDPLocChassisIDSubtype,
							Type:  gosnmp.Integer,
							Value: ChassisSubtypeNetworkAddress,
						},
						{
							Name:  oidLLDPLocChassisID,
							Type:  gosnmp.OctetString,
							Value: []byte{192, 168, 1, 1},
						},
					},
				}, nil)
			},
			expectedChassisID: "192.168.1.1",
			expectedSubtype:   ChassisSubtypeNetworkAddress,
			expectError:       false,
		},
		{
			name: "OID order reversed",
			setupMock: func(sess *snmputil.MockSession) {
				sess.On("Get", []string{oidLLDPLocChassisIDSubtype, oidLLDPLocChassisID}).Return(&gosnmp.SnmpPacket{
					Variables: []gosnmp.SnmpPDU{
						{
							Name:  oidLLDPLocChassisID,
							Type:  gosnmp.OctetString,
							Value: []byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF},
						},
						{
							Name:  oidLLDPLocChassisIDSubtype,
							Type:  gosnmp.Integer,
							Value: ChassisSubtypeMACAddress,
						},
					},
				}, nil)
			},
			expectedChassisID: "aa:bb:cc:dd:ee:ff",
			expectedSubtype:   ChassisSubtypeMACAddress,
			expectError:       false,
		},
		{
			name: "OID with leading dot",
			setupMock: func(sess *snmputil.MockSession) {
				sess.On("Get", []string{oidLLDPLocChassisIDSubtype, oidLLDPLocChassisID}).Return(&gosnmp.SnmpPacket{
					Variables: []gosnmp.SnmpPDU{
						{
							Name:  "." + oidLLDPLocChassisIDSubtype,
							Type:  gosnmp.Integer,
							Value: ChassisSubtypeMACAddress,
						},
						{
							Name:  "." + oidLLDPLocChassisID,
							Type:  gosnmp.OctetString,
							Value: []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66},
						},
					},
				}, nil)
			},
			expectedChassisID: "11:22:33:44:55:66",
			expectedSubtype:   ChassisSubtypeMACAddress,
			expectError:       false,
		},
		{
			name: "SNMP Get error",
			setupMock: func(sess *snmputil.MockSession) {
				sess.On("Get", []string{oidLLDPLocChassisIDSubtype, oidLLDPLocChassisID}).Return((*gosnmp.SnmpPacket)(nil), assert.AnError)
			},
			expectError:   true,
			errorContains: "failed to get",
		},
		{
			name: "NoSuchObject error for subtype",
			setupMock: func(sess *snmputil.MockSession) {
				sess.On("Get", []string{oidLLDPLocChassisIDSubtype, oidLLDPLocChassisID}).Return(&gosnmp.SnmpPacket{
					Variables: []gosnmp.SnmpPDU{
						{
							Name:  oidLLDPLocChassisIDSubtype,
							Type:  gosnmp.NoSuchObject,
							Value: nil,
						},
						{
							Name:  oidLLDPLocChassisID,
							Type:  gosnmp.OctetString,
							Value: []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55},
						},
					},
				}, nil)
			},
			expectError:   true,
			errorContains: "does not exist",
		},
		{
			name: "NoSuchInstance error for ChassisID",
			setupMock: func(sess *snmputil.MockSession) {
				sess.On("Get", []string{oidLLDPLocChassisIDSubtype, oidLLDPLocChassisID}).Return(&gosnmp.SnmpPacket{
					Variables: []gosnmp.SnmpPDU{
						{
							Name:  oidLLDPLocChassisIDSubtype,
							Type:  gosnmp.Integer,
							Value: ChassisSubtypeMACAddress,
						},
						{
							Name:  oidLLDPLocChassisID,
							Type:  gosnmp.NoSuchInstance,
							Value: nil,
						},
					},
				}, nil)
			},
			expectError:   true,
			errorContains: "does not exist",
		},
		{
			name: "Wrong number of variables",
			setupMock: func(sess *snmputil.MockSession) {
				sess.On("Get", []string{oidLLDPLocChassisIDSubtype, oidLLDPLocChassisID}).Return(&gosnmp.SnmpPacket{
					Variables: []gosnmp.SnmpPDU{
						{
							Name:  oidLLDPLocChassisIDSubtype,
							Type:  gosnmp.Integer,
							Value: ChassisSubtypeMACAddress,
						},
					},
				}, nil)
			},
			expectError:   true,
			errorContains: "expected 2 variables",
		},
		{
			name: "Nil session",
			setupMock: func(sess *snmputil.MockSession) {
				// No mock setup needed
			},
			expectError:   true,
			errorContains: "SNMP session is nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *T.T) {
			var session snmputil.Session
			if !tt.expectError || tt.errorContains != "SNMP session is nil" {
				sess := snmputil.CreateMockSession()
				tt.setupMock(sess)
				session = sess
			}

			chassisID, subtype, err := GetLocalChassisID(session)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedChassisID, chassisID)
				assert.Equal(t, tt.expectedSubtype, subtype)
			}
		})
	}
}

// 1. FormatChassisSubtype 测试
func Test_FormatChassisSubtype(t *T.T) {
	tests := []struct {
		name     string
		subtype  int
		expected string
	}{
		{name: "Component", subtype: ChassisSubtypeComponent, expected: ChassisSubtypeStrComponent},
		{name: "InterfaceAlias", subtype: ChassisSubtypeInterfaceAlias, expected: ChassisSubtypeStrInterfaceAlias},
		{name: "PortComponent", subtype: ChassisSubtypePortComponent, expected: ChassisSubtypeStrPortComponent},
		{name: "MACAddress", subtype: ChassisSubtypeMACAddress, expected: ChassisSubtypeStrMACAddress},
		{name: "NetworkAddress", subtype: ChassisSubtypeNetworkAddress, expected: ChassisSubtypeStrNetworkAddress},
		{name: "InterfaceName", subtype: ChassisSubtypeInterfaceName, expected: ChassisSubtypeStrInterfaceName},
		{name: "LocallyAssigned", subtype: ChassisSubtypeLocallyAssigned, expected: ChassisSubtypeStrLocallyAssigned},
		{name: "Unknown", subtype: 99, expected: SubtypeStrUnknown},
		{name: "Zero", subtype: 0, expected: SubtypeStrUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *T.T) {
			result := FormatChassisSubtype(tt.subtype)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// 2. FormatPortSubtype 测试
func Test_FormatPortSubtype(t *T.T) {
	tests := []struct {
		name     string
		subtype  int
		expected string
	}{
		{name: "InterfaceAlias", subtype: PortSubtypeInterfaceAlias, expected: PortSubtypeStrInterfaceAlias},
		{name: "PortComponent", subtype: PortSubtypePortComponent, expected: PortSubtypeStrPortComponent},
		{name: "MACAddress", subtype: PortSubtypeMACAddress, expected: PortSubtypeStrMACAddress},
		{name: "NetworkAddress", subtype: PortSubtypeNetworkAddress, expected: PortSubtypeStrNetworkAddress},
		{name: "InterfaceName", subtype: PortSubtypeInterfaceName, expected: PortSubtypeStrInterfaceName},
		{name: "AgentCircuitID", subtype: PortSubtypeAgentCircuitID, expected: PortSubtypeStrAgentCircuitID},
		{name: "LocallyAssigned", subtype: PortSubtypeLocallyAssigned, expected: PortSubtypeStrLocallyAssigned},
		{name: "Unknown", subtype: 99, expected: SubtypeStrUnknown},
		{name: "Zero", subtype: 0, expected: SubtypeStrUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *T.T) {
			result := FormatPortSubtype(tt.subtype)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// 3. GetLocalChassisID SNMPv1 Null 类型测试
func Test_GetLocalChassisID_SNMPv1_Null(t *T.T) {
	t.Run("SNMPv1 Null type for subtype", func(t *T.T) {
		sess := snmputil.CreateMockSession()
		sess.Version = gosnmp.Version1
		sess.On("Get", []string{oidLLDPLocChassisIDSubtype, oidLLDPLocChassisID}).Return(&gosnmp.SnmpPacket{
			Variables: []gosnmp.SnmpPDU{
				{
					Name:  oidLLDPLocChassisIDSubtype,
					Type:  gosnmp.Null,
					Value: nil,
				},
				{
					Name:  oidLLDPLocChassisID,
					Type:  gosnmp.OctetString,
					Value: []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55},
				},
			},
		}, nil)

		_, _, err := GetLocalChassisID(sess)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not exist")
	})

	t.Run("SNMPv1 Null type for ChassisID", func(t *T.T) {
		sess := snmputil.CreateMockSession()
		sess.Version = gosnmp.Version1
		sess.On("Get", []string{oidLLDPLocChassisIDSubtype, oidLLDPLocChassisID}).Return(&gosnmp.SnmpPacket{
			Variables: []gosnmp.SnmpPDU{
				{
					Name:  oidLLDPLocChassisIDSubtype,
					Type:  gosnmp.Integer,
					Value: ChassisSubtypeMACAddress,
				},
				{
					Name:  oidLLDPLocChassisID,
					Type:  gosnmp.Null,
					Value: nil,
				},
			},
		}, nil)

		_, _, err := GetLocalChassisID(sess)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not exist")
	})
}

// 4. walkSNMPTable 版本差异测试
func Test_walkSNMPTable_Version(t *T.T) {
	t.Run("SNMPv1 uses GetWalkAll", func(t *T.T) {
		sess := snmputil.CreateMockSession()
		sess.Version = gosnmp.Version1
		sess.On("GetWalkAll", "1.2.3.4").Return([]gosnmp.SnmpPDU{
			{Name: "1.2.3.4.1", Type: gosnmp.Integer, Value: 1},
		}, nil)

		result, err := walkSNMPTable(sess, "1.2.3.4")
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		sess.AssertExpectations(t)
	})

	t.Run("SNMPv2c uses BulkWalk", func(t *T.T) {
		sess := snmputil.CreateMockSession()
		sess.Version = gosnmp.Version2c
		sess.On("BulkWalk", "1.2.3.4", mock.AnythingOfType("gosnmp.WalkFunc")).Return(nil).Run(func(args mock.Arguments) {
			walkFn := args.Get(1).(gosnmp.WalkFunc)
			walkFn(gosnmp.SnmpPDU{Name: "1.2.3.4.1", Type: gosnmp.Integer, Value: 1})
		})

		result, err := walkSNMPTable(sess, "1.2.3.4")
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		sess.AssertExpectations(t)
	})
}
