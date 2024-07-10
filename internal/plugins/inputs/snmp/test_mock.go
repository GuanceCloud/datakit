// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmp

import (
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/gosnmp/gosnmp"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp/snmputil"
)

// testing helpers ...

func getMockInput01() *Input {
	conf := `
[key_mapping]
  CNTLR_NAME = "unit_name"
  DISK_NAME = "unit_name"
  ENT_CLASS = "unit_class"
  ENT_NAME = "unit_name"
  FAN_DESCR = "unit_desc"
  IF_OPERS_TATUS = "unit_status"
  IFADMINSTATUS = "unit_status"
  IFALIAS = "unit_alias"
  IFDESCR = "unit_desc"
  IFNAME = "unit_name"
  IFOPERSTATUS = "unit_status"
  IFTYPE = "unit_type"
  PSU_DESCR = "unit_desc"
  SENSOR_LOCALE = "unit_locale"
  SNMPINDEX = "snmp_index"
  SNMPVALUE = "snmp_value"
  TYPE = "unit_type"
  SENSOR_INFO = "unit_desc"
[oid_keys]
  "1.3.6.1.2.1.1.3.0" = "netUptime"
  "1.3.6.1.2.1.25.1.1.0" = "uptime"
  "1.3.6.1.2.1.2.2.1.13" = "ifInDiscards"
  "1.3.6.1.2.1.2.2.1.14" = "ifInErrors"
  "1.3.6.1.2.1.31.1.1.1.6" = "ifHCInOctets"
  "1.3.6.1.2.1.2.2.1.19" = "ifOutDiscards"
  "1.3.6.1.2.1.2.2.1.20" = "ifOutErrors"
  "1.3.6.1.2.1.31.1.1.1.10" = "ifHCOutOctets"
  "1.3.6.1.2.1.31.1.1.1.15" = "ifHighSpeed"
  "1.3.6.1.2.1.2.2.1.8" = "ifNetStatus"
  "1.3.6.1.4.1.2011.5.25.31.1.1.1.1.11" = "temperature"
  "1.3.6.1.4.1.2011.5.25.31.1.1.1.1.5" = "cpuUsage"
  "1.3.6.1.4.1.2011.5.25.31.1.1.1.1.7" = "memoryUsage"
  "1.3.6.1.4.1.2011.5.25.31.1.1.10.1.7" = "fanStatus"
`
	ipt := defaultInput()
	_, _ = toml.Decode(conf, ipt)

	return ipt
}

func getMockSession01() snmputil.Session {
	sess := snmputil.CreateMockSession()

	sess.On("GetWalkAll", "1.3.6.1.4.1.2011.5.25.31.1.1.1.1.5").Return(getCPUUsage01(), nil)
	sess.On("GetWalkAll", "1.3.6.1.2.1.10.7.2.1.19").Return(getSNMPVALUE01(), nil)

	sess.On("GetWalkAll", "1.3.6.1.2.1.2.2.1.8").Return(getSNMPVALUE02(), nil)
	sess.On("GetWalkAll", "1.3.6.1.2.1.2.2.1.7").Return(getIFADMINSTATUS02(), nil)
	sess.On("GetWalkAll", "1.3.6.1.2.1.31.1.1.1.18").Return(getIFALIAS02(), nil)
	sess.On("GetWalkAll", "1.3.6.1.2.1.31.1.1.1.1").Return(getIFNAME02(), nil)
	sess.On("GetWalkAll", "1.3.6.1.2.1.2.2.1.2").Return(getIFDESCR02(), nil)
	sess.On("GetWalkAll", "1.3.6.1.2.1.2.2.1.3").Return(getIFTYPE02(), nil)

	sess.On("GetWalkAll", "1.3.6.1.2.1.47.1.1.1.1.7").Return(getENTNAME03(), nil)
	sess.On("GetWalkAll", "1.3.6.1.2.1.31.1.1.1.15").Return(getIFSpeed03(), nil)
	sess.On("GetNext", []string{snmputil.DeviceReachableGetNextOid}).Return(&gosnmp.SnmpPacket{}, nil)
	sess.On("Get", []string{sysNameOID}).Return(&gosnmp.SnmpPacket{
		Variables: []gosnmp.SnmpPDU{{
			Name: sysNameOID, Type: gosnmp.OctetString, Value: []byte("ZY_WLC"),
		}},
	}, nil)
	sess.On("Get", []string{sysObjectOID}).Return(&gosnmp.SnmpPacket{
		Variables: []gosnmp.SnmpPDU{{
			Name: sysObjectOID, Type: gosnmp.ObjectIdentifier, Value: ".1.3.6.1.4.1.2011.2.240.12",
		}},
	}, nil)
	sess.On("Get", []string{"1.3.6.1.2.1.25.1.1.0"}).Return(&gosnmp.SnmpPacket{
		Variables: []gosnmp.SnmpPDU{{
			Name: ".1.3.6.1.2.1.25.1.1.0", Type: gosnmp.NoSuchObject, Value: nil,
		}},
	}, nil)
	sess.On("Get", []string{"1.3.6.1.2.1.1.3.0"}).Return(&gosnmp.SnmpPacket{
		Variables: []gosnmp.SnmpPDU{{
			Name: ".1.3.6.1.2.1.1.3.0", Type: gosnmp.TimeTicks, Value: 204572431,
		}},
	}, nil)
	sess.On("Get", []string{"1.3.6.1.2.1.1.6.0"}).Return(&gosnmp.SnmpPacket{
		Variables: []gosnmp.SnmpPDU{{
			Name: ".1.3.6.1.2.1.1.6.0", Type: gosnmp.OctetString, Value: []byte("Shenzhen China"),
		}},
	}, nil)

	sess.On("Get", []string{"1.3.6.1.4.1.2011.5.25.31.1.1.1.1.5.9"}).Return(&gosnmp.SnmpPacket{
		Variables: []gosnmp.SnmpPDU{{Name: ".1.3.6.1.4.1.2011.5.25.31.1.1.1.1.5.9", Type: gosnmp.Integer, Value: 37}},
	}, nil)

	// Add metrics batch.
	func(snmpPDUs []gosnmp.SnmpPDU) {
		sps := []gosnmp.SnmpPDU{}
		oids := []string{}

		for _, snmpPDU := range snmpPDUs {
			sp := gosnmp.SnmpPDU{Name: snmpPDU.Name, Type: gosnmp.Gauge32, Value: snmpPDU.Value}
			sps = append(sps, sp)
			oids = append(oids, FormatOID(snmpPDU.Name))
		}

		sess.On("Get", oids).Return(&gosnmp.SnmpPacket{
			Variables: sps,
		}, nil)
	}(getIFSpeed03())

	return sess
}

