// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package vsphere

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type Measurement struct {
	name     string
	tags     map[string]string
	fields   map[string]interface{}
	election bool
}

// Point implement MeasurementV2.
func (m *Measurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()

	if m.election {
		opts = append(opts, point.WithExtraTags(datakit.GlobalElectionTags()))
	}

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

func (m *Measurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "vsphere_cpu",
		Cat:  point.Metric,
		Tags: map[string]interface{}{},
	}
}

type clusterMeasurement struct {
	Measurement
}

func (m *clusterMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "vsphere_cluster",
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"cpu_usage_average": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.Percent,
				Desc:     "Percentage of CPU capacity being used.",
			},
			"cpu_usagemhz_average": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.Percent,
				Desc:     "CPU usage, as measured in megahertz.",
			},
			"mem_consumed_average": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeKB,
				Desc:     "Amount of host physical memory consumed by a virtual machine, host, or cluster.",
			},
			"mem_overhead_average": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeKB,
				Desc:     "Host physical memory consumed by the virtualization infrastructure for running the virtual machine.",
			},
			"mem_usage_average": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.Percent,
				Desc:     "Memory usage as percent of total configured or available memory.",
			},
			"mem_vmmemctl_average": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.Percent,
				Desc:     "Amount of memory allocated by the virtual machine memory control driver (`vmmemctl`).",
			},
			"vmop_numChangeDS_latest": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of datastore change operations for powered-off and suspended virtual machines.",
			},
			"vmop_numChangeHostDS_latest": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of host and datastore change operations for powered-off and suspended virtual machines.",
			},
			"vmop_numChangeHost_latest": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of host change operations for powered-off and suspended virtual machines.",
			},
			"vmop_numClone_latest": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of virtual machine clone operations.",
			},
			"vmop_numCreate_latest": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of virtual machine create operations.",
			},
			"vmop_numDeploy_latest": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of virtual machine template deploy operations.",
			},
			"vmop_numDestroy_latest": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of virtual machine delete operations.",
			},
			"vmop_numPoweroff_latest": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of virtual machine power off operations.",
			},
			"vmop_numPoweron_latest": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of virtual machine power on operations.",
			},
			"vmop_numRebootGuest_latest": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of virtual machine guest reboot operations.",
			},
			"vmop_numReconfigure_latest": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of virtual machine reconfigure operations.",
			},
			"vmop_numRegister_latest": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of virtual machine register operations.",
			},
			"vmop_numReset_latest": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of virtual machine reset operations.",
			},
			"vmop_numSVMotion_latest": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of migrations with Storage vMotion (datastore change operations for powered-on VMs).",
			},
			"vmop_numShutdownGuest_latest": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of virtual machine guest shutdown operations.",
			},
			"vmop_numStandbyGuest_latest": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of virtual machine standby guest operations.",
			},
			"vmop_numSuspend_latest": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of virtual machine suspend operations.",
			},
			"vmop_numUnregister_latest": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of virtual machine unregister operations.",
			},
			"vmop_numVMotion_latest": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of migrations with vMotion (host change operations for powered-on VMs).",
			},
			"vmop_numXVMotion_latest": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of host and datastore change operations for powered-on and suspended virtual machines.",
			},
		},
		Tags: map[string]interface{}{
			"host":         &inputs.TagInfo{Desc: "The host of the vCenter"},
			"dcname":       &inputs.TagInfo{Desc: "`Datacenter` name"},
			"cluster_name": &inputs.TagInfo{Desc: "Cluster name"},
			"moid":         &inputs.TagInfo{Desc: "The managed object id"},
		},
	}
}

type datastoreMeasurement struct {
	Measurement
}

//nolint:lll
func (m *datastoreMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "vsphere_datastore",
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"datastore_busResets_sum":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of SCSI-bus reset commands issued."},
			"datastore_commandsAborted_sum":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of SCSI commands aborted."},
			"datastore_numberReadAveraged_average":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Average number of read commands issued per second to the datastore."},
			"datastore_numberWriteAveraged_average": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Average number of write commands issued per second to the datastore during the collection interval."},
			"datastore_throughput_contention.avg":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Average amount of time for an I/O operation to the datastore or LUN across all ESX hosts accessing it."},
			"datastore_throughput_usage_average":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "The current bandwidth usage for the datastore or LUN."},
			"disk_busResets_sum":                    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of SCSI-bus reset commands issued."},
			"disk_capacity_contention_average":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "The amount of storage capacity overcommitment for the entity, measured in percent."},
			"disk_capacity_latest":                  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Configured size of the datastore."},
			"disk_capacity_provisioned_average":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Provisioned size of the entity."},
			"disk_capacity_usage.avg":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "The amount of storage capacity currently being consumed by or on the entity."},
			"disk_numberReadAveraged_average":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Average number of read commands issued per second to the datastore."},
			"disk_numberWriteAveraged_average":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Average number of write commands issued per second to the datastore."},
			"disk_provisioned_latest":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Amount of storage set aside for use by a datastore or a virtual machine. Files on the datastore and the virtual machine can expand to this size but not beyond it."},
			"disk_unshared_latest":                  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Amount of space associated exclusively with a virtual machine."},
			"disk_used_latest":                      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Amount of space actually used by the virtual machine or the datastore. May be less than the amount provisioned at any given time, depending on whether the virtual machine is powered-off, whether snapshots have been created or not, and other such factors."},
		},
		Tags: map[string]interface{}{
			"host":   &inputs.TagInfo{Desc: "The host of the vCenter"},
			"dcname": &inputs.TagInfo{Desc: "`Datacenter` name"},
			"moid":   &inputs.TagInfo{Desc: "The managed object id"},
			"dsname": &inputs.TagInfo{Desc: "The name of the datastore"},
		},
	}
}

