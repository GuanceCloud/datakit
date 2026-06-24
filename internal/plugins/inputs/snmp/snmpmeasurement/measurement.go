// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package snmpmeasurement constains snmp measurement definitions.
package snmpmeasurement

import (
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

//------------------------------------------------------------------------------

const (
	InputName      = "snmp"
	SNMPObjectName = "snmp_object"
	SNMPMetricName = "snmp_metric"
	SNMPLLDPName   = "snmp_lldp"
)

//------------------------------------------------------------------------------

// SNMPObject ...
type SNMPObject struct {
	Name   string
	Tags   map[string]string
	Fields map[string]interface{}
	TS     time.Time
}

// Point implement MeasurementV2.
func (m *SNMPObject) Point() *point.Point {
	opts := point.DefaultObjectOptions()
	opts = append(opts, point.WithTime(m.TS))

	return point.NewPoint(m.Name,
		append(point.NewTags(m.Tags), point.NewKVs(m.Fields)...),
		opts...)
}

//nolint:lll
func (m *SNMPObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   SNMPObjectName,
		Desc:   "SNMP device object data.",
		DescZh: "SNMP 设备对象数据。",
		Cat:    point.Object,
		Fields: map[string]interface{}{
			"device_meta":    &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit, Desc: "Device meta data (JSON format)."},
			"uptime":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationSecond, Desc: "Device uptime in seconds."},
			"interfaces":     &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit, Desc: "Device network interfaces (JSON format)."},
			"sensors":        &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit, Desc: "Device sensors (JSON format)."},
			"mems":           &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit, Desc: "Device memories (JSON format)."},
			"mem_pool_names": &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit, Desc: "Device memory pool names (JSON format)."},
			"cpus":           &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit, Desc: "Device CPUs (JSON format)."},
			"all":            &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit, Desc: "Device all data (JSON format). (Deprecated)"},
		},
		Tags: map[string]interface{}{
			"device_type":     inputs.NewTagInfo("Device type (e.g. router, switch, pdu)."),
			"device_vendor":   inputs.NewTagInfo("Device vendor."),
			"device_hostname": inputs.NewTagInfo("Device hostname from SNMP (e.g. sysName)."),
			"host":            inputs.NewTagInfo("Device host, replace with IP."),
			"ip":              inputs.NewTagInfo("Device IP."),
			"name":            inputs.NewTagInfo("Device name, replace with IP."),
			"snmp_profile":    inputs.NewTagInfo("Device SNMP profile file."),
			"snmp_host":       inputs.NewTagInfo("Device host."),
		},
	}
}

//------------------------------------------------------------------------------

// SNMPMetric ...
type SNMPMetric struct {
	Name   string
	Tags   map[string]string
	Fields map[string]interface{}
	TS     time.Time
}

// Point implement MeasurementV2.
func (m *SNMPMetric) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.TS))

	return point.NewPoint(m.Name,
		append(point.NewTags(m.Tags), point.NewKVs(m.Fields)...),
		opts...)
}