func getMockMacros01() map[string]snmputil.Macro {
	m := map[string]snmputil.Macro{
		"ENT_NAME": {"9": "SRU Board 0"},
	}

	return m
}

func getMockItem01() *snmputil.Item {
	it := &snmputil.Item{
		SnmpOID: "1.3.6.1.4.1.2011.5.25.31.1.1.1.1.5.{#SNMPINDEX}",
		Key:     "system.cpu.util[hwEntityCpuUsage.{#SNMPINDEX}]",
		Tags:    []*snmputil.Tag{{Tag: "component", Value: "cpu"}},
	}

	return it
}

// oid: 1.3.6.1.4.1.2011.5.25.31.1.1.1.1.5
func getCPUUsage01() []gosnmp.SnmpPDU {
	return []gosnmp.SnmpPDU{
		{Name: ".1.3.6.1.4.1.2011.5.25.31.1.1.1.1.5.3", Type: gosnmp.Integer, Value: 0},
		{Name: ".1.3.6.1.4.1.2011.5.25.31.1.1.1.1.5.5", Type: gosnmp.Integer, Value: 0},
		{Name: ".1.3.6.1.4.1.2011.5.25.31.1.1.1.1.5.9", Type: gosnmp.Integer, Value: 37},
		{Name: ".1.3.6.1.4.1.2011.5.25.31.1.1.1.1.5.14", Type: gosnmp.Integer, Value: 0},
		{Name: ".1.3.6.1.4.1.2011.5.25.31.1.1.1.1.5.78", Type: gosnmp.Integer, Value: 0},
		{Name: ".1.3.6.1.4.1.2011.5.25.31.1.1.1.1.5.142", Type: gosnmp.Integer, Value: 0},
		{Name: ".1.3.6.1.4.1.2011.5.25.31.1.1.1.1.5.206", Type: gosnmp.Integer, Value: 0},
		{Name: ".1.3.6.1.4.1.2011.5.25.31.1.1.1.1.5.270", Type: gosnmp.Integer, Value: 0},
		{Name: ".1.3.6.1.4.1.2011.5.25.31.1.1.1.1.5.334", Type: gosnmp.Integer, Value: 0},
		{Name: ".1.3.6.1.4.1.2011.5.25.31.1.1.1.1.5.398", Type: gosnmp.Integer, Value: 0},
		{Name: ".1.3.6.1.4.1.2011.5.25.31.1.1.1.1.5.462", Type: gosnmp.Integer, Value: 0},
	}
}

// oid: 1.3.6.1.2.1.10.7.2.1.19
func getSNMPVALUE01() []gosnmp.SnmpPDU {
	return []gosnmp.SnmpPDU{
		{Name: ".1.3.6.1.2.1.10.7.2.1.19.3", Type: gosnmp.Integer, Value: 1},
		{Name: ".1.3.6.1.2.1.10.7.2.1.19.4", Type: gosnmp.Integer, Value: 1},
		{Name: ".1.3.6.1.2.1.10.7.2.1.19.5", Type: gosnmp.Integer, Value: 1},
		{Name: ".1.3.6.1.2.1.10.7.2.1.19.6", Type: gosnmp.Integer, Value: 1},
		{Name: ".1.3.6.1.2.1.10.7.2.1.19.7", Type: gosnmp.Integer, Value: 1},
		{Name: ".1.3.6.1.2.1.10.7.2.1.19.8", Type: gosnmp.Integer, Value: 1},
		{Name: ".1.3.6.1.2.1.10.7.2.1.19.9", Type: gosnmp.Integer, Value: 1},
		{Name: ".1.3.6.1.2.1.10.7.2.1.19.10", Type: gosnmp.Integer, Value: 1},
		{Name: ".1.3.6.1.2.1.10.7.2.1.19.201", Type: gosnmp.Integer, Value: 1},
	}
}