type vmMeasurement struct {
	Measurement
}

//nolint:lll
func (m *vmMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "vsphere_vm",
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"cpu_costop_sum":                          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Time the virtual machine is ready to run, but is unable to run due to co-scheduling constraints."},
			"cpu_demand_average":                      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "The amount of CPU resources a virtual machine would use if there were no CPU contention or CPU limit."},
			"cpu_demandEntitlementRatio_latest":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "CPU resource entitlement to CPU demand ratio (in percents)."},
			"cpu_entitlement_latest":                  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "CPU resources devoted by the ESXi scheduler."},
			"cpu_idle_sum":                            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Total time that the CPU spent in an idle state."},
			"cpu_latency_average":                     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Percent of time the virtual machine is unable to run because it is contending for access to the physical CPU(s)."},
			"cpu_maxlimited_sum":                      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Time the virtual machine is ready to run, but is not running because it has reached its maximum CPU limit setting."},
			"cpu_overlap_sum":                         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Time the virtual machine was interrupted to perform system services on behalf of itself or other virtual machines."},
			"cpu_readiness_average":                   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Percentage of time that the virtual machine was ready, but could not get scheduled to run on the physical CPU."},
			"cpu_ready_sum":                           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Milliseconds of CPU time spent in ready state."},
			"cpu_run_sum":                             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Time the virtual machine is scheduled to run."},
			"cpu_swapwait_sum":                        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "CPU time spent waiting for swap-in."},
			"cpu_system_sum":                          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Amount of time spent on system processes on each virtual CPU in the virtual machine. This is the host view of the CPU usage, not the guest operating system view."},
			"cpu_usage_average":                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Percentage of CPU capacity being used."},
			"cpu_usagemhz_average":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "CPU usage, as measured in megahertz."},
			"cpu_used_sum":                            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Time accounted to the virtual machine. If a system service runs on behalf of this virtual machine, the time spent by that service (represented by cpu.system) should be charged to this virtual machine. If not, the time spent (represented by cpu.overlap) should not be charged against this virtual machine."},
			"cpu_wait_sum":                            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Total CPU time spent in wait state.The wait total includes time spent the CPU Idle, CPU Swap Wait, and CPU I/O Wait states."},
			"cpu_capacity_demand_average":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "The amount of CPU resources a virtual machine would use if there were no CPU contention or CPU limit."},
			"cpu_capacity_usage_average":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "CPU usage as a percent during the interval."},
			"cpu_capacity_contention_average":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Percent of time the virtual machine is unable to run because it is contending for access to the physical CPU(s)."},
			"datastore_maxTotalLatency_latest":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Highest latency value across all datastores used by the host."},
			"datastore_numberReadAveraged_average":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Average number of read commands issued per second to the datastore."},
			"datastore_numberWriteAveraged_average":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Average number of write commands issued per second to the datastore during the collection interval."},
			"datastore_read_average":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Rate of reading data from the datastore."},
			"datastore_totalReadLatency_average":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Average amount of time for a read operation from the datastore."},
			"datastore_totalWriteLatency_average":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Average amount of time for a write operation from the datastore."},
			"datastore_write_average":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Rate of writing data to the datastore."},
			"disk_busResets_sum":                      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of SCSI-bus reset commands issued."},
			"disk_commands_sum":                       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of SCSI commands issued"},
			"disk_commandsAborted_sum":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of SCSI commands aborted."},
			"disk_commandsAveraged_average":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Average number of SCSI commands issued per second."},
			"disk_maxTotalLatency_latest":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Highest latency value across all disks used by the host."},
			"disk_numberRead_sum":                     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of disk reads during the collection interval."},
			"disk_numberReadAveraged_average":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Average number of read commands issued per second to the datastore."},
			"disk_numberWrite_sum":                    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of disk writes during the collection interval."},
			"disk_numberWriteAveraged_average":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Average number of write commands issued per second to the datastore."},
			"disk_read_average":                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Average number of kilobytes read from the disk each second."},
			"disk_usage_average":                      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Aggregated disk I/O rate."},
			"disk_write_average":                      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Average number of kilobytes written to the disk each second."},
			"hbr_hbrNetRx_average":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Kilobytes per second of outgoing host-based replication network traffic (for this virtual machine or host)."},
			"hbr_hbrNetTx_average":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Average amount of data transmitted per second."},
			"mem_active_average":                      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Amount of memory that is actively used, as estimated by VMkernel based on recently touched memory pages."},
			"mem_activewrite_average":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Estimate for the amount of memory actively being written to by the virtual machine."},
			"mem_compressed_average":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Amount of memory reserved by `userworlds`."},
			"mem_compressionRate_average":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Rate of memory compression for the virtual machine."},
			"mem_consumed_average":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Amount of host physical memory consumed by a virtual machine, host, or cluster."},
			"mem_decompressionRate_average":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Rate of memory decompression for the virtual machine."},
			"mem_entitlement_average":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Amount of host physical memory the virtual machine is entitled to, as determined by the ESX scheduler."},
			"mem_granted_average":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Amount of host physical memory or physical memory that is mapped for a virtual machine or a host."},
			"mem_latency_average":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Percentage of time the virtual machine is waiting to access swapped or compressed memory."},
			"mem_llSwapInRate_average":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Rate at which memory is being swapped from host cache into active memory."},
			"mem_llSwapOutRate_average":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Rate at which memory is being swapped from active memory to host cache."},
			"mem_llSwapUsed_average":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Space used for caching swapped pages in the host cache."},
			"mem_overhead_average":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Host physical memory consumed by the virtualization infrastructure for running the virtual machine."},
			"mem_overheadMax_average":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Host physical memory reserved for use as the virtualization overhead for the virtual machine."},
			"mem_overheadTouched_average":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Actively touched overhead host physical memory (KB) reserved for use as the virtualization overhead for the virtual machine."},
			"mem_shared_average":                      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Amount of guest physical memory that is shared with other virtual machines, relative to a single virtual machine or to all powered-on virtual machines on a host."},
			"mem_swapin_average":                      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Amount of memory swapped-in from disk."},
			"mem_swapinRate_average":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Rate at which memory is swapped from disk into active memory."},
			"mem_swapout_average":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Amount of memory swapped-out to disk."},
			"mem_swapoutRate_average":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Rate at which memory is being swapped from active memory to disk."},
			"mem_swapped_average":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Current amount of guest physical memory swapped out to the virtual machine swap file by the VMkernel. Swapped memory stays on disk until the virtual machine needs it. This statistic refers to VMkernel swapping and not to guest OS swapping."},
			"mem_swaptarget_average":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Target size for the virtual machine swap file. The VMkernel manages swapping by comparing `swaptarget` against swapped."},
			"mem_usage_average":                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Memory usage as percent of total configured or available memory"},
			"mem_vmmemctl_average":                    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Amount of memory allocated by the virtual machine memory control driver (`vmmemctl`)."},
			"mem_vmmemctltarget_average":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Target value set by `VMkernal` for the virtual machine's memory balloon size. In conjunction with `vmmemctl` metric, this metric is used by VMkernel to inflate and deflate the balloon for a virtual machine."},
			"mem_zero_average":                        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Memory that contains 0s only. Included in shared amount. Through transparent page sharing, zero memory pages can be shared among virtual machines that run the same operating system."},
			"mem_zipSaved_latest":                     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Memory saved due to memory zipping."},
			"mem_zipped_latest":                       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Memory zipped"},
			"mem_capacity_usage_average":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Amount of physical memory actively used."},
			"mem_capacity_contention_average":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Percentage of time VMs are waiting to access swapped, compressed or ballooned memory."},
			"net_broadcastRx_sum":                     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of broadcast packets received."},
			"net_broadcastTx_sum":                     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of broadcast packets transmitted."},
			"net_bytesRx_average":                     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Average amount of data received per second."},
			"net_bytesTx_average":                     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Average amount of data transmitted per second."},
			"net_droppedRx_sum":                       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of received packets dropped."},
			"net_droppedTx_sum":                       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of transmitted packets dropped."},
			"net_multicastRx_sum":                     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of multicast packets received."},
			"net_multicastTx_sum":                     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of multicast packets transmitted."},
			"net_packetsRx_sum":                       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of packets received."},
			"net_packetsTx_sum":                       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of packets transmitted."},
			"net_pnicBytesRx_average":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Average number of bytes received per second by a physical network interface card (`PNIC`) on an ESXi host."},
			"net_pnicBytesTx_average":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Average number of bytes transmitted per second by a physical network interface card (`PNIC`) on an ESXi host."},
			"net_received_average":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Average rate at which data was received during the interval. This represents the bandwidth of the network."},
			"net_throughput_usage_average":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "The current network bandwidth usage for the host."},
			"net_transmitted_average":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Average rate at which data was transmitted during the interval. This represents the bandwidth of the network."},
			"net_usage_average":                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Network utilization (combined transmit- and receive-rates)."},
			"power_energy_sum":                        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Total energy (in joule) used since last stats reset."},
			"power_power_average":                     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Current power usage."},
			"rescpu_actav1_latest":                    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "CPU active average over 1 minute."},
			"rescpu_actav15_latest":                   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "CPU active average over 15 minutes."},
			"rescpu_actav5_latest":                    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "CPU active average over 5 minutes."},
			"rescpu_actpk1_latest":                    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "CPU active peak over 1 minute."},
			"rescpu_actpk15_latest":                   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "CPU active peak over 15 minutes."},
			"rescpu_actpk5_latest":                    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "CPU active peak over 5 minutes."},
			"rescpu_maxLimited1_latest":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Amount of CPU resources over the limit that were refused, average over 1 minute."},
			"rescpu_maxLimited15_latest":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Amount of CPU resources over the limit that were refused, average over 15 minutes."},
			"rescpu_maxLimited5_latest":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Amount of CPU resources over the limit that were refused, average over 5 minutes."},
			"rescpu_runav1_latest":                    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "CPU running average over 1 minute."},
			"rescpu_runav15_latest":                   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "CPU running average over 15 minutes."},
			"rescpu_runav5_latest":                    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "CPU running average over 5 minutes."},
			"rescpu_runpk1_latest":                    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "CPU running peak over 1 minute."},
			"rescpu_runpk15_latest":                   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "CPU running peak over 15 minutes."},
			"rescpu_runpk5_latest":                    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "CPU running peak over 5 minutes."},
			"rescpu_sampleCount_latest":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Group CPU sample count."},
			"rescpu_samplePeriod_latest":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Group CPU sample period."},
			"sys_heartbeat_latest":                    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of heartbeats issued per virtual machine."},
			"sys_heartbeat_sum":                       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of heartbeats issued per virtual machine."},
			"sys_osUptime_latest":                     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationSecond, Desc: "Total time elapsed, in seconds, since last operating system boot-up."},
			"sys_uptime_latest":                       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationSecond, Desc: "Total time elapsed since last system startup"},
			"virtualDisk_busResets_sum":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of SCSI-bus reset commands issued."},
			"virtualDisk_commandsAborted_sum":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of SCSI commands aborted"},
			"virtualDisk_largeSeeks_latest":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of seeks during the interval that were greater than 8192 LBNs apart."},
			"virtualDisk_mediumSeeks_latest":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of seeks during the interval that were between 64 and 8192 LBNs apart."},
			"virtualDisk_numberReadAveraged_average":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Average number of read commands issued per second to the virtual disk."},
			"virtualDisk_numberWriteAveraged_average": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Average number of write commands issued per second to the virtual disk."},
			"virtualDisk_read_average":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Average number of kilobytes read from the virtual disk each second."},
			"virtualDisk_readIOSize_latest":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Average read request size in bytes."},
			"virtualDisk_readLatencyUS_latest":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationUS, Desc: "Read latency in microseconds."},
			"virtualDisk_readLoadMetric_latest":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Storage DRS virtual disk metric for the read workload model."},
			"virtualDisk_readOIO_latest":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Average number of outstanding read requests to the virtual disk."},
			"virtualDisk_smallSeeks_latest":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of seeks during the interval that were less than 64 LBNs apart."},
			"virtualDisk_totalReadLatency_average":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Average amount of time for a read operation from the virtual disk."},
			"virtualDisk_totalWriteLatency_average":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Average amount of time for a write operation from the virtual disk."},
			"virtualDisk_write_average":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Average number of kilobytes written to the virtual disk each second."},
			"virtualDisk_writeIOSize_latest":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Average write request size in bytes."},
			"virtualDisk_writeLatencyUS_latest":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationUS, Desc: "Write latency in microseconds."},
			"virtualDisk_writeLoadMetric_latest":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Storage DRS virtual disk metric for the write workload model."},
			"virtualDisk_writeOIO_latest":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Average number of outstanding write requests to the virtual disk."},
		},
		Tags: map[string]interface{}{
			"dcname":       &inputs.TagInfo{Desc: "`Datacenter` name"},
			"cluster_name": &inputs.TagInfo{Desc: "Cluster name"},
			"esx_hostname": &inputs.TagInfo{Desc: "The name of the ESXi host"},
			"instance":     &inputs.TagInfo{Desc: "The name of the instance"},
			"host":         &inputs.TagInfo{Desc: "The host of the vCenter"},
			"moid":         &inputs.TagInfo{Desc: "The managed object id"},
			"vm_name":      &inputs.TagInfo{Desc: "The name of the resource"},
		},
	}
}

