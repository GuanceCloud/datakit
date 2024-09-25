// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package vsphere

import (
	"fmt"

	"github.com/vmware/govmomi/vim25/mo"
)

// nolint: funlen
func getHostTagsAndFields(host *mo.HostSystem) (map[string]string, map[string]interface{}) {
	tags := map[string]string{}
	fields := map[string]interface{}{}

	tags[Name] = host.Name

	summary := host.Summary

	if summary.Hardware != nil {
		tags[Vendor] = summary.Hardware.Vendor
		tags[Model] = summary.Hardware.Model
		tags[CPUModel] = summary.Hardware.CpuModel
		fields[NumCPUCores] = summary.Hardware.NumCpuCores
		fields[MemorySize] = summary.Hardware.MemorySize
		fields[NumNics] = summary.Hardware.NumNics
	}

	if summary.Runtime != nil {
		tags[ConnectionState] = string(summary.Runtime.ConnectionState)
		fields[BootTime] = summary.Runtime.BootTime.UnixNano()
	}

	return tags, fields
}

func getVMTagsAndFields(vm *mo.VirtualMachine) (map[string]string, map[string]interface{}) {
	tags := map[string]string{}
	fields := map[string]interface{}{}

	tags[Name] = vm.Name

	summary := vm.Summary

	// guest
	if summary.Guest != nil {
		tags[GuestFullName] = summary.Guest.GuestFullName
		tags[HostName] = summary.Guest.HostName
		tags[IPAddress] = summary.Guest.IpAddress
	}

	// config
	config := summary.Config
	fields[NumCPU] = config.NumCpu
	fields[NumEthernetCards] = config.NumEthernetCards
	fields[NumVirtualDisks] = config.NumVirtualDisks
	fields[MemorySizeMB] = config.MemorySizeMB
	tags[Template] = fmt.Sprintf("%v", config.Template)

	// runtime
	runtime := summary.Runtime
	tags[ConnectionState] = string(runtime.ConnectionState)
	fields[BootTime] = runtime.BootTime.UnixNano()
	fields[MaxCPUUsage] = runtime.MaxCpuUsage
	fields[MaxMemoryUsage] = runtime.MaxMemoryUsage

	return tags, fields
}

func getClusterTagsAndFields(cluster *mo.ClusterComputeResource) (map[string]string, map[string]interface{}) {
	tags := map[string]string{}
	fields := map[string]interface{}{}

	tags[Name] = cluster.Name

	if cluster.Summary != nil {
		s := cluster.Summary.GetComputeResourceSummary()
		if s != nil {
			fields[TotalCPU] = s.TotalCpu
			fields[NumHosts] = s.NumHosts
			fields[NumEffectiveHosts] = s.NumEffectiveHosts
			fields[TotalMemory] = s.TotalMemory
			fields[NumCPUCores] = s.NumCpuCores
			fields[NumCPUThreads] = s.NumCpuThreads
			fields[EffectiveCPU] = s.EffectiveCpu
			fields[EffectiveMemory] = s.EffectiveMemory
		}
	}

	return tags, fields
}

func getDatastoreTagsAndFields(ds *mo.Datastore) (map[string]string, map[string]interface{}) {
	tags := map[string]string{}
	fields := map[string]interface{}{}

	tags[Name] = ds.Name

	if ds.Info != nil {
		info := ds.Info.GetDatastoreInfo()
		if info != nil {
			tags[URL] = info.Url
			fields[FreeSpace] = info.FreeSpace
			fields[MaxFileSize] = info.MaxFileSize
			fields[MaxMemoryFileSize] = info.MaxMemoryFileSize
			fields[MaxVirtualDiskCapacity] = info.MaxVirtualDiskCapacity
		}
	}

	tags[Type] = ds.Summary.Type

	return tags, fields
}
