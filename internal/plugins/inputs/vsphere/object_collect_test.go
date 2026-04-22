// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package vsphere

import (
	"testing"
	"time"

	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

func TestGetHostTagsAndFields(t *testing.T) {
	boot := time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC)
	host := &mo.HostSystem{
		ManagedEntity: mo.ManagedEntity{Name: "esx-1"},
		Summary: types.HostListSummary{
			Hardware: &types.HostHardwareSummary{
				Vendor:      "Dell",
				Model:       "PowerEdge",
				CpuModel:    "Intel",
				NumCpuCores: 16,
				MemorySize:  128,
				NumNics:     4,
			},
			Runtime: &types.HostRuntimeInfo{
				ConnectionState:   types.HostSystemConnectionStateConnected,
				PowerState:        types.HostSystemPowerStatePoweredOn,
				BootTime:          &boot,
				InMaintenanceMode: false,
			},
		},
	}

	tags, fields := getHostTagsAndFields(host)

	if tags[Name] != "esx-1" || tags[Vendor] != "Dell" || tags[PowerState] != "poweredOn" {
		t.Fatalf("unexpected tags: %#v", tags)
	}
	if fields[NumCPUCores] != int16(16) || fields[MemorySize] != int64(128) || fields[BootTime] != boot.UnixNano() {
		t.Fatalf("unexpected fields: %#v", fields)
	}
}

func TestGetHostTagsAndFieldsNilBootTime(t *testing.T) {
	host := &mo.HostSystem{
		ManagedEntity: mo.ManagedEntity{Name: "esx-1"},
		Summary: types.HostListSummary{
			Runtime: &types.HostRuntimeInfo{
				ConnectionState:   types.HostSystemConnectionStateConnected,
				PowerState:        types.HostSystemPowerStatePoweredOn,
				InMaintenanceMode: true,
			},
		},
	}

	tags, fields := getHostTagsAndFields(host)

	if tags[Name] != "esx-1" || tags[InMaintenanceMode] != "true" {
		t.Fatalf("unexpected tags: %#v", tags)
	}
	if _, ok := fields[BootTime]; ok {
		t.Fatalf("boot time field exists for nil boot time: %#v", fields)
	}
}

func TestGetVMTagsAndFields(t *testing.T) {
	boot := time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC)
	vm := &mo.VirtualMachine{
		ManagedEntity: mo.ManagedEntity{Name: "vm-1"},
		Summary: types.VirtualMachineSummary{
			Guest: &types.VirtualMachineGuestSummary{
				GuestFullName: "Ubuntu Linux",
				HostName:      "vm-1.local",
				IpAddress:     "10.0.0.10",
			},
			Config: types.VirtualMachineConfigSummary{
				NumCpu:           4,
				NumEthernetCards: 2,
				NumVirtualDisks:  3,
				MemorySizeMB:     8192,
				Template:         false,
			},
			Runtime: types.VirtualMachineRuntimeInfo{
				ConnectionState: types.VirtualMachineConnectionStateConnected,
				PowerState:      types.VirtualMachinePowerStatePoweredOn,
				BootTime:        &boot,
				MaxCpuUsage:     1000,
				MaxMemoryUsage:  2048,
			},
		},
	}

	tags, fields := getVMTagsAndFields(vm)

	if tags[Name] != "vm-1" || tags[GuestFullName] != "Ubuntu Linux" || tags[PowerState] != "poweredOn" {
		t.Fatalf("unexpected tags: %#v", tags)
	}
	if fields[NumCPU] != int32(4) || fields[MemorySizeMB] != int32(8192) || fields[BootTime] != boot.UnixNano() {
		t.Fatalf("unexpected fields: %#v", fields)
	}
}

func TestGetVMTagsAndFieldsNilVM(t *testing.T) {
	tags, fields := getVMTagsAndFields(nil)

	if len(tags) != 0 || len(fields) != 0 {
		t.Fatalf("got tags=%#v fields=%#v, want empty maps", tags, fields)
	}
}

func TestGetClusterTagsAndFields(t *testing.T) {
	cluster := &mo.ClusterComputeResource{
		ComputeResource: mo.ComputeResource{
			ManagedEntity: mo.ManagedEntity{Name: "cluster-1"},
			Summary: &types.ComputeResourceSummary{
				TotalCpu:          100,
				NumHosts:          3,
				NumEffectiveHosts: 2,
				TotalMemory:       2048,
				NumCpuCores:       32,
				NumCpuThreads:     64,
				EffectiveCpu:      80,
				EffectiveMemory:   1024,
			},
		},
	}

	tags, fields := getClusterTagsAndFields(cluster)

	if tags[Name] != "cluster-1" {
		t.Fatalf("unexpected tags: %#v", tags)
	}
	if fields[TotalCPU] != int32(100) || fields[NumHosts] != int32(3) || fields[EffectiveMemory] != int64(1024) {
		t.Fatalf("unexpected fields: %#v", fields)
	}
}

func TestGetDatastoreTagsAndFields(t *testing.T) {
	ds := &mo.Datastore{
		ManagedEntity: mo.ManagedEntity{Name: "ds-1"},
		Summary:       types.DatastoreSummary{Type: "VMFS"},
		Info: &types.LocalDatastoreInfo{
			DatastoreInfo: types.DatastoreInfo{
				Url:                    "ds:///vmfs/volumes/datastore1",
				FreeSpace:              100,
				MaxFileSize:            200,
				MaxMemoryFileSize:      300,
				MaxVirtualDiskCapacity: 400,
			},
		},
	}

	tags, fields := getDatastoreTagsAndFields(ds)

	if tags[Name] != "ds-1" || tags[Type] != "VMFS" || tags[URL] != "ds:///vmfs/volumes/datastore1" {
		t.Fatalf("unexpected tags: %#v", tags)
	}
	if fields[FreeSpace] != int64(100) || fields[MaxVirtualDiskCapacity] != int64(400) {
		t.Fatalf("unexpected fields: %#v", fields)
	}
}
