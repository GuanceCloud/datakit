// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

// Package lldp provides LLDP (Link Layer Discovery Protocol) data collection via SNMP.
package lldp

import (
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/gosnmp/gosnmp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp/snmputil"
)

// SNMP OIDs for LLDP data collection.
const (
	oidIFDescription           = "1.3.6.1.2.1.2.2.1.2"    // IF-MIB::ifDescr - Interface description table
	oidLLDPRemoteTable         = "1.0.8802.1.1.2.1.4.1"   // LLDP-MIB::lldpRemTable - Remote device table
	oidLLDPLocChassisID        = "1.0.8802.1.1.2.1.3.2.0" // LLDP-MIB::lldpLocChassisId - Local device ChassisID (scalar OID with .0)
	oidLLDPLocChassisIDSubtype = "1.0.8802.1.1.2.1.3.1.0" // LLDP-MIB::lldpLocChassisIdSubtype - Local device ChassisID subtype (scalar OID with .0)
	packageName                = "snmp-lldp"
)

// LLDP subtype constants (RFC 8021).
const (
	// Chassis ID Subtype.
	ChassisSubtypeComponent       = 1 // chassis_component
	ChassisSubtypeInterfaceAlias  = 2 // interface_alias
	ChassisSubtypePortComponent   = 3 // port_component
	ChassisSubtypeMACAddress      = 4 // mac_address
	ChassisSubtypeNetworkAddress  = 5 // network_address
	ChassisSubtypeInterfaceName   = 6 // interface_name
	ChassisSubtypeLocallyAssigned = 7 // locally_assigned

	// Port ID Subtype.
	PortSubtypeInterfaceAlias  = 1 // interface_alias
	PortSubtypePortComponent   = 2 // port_component
	PortSubtypeMACAddress      = 3 // mac_address
	PortSubtypeNetworkAddress  = 4 // network_address
	PortSubtypeInterfaceName   = 5 // interface_name
	PortSubtypeAgentCircuitID  = 6 // agent_circuit_id
	PortSubtypeLocallyAssigned = 7 // locally_assigned
)

// LLDP subtype string constants (used as tag values).
const (
	// Chassis subtype strings.
	ChassisSubtypeStrComponent       = "chassis_component"
	ChassisSubtypeStrInterfaceAlias  = "interface_alias"
	ChassisSubtypeStrPortComponent   = "port_component"
	ChassisSubtypeStrMACAddress      = "mac_address"
	ChassisSubtypeStrNetworkAddress  = "network_address"
	ChassisSubtypeStrInterfaceName   = "interface_name"
	ChassisSubtypeStrLocallyAssigned = "locally_assigned"

	// Port subtype strings.
	PortSubtypeStrInterfaceAlias  = "interface_alias"
	PortSubtypeStrPortComponent   = "port_component"
	PortSubtypeStrMACAddress      = "mac_address"
	PortSubtypeStrNetworkAddress  = "network_address"
	PortSubtypeStrInterfaceName   = "interface_name"
	PortSubtypeStrAgentCircuitID  = "agent_circuit_id"
	PortSubtypeStrLocallyAssigned = "locally_assigned"

	SubtypeStrUnknown = "unknown"
)

var (
	l    = logger.DefaultSLogger(packageName)
	once sync.Once
)

func SetLog() {
	once.Do(func() {
		l = logger.SLogger(packageName)
	})
}

// FormatChassisSubtype converts chassis subtype number to human-readable string.
// Returns "unknown" for invalid subtype (0) or unknown values.
func FormatChassisSubtype(subtype int) string {
	switch subtype {
	case ChassisSubtypeComponent:
		return ChassisSubtypeStrComponent
	case ChassisSubtypeInterfaceAlias:
		return ChassisSubtypeStrInterfaceAlias
	case ChassisSubtypePortComponent:
		return ChassisSubtypeStrPortComponent
	case ChassisSubtypeMACAddress:
		return ChassisSubtypeStrMACAddress
	case ChassisSubtypeNetworkAddress:
		return ChassisSubtypeStrNetworkAddress
	case ChassisSubtypeInterfaceName:
		return ChassisSubtypeStrInterfaceName
	case ChassisSubtypeLocallyAssigned:
		return ChassisSubtypeStrLocallyAssigned
	default:
		return SubtypeStrUnknown
	}
}

