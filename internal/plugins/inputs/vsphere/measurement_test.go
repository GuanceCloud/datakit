// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package vsphere

import (
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func TestMeasurementInfo(t *testing.T) {
	tests := []struct {
		name       string
		m          inputs.Measurement
		wantName   string
		wantCat    point.Category
		wantFields []string
		wantTags   []string
	}{
		{
			name:       "base metric",
			m:          &Measurement{},
			wantName:   "vsphere_cpu",
			wantCat:    point.Metric,
			wantFields: nil,
			wantTags:   nil,
		},
		{
			name:       "cluster metric",
			m:          &clusterMeasurement{},
			wantName:   "vsphere_cluster",
			wantCat:    point.Metric,
			wantFields: []string{"cpu_usage_average", "mem_usage_average", "vmop_numPoweron_latest"},
			wantTags:   []string{"host", "dcname", "cluster_name", "moid"},
		},
		{
			name:       "datastore metric",
			m:          &datastoreMeasurement{},
			wantName:   "vsphere_datastore",
			wantCat:    point.Metric,
			wantFields: []string{"datastore_busResets_sum", "disk_capacity_latest", "disk_used_latest"},
			wantTags:   []string{"host", "dcname", "moid", "dsname"},
		},
		{
			name:       "vm metric",
			m:          &vmMeasurement{},
			wantName:   "vsphere_vm",
			wantCat:    point.Metric,
			wantFields: []string{"cpu_usage_average", "mem_usage_average", "virtualDisk_read_average"},
			wantTags:   []string{"host", "dcname", "vm_name", "moid"},
		},
		{
			name:       "host metric",
			m:          &hostMeasurement{},
			wantName:   "vsphere_host",
			wantCat:    point.Metric,
			wantFields: []string{"cpu_usage_average", "mem_usage_average", "sys_uptime_latest"},
			wantTags:   []string{"host", "dcname", "cluster_name", "esx_hostname", "moid"},
		},
		{
			name:       "host object",
			m:          &hostObject{},
			wantName:   "vsphere_host",
			wantCat:    point.Object,
			wantFields: []string{MemorySize, NumCPUCores, NumNics, BootTime},
			wantTags:   []string{Name, Vendor, Model, CPUModel, ConnectionState, InMaintenanceMode, PowerState},
		},
		{
			name:       "vm object",
			m:          &vmObject{},
			wantName:   "vsphere_vm",
			wantCat:    point.Object,
			wantFields: []string{BootTime, MaxCPUUsage, MaxMemoryUsage, NumCPU, NumEthernetCards, NumVirtualDisks, MemorySizeMB},
			wantTags:   []string{Name, GuestFullName, HostName, IPAddress, ConnectionState, Template, PowerState},
		},
		{
			name:       "cluster object",
			m:          &clusterObject{},
			wantName:   "vsphere_cluster",
			wantCat:    point.Object,
			wantFields: []string{TotalCPU, TotalMemory, NumHosts, NumEffectiveHosts, NumCPUCores, NumCPUThreads, EffectiveCPU, EffectiveMemory},
			wantTags:   []string{Name},
		},
		{
			name:       "datastore object",
			m:          &datastoreObject{},
			wantName:   "vsphere_datastore",
			wantCat:    point.Object,
			wantFields: []string{FreeSpace, MaxFileSize, MaxMemoryFileSize, MaxVirtualDiskCapacity},
			wantTags:   []string{Name, URL, Type},
		},
		{
			name:       "event",
			m:          &eventMeasurement{},
			wantName:   EventMeasurementName,
			wantCat:    point.Logging,
			wantFields: []string{Message, ChainID, EventKey},
			wantTags:   []string{Status, "host", EventTypeID, ObjectName, UserName, ResourceType, ChangeTag},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := tt.m.Info()
			if info == nil {
				t.Fatal("Info() = nil")
			}
			if info.Name != tt.wantName {
				t.Fatalf("Info().Name = %q, want %q", info.Name, tt.wantName)
			}
			if info.Cat != tt.wantCat {
				t.Fatalf("Info().Cat = %q, want %q", info.Cat, tt.wantCat)
			}
			for _, field := range tt.wantFields {
				if _, ok := info.Fields[field]; !ok {
					t.Fatalf("Info().Fields missing %q", field)
				}
			}
			for _, tag := range tt.wantTags {
				if _, ok := info.Tags[tag]; !ok {
					t.Fatalf("Info().Tags missing %q", tag)
				}
			}
		})
	}
}

func TestPointBuildersAndEmptyInfo(t *testing.T) {
	metric := (&Measurement{
		name:     "vsphere_vm",
		tags:     map[string]string{"host": "vcenter.local"},
		fields:   map[string]interface{}{"cpu_usage_average": float64(1.2)},
		election: true,
	}).Point()
	if metric.Name() != "vsphere_vm" {
		t.Fatalf("metric point name = %q, want vsphere_vm", metric.Name())
	}
	if got := metric.GetTag("host"); got != "vcenter.local" {
		t.Fatalf("metric host tag = %q, want vcenter.local", got)
	}

	obj := (&Object{
		class:    "vsphere_host",
		tags:     map[string]string{Name: "esx-1"},
		fields:   map[string]interface{}{NumCPUCores: int64(8)},
		election: true,
	}).Point()
	if obj.Name() != "vsphere_host" {
		t.Fatalf("object point name = %q, want vsphere_host", obj.Name())
	}
	if got := obj.GetTag(Name); got != "esx-1" {
		t.Fatalf("object name tag = %q, want esx-1", got)
	}

	log := (&Log{
		source:   EventMeasurementName,
		tags:     map[string]string{Status: Info},
		fields:   map[string]interface{}{Message: "event"},
		election: true,
	}).Point()
	if log.Name() != EventMeasurementName {
		t.Fatalf("log point name = %q, want %s", log.Name(), EventMeasurementName)
	}
	if got := log.GetTag(Status); got != Info {
		t.Fatalf("log status tag = %q, want %s", got, Info)
	}

	if info := (&Object{}).Info(); info == nil {
		t.Fatal("Object.Info() = nil")
	}
	if info := (&Log{}).Info(); info == nil {
		t.Fatal("Log.Info() = nil")
	}
}

func TestInputMetadataMethods(t *testing.T) {
	ipt := &Input{}

	if ipt.Catalog() != catalogName {
		t.Fatalf("Catalog() = %q, want %q", ipt.Catalog(), catalogName)
	}
	if len(ipt.AvailableArchs()) == 0 {
		t.Fatal("AvailableArchs() is empty")
	}
	if got := ipt.Dashboard(inputs.I18nZh)["title"]; got != "vSphere 监控视图" {
		t.Fatalf("Dashboard(zh) title = %q", got)
	}
	if got := ipt.Dashboard(inputs.I18nEn)["title"]; got != "vSphere Monitor View" {
		t.Fatalf("Dashboard(en) title = %q", got)
	}
	if got := ipt.Dashboard(inputs.I18n(99)); got != nil {
		t.Fatalf("Dashboard(missing) = %#v, want nil", got)
	}

	measurements := ipt.SampleMeasurement()
	if len(measurements) != 9 {
		t.Fatalf("SampleMeasurement() length = %d, want 9", len(measurements))
	}
	for _, m := range measurements {
		if m.Info() == nil {
			t.Fatalf("%T Info() = nil", m)
		}
	}

	ipt.Election = true
	if !ipt.ElectionEnabled() {
		t.Fatal("ElectionEnabled() = false, want true")
	}
}
