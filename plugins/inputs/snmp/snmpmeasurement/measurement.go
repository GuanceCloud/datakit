// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package snmpmeasurement constains snmp measurement definitions.
package snmpmeasurement

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

//------------------------------------------------------------------------------

const InputName = "snmp"

//------------------------------------------------------------------------------

// SNMPObject ...
type SNMPObject struct {
	Name     string
	Tags     map[string]string
	Fields   map[string]interface{}
	TS       time.Time
	Election bool
}

func (m *SNMPObject) LineProto() (*point.Point, error) {
	return point.NewPoint(m.Name, m.Tags, m.Fields, point.OOptElectionV2(m.Election))
}

//nolint:lll
func (m *SNMPObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: InputName,
		Desc: "采集 SNMP 设备对象的数据",
		Type: "object",
		Fields: map[string]interface{}{
			"device_meta":                       newOtherFieldInfo(inputs.String, inputs.String, inputs.UnknownUnit, "Device meta data(JSON format)."),
			"ifNumber":                          newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.UnknownUnit, "Number of interface."),
			"sysUpTimeInstance":                 newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.NCount, "The time (in hundredths of a second) since the network management portion of the system was last re-initialized."),
			"tcpActiveOpens":                    newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "The number of times that TCP connections have made a direct transition to the SYN-SENT state from the CLOSED state."),
			"tcpAttemptFails":                   newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "The number of times that TCP connections have made a direct transition to the CLOSED state from either the SYN-SENT state or the SYN-RCVD state, or to the LISTEN state from the SYN-RCVD state."),
			"tcpCurrEstab":                      newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.UnknownUnit, "The number of TCP connections for which the current state is either ESTABLISHED or CLOSE-WAIT."),
			"tcpEstabResets":                    newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "The number of times that TCP connections have made a direct transition to the CLOSED state from either the ESTABLISHED state or the CLOSE-WAIT state."),
			"tcpInErrs":                         newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "(Shown as segment) The total number of segments received in error (e.g., bad TCP checksums)."),
			"tcpOutRsts":                        newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "(Shown as segment) The number of TCP segments sent containing the RST flag."),
			"tcpPassiveOpens":                   newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "(Shown as connection) The number of times TCP connections have made a direct transition to the SYN-RCVD state from the LISTEN state."),
			"tcpRetransSegs":                    newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "(Shown as segment) The total number of segments retransmitted; that is, the number of TCP segments transmitted containing one or more previously transmitted octets."),
			"udpInErrors":                       newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "(Shown as datagram) The number of received UDP datagrams that could not be delivered for reasons other than the lack of an application at the destination port."),
			"udpNoPorts":                        newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "(Shown as datagram) The total number of received UDP datagrams for which there was no application at the destination port."),
			"ifAdminStatus":                     newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.UnknownUnit, "The desired state of the interface."),
			"ifHCInBroadcastPkts":               newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "(Shown as packet) The number of packets delivered by this sub-layer to a higher (sub-)layer that were addressed to a broadcast address at this sub-layer."),
			"ifHCInMulticastPkts":               newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "(Shown as packet) The number of packets delivered by this sub-layer to a higher (sub-)layer which were addressed to a multicast address at this sub-layer."),
			"ifHCInOctets":                      newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "(Shown as byte) The total number of octets received on the interface including framing characters."),
			"ifHCInOctetsRate":                  newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.UnknownUnit, "(Shown as byte) The total number of octets received on the interface including framing characters."),
			"ifHCInUcastPkts":                   newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "(Shown as packet) The number of packets delivered by this sub-layer to a higher (sub-)layer that were not addressed to a multicast or broadcast address at this sub-layer."),
			"ifHCOutBroadcastPkts":              newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "(Shown as packet) The total number of packets that higher-level protocols requested be transmitted that were addressed to a broadcast address at this sub-layer, including those that were discarded or not sent."),
			"ifHCOutMulticastPkts":              newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "(Shown as packet) The total number of packets that higher-level protocols requested be transmitted that were addressed to a multicast address at this sub-layer including those that were discarded or not sent."),
			"ifHCOutOctets":                     newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "(Shown as byte) The total number of octets transmitted out of the interface including framing characters."),
			"ifHCOutOctetsRate":                 newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.NCount, "(Shown as byte) The total number of octets transmitted out of the interface including framing characters."),
			"ifHCOutUcastPkts":                  newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "(Shown as packet) The total number of packets higher-level protocols requested be transmitted that were not addressed to a multicast or broadcast address at this sub-layer including those that were discarded or not sent."),
			"ifHighSpeed":                       newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.NCount, "An estimate of the interface's current bandwidth in units of 1,000,000 bits per second, or the nominal bandwidth."),
			"ifInDiscards":                      newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "(Shown as packet) The number of inbound packets chosen to be discarded even though no errors had been detected to prevent them being deliverable to a higher-layer protocol."),
			"ifInDiscardsRate":                  newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.NCount, "(Shown as packet) The number of inbound packets chosen to be discarded even though no errors had been detected to prevent them being deliverable to a higher-layer protocol."),
			"ifInErrors":                        newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "(Shown as packet) The number of inbound packets that contained errors preventing them from being deliverable to a higher-layer protocol."),
			"ifInErrorsRate":                    newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.NCount, "(Shown as packet) The number of inbound packets that contained errors preventing them from being deliverable to a higher-layer protocol."),
			"ifOperStatus":                      newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.NCount, "(Shown as packet) The current operational state of the interface."),
			"ifOutDiscards":                     newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "(Shown as packet) The number of outbound packets chosen to be discarded even though no errors had been detected to prevent them being transmitted."),
			"ifOutDiscardsRate":                 newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.NCount, "(Shown as packet) The number of outbound packets chosen to be discarded even though no errors had been detected to prevent them being transmitted."),
			"ifOutErrors":                       newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "(Shown as packet) The number of outbound packets that could not be transmitted because of errors."),
			"ifOutErrorsRate":                   newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.NCount, "(Shown as packet) The number of outbound packets that could not be transmitted because of errors."),
			"ifSpeed":                           newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.NCount, "An estimate of the interface's current bandwidth in bits per second, or the nominal bandwidth."),
			"ifBandwidthInUsageRate":            newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "(Shown as percent) The percent rate of used received bandwidth."),
			"ifBandwidthOutUsageRate":           newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "(Shown as percent) The percent rate of used sent bandwidth."),
			"cpuUsage":                          newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "(Shown as percent) Percentage of CPU currently being used."),
			"memoryUsed":                        newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.NCount, "(Shown as byte) Number of bytes of memory currently being used."),
			"memoryUsage":                       newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "(Shown as percent) The percentage of memory currently being used."),
			"memoryFree":                        newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "(Shown as percent) The percentage of memory not being used."),
			"cieIfLastOutTime":                  newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.DurationMS, "[Cisco only] (Shown as millisecond) The elapsed time in milliseconds since the last protocol output packet was transmitted."),
			"cieIfOutputQueueDrops":             newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.NCount, "[Cisco only] (Shown as packet) The number of output packets dropped by the interface even though no error was detected to prevent them being transmitted."),
			"ciscoMemoryPoolUsed":               newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.NCount, "[Cisco only] Indicates the number of bytes from the memory pool that are currently in use by applications on the managed device."),
			"cpmCPUTotalMonIntervalValue":       newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "[Cisco only] (Shown as percent) The overall CPU busy percentage in the last cpmCPUMonInterval period."),
			"cieIfLastInTime":                   newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.DurationMS, "[Cisco only] (Shown as millisecond) The elapsed time in milliseconds since the last protocol input packet was received."),
			"cieIfResetCount":                   newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "[Cisco only] The number of times the interface was internally reset and brought up."),
			"ciscoMemoryPoolLargestFree":        newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.NCount, "[Cisco only] Indicates the largest number of contiguous bytes from the memory pool that are currently unused on the managed device."),
			"ciscoEnvMonTemperatureStatusValue": newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.NCount, "[Cisco only] The current value of the testpoint being instrumented."),
			"ciscoEnvMonSupplyState":            newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.NCount, "[Cisco only] The current state of the power supply being instrumented."),
			"cswStackPortOperStatus":            newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.NCount, "[Cisco only] The state of the stackport."),
			"cpmCPUTotal1minRev":                newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "[Cisco only] [Shown as percent] The overall CPU busy percentage in the last 1 minute period."),
			"ciscoMemoryPoolFree":               newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.NCount, "[Cisco only] Indicates the number of bytes from the memory pool that are currently unused on the managed device."),
			"cieIfInputQueueDrops":              newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.NCount, "[Cisco only] (Shown as packet) The number of input packets dropped."),
			"ciscoEnvMonFanState":               newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.NCount, "[Cisco only] The current state of the fan being instrumented."),
			"cswSwitchState":                    newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.NCount, "[Cisco only] The current state of a switch."),
		},
		Tags: map[string]interface{}{
			"device_vendor":      inputs.NewTagInfo("Device vendor."),
			"host":               inputs.NewTagInfo("Device name."),
			"interface":          inputs.NewTagInfo("Device interface."),
			"interface_alias":    inputs.NewTagInfo("Device interface alias."),
			"snmp_profile":       inputs.NewTagInfo("Device SNMP profile file."),
			"mac_addr":           inputs.NewTagInfo("Device MAC address"),
			"entity_name":        inputs.NewTagInfo("Device entity name."),
			"power_source":       inputs.NewTagInfo("Power source."),
			"power_status_descr": inputs.NewTagInfo("Power status description."),
			"temp_index":         inputs.NewTagInfo("Temperature index."),
			"temp_state":         inputs.NewTagInfo("Temperature state."),
			"cpu":                inputs.NewTagInfo("CPU index."),
			"mem":                inputs.NewTagInfo("Memory index."),
			"mem_pool_name":      inputs.NewTagInfo("Memory pool name."),
		},
	}
}

