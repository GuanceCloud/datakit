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

	return point.NewPointV2(m.Name,
		append(point.NewTags(m.Tags), point.NewKVs(m.Fields)...),
		opts...)
}

//nolint:lll
func (m *SNMPObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: SNMPObjectName,
		Desc: "SNMP device object data.",
		Type: "object",
		Fields: map[string]interface{}{
			"device_meta":    newOtherFieldInfo(inputs.String, inputs.String, inputs.NoUnit, "Device meta data (JSON format)."),
			"interfaces":     newOtherFieldInfo(inputs.String, inputs.String, inputs.NoUnit, "Device network interfaces (JSON format)."),
			"sensors":        newOtherFieldInfo(inputs.String, inputs.String, inputs.NoUnit, "Device sensors (JSON format)."),
			"mems":           newOtherFieldInfo(inputs.String, inputs.String, inputs.NoUnit, "Device memories (JSON format)."),
			"mem_pool_names": newOtherFieldInfo(inputs.String, inputs.String, inputs.NoUnit, "Device memory pool names (JSON format)."),
			"cpus":           newOtherFieldInfo(inputs.String, inputs.String, inputs.NoUnit, "Device CPUs (JSON format)."),
			"all":            newOtherFieldInfo(inputs.String, inputs.String, inputs.NoUnit, "Device all data (JSON format)."),
		},
		Tags: map[string]interface{}{
			"device_vendor": inputs.NewTagInfo("Device vendor."),
			"host":          inputs.NewTagInfo("Device host, replace with IP."),
			"ip":            inputs.NewTagInfo("Device IP."),
			"name":          inputs.NewTagInfo("Device name, replace with IP."),
			"snmp_profile":  inputs.NewTagInfo("Device SNMP profile file."),
			"snmp_host":     inputs.NewTagInfo("Device host."),
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

	return point.NewPointV2(m.Name,
		append(point.NewTags(m.Tags), point.NewKVs(m.Fields)...),
		opts...)
}

//nolint:lll
func (m *SNMPMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: SNMPMetricName,
		Desc: "SNMP device metric data.",
		Type: "metric",
		Fields: map[string]interface{}{
			"ifNumber":                          newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.NCount, "Number of interface."),
			"sysUpTimeInstance":                 newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.NCount, "The time (in hundredths of a second) since the network management portion of the system was last re-initialized."),
			"tcpActiveOpens":                    newOtherFieldInfo(inputs.Float, inputs.Count, inputs.NCount, "The number of times that TCP connections have made a direct transition to the SYN-SENT state from the CLOSED state."),
			"tcpAttemptFails":                   newOtherFieldInfo(inputs.Float, inputs.Count, inputs.NCount, "The number of times that TCP connections have made a direct transition to the CLOSED state from either the SYN-SENT state or the SYN-RCVD state, or to the LISTEN state from the SYN-RCVD state."),
			"tcpCurrEstab":                      newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.NCount, "The number of TCP connections for which the current state is either ESTABLISHED or CLOSE-WAIT."),
			"tcpEstabResets":                    newOtherFieldInfo(inputs.Float, inputs.Count, inputs.NCount, "The number of times that TCP connections have made a direct transition to the CLOSED state from either the ESTABLISHED state or the CLOSE-WAIT state."),
			"tcpInErrs":                         newOtherFieldInfo(inputs.Float, inputs.Count, inputs.NCount, "(Shown as segment) The total number of segments received in error (e.g., bad TCP checksums)."),
			"tcpOutRsts":                        newOtherFieldInfo(inputs.Float, inputs.Count, inputs.NCount, "(Shown as segment) The number of TCP segments sent containing the RST flag."),
			"tcpPassiveOpens":                   newOtherFieldInfo(inputs.Float, inputs.Count, inputs.NCount, "(Shown as connection) The number of times TCP connections have made a direct transition to the SYN-RCVD state from the LISTEN state."),
			"tcpRetransSegs":                    newOtherFieldInfo(inputs.Float, inputs.Count, inputs.NCount, "(Shown as segment) The total number of segments retransmitted; that is, the number of TCP segments transmitted containing one or more previously transmitted octets."),
			"udpInErrors":                       newOtherFieldInfo(inputs.Float, inputs.Count, inputs.NCount, "(Shown as datagram) The number of received UDP datagram that could not be delivered for reasons other than the lack of an application at the destination port."),
			"udpNoPorts":                        newOtherFieldInfo(inputs.Float, inputs.Count, inputs.NCount, "(Shown as datagram) The total number of received UDP datagram for which there was no application at the destination port."),
			"ifAdminStatus":                     newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.NoUnit, "The desired state of the interface."),
			"ifHCInBroadcastPkts":               newOtherFieldInfo(inputs.Float, inputs.Count, inputs.NCount, "(Shown as packet) The number of packets delivered by this sub-layer to a higher (sub-)layer that were addressed to a broadcast address at this sub-layer."),
			"ifHCInMulticastPkts":               newOtherFieldInfo(inputs.Float, inputs.Count, inputs.NCount, "(Shown as packet) The number of packets delivered by this sub-layer to a higher (sub-)layer which were addressed to a multicast address at this sub-layer."),
			"ifHCInOctetsRate":                  newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.SizeByte, "(Shown as byte) The total number of octets received on the interface including framing characters."),
			"ifHCInUcastPkts":                   newOtherFieldInfo(inputs.Float, inputs.Count, inputs.NCount, "(Shown as packet) The number of packets delivered by this sub-layer to a higher (sub-)layer that were not addressed to a multicast or broadcast address at this sub-layer."),
			"ifHCOutBroadcastPkts":              newOtherFieldInfo(inputs.Float, inputs.Count, inputs.NCount, "(Shown as packet) The total number of packets that higher-level protocols requested be transmitted that were addressed to a broadcast address at this sub-layer, including those that were discarded or not sent."),
			"ifHCOutMulticastPkts":              newOtherFieldInfo(inputs.Float, inputs.Count, inputs.NCount, "(Shown as packet) The total number of packets that higher-level protocols requested be transmitted that were addressed to a multicast address at this sub-layer including those that were discarded or not sent."),
			"ifHCOutOctetsRate":                 newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.NCount, "(Shown as byte) The total number of octets transmitted out of the interface including framing characters."),
			"ifHCOutUcastPkts":                  newOtherFieldInfo(inputs.Float, inputs.Count, inputs.NCount, "(Shown as packet) The total number of packets higher-level protocols requested be transmitted that were not addressed to a multicast or broadcast address at this sub-layer including those that were discarded or not sent."),
			"ifInDiscardsRate":                  newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.NCount, "(Shown as packet) The number of inbound packets chosen to be discarded even though no errors had been detected to prevent them being deliverable to a higher-layer protocol."),
			"ifInErrorsRate":                    newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.NCount, "(Shown as packet) The number of inbound packets that contained errors preventing them from being deliverable to a higher-layer protocol."),
			"ifOutDiscardsRate":                 newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.NCount, "(Shown as packet) The number of outbound packets chosen to be discarded even though no errors had been detected to prevent them being transmitted."),
			"ifOutErrorsRate":                   newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.NCount, "(Shown as packet) The number of outbound packets that could not be transmitted because of errors."),
			"ifBandwidthInUsageRate":            newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "(Shown as percent) The percent rate of used received bandwidth."),
			"ifBandwidthOutUsageRate":           newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "(Shown as percent) The percent rate of used sent bandwidth."),
			"cieIfLastOutTime":                  newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.DurationMS, "[Cisco only] (Shown as millisecond) The elapsed time in milliseconds since the last protocol output packet was transmitted."),
			"cieIfOutputQueueDrops":             newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.NCount, "[Cisco only] (Shown as packet) The number of output packets dropped by the interface even though no error was detected to prevent them being transmitted."),
			"ciscoMemoryPoolUsed":               newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.NCount, "[Cisco only] Indicates the number of bytes from the memory pool that are currently in use by applications on the managed device."),
			"cpmCPUTotalMonIntervalValue":       newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "[Cisco only] (Shown as percent) The overall CPU busy percentage in the last cpmCPUMonInterval period."),
			"cieIfLastInTime":                   newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.DurationMS, "[Cisco only] (Shown as millisecond) The elapsed time in milliseconds since the last protocol input packet was received."),
			"cieIfResetCount":                   newOtherFieldInfo(inputs.Float, inputs.Count, inputs.NCount, "[Cisco only] The number of times the interface was internally reset and brought up."),
			"ciscoMemoryPoolLargestFree":        newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.NCount, "[Cisco only] Indicates the largest number of contiguous bytes from the memory pool that are currently unused on the managed device."),
			"ciscoEnvMonTemperatureStatusValue": newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.NCount, "[Cisco only] The current value of the test point being instrumented."),
			"ciscoEnvMonSupplyState":            newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.NCount, "[Cisco only] The current state of the power supply being instrumented."),
			"cswStackPortOperStatus":            newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.NCount, "[Cisco only] The state of the stack port."),
			"cpmCPUTotal1minRev":                newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "[Cisco only] [Shown as percent] The overall CPU busy percentage in the last 1 minute period."),
			"ciscoMemoryPoolFree":               newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.NCount, "[Cisco only] Indicates the number of bytes from the memory pool that are currently unused on the managed device."),
			"cieIfInputQueueDrops":              newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.NCount, "[Cisco only] (Shown as packet) The number of input packets dropped."),
			"ciscoEnvMonFanState":               newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.NCount, "[Cisco only] The current state of the fan being instrumented."),
			"cswSwitchState":                    newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.NCount, "[Cisco only] The current state of a switch."),
			"entSensorValue":                    newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.NCount, "[Cisco only] The most recent measurement seen by the sensor."),

			"uptime":          newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.DurationSecond, "(in second) uptime."),
			"netUptime":       newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.DurationSecond, "(in second) net uptime."),
			"uptimeTimestamp": newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.TimestampSec, "uptime timestamp."),
			"temperature":     newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Celsius, "The Temperature of item."),
			"voltage":         newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Volt, "The Volt of item."),
			"voltageStatus":   newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Bool, "The voltage status of item."),
			"current":         newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.UnknownType, "The current of item."),
			"power":           newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.UnknownType, "The power of item."),
			"powerStatus":     newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.UnknownType, "The power of item."),
			"fanSpeed":        newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.RotationRete, "The fan speed."),
			"fanStatus":       newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Bool, "The fan status."),

			"ifNetStatus":     newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Bool, "The net status."),
			"ifNetConnStatus": newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Bool, "The net connection status."),
			"ifOperStatus":    newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.NCount, "(Shown as packet) The current operational state of the interface."),
			"ifStatus":        newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Bool, "The interface status."),
			"ifSpeed":         newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.NCount, "An estimate of the interface's current bandwidth in bits per second, or the nominal bandwidth."),
			"ifHighSpeed":     newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.NCount, "An estimate of the interface's current bandwidth in units of 1,000,000 bits per second, or the nominal bandwidth."),
			"ifInDiscards":    newOtherFieldInfo(inputs.Float, inputs.Count, inputs.NCount, "(Shown as packet) The number of inbound packets chosen to be discarded even though no errors had been detected to prevent them being deliverable to a higher-layer protocol."),
			"ifOutDiscards":   newOtherFieldInfo(inputs.Float, inputs.Count, inputs.NCount, "(Shown as packet) The number of outbound packets chosen to be discarded even though no errors had been detected to prevent them being transmitted."),
			"ifInErrors":      newOtherFieldInfo(inputs.Float, inputs.Count, inputs.NCount, "(Shown as packet) The number of inbound packets that contained errors preventing them from being deliverable to a higher-layer protocol."),
			"ifOutErrors":     newOtherFieldInfo(inputs.Float, inputs.Count, inputs.NCount, "(Shown as packet) The number of outbound packets that could not be transmitted because of errors."),
			"ifHCInPkts":      newOtherFieldInfo(inputs.Float, inputs.Count, inputs.NCount, "(Shown as packet) The number of packets delivered by this sub-layer to a higher (sub-)layer that were not addressed to a multicast or broadcast address at this sub-layer."),
			"ifHCOutPkts":     newOtherFieldInfo(inputs.Float, inputs.Count, inputs.NCount, "(Shown as packet) The total number of packets higher-level protocols requested be transmitted that were not addressed to a multicast or broadcast address at this sub-layer including those that were discarded or not sent."),
			"ifHCInOctets":    newOtherFieldInfo(inputs.Float, inputs.Count, inputs.NCount, "(Shown as byte) The total number of octets received on the interface including framing characters."),
			"ifHCOutOctets":   newOtherFieldInfo(inputs.Float, inputs.Count, inputs.NCount, "(Shown as byte) The total number of octets transmitted out of the interface including framing characters."),

			"cpuUsage":       newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "(Shown as percent) Percentage of CPU currently being used."),
			"cpuTemperature": newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Celsius, "The Temperature of cpu."),
			"cpuStatus":      newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Bool, "CPU status."),

			"memoryTotal":     newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.SizeByte, "(Shown as byte) Number of bytes of memory."),
			"memoryUsage":     newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "(Shown as percent) The percentage of memory currently being used."),
			"memoryUsed":      newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.SizeByte, "(Shown as byte) Number of bytes of memory currently being used."),
			"memoryFree":      newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "(Shown as percent) The percentage of memory not being used."),
			"memoryAvailable": newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.SizeByte, "(Shown as byte) Number of memory available."),

			"diskTotal":     newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.SizeByte, "Total of disk size."),
			"diskUsage":     newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "(Shown as percent) The percentage of disk currently being used."),
			"diskUsed":      newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.SizeByte, "Number of disk currently being used."),
			"diskFree":      newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "(Shown as percent) The percentage of disk not being used."),
			"diskAvailable": newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.SizeByte, "Number of disk available."),

			"itemTotal":     newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.UnknownType, "Item total."),
			"itemUsage":     newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "(Shown as percent) Item being used."),
			"itemUsed":      newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.UnknownType, "Item being used."),
			"itemFree":      newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "(Shown as percent) Item not being used."),
			"itemAvailable": newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.UnknownType, "Item available."),
		},
		Tags: map[string]interface{}{
			"device_vendor":      inputs.NewTagInfo("Device vendor."),
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

func newOtherFieldInfo(datatype, ftype, unit, desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: datatype,
		Type:     ftype,
		Unit:     unit,
		Desc:     desc,
	}
}
