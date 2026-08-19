// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmputil

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
)

const (
	topologyLinkSourceTypeLLDP = "lldp"
	topologyLinkSourceTypeCDP  = "cdp"
	ciscoNetworkProtocolIPv4   = "1"
	ciscoNetworkProtocolIPv6   = "20"
	IDTypeMacAddress           = "mac_address"
	IDTypeNetworkAddress       = "network_address"
	IDTypeInterfaceName        = "interface_name"
	IDTypeInterfaceAlias       = "interface_alias"
	lldpAddressFamilyIPv4      = 1
	lldpAddressFamilyIPv6      = 2
)

// ChassisIDSubtypeMap maps LLDP chassis ID subtypes to human-readable values.
var ChassisIDSubtypeMap = map[string]string{
	"1": "chassis_component",
	"2": "interface_alias",
	"3": "port_component",
	"4": "mac_address",
	"5": "network_address",
	"6": "interface_name",
	"7": "local",
}

// PortIDSubTypeMap maps LLDP port ID subtypes to human-readable values.
var PortIDSubTypeMap = map[string]string{
	"1": "interface_alias",
	"2": "port_component",
	"3": "mac_address",
	"4": "network_address",
	"5": "interface_name",
	"6": "agent_circuit_id",
	"7": "local",
}

// TopologyColumnOIDs returns the LLDP/CDP column OIDs needed for topology collection.
func TopologyColumnOIDs() []string {
	return ParseColumnOids(nil, topologyMetadataConfig, true)
}

// BuildNetworkTopologyMetadata builds LLDP links, falling back to CDP when no
// LLDP entries are available.
func BuildNetworkTopologyMetadata(deviceID string, values *ResultValueStore, interfaces []InterfaceMetadata) []TopologyLinkMetadata {
	store := BuildMetadataStore(topologyMetadataConfig, values)
	links := buildNetworkTopologyMetadataWithLLDP(deviceID, store, interfaces)
	if len(links) == 0 {
		links = buildNetworkTopologyMetadataWithCDP(deviceID, store, interfaces)
	}
	return links
}

func buildNetworkTopologyMetadataWithLLDP(deviceID string, store *Store, interfaces []InterfaceMetadata) []TopologyLinkMetadata {
	links := make([]TopologyLinkMetadata, 0)
	interfaceIndexByIDType := buildInterfaceIndexByIDType(interfaces)
	remManAddrByLLDPRemIndexAndLLDPRemLocalPortNum := getRemManIPAddrByLLDPRemIndexAndLLDPRemLocalPortNum(
		store.GetColumnIndexes("lldp_remote_management.interface_id_type"),
	)

	indexes := store.GetColumnIndexes("lldp_remote.interface_id")
	if len(indexes) == 0 {
		l.Debugf("Unable to build links metadata: no lldp_remote indexes found")
		return links
	}
	sort.Strings(indexes)
	for _, strIndex := range indexes {
		indexElems := strings.Split(strIndex, ".")
		if len(indexElems) != 3 {
			l.Debugf("Expected 3 index elements, but got %d, index=`%s`", len(indexElems), strIndex)
			continue
		}

		localPortNum := indexElems[1]
		lldpRemIndex := indexElems[2]

		remoteDeviceIDType := ChassisIDSubtypeMap[store.GetColumnAsString("lldp_remote.chassis_id_type", strIndex)]
		remoteDeviceID := formatID(remoteDeviceIDType, store, "lldp_remote.chassis_id", strIndex)

		remoteInterfaceIDType := PortIDSubTypeMap[store.GetColumnAsString("lldp_remote.interface_id_type", strIndex)]
		remoteInterfaceID := formatID(remoteInterfaceIDType, store, "lldp_remote.interface_id", strIndex)

		localInterfaceIDType := PortIDSubTypeMap[store.GetColumnAsString("lldp_local.interface_id_type", localPortNum)]
		localInterfaceID := formatID(localInterfaceIDType, store, "lldp_local.interface_id", localPortNum)

		resolvedLocalInterfaceID := resolveLocalInterface(deviceID, interfaceIndexByIDType, localInterfaceIDType, localInterfaceID)
		remEntryUniqueID := localPortNum + "." + lldpRemIndex

		links = append(links, TopologyLinkMetadata{
			ID:          deviceID + ":" + remEntryUniqueID,
			SourceType:  topologyLinkSourceTypeLLDP,
			Integration: "snmp",
			Remote: &TopologyLinkSide{
				Device: &TopologyLinkDevice{
					Name:        store.GetColumnAsString("lldp_remote.device_name", strIndex),
					Description: store.GetColumnAsString("lldp_remote.device_desc", strIndex),
					ID:          remoteDeviceID,
					IDType:      remoteDeviceIDType,
					IPAddress:   remManAddrByLLDPRemIndexAndLLDPRemLocalPortNum[buildLLDPRemoteKey(localPortNum, lldpRemIndex)],
				},
				Interface: &TopologyLinkInterface{
					ID:          remoteInterfaceID,
					IDType:      remoteInterfaceIDType,
					Description: store.GetColumnAsString("lldp_remote.interface_desc", strIndex),
				},
			},
			Local: &TopologyLinkSide{
				Interface: &TopologyLinkInterface{
					ResolvedID: resolvedLocalInterfaceID,
					ID:         localInterfaceID,
					IDType:     localInterfaceIDType,
				},
				Device: &TopologyLinkDevice{ResolvedID: deviceID},
			},
		})
	}
	return links
}