//------------------------------------------------------------------------------

// SNMPMetric ...
type SNMPMetric struct {
	Name     string
	Tags     map[string]string
	Fields   map[string]interface{}
	TS       time.Time
	Election bool
}

func (m *SNMPMetric) LineProto() (*point.Point, error) {
	return point.NewPoint(m.Name, m.Tags, m.Fields, point.MOptElectionV2(m.Election))
}

//nolint:lll
func (m *SNMPMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: InputName,
		Desc: "采集 SNMP 设备指标的数据",
		Type: "metric",
		Fields: map[string]interface{}{
			"ifNumber":                          newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.UnknownUnit, "Number of interface."),
			"sysUpTimeInstance":                 newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.NCount, "The time (in hundredths of a second) since the network management portion of the system was last re-initialized."),
			"tcpActiveOpens":                    newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "The number of times that TCP connections have made a direct transition to the SYN-SENT state from the CLOSED state."),
			"tcpAttemptFails":                   newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "The number of times that TCP connections have made a direct transition to the CLOSED state from either the SYN-SENT state or the SYN-RCVD state, or to the LISTEN state from the SYN-RCVD state."),
			"tcpCurrEstab":                      newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.UnknownUnit, "The number of TCP connections for which the current state is either ESTABLISHED or CLOSE-WAIT."),
			"tcpEstabResets":                    newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "The number of times that TCP connections have made a direct transition to the CLOSED state from either the ESTABLISHED state or the CLOSE-WAIT state."),
			"tcpInErrs":                         newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "(Shown as segment) The total number of segments received in error (e.g., bad TCP checksums)."),
			"tcpOutRsts":                        newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "(Shown as segment) The number of TCP segments sent containing the RST flag."),
			"tcpPassiveOpens":                   newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "(Shown as connection) The number of times TCP connections have made a direct transition to the SYN-RCVD state from the LISTEN state."),
			"tcpRetransSegs":                    newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "(Shown as segment) The total number of segments retransmitted; that is, the number of TCP segments transmitted containing one or more previously transmitted octets."),
			"udpInErrors":                       newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "(Shown as datagram) The number of received UDP datagrams that could not be delivered for reasons other than the lack of an application at the destination port."),
			"udpNoPorts":                        newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "(Shown as datagram) The total number of received UDP datagrams for which there was no application at the destination port."),
			"ifAdminStatus":                     newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.UnknownUnit, "The desired state of the interface."),
			"ifHCInBroadcastPkts":               newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "(Shown as packet) The number of packets delivered by this sub-layer to a higher (sub-)layer that were addressed to a broadcast address at this sub-layer."),
			"ifHCInMulticastPkts":               newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "(Shown as packet) The number of packets delivered by this sub-layer to a higher (sub-)layer which were addressed to a multicast address at this sub-layer."),
			"ifHCInOctets":                      newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "(Shown as byte) The total number of octets received on the interface including framing characters."),
			"ifHCInOctetsRate":                  newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.UnknownUnit, "(Shown as byte) The total number of octets received on the interface including framing characters."),
			"ifHCInUcastPkts":                   newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "(Shown as packet) The number of packets delivered by this sub-layer to a higher (sub-)layer that were not addressed to a multicast or broadcast address at this sub-layer."),
			"ifHCOutBroadcastPkts":              newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "(Shown as packet) The total number of packets that higher-level protocols requested be transmitted that were addressed to a broadcast address at this sub-layer, including those that were discarded or not sent."),
			"ifHCOutMulticastPkts":              newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "(Shown as packet) The total number of packets that higher-level protocols requested be transmitted that were addressed to a multicast address at this sub-layer including those that were discarded or not sent."),
			"ifHCOutOctets":                     newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "(Shown as byte) The total number of octets transmitted out of the interface including framing characters."),
			"ifHCOutOctetsRate":                 newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.NCount, "(Shown as byte) The total number of octets transmitted out of the interface including framing characters."),
			"ifHCOutUcastPkts":                  newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "(Shown as packet) The total number of packets higher-level protocols requested be transmitted that were not addressed to a multicast or broadcast address at this sub-layer including those that were discarded or not sent."),
			"ifHighSpeed":                       newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.NCount, "An estimate of the interface's current bandwidth in units of 1,000,000 bits per second, or the nominal bandwidth."),
			"ifInDiscards":                      newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "(Shown as packet) The number of inbound packets chosen to be discarded even though no errors had been detected to prevent them being deliverable to a higher-layer protocol."),
			"ifInDiscardsRate":                  newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.NCount, "(Shown as packet) The number of inbound packets chosen to be discarded even though no errors had been detected to prevent them being deliverable to a higher-layer protocol."),
			"ifInErrors":                        newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "(Shown as packet) The number of inbound packets that contained errors preventing them from being deliverable to a higher-layer protocol."),
			"ifInErrorsRate":                    newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.NCount, "(Shown as packet) The number of inbound packets that contained errors preventing them from being deliverable to a higher-layer protocol."),
			"ifOperStatus":                      newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.NCount, "(Shown as packet) The current operational state of the interface."),
			"ifOutDiscards":                     newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "(Shown as packet) The number of outbound packets chosen to be discarded even though no errors had been detected to prevent them being transmitted."),
			"ifOutDiscardsRate":                 newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.NCount, "(Shown as packet) The number of outbound packets chosen to be discarded even though no errors had been detected to prevent them being transmitted."),
			"ifOutErrors":                       newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "(Shown as packet) The number of outbound packets that could not be transmitted because of errors."),
			"ifOutErrorsRate":                   newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.NCount, "(Shown as packet) The number of outbound packets that could not be transmitted because of errors."),
			"ifSpeed":                           newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.NCount, "An estimate of the interface's current bandwidth in bits per second, or the nominal bandwidth."),
			"ifBandwidthInUsageRate":            newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "(Shown as percent) The percent rate of used received bandwidth."),
			"ifBandwidthOutUsageRate":           newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "(Shown as percent) The percent rate of used sent bandwidth."),
			"cpuUsage":                          newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "(Shown as percent) Percentage of CPU currently being used."),
			"memoryUsed":                        newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.NCount, "(Shown as byte) Number of bytes of memory currently being used."),
			"memoryUsage":                       newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "(Shown as percent) The percentage of memory currently being used."),
			"memoryFree":                        newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "(Shown as percent) The percentage of memory not being used."),
			"cieIfLastOutTime":                  newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.DurationMS, "[Cisco only] (Shown as millisecond) The elapsed time in milliseconds since the last protocol output packet was transmitted."),
			"cieIfOutputQueueDrops":             newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.NCount, "[Cisco only] (Shown as packet) The number of output packets dropped by the interface even though no error was detected to prevent them being transmitted."),
			"ciscoMemoryPoolUsed":               newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.NCount, "[Cisco only] Indicates the number of bytes from the memory pool that are currently in use by applications on the managed device."),
			"cpmCPUTotalMonIntervalValue":       newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "[Cisco only] (Shown as percent) The overall CPU busy percentage in the last cpmCPUMonInterval period."),
			"cieIfLastInTime":                   newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.DurationMS, "[Cisco only] (Shown as millisecond) The elapsed time in milliseconds since the last protocol input packet was received."),
			"cieIfResetCount":                   newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "[Cisco only] The number of times the interface was internally reset and brought up."),
			"ciscoMemoryPoolLargestFree":        newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.NCount, "[Cisco only] Indicates the largest number of contiguous bytes from the memory pool that are currently unused on the managed device."),
			"ciscoEnvMonTemperatureStatusValue": newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.NCount, "[Cisco only] The current value of the testpoint being instrumented."),
			"ciscoEnvMonSupplyState":            newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.NCount, "[Cisco only] The current state of the power supply being instrumented."),
			"cswStackPortOperStatus":            newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.NCount, "[Cisco only] The state of the stackport."),
			"cpmCPUTotal1minRev":                newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "[Cisco only] [Shown as percent] The overall CPU busy percentage in the last 1 minute period."),
			"ciscoMemoryPoolFree":               newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.NCount, "[Cisco only] Indicates the number of bytes from the memory pool that are currently unused on the managed device."),
			"cieIfInputQueueDrops":              newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.NCount, "[Cisco only] (Shown as packet) The number of input packets dropped."),
			"ciscoEnvMonFanState":               newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.NCount, "[Cisco only] The current state of the fan being instrumented."),
			"cswSwitchState":                    newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.NCount, "[Cisco only] The current state of a switch."),
		},
		Tags: map[string]interface{}{
			"device_vendor":      inputs.NewTagInfo("Device vendor."),
			"host":               inputs.NewTagInfo("Device name."),
			"interface":          inputs.NewTagInfo("Device interface."),
			"interface_alias":    inputs.NewTagInfo("Device interface alias."),
			"snmp_profile":       inputs.NewTagInfo("Device SNMP profile file."),
			"mac_addr":           inputs.NewTagInfo("Device MAC address"),
			"entity_name":        inputs.NewTagInfo("Device entity name."),
			"power_source":       inputs.NewTagInfo("Power source."),
			"power_status_descr": inputs.NewTagInfo("Power status description."),
			"temp_index":         inputs.NewTagInfo("Temperature index."),
			"temp_state":         inputs.NewTagInfo("Temperature state."),
			"cpu":                inputs.NewTagInfo("CPU index."),
			"mem":                inputs.NewTagInfo("Memory index."),
			"mem_pool_name":      inputs.NewTagInfo("Memory pool name."),
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