// oid: 1.3.6.1.2.1.2.2.1.8
func getSNMPVALUE02() []gosnmp.SnmpPDU {
	return []gosnmp.SnmpPDU{
		{Name: ".1.3.6.1.2.1.2.2.1.8.1", Type: gosnmp.Integer, Value: 1},
		{Name: ".1.3.6.1.2.1.2.2.1.8.2", Type: gosnmp.Integer, Value: 1},
		{Name: ".1.3.6.1.2.1.2.2.1.8.3", Type: gosnmp.Integer, Value: 1},
		{Name: ".1.3.6.1.2.1.2.2.1.8.4", Type: gosnmp.Integer, Value: 2},
		{Name: ".1.3.6.1.2.1.2.2.1.8.5", Type: gosnmp.Integer, Value: 2},
		{Name: ".1.3.6.1.2.1.2.2.1.8.6", Type: gosnmp.Integer, Value: 2},
		{Name: ".1.3.6.1.2.1.2.2.1.8.7", Type: gosnmp.Integer, Value: 2},
		{Name: ".1.3.6.1.2.1.2.2.1.8.8", Type: gosnmp.Integer, Value: 1},
		{Name: ".1.3.6.1.2.1.2.2.1.8.9", Type: gosnmp.Integer, Value: 2},
		{Name: ".1.3.6.1.2.1.2.2.1.8.10", Type: gosnmp.Integer, Value: 2},
		{Name: ".1.3.6.1.2.1.2.2.1.8.11", Type: gosnmp.Integer, Value: 1},
		{Name: ".1.3.6.1.2.1.2.2.1.8.12", Type: gosnmp.Integer, Value: 2},
		{Name: ".1.3.6.1.2.1.2.2.1.8.13", Type: gosnmp.Integer, Value: 2},
		{Name: ".1.3.6.1.2.1.2.2.1.8.196", Type: gosnmp.Integer, Value: 2},
		{Name: ".1.3.6.1.2.1.2.2.1.8.197", Type: gosnmp.Integer, Value: 2},
		{Name: ".1.3.6.1.2.1.2.2.1.8.198", Type: gosnmp.Integer, Value: 2},
		{Name: ".1.3.6.1.2.1.2.2.1.8.199", Type: gosnmp.Integer, Value: 2},
		{Name: ".1.3.6.1.2.1.2.2.1.8.200", Type: gosnmp.Integer, Value: 2},
		{Name: ".1.3.6.1.2.1.2.2.1.8.201", Type: gosnmp.Integer, Value: 1},
	}
}

// oid: 1.3.6.1.2.1.2.2.1.7
func getIFADMINSTATUS02() []gosnmp.SnmpPDU {
	return []gosnmp.SnmpPDU{
		{Name: ".1.3.6.1.2.1.2.2.1.7.1", Type: gosnmp.Integer, Value: 1},
		{Name: ".1.3.6.1.2.1.2.2.1.7.2", Type: gosnmp.Integer, Value: 1},
		{Name: ".1.3.6.1.2.1.2.2.1.7.3", Type: gosnmp.Integer, Value: 1},
		{Name: ".1.3.6.1.2.1.2.2.1.7.4", Type: gosnmp.Integer, Value: 1},
		{Name: ".1.3.6.1.2.1.2.2.1.7.5", Type: gosnmp.Integer, Value: 1},
		{Name: ".1.3.6.1.2.1.2.2.1.7.6", Type: gosnmp.Integer, Value: 1},
		{Name: ".1.3.6.1.2.1.2.2.1.7.7", Type: gosnmp.Integer, Value: 1},
		{Name: ".1.3.6.1.2.1.2.2.1.7.8", Type: gosnmp.Integer, Value: 1},
		{Name: ".1.3.6.1.2.1.2.2.1.7.9", Type: gosnmp.Integer, Value: 1},
		{Name: ".1.3.6.1.2.1.2.2.1.7.10", Type: gosnmp.Integer, Value: 1},
		{Name: ".1.3.6.1.2.1.2.2.1.7.11", Type: gosnmp.Integer, Value: 1},
		{Name: ".1.3.6.1.2.1.2.2.1.7.12", Type: gosnmp.Integer, Value: 1},
		{Name: ".1.3.6.1.2.1.2.2.1.7.13", Type: gosnmp.Integer, Value: 1},
		{Name: ".1.3.6.1.2.1.2.2.1.7.196", Type: gosnmp.Integer, Value: 1},
		{Name: ".1.3.6.1.2.1.2.2.1.7.197", Type: gosnmp.Integer, Value: 1},
		{Name: ".1.3.6.1.2.1.2.2.1.7.198", Type: gosnmp.Integer, Value: 1},
		{Name: ".1.3.6.1.2.1.2.2.1.7.199", Type: gosnmp.Integer, Value: 1},
		{Name: ".1.3.6.1.2.1.2.2.1.7.200", Type: gosnmp.Integer, Value: 1},
		{Name: ".1.3.6.1.2.1.2.2.1.7.201", Type: gosnmp.Integer, Value: 1},
	}
}