// FormatPortSubtype converts port subtype number to human-readable string.
// Returns "unknown" for invalid subtype (0) or unknown values.
func FormatPortSubtype(subtype int) string {
	switch subtype {
	case PortSubtypeInterfaceAlias:
		return PortSubtypeStrInterfaceAlias
	case PortSubtypePortComponent:
		return PortSubtypeStrPortComponent
	case PortSubtypeMACAddress:
		return PortSubtypeStrMACAddress
	case PortSubtypeNetworkAddress:
		return PortSubtypeStrNetworkAddress
	case PortSubtypeInterfaceName:
		return PortSubtypeStrInterfaceName
	case PortSubtypeAgentCircuitID:
		return PortSubtypeStrAgentCircuitID
	case PortSubtypeLocallyAssigned:
		return PortSubtypeStrLocallyAssigned
	default:
		return SubtypeStrUnknown
	}
}

// formatColonSepBytes formats byte array as colon-separated hexadecimal string.
func formatColonSepBytes(val []byte) string {
	octetsList := make([]string, 0, len(val)*3)
	for _, b := range val {
		octetsList = append(octetsList, hex.EncodeToString([]byte{b}))
	}
	return strings.Join(octetsList, ":")
}

// walkSNMPTable performs SNMP walk operation using the appropriate method based on SNMP version.
func walkSNMPTable(session snmputil.Session, oid string) ([]gosnmp.SnmpPDU, error) {
	var pdus []gosnmp.SnmpPDU
	var err error

	// Use BulkWalk for SNMPv2c/v3, fallback to GetWalkAll for SNMPv1
	if session.GetVersion() == gosnmp.Version1 {
		pdus, err = session.GetWalkAll(oid)
	} else {
		err = session.BulkWalk(oid, func(pdu gosnmp.SnmpPDU) error {
			pdus = append(pdus, pdu)
			return nil
		})
	}

	if err != nil {
		return nil, fmt.Errorf("snmp walk failed for oid %s: %w", oid, err)
	}

	return pdus, nil
}

func getLocalPortMap(session snmputil.Session, oid string) (map[string]string, error) {
	pdus, err := walkSNMPTable(session, oid)
	if err != nil {
		return nil, err
	}

	portMap := map[string]string{}
	for _, pdu := range pdus {
		oidParts := strings.Split(pdu.Name, ".")
		ifIdx := oidParts[len(oidParts)-1]
		if val, ok := pdu.Value.([]byte); ok {
			portMap[ifIdx] = string(val)
		} else {
			l.Warnf("Unexpected ifDescr type %T for index %s, skipping", pdu.Value, ifIdx)
		}
	}
	return portMap, nil
}

type neighborInfo struct {
	LocalPort string `json:"local_port"`

	ChassisSubType int    `json:"chassis_subtype"`
	ChassisID      string `json:"chassis_id"`
	rawChassisID   []byte

	PortSubType int    `json:"port_subtype"`
	PortID      string `json:"port_id"`
	rawPortID   []byte
	SystemName  string `json:"system_name"`
	SystemDesc  string `json:"system_desc"`
}

//nolint:exhaustive
func pduPretty(pdu *gosnmp.SnmpPDU) string {
	switch pdu.Type {
	case gosnmp.OctetString:
		return fmt.Sprintf("type:%d|name:%s|val: %q", pdu.Type, pdu.Name, string(pdu.Value.([]byte)))
	case gosnmp.Integer:
		return fmt.Sprintf("type:%d|name:%s|val: %d", pdu.Type, pdu.Name, pdu.Value.(int))
	}
	return fmt.Sprintf("type:%d|name:%s | val: %v", pdu.Type, pdu.Name, pdu.Value)
}

