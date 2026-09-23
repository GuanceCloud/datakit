// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmputil

import (
	"net"
	"sort"
	"strconv"
)

func BuildMetadataStore(metadataConfigs MetadataConfig, values *ResultValueStore) *Store {
	metadataStore := NewMetadataStore()
	if values == nil {
		return metadataStore
	}

	for resourceName, metadataConfig := range metadataConfigs {
		for fieldName, field := range metadataConfig.Fields {
			fieldFullName := resourceName + "." + fieldName

			var symbols []SymbolConfig
			if field.Symbol.OID != "" {
				symbols = append(symbols, field.Symbol)
			}
			symbols = append(symbols, field.Symbols...)

			if IsMetadataResourceWithScalarOids(resourceName) {
				for _, symbol := range symbols {
					if metadataStore.ScalarFieldHasValue(fieldFullName) {
						break
					}
					value, err := getScalarValueFromSymbol(values, symbol)
					if err != nil {
						l.Debugf("error getting scalar value: %v", err)
						continue
					}
					metadataStore.AddScalarValue(fieldFullName, value)
				}
				if field.Value != "" && !metadataStore.ScalarFieldHasValue(fieldFullName) {
					metadataStore.AddScalarValue(fieldFullName, ResultValue{Value: field.Value})
				}
			} else {
				for _, symbol := range symbols {
					metricValues, err := getColumnValueFromSymbol(values, symbol)
					if err != nil {
						continue
					}
					for fullIndex, value := range metricValues {
						metadataStore.AddColumnValue(fieldFullName, fullIndex, value)
					}
				}
			}
		}
		indexOid := GetIndexOIDForResource(resourceName)
		if indexOid != "" {
			indexes, err := values.GetColumnIndexes(indexOid)
			if err != nil {
				continue
			}
			for _, fullIndex := range indexes {
				// TODO: Support extract value see II-635
				idTags := getTagsFromMetricTagConfigList(metadataConfig.IDTags, fullIndex, values)
				metadataStore.AddIDTags(resourceName, fullIndex, idTags)
			}
		}
	}
	return metadataStore
}

func BuildNetworkInterfacesMetadata(deviceID string, store *Store) []InterfaceMetadata {
	if store == nil {
		// it's expected that the value store is nil if we can't reach the device
		// in that case, we just return a nil slice.
		return nil
	}
	indexes := store.GetColumnIndexes("interface.name")
	if len(indexes) == 0 {
		l.Debugf("Unable to build interfaces metadata: no interface indexes found")
		return nil
	}
	sort.Strings(indexes)
	var interfaces []InterfaceMetadata
	for _, strIndex := range indexes {
		index, err := strconv.ParseInt(strIndex, 10, 32)
		if err != nil {
			l.Warnf("interface metadata: invalid index: %d", index)
			continue
		}

		ifIDTags := store.GetIDTags("interface", strIndex)

		name := store.GetColumnAsString("interface.name", strIndex)
		ifType := int32(store.GetColumnAsFloat("interface.type", strIndex))

		var isPhysical *bool
		if ifType != 0 {
			physical := ifType == 6 || ifType == 62 || ifType == 69 || ifType == 117
			isPhysical = &physical
		}

		networkInterface := InterfaceMetadata{
			DeviceID:    deviceID,
			Index:       int32(index),
			Name:        name,
			Alias:       store.GetColumnAsString("interface.alias", strIndex),
			Description: store.GetColumnAsString("interface.description", strIndex),
			MacAddress:  store.GetColumnAsString("interface.mac_address", strIndex),
			AdminStatus: int32(store.GetColumnAsFloat("interface.admin_status", strIndex)),
			OperStatus:  int32(store.GetColumnAsFloat("interface.oper_status", strIndex)),
			Type:        ifType,
			IsPhysical:  isPhysical,
			IDTags:      ifIDTags,
		}
		interfaces = append(interfaces, networkInterface)
	}
	return interfaces
}

func BuildNetworkIPAddressesMetadata(deviceID string, store *Store) []IPAddressMetadata {
	if store == nil {
		// it's expected that the value store is nil if we can't reach the device
		// in that case, we just return a nil slice.
		return nil
	}
	indexes := store.GetColumnIndexes("ip_addresses.if_index")
	if len(indexes) == 0 {
		l.Debugf("Unable to build ip addresses metadata: no ip_addresses.if_index found")
		return nil
	}
	sort.Strings(indexes)
	var ipAddresses []IPAddressMetadata
	for _, strIndex := range indexes {
		index := store.GetColumnAsString("ip_addresses.if_index", strIndex)
		Netmask := store.GetColumnAsString("ip_addresses.netmask", strIndex)
		ipAddress := IPAddressMetadata{
			InterfaceID: deviceID + ":" + index,
			IPAddress:   strIndex,
			Prefixlen:   int32(netmaskToPrefixlen(Netmask)),
		}
		ipAddresses = append(ipAddresses, ipAddress)
	}
	return ipAddresses
}

func netmaskToPrefixlen(netmask string) int {
	if netmask == "" {
		return 0
	}
	ip := net.ParseIP(netmask)
	if ip == nil {
		return 0
	}
	ipv4 := ip.To4()
	if ipv4 == nil {
		return 0
	}
	stringMask := net.IPMask(ipv4)
	length, _ := stringMask.Size()
	return length
}
