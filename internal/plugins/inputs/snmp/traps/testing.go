// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

//go:build !serverless
// +build !serverless

package traps

import (
	"github.com/gosnmp/gosnmp"
)

// List of variables for a NetSNMP::ExampleHeartBeatNotification trap message.
// See: http://www.circitor.fr/Mibs/Html/N/NET-SNMP-EXAMPLES-MIB.php#netSnmpExampleHeartbeatNotification
var (
	NetSNMPExampleHeartbeatNotification = gosnmp.SnmpTrap{
		Variables: []gosnmp.SnmpPDU{
			// sysUpTimeInstance
			{Name: "1.3.6.1.2.1.1.3.0", Type: gosnmp.TimeTicks, Value: uint32(1000)},
			// snmpTrapOID
			{Name: "1.3.6.1.6.3.1.1.4.1.0", Type: gosnmp.OctetString, Value: "1.3.6.1.4.1.8072.2.3.0.1"},
			// heartBeatRate
			{Name: "1.3.6.1.4.1.8072.2.3.2.1", Type: gosnmp.Integer, Value: 1024},
			// heartBeatName
			{Name: "1.3.6.1.4.1.8072.2.3.2.2", Type: gosnmp.OctetString, Value: "test"},
		},
	}
	LinkDownv1GenericTrap = gosnmp.SnmpTrap{
		AgentAddress: "127.0.0.1",
		Enterprise:   ".1.3.6.1.6.3.1.1.5",
		GenericTrap:  2,
		SpecificTrap: 0,
		Timestamp:    1000,
		Variables: []gosnmp.SnmpPDU{
			// ifIndex
			{Name: ".1.3.6.1.2.1.2.2.1.1", Type: gosnmp.Integer, Value: 2},
			// ifAdminStatus
			{Name: ".1.3.6.1.2.1.2.2.1.7", Type: gosnmp.Integer, Value: 1},
			// ifOperStatus
			{Name: ".1.3.6.1.2.1.2.2.1.8", Type: gosnmp.Integer, Value: 2},
			// myFakeVarType 0, 1, 2, 3, 12, 13, 14, 15, 95, and 130 are set
			{Name: ".1.3.6.1.2.1.200.1.3.1.5", Type: gosnmp.OctetString, Value: []uint8{0xf0, 0x0f, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x01, 0, 0, 0, 0, 0x20}},
		},
	}
	AlarmActiveStatev1SpecificTrap = gosnmp.SnmpTrap{
		AgentAddress: "127.0.0.1",
		Enterprise:   ".1.3.6.1.2.1.118",
		GenericTrap:  6,
		SpecificTrap: 2,
		Timestamp:    1000,
		Variables: []gosnmp.SnmpPDU{
			// alarmActiveModelPointer
			{Name: ".1.3.6.1.2.1.118.1.2.2.1.13", Type: gosnmp.OctetString, Value: []uint8{0x66, 0x6f, 0x6f}},
			// alarmActiveResourceId
			{Name: ".1.3.6.1.2.1.118.1.2.2.1.10", Type: gosnmp.OctetString, Value: []uint8{0x62, 0x61, 0x72}},
			// myFakeVarType 0, 1, 2, 3, 12, 13, 14, 15, 95, and 130 are set
			{Name: ".1.3.6.1.2.1.200.1.3.1.5", Type: gosnmp.OctetString, Value: []uint8{0xf0, 0x0f, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x01, 0, 0, 0, 0, 0x20}},
		},
	}
)
