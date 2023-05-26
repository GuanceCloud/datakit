// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmputil

import (
	"fmt"
	"testing"

	"github.com/gosnmp/gosnmp"
	"github.com/stretchr/testify/assert"
)

func Test_getAuthProtocol(t *testing.T) {
	tests := []struct {
		authProtocolStr      string
		expectedAuthProtocol gosnmp.SnmpV3AuthProtocol
		expectedError        error
	}{
		{
			"invalid",
			gosnmp.NoAuth,
			fmt.Errorf("unsupported authentication protocol: invalid"),
		},
		{
			"",
			gosnmp.NoAuth,
			nil,
		},
		{
			"md5",
			gosnmp.MD5,
			nil,
		},
		{
			"MD5",
			gosnmp.MD5,
			nil,
		},
		{
			"sha",
			gosnmp.SHA,
			nil,
		},
		{
			"sha224",
			gosnmp.SHA224,
			nil,
		},
		{
			"sha256",
			gosnmp.SHA256,
			nil,
		},
		{
			"sha384",
			gosnmp.SHA384,
			nil,
		},
		{
			"sha512",
			gosnmp.SHA512,
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.authProtocolStr, func(t *testing.T) {
			authProtocol, err := GetAuthProtocol(tt.authProtocolStr)
			assert.Equal(t, tt.expectedError, err)
			assert.Equal(t, tt.expectedAuthProtocol, authProtocol)
		})
	}
}

func Test_getPrivProtocol(t *testing.T) {
	tests := []struct {
		privProtocolStr  string
		expectedProtocol gosnmp.SnmpV3PrivProtocol
		expectedError    error
	}{
		{
			"invalid",
			gosnmp.NoPriv,
			fmt.Errorf("unsupported privacy protocol: invalid"),
		},
		{
			"",
			gosnmp.NoPriv,
			nil,
		},
		{
			"des",
			gosnmp.DES,
			nil,
		},
		{
			"DES",
			gosnmp.DES,
			nil,
		},
		{
			"aes",
			gosnmp.AES,
			nil,
		},
		{
			"aes192",
			gosnmp.AES192,
			nil,
		},
		{
			"aes256",
			gosnmp.AES256,
			nil,
		},
		{
			"aes192c",
			gosnmp.AES192C,
			nil,
		},
		{
			"aes256c",
			gosnmp.AES256C,
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.privProtocolStr, func(t *testing.T) {
			privProtocol, err := GetPrivProtocol(tt.privProtocolStr)
			assert.Equal(t, tt.expectedError, err)
			assert.Equal(t, tt.expectedProtocol, privProtocol)
		})
	}
}

//------------------------------------------------------------------------------

func TestPacketToString(t *testing.T) {
	tests := []struct {
		name        string
		packet      *gosnmp.SnmpPacket
		expectedStr string
	}{
		{
			name: "to string",
			packet: &gosnmp.SnmpPacket{
				Variables: []gosnmp.SnmpPDU{
					{
						Name:  "1.3.6.1.2.1.1.2.0",
						Type:  gosnmp.ObjectIdentifier,
						Value: "1.3.6.1.4.1.3375.2.1.3.4.1",
					},
					{
						Name:  "1.3.6.1.2.1.1.3.0",
						Type:  gosnmp.Counter32,
						Value: 10,
					},
				},
			},
			expectedStr: "error=NoError(code:0, idx:0), values=[{\"oid\":\"1.3.6.1.2.1.1.2.0\",\"type\":\"ObjectIdentifier\",\"value\":\"1.3.6.1.4.1.3375.2.1.3.4.1\"},{\"oid\":\"1.3.6.1.2.1.1.3.0\",\"type\":\"Counter32\",\"value\":\"10\"}]",
		},
		{
			name: "invalid ipaddr",
			packet: &gosnmp.SnmpPacket{
				Variables: []gosnmp.SnmpPDU{
					{
						Name:  "1.3.6.1.2.1.1.2.0",
						Type:  gosnmp.IPAddress,
						Value: 10,
					},
				},
			},
			expectedStr: "error=NoError(code:0, idx:0), values=[{\"oid\":\"1.3.6.1.2.1.1.2.0\",\"type\":\"IPAddress\",\"value\":\"10\",\"parse_err\":\"`oid 1.3.6.1.2.1.1.2.0: IPAddress should be string type but got type `int` and value `10``\"}]",
		},
		{
			name:        "nil packet loglevel",
			expectedStr: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str := PacketAsString(tt.packet)
			assert.Equal(t, tt.expectedStr, str)
		})
	}
}

//------------------------------------------------------------------------------