func buildNetworkTopologyMetadataWithCDP(deviceID string, store *Store, _ []InterfaceMetadata) []TopologyLinkMetadata {
	links := make([]TopologyLinkMetadata, 0)
	indexes := store.GetColumnIndexes("cdp_remote.interface_id")
	if len(indexes) == 0 {
		l.Debugf("Unable to build links metadata: no cdp_remote indexes found")
		return links
	}
	sort.Strings(indexes)
	for _, strIndex := range indexes {
		indexElems := strings.Split(strIndex, ".")
		if len(indexElems) != 2 {
			l.Debugf("Expected 2 index elements, but got %d, index=`%s`", len(indexElems), strIndex)
			continue
		}

		cdpCacheIfIndex := indexElems[0]
		cdpCacheDeviceIndex := indexElems[1]
		remoteDeviceAddress := getRemDeviceAddressByCDPRemIndex(store, strIndex)
		resolvedLocalInterfaceID := deviceID + ":" + cdpCacheIfIndex
		remEntryUniqueID := cdpCacheIfIndex + "." + cdpCacheDeviceIndex

		links = append(links, TopologyLinkMetadata{
			ID:          deviceID + ":" + remEntryUniqueID,
			SourceType:  topologyLinkSourceTypeCDP,
			Integration: "snmp",
			Remote: &TopologyLinkSide{
				Device: &TopologyLinkDevice{
					Name:        store.GetColumnAsString("cdp_remote.device_name", strIndex),
					Description: store.GetColumnAsString("cdp_remote.device_desc", strIndex),
					ID:          store.GetColumnAsString("cdp_remote.device_id", strIndex),
					IPAddress:   remoteDeviceAddress,
				},
				Interface: &TopologyLinkInterface{
					ID:     store.GetColumnAsString("cdp_remote.interface_id", strIndex),
					IDType: IDTypeInterfaceName,
				},
			},
			Local: &TopologyLinkSide{
				Interface: &TopologyLinkInterface{ResolvedID: resolvedLocalInterfaceID, ID: ""},
				Device:    &TopologyLinkDevice{ResolvedID: deviceID},
			},
		})
	}
	return links
}

func getRemDeviceAddressByCDPRemIndex(store *Store, strIndex string) string {
	remoteDeviceAddress := getRemDeviceAddressIfIPType(store, strIndex, "device_address_type", "device_address")
	if remoteDeviceAddress != "" {
		return remoteDeviceAddress
	}
	remoteDeviceSecondaryAddress := getRemDeviceAddressIfIPType(store, strIndex, "device_secondary_address_type", "device_secondary_address")
	if remoteDeviceSecondaryAddress != "" {
		return remoteDeviceSecondaryAddress
	}
	return getRemDeviceAddressIfIPType(store, strIndex, "device_cache_address_type", "device_cache_address")
}

func getRemDeviceAddressIfIPType(store *Store, strIndex string, addressTypeField string, addressField string) string {
	remoteDeviceAddressType := store.GetColumnAsString("cdp_remote."+addressTypeField, strIndex)
	if remoteDeviceAddressType == ciscoNetworkProtocolIPv4 || remoteDeviceAddressType == ciscoNetworkProtocolIPv6 {
		return store.GetColumnAsIPString("cdp_remote."+addressField, strIndex)
	}
	return ""
}

type interfaceCandidate struct {
	ifIndex    int32
	isPhysical bool
	macAddress string
}

func singlePhysicalCandidateSharingMAC(candidates map[int32]interfaceCandidate) (interfaceCandidate, bool) {
	var found interfaceCandidate
	var physicalCount int
	var sharedMAC string
	for _, candidate := range candidates {
		if candidate.macAddress == "" {
			return interfaceCandidate{}, false
		}
		if sharedMAC == "" {
			sharedMAC = candidate.macAddress
		} else if candidate.macAddress != sharedMAC {
			return interfaceCandidate{}, false
		}
		if candidate.isPhysical {
			found = candidate
			physicalCount++
			if physicalCount > 1 {
				return interfaceCandidate{}, false
			}
		}
	}
	return found, physicalCount == 1
}