// oid: 1.3.6.1.2.1.31.1.1.1.18
func getIFALIAS02() []gosnmp.SnmpPDU {
	return []gosnmp.SnmpPDU{
		{Name: ".1.3.6.1.2.1.31.1.1.1.18.1", Type: gosnmp.OctetString, Value: []byte("HUAWEI, AC Series, InLoopBack0 Interface")},
		{Name: ".1.3.6.1.2.1.31.1.1.1.18.2", Type: gosnmp.OctetString, Value: []byte("HUAWEI, AC Series, NULL0 Interface")},
		{Name: ".1.3.6.1.2.1.31.1.1.1.18.3", Type: gosnmp.OctetString, Value: []byte("HUAWEI, AC Series, GigabitEthernet0/0/1 Interface")},
		{Name: ".1.3.6.1.2.1.31.1.1.1.18.4", Type: gosnmp.OctetString, Value: []byte("HUAWEI, AC Series, GigabitEthernet0/0/2 Interface")},
		{Name: ".1.3.6.1.2.1.31.1.1.1.18.5", Type: gosnmp.OctetString, Value: []byte("HUAWEI, AC Series, GigabitEthernet0/0/3 Interface")},
		{Name: ".1.3.6.1.2.1.31.1.1.1.18.6", Type: gosnmp.OctetString, Value: []byte("HUAWEI, AC Series, GigabitEthernet0/0/4 Interface")},
		{Name: ".1.3.6.1.2.1.31.1.1.1.18.7", Type: gosnmp.OctetString, Value: []byte("HUAWEI, AC Series, GigabitEthernet0/0/5 Interface")},
		{Name: ".1.3.6.1.2.1.31.1.1.1.18.8", Type: gosnmp.OctetString, Value: []byte("HUAWEI, AC Series, GigabitEthernet0/0/6 Interface")},
		{Name: ".1.3.6.1.2.1.31.1.1.1.18.9", Type: gosnmp.OctetString, Value: []byte("HUAWEI, AC Series, GigabitEthernet0/0/7 Interface")},
		{Name: ".1.3.6.1.2.1.31.1.1.1.18.10", Type: gosnmp.OctetString, Value: []byte("HUAWEI, AC Series, GigabitEthernet0/0/8 Interface")},
		{Name: ".1.3.6.1.2.1.31.1.1.1.18.11", Type: gosnmp.OctetString, Value: []byte("HUAWEI, AC Series, Vlanif1 Interface")},
		{Name: ".1.3.6.1.2.1.31.1.1.1.18.12", Type: gosnmp.OctetString, Value: []byte("HUAWEI, AC Series, Vlanif305 Interface")},
		{Name: ".1.3.6.1.2.1.31.1.1.1.18.13", Type: gosnmp.OctetString, Value: []byte("connect ZY_DSW")},
		{Name: ".1.3.6.1.2.1.31.1.1.1.18.196", Type: gosnmp.OctetString, Value: []byte("wireless-zhuyun-2F")},
		{Name: ".1.3.6.1.2.1.31.1.1.1.18.197", Type: gosnmp.OctetString, Value: []byte("wireless-guest")},
		{Name: ".1.3.6.1.2.1.31.1.1.1.18.198", Type: gosnmp.OctetString, Value: []byte("wireless-zhuyun-4F")},
		{Name: ".1.3.6.1.2.1.31.1.1.1.18.199", Type: gosnmp.OctetString, Value: []byte("wireless-zhuyun-1F")},
		{Name: ".1.3.6.1.2.1.31.1.1.1.18.200", Type: gosnmp.OctetString, Value: []byte("wireless-zhuyun-3F")},
		{Name: ".1.3.6.1.2.1.31.1.1.1.18.201", Type: gosnmp.OctetString, Value: []byte("HUAWEI, AC Series, Ethernet0/0/47 Interface")},
	}
}

