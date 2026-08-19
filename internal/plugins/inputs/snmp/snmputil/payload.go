// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmputil

// PayloadMetadataBatchSize is the number of resources per event payload
// Resources are devices, interfaces, etc.
const PayloadMetadataBatchSize = 100

// DeviceStatus enum type.
type DeviceStatus int32

const (
	// DeviceStatusReachable means the device can be reached by snmp integration.
	DeviceStatusReachable = DeviceStatus(1)
	// DeviceStatusUnreachable means the device cannot be reached by snmp integration.
	DeviceStatusUnreachable = DeviceStatus(2)
)

// NetworkDevicesMetadata contains network devices metadata.
type NetworkDevicesMetadata struct {
	Subnet           string              `json:"subnet"`
	Namespace        string              `json:"namespace"`
	Devices          []DeviceMetadata    `json:"devices,omitempty"`
	Interfaces       []InterfaceMetadata `json:"interfaces,omitempty"`
	IPAddresses      []IPAddressMetadata `json:"ip_addresses,omitempty"`
	CollectTimestamp int64               `json:"collect_timestamp"`
}

// DeviceMetadata contains device metadata.
type DeviceMetadata struct {
	ID           string       `json:"id"`
	IDTags       []string     `json:"id_tags"` // id_tags is the input to produce device.id, it's also used to correlated with device metrics.
	Tags         []string     `json:"tags"`
	IPAddress    string       `json:"ip_address"`
	Status       DeviceStatus `json:"status"`
	Name         string       `json:"name,omitempty"`
	Description  string       `json:"description,omitempty"`
	SysObjectID  string       `json:"sys_object_id,omitempty"`
	Location     string       `json:"location,omitempty"`
	Profile      string       `json:"profile,omitempty"`
	Vendor       string       `json:"vendor,omitempty"`
	Subnet       string       `json:"subnet,omitempty"`
	SerialNumber string       `json:"serial_number,omitempty"`
	Version      string       `json:"version,omitempty"`
	ProductName  string       `json:"product_name,omitempty"`
	Model        string       `json:"model,omitempty"`
	OsName       string       `json:"os_name,omitempty"`
	OsVersion    string       `json:"os_version,omitempty"`
	OsHostname   string       `json:"os_hostname,omitempty"`
	Type         string       `json:"type,omitempty"`
	Uptime       float64      `json:"uptime,omitempty"`
}

// InterfaceMetadata contains interface metadata.
type InterfaceMetadata struct {
	DeviceID    string   `json:"device_id"`
	IDTags      []string `json:"id_tags"` // used to correlate with interface metrics
	Index       int32    `json:"index"`   // IF-MIB ifIndex type is InterfaceIndex (Integer32 (1..2147483647))
	Name        string   `json:"name,omitempty"`
	Alias       string   `json:"alias,omitempty"`
	Description string   `json:"description,omitempty"`
	MacAddress  string   `json:"mac_address,omitempty"`
	AdminStatus int32    `json:"admin_status,omitempty"` // IF-MIB ifAdminStatus type is INTEGER
	OperStatus  int32    `json:"oper_status,omitempty"`  // IF-MIB ifOperStatus type is INTEGER
	Type        int32    `json:"type,omitempty"`         // IF-MIB ifType (RFC7224 IANAifType)
	IsPhysical  *bool    `json:"is_physical,omitempty"`  // true for physical ethernet interface types (6, 62, 69, 117)
}

// IPAddressMetadata contains ip address metadata.
type IPAddressMetadata struct {
	InterfaceID string `json:"interface_id"`
	IPAddress   string `json:"ip_address"`
	Prefixlen   int32  `json:"prefixlen,omitempty"`
}

// TopologyLinkDevice contains device link data.
type TopologyLinkDevice struct {
	ResolvedID  string `json:"resolved_id,omitempty"`
	ID          string `json:"id,omitempty"`
	IDType      string `json:"id_type,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	IPAddress   string `json:"ip_address,omitempty"`
}

// TopologyLinkInterface contains interface link data.
type TopologyLinkInterface struct {
	ResolvedID  string `json:"resolved_id,omitempty"`
	ID          string `json:"id"`
	IDType      string `json:"id_type,omitempty"`
	Description string `json:"description,omitempty"`
}

// TopologyLinkSide contains data for the remote or local side of a link.
type TopologyLinkSide struct {
	Device    *TopologyLinkDevice    `json:"device,omitempty"`
	Interface *TopologyLinkInterface `json:"interface,omitempty"`
}

// TopologyLinkMetadata contains topology interface-to-interface link metadata.
type TopologyLinkMetadata struct {
	ID          string            `json:"id"`
	SourceType  string            `json:"source_type"`
	Integration string            `json:"integration,omitempty"`
	Local       *TopologyLinkSide `json:"local"`
	Remote      *TopologyLinkSide `json:"remote"`
}