//nolint:lll
func (m *SNMPMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   SNMPMetricName,
		Desc:   "SNMP device metric data.",
		DescZh: "SNMP 设备指标数据。",
		Cat:    point.Metric,
		Fields: map[string]interface{}{
			"ifNumber":                          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of interface."},
			"sysUpTimeInstance":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The time (in hundredths of a second) since the network management portion of the system was last re-initialized."},
			"tcpActiveOpens":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of times that TCP connections have made a direct transition to the SYN-SENT state from the CLOSED state."},
			"tcpAttemptFails":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of times that TCP connections have made a direct transition to the CLOSED state from either the SYN-SENT state or the SYN-RCVD state, or to the LISTEN state from the SYN-RCVD state."},
			"tcpCurrEstab":                      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of TCP connections for which the current state is either ESTABLISHED or CLOSE-WAIT."},
			"tcpEstabResets":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of times that TCP connections have made a direct transition to the CLOSED state from either the ESTABLISHED state or the CLOSE-WAIT state."},
			"tcpInErrs":                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "(Shown as segment) The total number of segments received in error (e.g., bad TCP checksums)."},
			"tcpOutRsts":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "(Shown as segment) The number of TCP segments sent containing the RST flag."},
			"tcpPassiveOpens":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "(Shown as connection) The number of times TCP connections have made a direct transition to the SYN-RCVD state from the LISTEN state."},
			"tcpRetransSegs":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "(Shown as segment) The total number of segments retransmitted; that is, the number of TCP segments transmitted containing one or more previously transmitted octets."},
			"udpInErrors":                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "(Shown as datagram) The number of received UDP datagram that could not be delivered for reasons other than the lack of an application at the destination port."},
			"udpNoPorts":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "(Shown as datagram) The total number of received UDP datagram for which there was no application at the destination port."},
			"ifAdminStatus":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NoUnit, Desc: "The desired state of the interface."},
			"ifHCInBroadcastPkts":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "(Shown as packet) The number of packets delivered by this sub-layer to a higher (sub-)layer that were addressed to a broadcast address at this sub-layer."},
			"ifHCInMulticastPkts":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "(Shown as packet) The number of packets delivered by this sub-layer to a higher (sub-)layer which were addressed to a multicast address at this sub-layer."},
			"ifHCInOctetsRate":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "(Shown as byte) The total number of octets received on the interface including framing characters."},
			"ifHCInUcastPkts":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "(Shown as packet) The number of packets delivered by this sub-layer to a higher (sub-)layer that were not addressed to a multicast or broadcast address at this sub-layer."},
			"ifHCOutBroadcastPkts":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "(Shown as packet) The total number of packets that higher-level protocols requested be transmitted that were addressed to a broadcast address at this sub-layer, including those that were discarded or not sent."},
			"ifHCOutMulticastPkts":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "(Shown as packet) The total number of packets that higher-level protocols requested be transmitted that were addressed to a multicast address at this sub-layer including those that were discarded or not sent."},
			"ifHCOutOctetsRate":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "(Shown as byte) The total number of octets transmitted out of the interface including framing characters."},
			"ifHCOutUcastPkts":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "(Shown as packet) The total number of packets higher-level protocols requested be transmitted that were not addressed to a multicast or broadcast address at this sub-layer including those that were discarded or not sent."},
			"ifInDiscardsRate":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "(Shown as packet) The number of inbound packets chosen to be discarded even though no errors had been detected to prevent them being deliverable to a higher-layer protocol."},
			"ifInErrorsRate":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "(Shown as packet) The number of inbound packets that contained errors preventing them from being deliverable to a higher-layer protocol."},
			"ifOutDiscardsRate":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "(Shown as packet) The number of outbound packets chosen to be discarded even though no errors had been detected to prevent them being transmitted."},
			"ifOutErrorsRate":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "(Shown as packet) The number of outbound packets that could not be transmitted because of errors."},
			"ifBandwidthInUsageRate":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "(Shown as percent) The percent rate of used received bandwidth."},
			"ifBandwidthOutUsageRate":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "(Shown as percent) The percent rate of used sent bandwidth."},
			"cieIfLastOutTime":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "[Cisco only] (Shown as millisecond) The elapsed time in milliseconds since the last protocol output packet was transmitted."},
			"cieIfOutputQueueDrops":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "[Cisco only] (Shown as packet) The number of output packets dropped by the interface even though no error was detected to prevent them being transmitted."},
			"ciscoMemoryPoolUsed":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "[Cisco only] Indicates the number of bytes from the memory pool that are currently in use by applications on the managed device."},
			"cpmCPUTotalMonIntervalValue":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "[Cisco only] (Shown as percent) The overall CPU busy percentage in the last cpmCPUMonInterval period."},
			"cieIfLastInTime":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "[Cisco only] (Shown as millisecond) The elapsed time in milliseconds since the last protocol input packet was received."},
			"cieIfResetCount":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "[Cisco only] The number of times the interface was internally reset and brought up."},
			"ciscoMemoryPoolLargestFree":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "[Cisco only] Indicates the largest number of contiguous bytes from the memory pool that are currently unused on the managed device."},
			"ciscoEnvMonTemperatureStatusValue": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "[Cisco only] The current value of the test point being instrumented."},
			"ciscoEnvMonSupplyState":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "[Cisco only] The current state of the power supply being instrumented."},
			"cswStackPortOperStatus":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "[Cisco only] The state of the stack port."},
			"cpmCPUTotal1minRev":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "[Cisco only] [Shown as percent] The overall CPU busy percentage in the last 1 minute period."},
			"ciscoMemoryPoolFree":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "[Cisco only] Indicates the number of bytes from the memory pool that are currently unused on the managed device."},
			"cieIfInputQueueDrops":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "[Cisco only] (Shown as packet) The number of input packets dropped."},
			"ciscoEnvMonFanState":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "[Cisco only] The current state of the fan being instrumented."},
			"cswSwitchState":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "[Cisco only] The current state of a switch."},
			"entSensorValue":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "[Cisco only] The most recent measurement seen by the sensor."},

			"uptime":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "(in hundredths of a second, sysUpTime raw) uptime."},
			"netUptime":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "(in hundredths of a second, sysUpTime raw) net uptime."},
			"uptimeTimestamp": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.TimestampSec, Desc: "uptime timestamp."},
			"temperature":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Celsius, Desc: "The Temperature of item."},
			"voltage":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Volt, Desc: "The Volt of item."},
			"voltageStatus":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Bool, Desc: "The voltage status of item."},
			"current":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownType, Desc: "The current of item."},
			"power":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownType, Desc: "The power of item."},
			"powerStatus":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownType, Desc: "The power of item."},
			"fanSpeed":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RotationRete, Desc: "The fan speed."},
			"fanStatus":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Bool, Desc: "The fan status."},

			"ifNetStatus":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Bool, Desc: "The net status."},
			"ifNetConnStatus":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Bool, Desc: "The net connection status."},
			"ifOperStatus":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "(Shown as packet) The current operational state of the interface."},
			"ifStatus":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Bool, Desc: "The interface status."},
			"ifSpeed":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "An estimate of the interface's current bandwidth in bits per second, or the nominal bandwidth."},
			"ifHighSpeed":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "An estimate of the interface's current bandwidth in units of 1,000,000 bits per second, or the nominal bandwidth."},
			"ifInDiscards":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "(Shown as packet) The number of inbound packets chosen to be discarded even though no errors had been detected to prevent them being deliverable to a higher-layer protocol."},
			"ifOutDiscards":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "(Shown as packet) The number of outbound packets chosen to be discarded even though no errors had been detected to prevent them being transmitted."},
			"ifInErrors":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "(Shown as packet) The number of inbound packets that contained errors preventing them from being deliverable to a higher-layer protocol."},
			"ifOutErrors":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "(Shown as packet) The number of outbound packets that could not be transmitted because of errors."},
			"ifHCInPkts":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "(Shown as packet) The number of packets delivered by this sub-layer to a higher (sub-)layer that were not addressed to a multicast or broadcast address at this sub-layer."},
			"ifHCOutPkts":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "(Shown as packet) The total number of packets higher-level protocols requested be transmitted that were not addressed to a multicast or broadcast address at this sub-layer including those that were discarded or not sent."},
			"ifHCInOctets":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "(Shown as byte) The total number of octets received on the interface including framing characters."},
			"ifHCOutOctets":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "(Shown as byte) The total number of octets transmitted out of the interface including framing characters."},
			"ifHCInOctetsDiff":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Inbound octet increment in the collection interval."},
			"ifHCOutOctetsDiff":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Outbound octet increment in the collection interval."},
			"ifInErrorsDiff":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Inbound error packet increment in the collection interval."},
			"ifOutErrorsDiff":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Outbound error packet increment in the collection interval."},
			"ifInDiscardsDiff":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Inbound discard packet increment in the collection interval."},
			"ifOutDiscardsDiff":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Outbound discard packet increment in the collection interval."},
			"ifBandwidthInUsageDiff":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NoUnit, Desc: "Inbound bandwidth usage scaled-value increment in the collection interval."},
			"ifBandwidthOutUsageDiff": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NoUnit, Desc: "Outbound bandwidth usage scaled-value increment in the collection interval."},

			"cpuUsage":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "(Shown as percent) Percentage of CPU currently being used."},
			"cpuTemperature": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Celsius, Desc: "The Temperature of cpu."},
			"cpuStatus":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Bool, Desc: "CPU status."},

			"memoryTotal":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "(Shown as byte) Number of bytes of memory."},
			"memoryUsage":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "(Shown as percent) The percentage of memory currently being used."},
			"memoryUsed":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "(Shown as byte) Number of bytes of memory currently being used."},
			"memoryFree":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "(Shown as percent) The percentage of memory not being used."},
			"memoryAvailable": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "(Shown as byte) Number of memory available."},

			"diskTotal":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total of disk size."},
			"diskUsage":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "(Shown as percent) The percentage of disk currently being used."},
			"diskUsed":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Number of disk currently being used."},
			"diskFree":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "(Shown as percent) The percentage of disk not being used."},
			"diskAvailable": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Number of disk available."},

			"itemTotal":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownType, Desc: "Item total."},
			"itemUsage":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "(Shown as percent) Item being used."},
			"itemUsed":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownType, Desc: "Item being used."},
			"itemFree":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "(Shown as percent) Item not being used."},
			"itemAvailable": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownType, Desc: "Item available."},
		},
		Tags: map[string]interface{}{
			"device_vendor":      inputs.NewTagInfo("Device vendor."),
			"device_namespace":   inputs.NewTagInfo("Device namespace."),
			"host":               inputs.NewTagInfo("Device host, replace with IP."),
			"ip":                 inputs.NewTagInfo("Device IP."),
			"name":               inputs.NewTagInfo("Device name and IP."),
			"snmp_profile":       inputs.NewTagInfo("Device SNMP profile file."),
			"snmp_host":          inputs.NewTagInfo("Device host."),
			"interface":          inputs.NewTagInfo("Device interface. Optional."),
			"interface_alias":    inputs.NewTagInfo("Device interface alias. Optional."),
			"mac_addr":           inputs.NewTagInfo("Device MAC address. Optional."),
			"entity_name":        inputs.NewTagInfo("Device entity name. Optional."),
			"power_source":       inputs.NewTagInfo("Power source. Optional."),
			"power_status_descr": inputs.NewTagInfo("Power status description. Optional."),
			"temp_index":         inputs.NewTagInfo("Temperature index. Optional."),
			"temp_state":         inputs.NewTagInfo("Temperature state. Optional."),
			"cpu":                inputs.NewTagInfo("CPU index. Optional."),
			"mem":                inputs.NewTagInfo("Memory index. Optional."),
			"mem_pool_name":      inputs.NewTagInfo("Memory pool name. Optional."),
			"sensor_id":          inputs.NewTagInfo("Sensor ID. Optional."),
			"sensor_type":        inputs.NewTagInfo("Sensor type. Optional."),
			"snmp_index":         inputs.NewTagInfo("Macro value. Optional."),
			"snmp_value":         inputs.NewTagInfo("Macro value. Optional."),
			"unit_class":         inputs.NewTagInfo("Macro value. Optional."),
			"unit_name":          inputs.NewTagInfo("Macro value. Optional."),
			"unit_alias":         inputs.NewTagInfo("Macro value. Optional."),
			"unit_type":          inputs.NewTagInfo("Macro value. Optional."),
			"unit_desc":          inputs.NewTagInfo("Macro value. Optional."),
			"unit_status":        inputs.NewTagInfo("Macro value. Optional."),
			"unit_locale":        inputs.NewTagInfo("Macro value. Optional."),
			"oid":                inputs.NewTagInfo("OID."),
			"sys_name":           inputs.NewTagInfo("System name."),
			"sys_object_id":      inputs.NewTagInfo("System object id."),
			"device_type":        inputs.NewTagInfo("Device vendor."),
		},
	}
}