// oid: 1.3.6.1.2.1.31.1.1.1.1
func getIFNAME02() []gosnmp.SnmpPDU {
	return []gosnmp.SnmpPDU{
		{Name: ".1.3.6.1.2.1.31.1.1.1.1.1", Type: gosnmp.OctetString, Value: []byte("InLoopBack0")},
		{Name: ".1.3.6.1.2.1.31.1.1.1.1.2", Type: gosnmp.OctetString, Value: []byte("NULL0")},
		{Name: ".1.3.6.1.2.1.31.1.1.1.1.3", Type: gosnmp.OctetString, Value: []byte("GigabitEthernet0/0/1")},
		{Name: ".1.3.6.1.2.1.31.1.1.1.1.4", Type: gosnmp.OctetString, Value: []byte("GigabitEthernet0/0/2")},
		{Name: ".1.3.6.1.2.1.31.1.1.1.1.5", Type: gosnmp.OctetString, Value: []byte("GigabitEthernet0/0/3")},
		{Name: ".1.3.6.1.2.1.31.1.1.1.1.6", Type: gosnmp.OctetString, Value: []byte("GigabitEthernet0/0/4")},
		{Name: ".1.3.6.1.2.1.31.1.1.1.1.7", Type: gosnmp.OctetString, Value: []byte("GigabitEthernet0/0/5")},
		{Name: ".1.3.6.1.2.1.31.1.1.1.1.8", Type: gosnmp.OctetString, Value: []byte("GigabitEthernet0/0/6")},
		{Name: ".1.3.6.1.2.1.31.1.1.1.1.9", Type: gosnmp.OctetString, Value: []byte("GigabitEthernet0/0/7")},
		{Name: ".1.3.6.1.2.1.31.1.1.1.1.10", Type: gosnmp.OctetString, Value: []byte("GigabitEthernet0/0/8")},
		{Name: ".1.3.6.1.2.1.31.1.1.1.1.11", Type: gosnmp.OctetString, Value: []byte("Vlanif1")},
		{Name: ".1.3.6.1.2.1.31.1.1.1.1.12", Type: gosnmp.OctetString, Value: []byte("Vlanif305")},
		{Name: ".1.3.6.1.2.1.31.1.1.1.1.13", Type: gosnmp.OctetString, Value: []byte("Eth-Trunk1")},
		{Name: ".1.3.6.1.2.1.31.1.1.1.1.196", Type: gosnmp.OctetString, Value: []byte("Vlanif302")},
		{Name: ".1.3.6.1.2.1.31.1.1.1.1.197", Type: gosnmp.OctetString, Value: []byte("Vlanif300")},
		{Name: ".1.3.6.1.2.1.31.1.1.1.1.198", Type: gosnmp.OctetString, Value: []byte("Vlanif304")},
		{Name: ".1.3.6.1.2.1.31.1.1.1.1.199", Type: gosnmp.OctetString, Value: []byte("Vlanif301")},
		{Name: ".1.3.6.1.2.1.31.1.1.1.1.200", Type: gosnmp.OctetString, Value: []byte("Vlanif303")},
		{Name: ".1.3.6.1.2.1.31.1.1.1.1.201", Type: gosnmp.OctetString, Value: []byte("Ethernet0/0/47")},
	}
}

// oid: 1.3.6.1.2.1.2.2.1.2
func getIFDESCR02() []gosnmp.SnmpPDU {
	return []gosnmp.SnmpPDU{
		{Name: ".1.3.6.1.2.1.2.2.1.2.1", Type: gosnmp.OctetString, Value: []byte("InLoopBack0")},
		{Name: ".1.3.6.1.2.1.2.2.1.2.2", Type: gosnmp.OctetString, Value: []byte("NULL0")},
		{Name: ".1.3.6.1.2.1.2.2.1.2.3", Type: gosnmp.OctetString, Value: []byte("GigabitEthernet0/0/1")},
		{Name: ".1.3.6.1.2.1.2.2.1.2.4", Type: gosnmp.OctetString, Value: []byte("GigabitEthernet0/0/2")},
		{Name: ".1.3.6.1.2.1.2.2.1.2.5", Type: gosnmp.OctetString, Value: []byte("GigabitEthernet0/0/3")},
		{Name: ".1.3.6.1.2.1.2.2.1.2.6", Type: gosnmp.OctetString, Value: []byte("GigabitEthernet0/0/4")},
		{Name: ".1.3.6.1.2.1.2.2.1.2.7", Type: gosnmp.OctetString, Value: []byte("GigabitEthernet0/0/5")},
		{Name: ".1.3.6.1.2.1.2.2.1.2.8", Type: gosnmp.OctetString, Value: []byte("GigabitEthernet0/0/6")},
		{Name: ".1.3.6.1.2.1.2.2.1.2.9", Type: gosnmp.OctetString, Value: []byte("GigabitEthernet0/0/7")},
		{Name: ".1.3.6.1.2.1.2.2.1.2.10", Type: gosnmp.OctetString, Value: []byte("GigabitEthernet0/0/8")},
		{Name: ".1.3.6.1.2.1.2.2.1.2.11", Type: gosnmp.OctetString, Value: []byte("Vlanif1")},
		{Name: ".1.3.6.1.2.1.2.2.1.2.12", Type: gosnmp.OctetString, Value: []byte("Vlanif305")},
		{Name: ".1.3.6.1.2.1.2.2.1.2.13", Type: gosnmp.OctetString, Value: []byte("Eth-Trunk1")},
		{Name: ".1.3.6.1.2.1.2.2.1.2.196", Type: gosnmp.OctetString, Value: []byte("Vlanif302")},
		{Name: ".1.3.6.1.2.1.2.2.1.2.197", Type: gosnmp.OctetString, Value: []byte("Vlanif300")},
		{Name: ".1.3.6.1.2.1.2.2.1.2.198", Type: gosnmp.OctetString, Value: []byte("Vlanif304")},
		{Name: ".1.3.6.1.2.1.2.2.1.2.199", Type: gosnmp.OctetString, Value: []byte("Vlanif301")},
		{Name: ".1.3.6.1.2.1.2.2.1.2.200", Type: gosnmp.OctetString, Value: []byte("Vlanif303")},
		{Name: ".1.3.6.1.2.1.2.2.1.2.201", Type: gosnmp.OctetString, Value: []byte("Ethernet0/0/47")},
	}
}