type hostMeasurement struct {
	Measurement
}

//nolint:lll
func (m *hostMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "vsphere_host",
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"cpu_costop_sum":                                   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Time the virtual machine is ready to run, but is unable to run due to co-scheduling constraints."},
			"cpu_demand_average":                               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "The amount of CPU resources a virtual machine would use if there were no CPU contention or CPU limit."},
			"cpu_idle_sum":                                     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Total time that the CPU spent in an idle state."},
			"cpu_latency_average":                              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Percent of time the virtual machine is unable to run because it is contending for access to the physical CPU(s)."},
			"cpu_readiness_average":                            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Percentage of time that the virtual machine was ready, but could not get scheduled to run on the physical CPU."},
			"cpu_ready_sum":                                    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Milliseconds of CPU time spent in ready state."},
			"cpu_swapwait_sum":                                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "CPU time spent waiting for swap-in."},
			"cpu_usage_average":                                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Percentage of CPU capacity being used."},
			"cpu_usagemhz_average":                             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "CPU usage, as measured in megahertz."},
			"cpu_used_sum":                                     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Time accounted to the virtual machine. If a system service runs on behalf of this virtual machine, the time spent by that service (represented by cpu.system) should be charged to this virtual machine. If not, the time spent (represented by cpu.overlap) should not be charged against this virtual machine."},
			"cpu_wait_sum":                                     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Total CPU time spent in wait state.The wait total includes time spent the CPU Idle, CPU Swap Wait, and CPU I/O Wait states."},
			"cpu_capacity_usage_average":                       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "CPU usage as a percent during the interval."},
			"cpu_capacity_contention_average":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Percent of time the virtual machine is unable to run because it is contending for access to the physical CPU(s)."},
			"datastore_maxTotalLatency_latest":                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Highest latency value across all datastores used by the host."},
			"datastore_numberReadAveraged_average":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Average number of read commands issued per second to the datastore."},
			"datastore_numberWriteAveraged_average":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Average number of write commands issued per second to the datastore during the collection interval."},
			"datastore_read_average":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Rate of reading data from the datastore."},
			"datastore_totalReadLatency_average":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Average amount of time for a read operation from the datastore."},
			"datastore_totalWriteLatency_average":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Average amount of time for a write operation from the datastore."},
			"datastore_write_average":                          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Rate of writing data to the datastore."},
			"disk_busResets_sum":                               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of SCSI-bus reset commands issued."},
			"disk_commands_sum":                                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of SCSI commands issued"},
			"disk_commandsAborted_sum":                         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of SCSI commands aborted."},
			"disk_commandsAveraged_average":                    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Average number of SCSI commands issued per second."},
			"disk_maxTotalLatency_latest":                      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Highest latency value across all disks used by the host."},
			"disk_numberRead_sum":                              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of disk reads during the collection interval."},
			"disk_numberReadAveraged_average":                  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Average number of read commands issued per second to the datastore."},
			"disk_numberWrite_sum":                             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of disk writes during the collection interval."},
			"disk_numberWriteAveraged_average":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Average number of write commands issued per second to the datastore."},
			"disk_read_average":                                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Average number of kilobytes read from the disk each second."},
			"disk_usage_average":                               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Aggregated disk I/O rate."},
			"disk_write_average":                               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Average number of kilobytes written to the disk each second."},
			"hbr_hbrNetRx_average":                             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Kilobytes per second of outgoing host-based replication network traffic (for this virtual machine or host)."},
			"hbr_hbrNetTx_average":                             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Average amount of data transmitted per second."},
			"mem_active_average":                               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Amount of memory that is actively used, as estimated by VMkernel based on recently touched memory pages."},
			"mem_activewrite_average":                          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Estimate for the amount of memory actively being written to by the virtual machine."},
			"mem_compressed_average":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Amount of memory reserved by `userworlds`."},
			"mem_compressionRate_average":                      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Rate of memory compression for the virtual machine."},
			"mem_consumed_average":                             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Amount of host physical memory consumed by a virtual machine, host, or cluster."},
			"mem_decompressionRate_average":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Rate of memory decompression for the virtual machine."},
			"mem_granted_average":                              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Amount of host physical memory or physical memory that is mapped for a virtual machine or a host."},
			"mem_latency_average":                              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Percentage of time the virtual machine is waiting to access swapped or compressed memory."},
			"mem_llSwapInRate_average":                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Rate at which memory is being swapped from host cache into active memory."},
			"mem_llSwapOutRate_average":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Rate at which memory is being swapped from active memory to host cache."},
			"mem_llSwapUsed_average":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Space used for caching swapped pages in the host cache."},
			"mem_overhead_average":                             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Host physical memory consumed by the virtualization infrastructure for running the virtual machine."},
			"mem_shared_average":                               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Amount of guest physical memory that is shared with other virtual machines, relative to a single virtual machine or to all powered-on virtual machines on a host."},
			"mem_swapin_average":                               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Amount of memory swapped-in from disk."},
			"mem_swapinRate_average":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Rate at which memory is swapped from disk into active memory."},
			"mem_swapout_average":                              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Amount of memory swapped-out to disk."},
			"mem_swapoutRate_average":                          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Rate at which memory is being swapped from active memory to disk."},
			"mem_usage_average":                                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Memory usage as percent of total configured or available memory"},
			"mem_vmmemctl_average":                             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Amount of memory allocated by the virtual machine memory control driver (`vmmemctl`)."},
			"mem_zero_average":                                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Memory that contains 0s only. Included in shared amount. Through transparent page sharing, zero memory pages can be shared among virtual machines that run the same operating system."},
			"mem_capacity_usage_average":                       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Amount of physical memory actively used."},
			"mem_capacity_contention_average":                  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Percentage of time VMs are waiting to access swapped, compressed or ballooned memory."},
			"net_broadcastRx_sum":                              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of broadcast packets received."},
			"net_broadcastTx_sum":                              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of broadcast packets transmitted."},
			"net_bytesRx_average":                              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Average amount of data received per second."},
			"net_bytesTx_average":                              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Average amount of data transmitted per second."},
			"net_droppedRx_sum":                                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of received packets dropped."},
			"net_droppedTx_sum":                                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of transmitted packets dropped."},
			"net_multicastRx_sum":                              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of multicast packets received."},
			"net_multicastTx_sum":                              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of multicast packets transmitted."},
			"net_packetsRx_sum":                                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of packets received."},
			"net_packetsTx_sum":                                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of packets transmitted."},
			"net_received_average":                             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Average rate at which data was received during the interval. This represents the bandwidth of the network."},
			"net_throughput_usage_average":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "The current network bandwidth usage for the host."},
			"net_transmitted_average":                          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Average rate at which data was transmitted during the interval. This represents the bandwidth of the network."},
			"net_usage_average":                                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Network utilization (combined transmit- and receive-rates)."},
			"power_energy_sum":                                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Total energy (in joule) used since last stats reset."},
			"power_power_average":                              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Current power usage."},
			"rescpu_actav1_latest":                             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "CPU active average over 1 minute."},
			"rescpu_actav15_latest":                            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "CPU active average over 15 minutes."},
			"rescpu_actav5_latest":                             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "CPU active average over 5 minutes."},
			"rescpu_actpk1_latest":                             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "CPU active peak over 1 minute."},
			"rescpu_actpk15_latest":                            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "CPU active peak over 15 minutes."},
			"rescpu_actpk5_latest":                             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "CPU active peak over 5 minutes."},
			"rescpu_maxLimited1_latest":                        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Amount of CPU resources over the limit that were refused, average over 1 minute."},
			"rescpu_maxLimited15_latest":                       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Amount of CPU resources over the limit that were refused, average over 15 minutes."},
			"rescpu_maxLimited5_latest":                        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Amount of CPU resources over the limit that were refused, average over 5 minutes."},
			"rescpu_runav1_latest":                             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "CPU running average over 1 minute."},
			"rescpu_runav15_latest":                            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "CPU running average over 15 minutes."},
			"rescpu_runav5_latest":                             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "CPU running average over 5 minutes."},
			"rescpu_runpk1_latest":                             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "CPU running peak over 1 minute."},
			"rescpu_runpk15_latest":                            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "CPU running peak over 15 minutes."},
			"rescpu_runpk5_latest":                             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "CPU running peak over 5 minutes."},
			"rescpu_sampleCount_latest":                        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Group CPU sample count."},
			"rescpu_samplePeriod_latest":                       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Group CPU sample period."},
			"sys_uptime_latest":                                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationSecond, Desc: "Total time elapsed since last system startup"},
			"virtualDisk_busResets_sum":                        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of SCSI-bus reset commands issued."},
			"virtualDisk_commandsAborted_sum":                  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of SCSI commands aborted"},
			"cpu_coreUtilization_average":                      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "CPU utilization of the corresponding core (if hyper-threading is enabled) as a percentage."},
			"cpu_reservedCapacity_average":                     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Total CPU capacity reserved by virtual machines."},
			"cpu_totalCapacity_average":                        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Total CPU capacity reserved by and available for virtual machines."},
			"cpu_utilization_average":                          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "CPU utilization as a percentage during the interval (CPU usage and CPU utilization might be different due to power management technologies or hyper-threading)."},
			"datastore_datastoreIops_average":                  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Storage I/O Control aggregated IOPS."},
			"datastore_datastoreMaxQueueDepth_latest":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Storage I/O Control datastore maximum queue depth."},
			"datastore_datastoreNormalReadLatency_latest":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Storage DRS datastore normalized read latency."},
			"datastore_datastoreNormalWriteLatency_latest":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Storage DRS datastore normalized write latency."},
			"datastore_datastoreReadBytes_latest":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Storage DRS datastore bytes read."},
			"datastore_datastoreReadIops_latest":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Storage DRS datastore read I/O rate."},
			"datastore_datastoreReadLoadMetric_latest":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Storage DRS datastore metric for read workload model."},
			"datastore_datastoreReadOIO_latest":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Storage DRS datastore outstanding read requests."},
			"datastore_datastoreVMObservedLatency_latest":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationUS, Desc: "The average datastore latency as seen by virtual machines."},
			"datastore_datastoreWriteBytes_latest":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Storage DRS datastore bytes written."},
			"datastore_datastoreWriteIops_latest":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Storage DRS datastore write I/O rate."},
			"datastore_datastoreWriteLoadMetric_latest":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Storage DRS datastore metric for write workload model."},
			"datastore_datastoreWriteOIO_latest":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Storage DRS datastore outstanding write requests."},
			"datastore_siocActiveTimePercentage_average":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Percentage of time Storage I/O Control actively controlled datastore latency."},
			"datastore_sizeNormalizedDatastoreLatency_average": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationUS, Desc: "Storage I/O Control size-normalized I/O latency."},
			"disk_deviceLatency_average":                       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Average amount of time it takes to complete an SCSI command from physical device."},
			"disk_deviceReadLatency_average":                   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Average amount of time to read from the physical device."},
			"disk_deviceWriteLatency_average":                  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Average amount of time to write from the physical device."},
			"disk_kernelLatency_average":                       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Average amount of time spent by VMkernel to process each SCSI command."},
			"disk_kernelReadLatency_average":                   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Average amount of time spent by VMkernel to process each SCSI read command."},
			"disk_kernelWriteLatency_average":                  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Average amount of time spent by VMkernel to process each SCSI write command."},
			"disk_maxQueueDepth_average":                       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Maximum queue depth."},
			"disk_queueLatency_average":                        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Average amount of time spent in the VMkernel queue per SCSI command."},
			"disk_queueReadLatency_average":                    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Average amount of time spent in the VMkernel queue per SCSI read command."},
			"disk_queueWriteLatency_average":                   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Average amount of time spent in the VMkernel queue per SCSI write command."},
			"disk_scsiReservationCnflctsPct_average":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Number of SCSI reservation conflicts for the LUN as a percent of total commands during the collection interval."},
			"disk_scsiReservationConflicts_sum":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of SCSI reservation conflicts for the LUN during the collection interval."},
			"disk_totalLatency_average":                        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Average amount of time taken during the collection interval to process a SCSI command issued by the guest OS to the virtual machine."},
			"disk_totalReadLatency_average":                    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Average amount of time taken to process a SCSI read command issued from the guest OS to the virtual machine."},
			"disk_totalWriteLatency_average":                   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Average amount of time taken to process a SCSI write command issued by the guest OS to the virtual machine."},
			"hbr_hbrNumVms_average":                            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of powered-on virtual machines running on this host that currently have host-based replication protection enabled."},
			"mem_consumed_userworlds.avg":                      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Amount of physical memory consumed by `userworlds` on this host."},
			"mem_consumed_vms.avg":                             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Amount of physical memory consumed by VMs on this host."},
			"mem_heap_average":                                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "VMkernel virtual address space dedicated to VMkernel main heap and related data."},
			"mem_heapfree_average":                             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Free address space in the VMkernel main heap.Varies based on number of physical devices and configuration options. There is no direct way for the user to increase or decrease this statistic. For informational purposes only: not useful for performance monitoring."},
			"mem_llSwapIn_average":                             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Amount of memory swapped-in from host cache."},
			"mem_llSwapOut_average":                            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Amount of memory swapped-out to host cache."},
			"mem_lowfreethreshold_average":                     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Threshold of free host physical memory below which ESX/ESXi will begin reclaiming memory from virtual machines through ballooning and swapping."},
			"mem_reservedCapacity_average":                     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeMB, Desc: "Total amount of memory reservation used by powered-on virtual machines and vSphere services on the host."},
			"mem_sharedcommon_average":                         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Amount of machine memory that is shared by all powered-on virtual machines and vSphere services on the host."},
			"mem_state_latest":                                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "One of four threshold levels representing the percentage of free memory on the host. The counter value determines swapping and ballooning behavior for memory reclamation."},
			"mem_swapused_average":                             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Amount of memory that is used by swap. Sum of memory swapped of all powered on VMs and vSphere services on the host."},
			"mem_sysUsage_average":                             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Amount of host physical memory used by VMkernel for core functionality, such as device drivers and other internal uses. Does not include memory used by virtual machines or vSphere services."},
			"mem_totalCapacity_average":                        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeMB, Desc: "Total amount of memory reservation used by and available for powered-on virtual machines and vSphere services on the host."},
			"mem_unreserved_average":                           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Amount of memory that is unreserved. Memory reservation not used by the Service Console, VMkernel, vSphere services and other powered on VMs user-specified memory reservations and overhead memory."},
			"mem_vmfs_pbc_capMissRatio_latest":                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Trailing average of the ratio of capacity misses to compulsory misses for the `VMFS` PB Cache."},
			"mem_vmfs_pbc_overhead_latest":                     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Amount of `VMFS` heap used by the `VMFS` PB Cache."},
			"mem_vmfs_pbc_size_latest":                         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeMB, Desc: "Space used for holding `VMFS` Pointer Blocks in memory."},
			"mem_vmfs_pbc_sizeMax_latest":                      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeMB, Desc: "Maximum size the `VMFS` Pointer Block Cache can grow to."},
			"mem_vmfs_pbc_workingSet_latest":                   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeTB, Desc: "Amount of file blocks whose addresses are cached in the `VMFS` PB Cache."},
			"mem_vmfs_pbc_workingSetMax_latest":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeTB, Desc: "Maximum amount of file blocks whose addresses are cached in the `VMFS` PB Cache."},
			"net_errorsRx_sum":                                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of packets with errors received."},
			"net_errorsTx_sum":                                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of packets with errors transmitted."},
			"net_unknownProtos_sum":                            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Number of frames with unknown protocol received."},
			"power_powerCap_average":                           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Maximum allowed power usage."},
			"storageAdapter_commandsAveraged_average":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Average number of commands issued per second by the storage adapter."},
			"storageAdapter_maxTotalLatency_latest":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Highest latency value across all storage adapters used by the host."},
			"storageAdapter_numberReadAveraged_average":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Average number of read commands issued per second by the storage adapter."},
			"storageAdapter_numberWriteAveraged_average":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Average number of write commands issued per second by the storage adapter."},
			"storageAdapter_outstandingIOs_average":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "The number of I/Os that have been issued but have not yet completed."},
			"storageAdapter_queueDepth_average":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "The maximum number of I/Os that can be outstanding at a given time."},
			"storageAdapter_queueLatency_average":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Average amount of time spent in the VMkernel queue per SCSI command."},
			"storageAdapter_queued_average":                    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "The current number of I/Os that are waiting to be issued."},
			"storageAdapter_read_average":                      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Rate of reading data by the storage adapter."},
			"storageAdapter_totalReadLatency_average":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Average amount of time for a read operation by the storage adapter."},
			"storageAdapter_totalWriteLatency_average":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Average amount of time for a write operation by the storage adapter."},
			"storageAdapter_write_average":                     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Rate of writing data by the storage adapter."},
			"storagePath_busResets_sum":                        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of SCSI-bus reset commands issued."},
			"storagePath_commandsAborted_sum":                  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of SCSI commands aborted."},
			"storagePath_commandsAveraged_average":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Average number of commands issued per second on the storage path during the collection interval."},
			"storagePath_maxTotalLatency_latest":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Highest latency value across all storage paths used by the host."},
			"storagePath_numberReadAveraged_average":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Average number of read commands issued per second on the storage path during the collection interval."},
			"storagePath_numberWriteAveraged_average":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Average number of write commands issued per second on the storage path during the collection interval."},
			"storagePath_read_average":                         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Rate of reading data on the storage path."},
			"storagePath_totalReadLatency_average":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Average amount of time for a read issued on the storage path. Total latency = kernel latency + device latency."},
			"storagePath_totalWriteLatency_average":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Average amount of time for a write issued on the storage path. Total latency = kernel latency + device latency."},
			"storagePath_write_average":                        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Rate of writing data on the storage path."},
			"sys_resourceCpuAct1_latest":                       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "CPU active average over 1 minute of the system resource group."},
			"sys_resourceCpuAct5_latest":                       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "CPU active average over 5 minutes of the system resource group."},
			"sys_resourceCpuAllocMax_latest":                   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "CPU allocation limit (in MHz) of the system resource group."},
			"sys_resourceCpuAllocMin_latest":                   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "CPU allocation reservation (in MHz) of the system resource group."},
			"sys_resourceCpuAllocShares_latest":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "CPU allocation shares of the system resource group."},
			"sys_resourceCpuMaxLimited1_latest":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "CPU maximum limited over 1 minute of the system resource group."},
			"sys_resourceCpuMaxLimited5_latest":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "CPU maximum limited over 5 minutes of the system resource group."},
			"sys_resourceCpuRun1_latest":                       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "CPU running average over 1 minute of the system."},
			"sys_resourceCpuRun5_latest":                       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "CPU running average over 5 minutes of the system resource group."},
			"sys_resourceCpuUsage_average":                     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Amount of CPU used by the Service Console and other applications during the interval by the Service Console and other applications."},
			"sys_resourceFdUsage_latest":                       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of file descriptors used by the system resource group."},
			"sys_resourceMemAllocMax_latest":                   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Memory allocation limit (in KB) of the system resource group."},
			"sys_resourceMemAllocMin_latest":                   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Memory allocation reservation (in KB) of the system resource group."},
			"sys_resourceMemAllocShares_latest":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Memory allocation shares of the system resource group."},
			"sys_resourceMemConsumed_latest":                   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Memory consumed by the system resource group."},
			"sys_resourceMemCow_latest":                        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Memory shared by the system resource group."},
			"sys_resourceMemMapped_latest":                     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Memory mapped by the system resource group."},
			"sys_resourceMemOverhead_latest":                   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Overhead memory consumed by the system resource group."},
			"sys_resourceMemShared_latest":                     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Memory saved due to sharing by the system resource group."},
			"sys_resourceMemSwapped_latest":                    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Memory swapped out by the system resource group."},
			"sys_resourceMemTouched_latest":                    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Memory touched by the system resource group."},
			"sys_resourceMemZero_latest":                       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Zero filled memory used by the system resource group."},
		},
		Tags: map[string]interface{}{
			"cluster_name": &inputs.TagInfo{Desc: "Cluster name"},
			"dcname":       &inputs.TagInfo{Desc: "`Datacenter` name"},
			"esx_hostname": &inputs.TagInfo{Desc: "The name of the ESXi host"},
			"instance":     &inputs.TagInfo{Desc: "The name of the instance"},
			"host":         &inputs.TagInfo{Desc: "The host of the vCenter"},
			"moid":         &inputs.TagInfo{Desc: "The managed object id"},
		},
	}
}