func isPrintableASCII(data []byte) bool {
	for _, b := range data {
		if b < 32 || b > 126 {
			return false
		}
	}
	return len(data) > 0
}

// encodeChassisIDBySubtype encodes the ChassisID according to its subtype.
// Returns the encoded ChassisID string.
func encodeChassisIDBySubtype(subtype int, rawChassisID []byte) string {
	if rawChassisID == nil {
		return ""
	}

	switch subtype {
	case ChassisSubtypeComponent:
		return string(rawChassisID)

	case ChassisSubtypeMACAddress:
		return formatColonSepBytes(rawChassisID)

	case ChassisSubtypeNetworkAddress:
		switch len(rawChassisID) {
		case 4:
			return net.IP(rawChassisID).String() // IPv4
		case 16:
			return net.IP(rawChassisID).String() // IPv6
		default:
			return formatColonSepBytes(rawChassisID)
		}

	case ChassisSubtypeInterfaceAlias, ChassisSubtypeInterfaceName:
		return string(rawChassisID)

	case ChassisSubtypeLocallyAssigned:
		return string(rawChassisID)

	default:
		// Unknown subtype: try printable ASCII first, fallback to hex
		if isPrintableASCII(rawChassisID) {
			return string(rawChassisID)
		}
		l.Debugf("Unsupported ChassisSubType: %d", subtype)
		return formatColonSepBytes(rawChassisID)
	}
}

// encodePortIDBySubtype encodes the PortID according to its subtype with intelligent fallback.
// Returns the encoded PortID string.
func encodePortIDBySubtype(subtype int, rawPortID []byte) string {
	if rawPortID == nil {
		return ""
	}

	switch subtype {
	case PortSubtypeInterfaceAlias, PortSubtypePortComponent, PortSubtypeInterfaceName:
		return string(rawPortID)

	case PortSubtypeAgentCircuitID, PortSubtypeLocallyAssigned:
		return string(rawPortID)

	case PortSubtypeMACAddress:
		return formatColonSepBytes(rawPortID)

	case PortSubtypeNetworkAddress:
		switch len(rawPortID) {
		case 6:
			return formatColonSepBytes(rawPortID) // MAC
		case 4:
			return net.IP(rawPortID).String() // IPv4
		case 16:
			return net.IP(rawPortID).String() // IPv6
		default:
			return formatColonSepBytes(rawPortID)
		}

	default:
		// Unknown subtype: try printable ASCII first, fallback to hex
		if isPrintableASCII(rawPortID) {
			return string(rawPortID)
		}
		l.Debugf("Unsupported PortSubType: %d", subtype)
		return formatColonSepBytes(rawPortID)
	}
}