// oid: 1.3.6.1.2.1.2.2.1.3
func getIFTYPE02() []gosnmp.SnmpPDU {
	return []gosnmp.SnmpPDU{
		{Name: ".1.3.6.1.2.1.2.2.1.3.1", Type: gosnmp.Integer, Value: 24},
		{Name: ".1.3.6.1.2.1.2.2.1.3.2", Type: gosnmp.Integer, Value: 1},
		{Name: ".1.3.6.1.2.1.2.2.1.3.3", Type: gosnmp.Integer, Value: 6},
		{Name: ".1.3.6.1.2.1.2.2.1.3.4", Type: gosnmp.Integer, Value: 6},
		{Name: ".1.3.6.1.2.1.2.2.1.3.5", Type: gosnmp.Integer, Value: 6},
		{Name: ".1.3.6.1.2.1.2.2.1.3.6", Type: gosnmp.Integer, Value: 6},
		{Name: ".1.3.6.1.2.1.2.2.1.3.7", Type: gosnmp.Integer, Value: 6},
		{Name: ".1.3.6.1.2.1.2.2.1.3.8", Type: gosnmp.Integer, Value: 6},
		{Name: ".1.3.6.1.2.1.2.2.1.3.9", Type: gosnmp.Integer, Value: 6},
		{Name: ".1.3.6.1.2.1.2.2.1.3.10", Type: gosnmp.Integer, Value: 6},
		{Name: ".1.3.6.1.2.1.2.2.1.3.11", Type: gosnmp.Integer, Value: 53},
		{Name: ".1.3.6.1.2.1.2.2.1.3.12", Type: gosnmp.Integer, Value: 53},
		{Name: ".1.3.6.1.2.1.2.2.1.3.13", Type: gosnmp.Integer, Value: 53},
		{Name: ".1.3.6.1.2.1.2.2.1.3.196", Type: gosnmp.Integer, Value: 53},
		{Name: ".1.3.6.1.2.1.2.2.1.3.197", Type: gosnmp.Integer, Value: 53},
		{Name: ".1.3.6.1.2.1.2.2.1.3.198", Type: gosnmp.Integer, Value: 53},
		{Name: ".1.3.6.1.2.1.2.2.1.3.199", Type: gosnmp.Integer, Value: 53},
		{Name: ".1.3.6.1.2.1.2.2.1.3.200", Type: gosnmp.Integer, Value: 53},
		{Name: ".1.3.6.1.2.1.2.2.1.3.201", Type: gosnmp.Integer, Value: 6},
	}
}

// oid: 1.3.6.1.2.1.47.1.1.1.1.7
func getENTNAME03() []gosnmp.SnmpPDU {
	return []gosnmp.SnmpPDU{
		{Name: ".1.3.6.1.2.1.47.1.1.1.1.7.3", Type: gosnmp.OctetString, Value: []byte("AC6003-8")},
		{Name: ".1.3.6.1.2.1.47.1.1.1.1.7.5", Type: gosnmp.OctetString, Value: []byte("Board slot 0")},
		{Name: ".1.3.6.1.2.1.47.1.1.1.1.7.9", Type: gosnmp.OctetString, Value: []byte("SRU Board 0")},
		{Name: ".1.3.6.1.2.1.47.1.1.1.1.7.14", Type: gosnmp.OctetString, Value: []byte("GigabitEthernet0/0/1")},
		{Name: ".1.3.6.1.2.1.47.1.1.1.1.7.78", Type: gosnmp.OctetString, Value: []byte("GigabitEthernet0/0/2")},
		{Name: ".1.3.6.1.2.1.47.1.1.1.1.7.142", Type: gosnmp.OctetString, Value: []byte("GigabitEthernet0/0/3")},
		{Name: ".1.3.6.1.2.1.47.1.1.1.1.7.206", Type: gosnmp.OctetString, Value: []byte("GigabitEthernet0/0/4")},
		{Name: ".1.3.6.1.2.1.47.1.1.1.1.7.270", Type: gosnmp.OctetString, Value: []byte("GigabitEthernet0/0/5")},
		{Name: ".1.3.6.1.2.1.47.1.1.1.1.7.334", Type: gosnmp.OctetString, Value: []byte("GigabitEthernet0/0/6")},
		{Name: ".1.3.6.1.2.1.47.1.1.1.1.7.398", Type: gosnmp.OctetString, Value: []byte("GigabitEthernet0/0/7")},
		{Name: ".1.3.6.1.2.1.47.1.1.1.1.7.462", Type: gosnmp.OctetString, Value: []byte("GigabitEthernet0/0/8")},
	}
}