type hostObject struct{}

// nolint: lll
func (*hostObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "vsphere_host",
		Desc: "The object of the ESXi host.",
		Cat:  point.Object,
		Tags: map[string]interface{}{
			Name:            &inputs.TagInfo{Desc: "The name of the ESXi host"},
			Vendor:          &inputs.TagInfo{Desc: "The hardware vendor identification"},
			Model:           &inputs.TagInfo{Desc: "The system model identification"},
			CPUModel:        &inputs.TagInfo{Desc: "The CPU model"},
			ConnectionState: &inputs.TagInfo{Desc: "The host connection state"},
		},
		Fields: map[string]interface{}{
			MemorySize:  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The physical memory size in bytes."},
			NumCPUCores: &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of physical CPU cores on the host. Physical CPU cores are the processors contained by a CPU package."},
			NumNics:     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of network adapters."},
			BootTime:    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.TimestampNS, Desc: "The time when the host was booted."},
		},
	}
}

type vmObject struct{}

// nolint: lll
func (*vmObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "vsphere_vm",
		Desc: "The object of the virtual machine.",
		Cat:  point.Object,
		Tags: map[string]interface{}{
			Name:            &inputs.TagInfo{Desc: "The name of the virtual machine"},
			GuestFullName:   &inputs.TagInfo{Desc: "Guest operating system full name, if known"},
			HostName:        &inputs.TagInfo{Desc: "Hostname of the guest operating system, if known"},
			IPAddress:       &inputs.TagInfo{Desc: "Primary IP address assigned to the guest operating system, if known"},
			ConnectionState: &inputs.TagInfo{Desc: "Indicates whether or not the virtual machine is available for management"},
			Template:        &inputs.TagInfo{Desc: "Flag to determine whether or not this virtual machine is a template."},
		},
		Fields: map[string]interface{}{
			BootTime:         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.TimestampNS, Desc: "The timestamp when the virtual machine was most recently powered on."},
			MaxCPUUsage:      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.FrequencyMHz, Desc: "Current upper-bound on CPU usage."},
			MaxMemoryUsage:   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeMB, Desc: "Current upper-bound on memory usage."},
			NumCPU:           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of processors in the virtual machine."},
			NumEthernetCards: &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of virtual network adapters."},
			NumVirtualDisks:  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of virtual disks attached to the virtual machine."},
			MemorySizeMB:     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Memory size of the virtual machine, in megabytes."},
		},
	}
}