func doParseLLDPData(pdu *gosnmp.SnmpPDU, localPorts map[string]string, tmp map[string]*neighborInfo) {
	if pdu == nil {
		return
	}

	oid := pdu.Name

	// The OID structure is: .1.0.8802.1.1.2.1.4.1.1.{column}.{local_port_index}.{...}
	// We need to extract the column and the local port index.
	oidParts := strings.Split(oid, ".")
	if len(oidParts) < 15 {
		l.Warnf("ignore oid %q", oid)
		return // not a valid LLDP remote entry OID
	}

	offset := 1
	if oidParts[0] == "" { // oid starts with `.'
		offset = 0
	}

	col, localPortIdx := oidParts[11+offset], oidParts[13+offset]
	l.Debugf("col: %s, portIdx: %s, pdu: %s", col, localPortIdx, pduPretty(pdu))

	idxParts := oidParts[12+offset:]
	neighborKey := strings.Join(idxParts, ".")

	// create if not seen before, create it
	if _, ok := tmp[neighborKey]; !ok {
		localPortName, ok := localPorts[localPortIdx]
		if !ok {
			localPortName = "ifIndex_" + localPortIdx // Fallback name
			l.Warnf("Could not find local port name for index %s, using fallback name %s",
				localPortIdx, localPortName)
		}

		tmp[neighborKey] = &neighborInfo{
			LocalPort: localPortName,
		}
	}

	// see reference: https://mibs.observium.org/mib/LLDP-MIB/
	switch col {
	case "4": // lldpRemChassisIdSubtype
		if val, ok := pdu.Value.(int); ok {
			tmp[neighborKey].ChassisSubType = val
		} else {
			l.Warnf("Unexpected ChassisSubType type: %T for neighbor %s", pdu.Value, neighborKey)
		}

	case "5": // lldpRemChassisID
		switch v := pdu.Value.(type) {
		case []byte:
			tmp[neighborKey].rawChassisID = v
		case string:
			tmp[neighborKey].rawChassisID = []byte(v)
		default:
			l.Warnf("Unexpected ChassisID type: %T, value: %v", v, v)
		}

	case "6": // lldpRemPortIdSubtype
		if val, ok := pdu.Value.(int); ok {
			tmp[neighborKey].PortSubType = val
		} else {
			l.Warnf("Unexpected PortSubType type: %T for neighbor %s", pdu.Value, neighborKey)
		}

	case "7": // lldpRemPortId
		switch v := pdu.Value.(type) {
		case []byte:
			tmp[neighborKey].rawPortID = v
		case string:
			tmp[neighborKey].rawPortID = []byte(v)
		default:
			l.Warnf("Unexpected PortID type: %T, value: %v", v, v)
		}

	case "9": // lldpRemSysName
		if val, ok := pdu.Value.([]byte); ok {
			tmp[neighborKey].SystemName = string(val)
		} else {
			l.Warnf("Unexpected SystemName type: %T for neighbor %s", pdu.Value, neighborKey)
		}
	case "10": // lldpRemSysDesc
		if val, ok := pdu.Value.([]byte); ok {
			tmp[neighborKey].SystemDesc = string(val)
		} else {
			l.Warnf("Unexpected SystemDesc type: %T for neighbor %s", pdu.Value, neighborKey)
		}
	default:
		// pass
	}
}

func parseLLDPData(pdus []gosnmp.SnmpPDU, localPorts map[string]string) []*neighborInfo {
	tmp := map[string]*neighborInfo{}

	for i := range pdus {
		doParseLLDPData(&pdus[i], localPorts, tmp)
	}

	for _, ni := range tmp {
		ni.ChassisID = encodeChassisIDBySubtype(ni.ChassisSubType, ni.rawChassisID)
		ni.PortID = encodePortIDBySubtype(ni.PortSubType, ni.rawPortID)
	}

	var res []*neighborInfo
	for _, ni := range tmp {
		res = append(res, ni)
	}

	return res
}

// CollectNeighbors collects LLDP neighbors from a device using SNMP.
func CollectNeighbors(session snmputil.Session, deviceIP string) ([]*neighborInfo, error) {
	if session == nil {
		return nil, fmt.Errorf("SNMP session is nil")
	}

	if deviceIP == "" {
		return nil, fmt.Errorf("device IP is empty")
	}

	// Get local port mapping (ifIndex -> interface name)
	localPorts, err := getLocalPortMap(session, oidIFDescription)
	if err != nil {
		return nil, fmt.Errorf("failed to get local port map for device %s: %w", deviceIP, err)
	}

	l.Infof("Discovered %d local interfaces on device %s", len(localPorts), deviceIP)
	if len(localPorts) == 0 {
		return nil, fmt.Errorf("no local interfaces found on device %s", deviceIP)
	}

	// Walk LLDP remote table
	rawNeighborData, err := walkSNMPTable(session, oidLLDPRemoteTable)
	if err != nil {
		return nil, fmt.Errorf("failed to walk LLDP remote table: %w", err)
	}

	l.Infof("Retrieved %d LLDP PDUs from device %s", len(rawNeighborData), deviceIP)
	if len(rawNeighborData) == 0 {
		l.Debugf("No LLDP neighbors found on device %s", deviceIP)
		return nil, nil
	}

	// Parse LLDP data
	neighbors := parseLLDPData(rawNeighborData, localPorts)

	// Filter invalid entries
	result := make([]*neighborInfo, 0, len(neighbors))
	for _, n := range neighbors {
		if n == nil {
			continue
		}
		result = append(result, n)
	}

	l.Infof("Parsed %d LLDP neighbors from device %s", len(result), deviceIP)

	return result, nil
}