// oid: 1.3.6.1.2.1.31.1.1.1.15
func getIFSpeed03() []gosnmp.SnmpPDU {
	return []gosnmp.SnmpPDU{
		{Name: ".1.3.6.1.2.1.31.1.1.1.15.1", Type: gosnmp.Gauge32, Value: 0},
		{Name: ".1.3.6.1.2.1.31.1.1.1.15.2", Type: gosnmp.Gauge32, Value: 0},
		{Name: ".1.3.6.1.2.1.31.1.1.1.15.3", Type: gosnmp.Gauge32, Value: 1000},
		{Name: ".1.3.6.1.2.1.31.1.1.1.15.4", Type: gosnmp.Gauge32, Value: 1000},
		{Name: ".1.3.6.1.2.1.31.1.1.1.15.5", Type: gosnmp.Gauge32, Value: 1000},
		{Name: ".1.3.6.1.2.1.31.1.1.1.15.6", Type: gosnmp.Gauge32, Value: 1000},
		{Name: ".1.3.6.1.2.1.31.1.1.1.15.7", Type: gosnmp.Gauge32, Value: 1000},
		{Name: ".1.3.6.1.2.1.31.1.1.1.15.8", Type: gosnmp.Gauge32, Value: 100},
		{Name: ".1.3.6.1.2.1.31.1.1.1.15.9", Type: gosnmp.Gauge32, Value: 1000},
		{Name: ".1.3.6.1.2.1.31.1.1.1.15.10", Type: gosnmp.Gauge32, Value: 1000},
		{Name: ".1.3.6.1.2.1.31.1.1.1.15.11", Type: gosnmp.Gauge32, Value: 1000},
		{Name: ".1.3.6.1.2.1.31.1.1.1.15.12", Type: gosnmp.Gauge32, Value: 1000},
		{Name: ".1.3.6.1.2.1.31.1.1.1.15.13", Type: gosnmp.Gauge32, Value: 0},
		{Name: ".1.3.6.1.2.1.31.1.1.1.15.196", Type: gosnmp.Gauge32, Value: 1000},
		{Name: ".1.3.6.1.2.1.31.1.1.1.15.197", Type: gosnmp.Gauge32, Value: 1000},
		{Name: ".1.3.6.1.2.1.31.1.1.1.15.198", Type: gosnmp.Gauge32, Value: 1000},
		{Name: ".1.3.6.1.2.1.31.1.1.1.15.199", Type: gosnmp.Gauge32, Value: 1000},
		{Name: ".1.3.6.1.2.1.31.1.1.1.15.200", Type: gosnmp.Gauge32, Value: 1000},
		{Name: ".1.3.6.1.2.1.31.1.1.1.15.201", Type: gosnmp.Gauge32, Value: 1000},
	}
}

func getOIDsFromMock(snmpPDUs []gosnmp.SnmpPDU) []string {
	arr := []string{}
	for _, snmpPDU := range snmpPDUs {
		arr = append(arr, FormatOID(snmpPDU.Name))
	}
	return arr
}

func getMacroFromMock(snmpPDUs []gosnmp.SnmpPDU) map[string]string {
	m := map[string]string{}
	for _, snmpPDU := range snmpPDUs {
		k := getSnmpIndex(snmpPDU.Name)
		v, _ := assertString(snmpPDU.Value)
		m[k] = v
	}
	return m
}

func GetPointLineProtos(pts []*point.Point) []string {
	arr := []string{}
	for _, p := range pts {
		s := p.LineProto()
		// remove timestamp
		s = s[:strings.LastIndex(s, " ")]
		arr = append(arr, s)
	}
	return arr
}