type clusterObject struct{}

// nolint: lll
func (*clusterObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "vsphere_cluster",
		Desc: "The object of the cluster.",
		Cat:  point.Object,
		Tags: map[string]interface{}{
			Name: &inputs.TagInfo{Desc: "The name of the cluster"},
		},
		Fields: map[string]interface{}{
			TotalCPU:          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.FrequencyMHz, Desc: "Aggregated CPU resources of all hosts, in MHz."},
			TotalMemory:       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Aggregated memory resources of all hosts, in bytes."},
			NumHosts:          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of hosts."},
			NumEffectiveHosts: &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of effective hosts."},
			NumCPUCores:       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of physical CPU cores. Physical CPU cores are the processors contained by a CPU package."},
			NumCPUThreads:     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Aggregated number of CPU threads."},
			EffectiveCPU:      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.FrequencyMHz, Desc: "Effective CPU resources (in MHz) available to run virtual machines."},
			EffectiveMemory:   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeMB, Desc: "Effective memory resources (in MB) available to run virtual machines. "},
		},
	}
}

type datastoreObject struct{}

// nolint: lll
func (*datastoreObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "vsphere_datastore",
		Desc: "The object of the datastore.",
		Cat:  point.Object,
		Tags: map[string]interface{}{
			Name: &inputs.TagInfo{Desc: "The name of the datastore"},
			URL:  &inputs.TagInfo{Desc: "The unique locator for the datastore"},
			Type: &inputs.TagInfo{Desc: "Type of file system volume, such as `VMFS` or NFS"},
		},
		Fields: map[string]interface{}{
			FreeSpace:              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Free space of this datastore, in bytes. The server periodically updates this value. It can be explicitly refreshed with the Refresh operation."},
			MaxFileSize:            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The maximum size of a file that can reside on this file system volume."},
			MaxMemoryFileSize:      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The maximum size of a snapshot or a swap file that can reside on this file system volume."},
			MaxVirtualDiskCapacity: &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The maximum capacity of a virtual disk which can be created on this volume."},
		},
	}
}

type eventMeasurement struct{}

// nolint: lll
func (*eventMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: EventMeasurementName,
		Desc: "The event of the vSphere.",
		Cat:  point.Logging,
		Tags: map[string]interface{}{
			Status:       &inputs.TagInfo{Desc: "The status of the logging"},
			"host":       &inputs.TagInfo{Desc: "The host of the vCenter"},
			EventTypeID:  &inputs.TagInfo{Desc: "The type of the event"},
			ObjectName:   &inputs.TagInfo{Desc: "The name of the object"},
			UserName:     &inputs.TagInfo{Desc: "The user who caused the event"},
			ResourceType: &inputs.TagInfo{Desc: "The resource type, such as host, vm, datastore"},
			ChangeTag:    &inputs.TagInfo{Desc: "The user entered tag to identify the operations and their side effects"},
		},
		Fields: map[string]interface{}{
			Message:  &inputs.FieldInfo{DataType: inputs.String, Type: inputs.UnknownType, Unit: inputs.NoUnit, Desc: "A formatted text message describing the event. The message may be localized."},
			ChainID:  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NoUnit, Desc: "The parent or group ID."},
			EventKey: &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NoUnit, Desc: "The event ID."},
		},
	}
}