func resolveLocalInterface(
	deviceID string,
	interfaceIndexByIDType map[string]map[string][]interfaceCandidate,
	localInterfaceIDType string,
	localInterfaceID string,
) string {
	if localInterfaceID == "" {
		return ""
	}

	var typesToTry []string
	if localInterfaceIDType == "" {
		typesToTry = []string{IDTypeMacAddress, IDTypeInterfaceName, IDTypeInterfaceAlias, "interface_index"}
	} else {
		typesToTry = []string{localInterfaceIDType}
	}
	matchedCandidates := make(map[int32]interfaceCandidate)
	for _, idType := range typesToTry {
		interfaceIndexByIDValue, ok := interfaceIndexByIDType[idType]
		if !ok {
			continue
		}
		for _, candidate := range interfaceIndexByIDValue[localInterfaceID] {
			matchedCandidates[candidate.ifIndex] = candidate
		}
	}
	if len(matchedCandidates) == 1 {
		for ifIndex := range matchedCandidates {
			return deviceID + ":" + strconv.Itoa(int(ifIndex))
		}
	} else if len(matchedCandidates) > 1 {
		if physical, ok := singlePhysicalCandidateSharingMAC(matchedCandidates); ok {
			return deviceID + ":" + strconv.Itoa(int(physical.ifIndex))
		}
	}
	return ""
}

func buildInterfaceIndexByIDType(interfaces []InterfaceMetadata) map[string]map[string][]interfaceCandidate {
	interfaceIndexByIDType := make(map[string]map[string][]interfaceCandidate)
	for _, idType := range []string{IDTypeMacAddress, IDTypeInterfaceName, IDTypeInterfaceAlias, "interface_index"} {
		interfaceIndexByIDType[idType] = make(map[string][]interfaceCandidate)
	}
	for _, devInterface := range interfaces {
		isPhysical := devInterface.IsPhysical != nil && *devInterface.IsPhysical
		candidate := interfaceCandidate{ifIndex: devInterface.Index, isPhysical: isPhysical, macAddress: devInterface.MacAddress}
		interfaceIndexByIDType[IDTypeMacAddress][devInterface.MacAddress] = append(
			interfaceIndexByIDType[IDTypeMacAddress][devInterface.MacAddress], candidate)
		interfaceIndexByIDType[IDTypeInterfaceName][devInterface.Name] = append(interfaceIndexByIDType[IDTypeInterfaceName][devInterface.Name], candidate)
		interfaceIndexByIDType[IDTypeInterfaceAlias][devInterface.Alias] = append(
			interfaceIndexByIDType[IDTypeInterfaceAlias][devInterface.Alias], candidate)
		strIndex := strconv.Itoa(int(devInterface.Index))
		interfaceIndexByIDType["interface_index"][strIndex] = append(interfaceIndexByIDType["interface_index"][strIndex], candidate)
	}
	return interfaceIndexByIDType
}

func buildLLDPRemoteKey(localPortNum, lldpRemIndex string) string {
	return fmt.Sprintf("%s.%s", localPortNum, lldpRemIndex)
}

func getRemManIPAddrByLLDPRemIndexAndLLDPRemLocalPortNum(remManIndexes []string) map[string]string {
	// Keep the selected address independent of the metadata store's map iteration order.
	sort.Strings(remManIndexes)
	remManAddrByRemIndex := make(map[string]string)
	for _, fullIndex := range remManIndexes {
		indexElems := strings.Split(fullIndex, ".")
		if len(indexElems) != 9 {
			continue
		}
		lldpRemLocalPortNum := indexElems[1]
		lldpRemIndex := indexElems[2]
		lldpRemManAddrSubtype := indexElems[3]
		addressLength := indexElems[4]
		if lldpRemManAddrSubtype != "1" || addressLength != "4" {
			continue
		}
		key := buildLLDPRemoteKey(lldpRemLocalPortNum, lldpRemIndex)
		if _, found := remManAddrByRemIndex[key]; !found {
			remManAddrByRemIndex[key] = strings.Join(indexElems[5:], ".")
		}
	}
	return remManAddrByRemIndex
}

func formatID(idType string, store *Store, field string, strIndex string) string {
	if idType == IDTypeMacAddress {
		return formatColonSepBytes(store.GetColumnAsByteArray(field, strIndex))
	}
	if idType == IDTypeNetworkAddress {
		if address := formatNetworkAddress(store.GetColumnAsByteArray(field, strIndex)); address != "" {
			return address
		}
	}
	return store.GetColumnAsString(field, strIndex)
}

func formatNetworkAddress(raw []byte) string {
	if len(raw) < 2 {
		return ""
	}

	address := raw[1:]
	switch raw[0] {
	case lldpAddressFamilyIPv4:
		if len(address) != net.IPv4len {
			return ""
		}
	case lldpAddressFamilyIPv6:
		if len(address) != net.IPv6len {
			return ""
		}
	default:
		return ""
	}
	return net.IP(address).String()
}
