// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package snmputil

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveLocalInterfaceWithSharedMAC(t *testing.T) {
	physical := true
	logical := false
	interfaces := []InterfaceMetadata{
		{Index: 12, MacAddress: "82:a5:6e:a5:c9:01", IsPhysical: &physical},
		{Index: 120, MacAddress: "82:a5:6e:a5:c9:01", IsPhysical: &logical},
	}

	resolvedID := resolveLocalInterface(
		"prod:192.0.2.10",
		buildInterfaceIndexByIDType(interfaces),
		IDTypeMacAddress,
		"82:a5:6e:a5:c9:01",
	)

	assert.Equal(t, "prod:192.0.2.10:12", resolvedID)
}

func TestResolveLocalInterfaceRejectsAmbiguousPhysicalCandidates(t *testing.T) {
	physical := true
	interfaces := []InterfaceMetadata{
		{Index: 12, MacAddress: "82:a5:6e:a5:c9:01", IsPhysical: &physical},
		{Index: 13, MacAddress: "82:a5:6e:a5:c9:01", IsPhysical: &physical},
	}

	resolvedID := resolveLocalInterface(
		"prod:192.0.2.10",
		buildInterfaceIndexByIDType(interfaces),
		IDTypeMacAddress,
		"82:a5:6e:a5:c9:01",
	)

	assert.Empty(t, resolvedID)
}

func TestBuildNetworkTopologyMetadataSkipsMalformedIndexes(t *testing.T) {
	t.Run("lldp", func(t *testing.T) {
		store := NewMetadataStore()
		store.AddColumnValue("lldp_remote.interface_id", "7.1", ResultValue{Value: "Ethernet1/7"})

		links := buildNetworkTopologyMetadataWithLLDP("prod:192.0.2.10", store, nil)

		assert.Empty(t, links)
	})

	t.Run("cdp", func(t *testing.T) {
		store := NewMetadataStore()
		store.AddColumnValue("cdp_remote.interface_id", "12.3.1", ResultValue{Value: "Ethernet1/1"})

		links := buildNetworkTopologyMetadataWithCDP("prod:192.0.2.10", store, nil)

		assert.Empty(t, links)
	})
}

func TestGetRemoteDeviceAddressByCDPFallbackOrder(t *testing.T) {
	const index = "12.3"

	tests := []struct {
		name     string
		populate func(*Store)
		expected string
	}{
		{
			name: "primary management address",
			populate: func(store *Store) {
				addCDPAddress(store, index, "device_address_type", "device_address", ciscoNetworkProtocolIPv4, net.ParseIP("192.0.2.1").To4())
				addCDPAddress(store, index, "device_secondary_address_type", "device_secondary_address", ciscoNetworkProtocolIPv4, net.ParseIP("192.0.2.2").To4())
			},
			expected: "192.0.2.1",
		},
		{
			name: "secondary management address",
			populate: func(store *Store) {
				addCDPAddress(store, index, "device_address_type", "device_address", "0", nil)
				addCDPAddress(store, index, "device_secondary_address_type", "device_secondary_address", ciscoNetworkProtocolIPv4, net.ParseIP("192.0.2.2").To4())
				addCDPAddress(store, index, "device_cache_address_type", "device_cache_address", ciscoNetworkProtocolIPv4, net.ParseIP("192.0.2.3").To4())
			},
			expected: "192.0.2.2",
		},
		{
			name: "cached IPv6 address",
			populate: func(store *Store) {
				addCDPAddress(store, index, "device_address_type", "device_address", "0", nil)
				addCDPAddress(store, index, "device_secondary_address_type", "device_secondary_address", "0", nil)
				addCDPAddress(store, index, "device_cache_address_type", "device_cache_address", ciscoNetworkProtocolIPv6, net.ParseIP("2001:db8::3").To16())
			},
			expected: "2001:db8::3",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := NewMetadataStore()
			test.populate(store)

			assert.Equal(t, test.expected, getRemDeviceAddressByCDPRemIndex(store, index))
		})
	}
}

func TestGetRemoteManagementAddressByLLDPIndex(t *testing.T) {
	addresses := getRemManIPAddrByLLDPRemIndexAndLLDPRemLocalPortNum([]string{
		"malformed",
		"0.7.1.1.4.10.250.0.6",
		"0.8.2.2.16.32.1.13.184.0.0.0.0.0.0.0.0.0.0.0.1",
	})

	assert.Equal(t, map[string]string{"7.1": "10.250.0.6"}, addresses)
}

func TestGetRemoteManagementAddressByLLDPIndexIsStable(t *testing.T) {
	indexes := []string{
		"100.7.1.1.4.192.0.2.20",
		"100.7.1.1.4.192.0.2.10",
		"200.7.1.1.4.192.0.2.30",
		"200.7.1.1.4.192.0.2.25",
	}
	reversedIndexes := []string{indexes[3], indexes[2], indexes[1], indexes[0]}
	expected := map[string]string{"7.1": "192.0.2.10"}

	assert.Equal(t, expected, getRemManIPAddrByLLDPRemIndexAndLLDPRemLocalPortNum(indexes))
	assert.Equal(t, expected, getRemManIPAddrByLLDPRemIndexAndLLDPRemLocalPortNum(reversedIndexes))
}

func TestFormatNetworkAddressID(t *testing.T) {
	store := NewMetadataStore()
	store.AddColumnValue("device.id", "ipv4", ResultValue{Value: []byte{1, 192, 0, 2, 1}})
	store.AddColumnValue("device.id", "ipv6", ResultValue{Value: append([]byte{2}, net.ParseIP("2001:db8::1").To16()...)})
	store.AddColumnValue("device.id", "invalid", ResultValue{Value: []byte{1, 192, 0, 2}})
	store.AddColumnValue("device.id", "string", ResultValue{Value: "device-id"})

	assert.Equal(t, "192.0.2.1", formatID(IDTypeNetworkAddress, store, "device.id", "ipv4"))
	assert.Equal(t, "2001:db8::1", formatID(IDTypeNetworkAddress, store, "device.id", "ipv6"))
	assert.Equal(t, "0x01c00002", formatID(IDTypeNetworkAddress, store, "device.id", "invalid"))
	assert.Equal(t, "device-id", formatID(IDTypeNetworkAddress, store, "device.id", "string"))
}

func addCDPAddress(store *Store, index, typeField, addressField, addressType string, address []byte) {
	store.AddColumnValue("cdp_remote."+typeField, index, ResultValue{Value: addressType})
	store.AddColumnValue("cdp_remote."+addressField, index, ResultValue{Value: address})
}