//nolint:lll
var zabbixYaml = `zabbix_export:
  version: '6.0'
  date: '2024-05-24T02:19:33Z'
  groups:
    - uuid: 36bff6c29af64692839d077febfc7079
      name: 'Templates/Network devices'
  templates:
    - uuid: ad4c3dad4b7b492685d1fd3bd3a664f9
      template: 'Huawei VRP by SNMP'
      name: 'Huawei VRP by SNMP'
      description: |
        Template Net Huawei VRP
        
        MIBs used:
        EtherLike-MIB
        HUAWEI-ENTITY-EXTENT-MIB
        HOST-RESOURCES-MIB
        SNMPv2-MIB
        ENTITY-MIB
        IF-MIB
        
        Generated by official Zabbix template tool "Templator"
      groups:
        - name: 'Templates/Network devices'
      items:
        - uuid: 78041d1a10ed4cb6b89a77bc1478b009
          name: 'Huawei VRP: System location'
          type: SNMP_AGENT
          snmp_oid: 1.3.6.1.2.1.1.6.0
          key: 'system.location[sysLocation.0]'
          delay: 15m
          history: 7d
          trends: '0'
          value_type: CHAR
          description: |
            MIB: SNMPv2-MIB
            The physical location of this node (e.g., ).  If the location is unknown, the value is the zero-length string.
          inventory_link: LOCATION
          preprocessing:
            - type: DISCARD_UNCHANGED_HEARTBEAT
              parameters:
                - 12h
          tags:
            - tag: component
              value: system
        - uuid: 202c91a61d644285a0d74d88ae1b1a79
          name: 'Huawei VRP: Uptime (hardware)'
          type: SNMP_AGENT
          snmp_oid: 1.3.6.1.2.1.25.1.1.0
          key: 'system.hw.uptime[hrSystemUptime.0]'
          delay: 30s
          history: 7d
          trends: '0'
          units: uptime
          description: |
            MIB: HOST-RESOURCES-MIB
            The amount of time since this host was last initialized. Note that this is different from sysUpTime in the SNMPv2-MIB [RFC1907] because sysUpTime is the uptime of the network management portion of the system.
          preprocessing:
            - type: CHECK_NOT_SUPPORTED
              parameters:
                - ''
              error_handler: CUSTOM_VALUE
              error_handler_params: '0'
            - type: MULTIPLIER
              parameters:
                - '0.01'
          tags:
            - tag: component
              value: system
        - uuid: e9dd0d849aaa499da9e2ef9e21ccd50e
          name: 'Huawei VRP: Uptime (network)'
          type: SNMP_AGENT
          snmp_oid: 1.3.6.1.2.1.1.3.0
          key: 'system.net.uptime[sysUpTime.0]'
          delay: 30s
          history: 7d
          trends: '0'
          units: uptime
          description: |
            MIB: SNMPv2-MIB
            The time (in hundredths of a second) since the network management portion of the system was last re-initialized.
          preprocessing:
            - type: MULTIPLIER
              parameters:
                - '0.01'
          tags:
            - tag: component
              value: system
      discovery_rules:
        - uuid: 597d528e3de944a5807ff37a97763a7e
          name: 'MPU Discovery'
          type: SNMP_AGENT
          snmp_oid: 'discovery[{#ENT_NAME},1.3.6.1.2.1.47.1.1.1.1.7]'
          key: mpu.discovery
          delay: 1h
          filter:
            conditions:
              - macro: '{#ENT_NAME}'
                value: (MPU.*|SRU.*)
                formulaid: A
          description: 'http://support.huawei.com/enterprise/KnowledgebaseReadAction.action?contentId=KB1000090234. Filter limits results to Main Processing Units'
          item_prototypes:
            - uuid: 6fae73ea06ff4e99956085fb807d75ae
              name: '{#ENT_NAME}: CPU utilization'
              type: SNMP_AGENT
              snmp_oid: '1.3.6.1.4.1.2011.5.25.31.1.1.1.1.5.{#SNMPINDEX}'
              key: 'system.cpu.util[hwEntityCpuUsage.{#SNMPINDEX}]'
              history: 7d
              value_type: FLOAT
              units: '%'
              description: |
                MIB: HUAWEI-ENTITY-EXTENT-MIB
                The CPU usage for this entity. Generally, the CPU usage will calculate the overall CPU usage on the entity, and itis not sensible with the number of CPU on the entity.
                Reference: http://support.huawei.com/enterprise/KnowledgebaseReadAction.action?contentId=KB1000090234
              tags:
                - tag: component
                  value: cpu
              trigger_prototypes:
                - uuid: 66bfe7805cae4447a32a6e5877e1231b
                  expression: 'min(/Huawei VRP by SNMP/system.cpu.util[hwEntityCpuUsage.{#SNMPINDEX}],5m)>{$CPU.UTIL.CRIT}'
                  name: '{#ENT_NAME}: High CPU utilization'
                  event_name: '{#ENT_NAME}: High CPU utilization (over {$CPU.UTIL.CRIT}% for 5m)'
                  opdata: 'Current utilization: {ITEM.LASTVALUE1}'
                  priority: WARNING
                  description: 'The CPU utilization is too high. The system might be slow to respond.'
                  tags:
                    - tag: scope
                      value: performance
        - uuid: f6ee388bf9f9484295f800a6ca71b63f
          name: 'Network Interfaces Discovery'
          type: SNMP_AGENT
          snmp_oid: 'discovery[{#SNMPVALUE},1.3.6.1.2.1.2.2.1.8,{#IFADMINSTATUS},1.3.6.1.2.1.2.2.1.7,{#IFALIAS},1.3.6.1.2.1.31.1.1.1.18,{#IFNAME},1.3.6.1.2.1.31.1.1.1.1,{#IFDESCR},1.3.6.1.2.1.2.2.1.2,{#IFTYPE},1.3.6.1.2.1.2.2.1.3]'
          key: net.if.discovery
          delay: '3600'
          filter:
            evaltype: AND
            conditions:
              - macro: '{#IFADMINSTATUS}'
                value: (1|3)
                formulaid: A
              - macro: '{#IFNAME}'
                value: '@Network interfaces for discovery'
                formulaid: B
          description: 'Discovering interfaces from IF-MIB. Interfaces with down(2) Administrative Status are not discovered.'
          item_prototypes:
            - uuid: 8710321c25f54a26a1ad8cfcd867f0a8
              name: 'Interface {#IFNAME}({#IFALIAS}): Speed'
              type: SNMP_AGENT
              snmp_oid: '1.3.6.1.2.1.31.1.1.1.15.{#SNMPINDEX}'
              key: 'net.if.speed[ifHighSpeed.{#SNMPINDEX}]'
              delay: '300'
              history: 1w
              trends: 0d
              units: bps
              description: |
                MIB: IF-MIB
                An estimate of the interface's current bandwidth in units of 1,000,000 bits per second.  If this object reports a value of n+499,999'.  For interfaces which do not vary in bandwidth or for those where no accurate estimation can be made, this object should contain the nominal bandwidth.  For a sub-layer which has no concept of bandwidth, this object should be zero.
              preprocessing:
                - type: MULTIPLIER
                  parameters:
                    - '1000000'
              tags:
                - tag: Application
                  value: 'Network Interfaces'
`