//------------------------------------------------------------------------------

// SNMPLLDP ...
type SNMPLLDP struct {
	Name   string
	Tags   map[string]string
	Fields map[string]interface{}
	TS     time.Time
}

// Point implement MeasurementV2.
func (m *SNMPLLDP) Point() *point.Point {
	opts := point.DefaultLoggingOptions()
	opts = append(opts, point.WithTime(m.TS))

	return point.NewPoint(m.Name,
		append(point.NewTags(m.Tags), point.NewKVs(m.Fields)...),
		opts...)
}

//nolint:lll
func (m *SNMPLLDP) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   SNMPLLDPName,
		Desc:   "SNMP LLDP (Link Layer Discovery Protocol) topology data.",
		DescZh: "SNMP LLDP（链路层发现协议）拓扑数据。",
		Cat:    point.Logging,
		Fields: map[string]interface{}{
			"remote_system":      &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit, Desc: "Name of the remote system."},
			"remote_system_desc": &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit, Desc: "Description of the remote system."},
		},
		Tags: map[string]interface{}{
			"local_ip":               inputs.NewTagInfo("IP address of the local device (string)."),
			"local_chassis_id":       inputs.NewTagInfo("Chassis ID of the local device (string)."),
			"local_interface":        inputs.NewTagInfo("Local interface name."),
			"local_chassis_subtype":  inputs.NewTagInfo("Local chassis ID subtype string (e.g., 'mac_address', 'network_address', 'chassis_component', 'locally_assigned', etc.)."),
			"remote_chassis_id":      inputs.NewTagInfo("Chassis ID of the remote device (string)."),
			"remote_interface":       inputs.NewTagInfo("Interface ID of the remote device."),
			"remote_chassis_subtype": inputs.NewTagInfo("Remote chassis ID subtype string (e.g., 'mac_address', 'network_address', 'chassis_component', 'locally_assigned', etc.)."),
			"remote_port_subtype":    inputs.NewTagInfo("Remote port ID subtype string (e.g., 'mac_address', 'network_address', 'interface_alias', 'agent_circuit_id', 'locally_assigned', etc.)."),
		},
	}
}

//------------------------------------------------------------------------------