// GetLocalChassisID gets the local device's ChassisID and encodes it according to subtype.
func GetLocalChassisID(session snmputil.Session) (string, int, error) {
	if session == nil {
		return "", 0, fmt.Errorf("SNMP session is nil")
	}

	// Get both ChassisID subtype and ChassisID in a single request
	packet, err := session.Get([]string{oidLLDPLocChassisIDSubtype, oidLLDPLocChassisID})
	if err != nil {
		return "", 0, fmt.Errorf("failed to get lldpLocChassisIdSubtype and lldpLocChassisId: %w", err)
	}
	if len(packet.Variables) != 2 {
		return "", 0, fmt.Errorf("expected 2 variables, but got %d", len(packet.Variables))
	}

	// Find subtype and chassisID variables by OID name (order may vary)
	var subtypeVar, chassisVar *gosnmp.SnmpPDU
	for i := range packet.Variables {
		// Normalize OID name (remove leading dot if present)
		oidName := packet.Variables[i].Name
		if len(oidName) > 0 && oidName[0] == '.' {
			oidName = oidName[1:]
		}

		switch oidName {
		case oidLLDPLocChassisIDSubtype:
			subtypeVar = &packet.Variables[i]
		case oidLLDPLocChassisID:
			chassisVar = &packet.Variables[i]
		}
	}

	if subtypeVar == nil {
		return "", 0, fmt.Errorf("lldpLocChassisIdSubtype not found in response")
	}
	if chassisVar == nil {
		return "", 0, fmt.Errorf("lldpLocChassisId not found in response")
	}

	// Check for SNMP error types (NoSuchObject, NoSuchInstance, Null)
	if subtypeVar.Type == gosnmp.NoSuchObject || subtypeVar.Type == gosnmp.NoSuchInstance || subtypeVar.Type == gosnmp.Null {
		return "", 0, fmt.Errorf("lldpLocChassisIdSubtype does not exist on device")
	}
	if chassisVar.Type == gosnmp.NoSuchObject || chassisVar.Type == gosnmp.NoSuchInstance || chassisVar.Type == gosnmp.Null {
		return "", 0, fmt.Errorf("lldpLocChassisId does not exist on device")
	}

	// Parse subtype
	if subtypeVar.Value == nil {
		return "", 0, fmt.Errorf("lldpLocChassisIdSubtype value is nil")
	}
	subtype, ok := subtypeVar.Value.(int)
	if !ok {
		return "", 0, fmt.Errorf("unexpected subtype type: %T", subtypeVar.Value)
	}

	// Parse ChassisID
	if chassisVar.Value == nil {
		return "", 0, fmt.Errorf("lldpLocChassisId value is nil")
	}
	var rawChassisID []byte
	switch v := chassisVar.Value.(type) {
	case []byte:
		rawChassisID = v
	case string:
		rawChassisID = []byte(v)
	default:
		return "", 0, fmt.Errorf("unexpected ChassisID type: %T", v)
	}

	// Encode ChassisID using existing logic
	encodedChassisID := encodeChassisIDBySubtype(subtype, rawChassisID)

	return encodedChassisID, subtype, nil
}
