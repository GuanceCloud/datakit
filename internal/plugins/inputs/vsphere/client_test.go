// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package vsphere

import (
	"context"
	"testing"
	"time"

	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	dknet "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
)

func TestMakeMetricIdentifier(t *testing.T) {
	client := &Client{}

	tests := []struct {
		metric          string
		wantMeasurement string
		wantField       string
	}{
		{
			metric:          "cpu.usage.average",
			wantMeasurement: "vsphere_vm",
			wantField:       "cpu_usage_average",
		},
		{
			metric:          "sys.uptime.latest",
			wantMeasurement: "vsphere_vm",
			wantField:       "sys_uptime_latest",
		},
		{
			metric:          "uptime",
			wantMeasurement: "vsphere_vm",
			wantField:       "uptime",
		},
	}

	for _, tt := range tests {
		t.Run(tt.metric, func(t *testing.T) {
			gotMeasurement, gotField := client.makeMetricIdentifier("vsphere_vm", tt.metric)
			if gotMeasurement != tt.wantMeasurement || gotField != tt.wantField {
				t.Fatalf(
					"makeMetricIdentifier(%q) = (%q, %q), want (%q, %q)",
					tt.metric,
					gotMeasurement,
					gotField,
					tt.wantMeasurement,
					tt.wantField,
				)
			}
		})
	}
}

func TestCleanGuestID(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want string
	}{
		{
			name: "trims Guest suffix",
			id:   "ubuntu64Guest",
			want: "ubuntu64",
		},
		{
			name: "keeps non guest suffix",
			id:   "other",
			want: "other",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cleanGuestID(tt.id); got != tt.want {
				t.Fatalf("cleanGuestID(%q) = %q, want %q", tt.id, got, tt.want)
			}
		})
	}
}

func TestAlignSamples(t *testing.T) {
	client := &Client{}

	base := time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC)
	info := []types.PerfSampleInfo{
		{Timestamp: base.Add(10 * time.Second), Interval: 20},
		{Timestamp: base.Add(20 * time.Second), Interval: 20},
		{Timestamp: base.Add(70 * time.Second), Interval: 20},
	}
	values := []int64{10, 20, 30}

	gotInfo, gotValues := client.alignSamples(info, values, time.Minute)

	if len(gotInfo) != 2 || len(gotValues) != 2 {
		t.Fatalf("got %d infos/%d values, want 2/2", len(gotInfo), len(gotValues))
	}
	if gotInfo[0].Timestamp != base || gotValues[0] != 15 {
		t.Fatalf("first bucket = (%s, %v), want (%s, 15)", gotInfo[0].Timestamp, gotValues[0], base)
	}
	if gotInfo[1].Timestamp != base.Add(time.Minute) || gotValues[1] != 30 {
		t.Fatalf(
			"second bucket = (%s, %v), want (%s, 30)",
			gotInfo[1].Timestamp,
			gotValues[1],
			base.Add(time.Minute),
		)
	}
}

func TestAlignSamplesSkipsNegativeAndMissingValues(t *testing.T) {
	client := &Client{}

	base := time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC)
	info := []types.PerfSampleInfo{
		{Timestamp: base.Add(10 * time.Second), Interval: 20},
		{Timestamp: base.Add(20 * time.Second), Interval: 20},
		{Timestamp: base.Add(70 * time.Second), Interval: 20},
	}
	values := []int64{-1, 20}

	gotInfo, gotValues := client.alignSamples(info, values, time.Minute)

	if len(gotInfo) != 1 || len(gotValues) != 1 {
		t.Fatalf("got %d infos/%d values, want 1/1", len(gotInfo), len(gotValues))
	}
	if gotInfo[0].Timestamp != base || gotValues[0] != 20 {
		t.Fatalf("bucket = (%s, %v), want (%s, 20)", gotInfo[0].Timestamp, gotValues[0], base)
	}
}

func TestGetParentNilParentRef(t *testing.T) {
	client := &Client{
		resourceKinds: map[string]*resourceKind{
			"host": {
				objects: objectMap{
					"host-1": {
						name: "host-1",
					},
				},
			},
		},
	}

	parent, ok := client.getParent(&objectRef{name: "vm-1"}, &resourceKind{parent: "host"})

	if ok {
		t.Fatal("getParent() ok = true, want false")
	}
	if parent != nil {
		t.Fatalf("getParent() parent = %#v, want nil", parent)
	}
}

func TestGetParentMissingParentKind(t *testing.T) {
	parentRef := types.ManagedObjectReference{Type: "HostSystem", Value: "host-1"}
	client := &Client{resourceKinds: map[string]*resourceKind{}}

	parent, ok := client.getParent(
		&objectRef{name: "vm-1", parentRef: &parentRef},
		&resourceKind{parent: "host"},
	)

	if ok {
		t.Fatal("getParent() ok = true, want false")
	}
	if parent != nil {
		t.Fatalf("getParent() parent = %#v, want nil", parent)
	}
}

func TestPopulateTagsParentChain(t *testing.T) {
	dcRef := types.ManagedObjectReference{Type: "Datacenter", Value: "dc-1"}
	clusterRef := types.ManagedObjectReference{Type: "ClusterComputeResource", Value: "domain-c1"}
	hostRef := types.ManagedObjectReference{Type: "HostSystem", Value: "host-1"}
	client := &Client{
		resourceKinds: map[string]*resourceKind{
			"datacenter": {
				name:    "datacenter",
				pKey:    "dcname",
				objects: objectMap{dcRef.Value: {name: "dc-a", ref: dcRef}},
			},
			"cluster": {
				name:      "cluster",
				pKey:      "cluster_name",
				parent:    "datacenter",
				parentTag: "dcname",
				objects: objectMap{
					clusterRef.Value: {
						name:      "cluster-a",
						ref:       clusterRef,
						parentRef: &dcRef,
					},
				},
			},
			"host": {
				name:      "host",
				pKey:      "esx_hostname",
				parent:    "cluster",
				parentTag: "cluster_name",
				objects:   objectMap{},
			},
		},
	}
	host := &objectRef{
		name:      "esx-1",
		ref:       hostRef,
		parentRef: &clusterRef,
		dcname:    "dc-a",
	}
	tags := map[string]string{"source": "old-source"}

	client.populateTags(host, client.resourceKinds["host"], tags, performance.MetricSeries{Instance: "vmnic0"})

	if _, ok := tags["source"]; ok {
		t.Fatalf("source tag should be removed when pKey is set: %#v", tags)
	}
	if tags["esx_hostname"] != "esx-1" || tags["cluster_name"] != "cluster-a" || tags["dcname"] != "dc-a" {
		t.Fatalf("unexpected hierarchy tags: %#v", tags)
	}
	if tags["instance"] != "vmnic0" {
		t.Fatalf("instance = %q, want vmnic0", tags["instance"])
	}
}

func TestPopulateTagsNilTagsNoop(t *testing.T) {
	client := &Client{}

	client.populateTags(&objectRef{name: "vm-1"}, &resourceKind{pKey: "vm_name"}, nil, performance.MetricSeries{})
}

func TestGetResourcePoolName(t *testing.T) {
	rpRef := types.ManagedObjectReference{Type: "ResourcePool", Value: "resgroup-1"}
	pools := objectMap{
		"resgroup-1": {
			name: "pool-a",
			ref:  rpRef,
		},
	}

	if got := getResourcePoolName(rpRef, pools); got != "pool-a" {
		t.Fatalf("getResourcePoolName() = %q, want pool-a", got)
	}
	if got := getResourcePoolName(types.ManagedObjectReference{Type: "ResourcePool", Value: "missing"}, pools); got != "Resources" {
		t.Fatalf("getResourcePoolName() = %q, want Resources", got)
	}
}

func TestTimeoutContextAppliesClientTimeout(t *testing.T) {
	client := &Client{Timeout: time.Second}

	ctx, cancel := client.timeoutContext(context.Background())
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("timeoutContext() has no deadline")
	}

	remaining := time.Until(deadline)
	if remaining <= 0 || remaining > client.Timeout {
		t.Fatalf("timeoutContext() deadline remaining = %s, want within %s", remaining, client.Timeout)
	}
}

func TestQueryEventsAppliesClientTimeout(t *testing.T) {
	client := &Client{Timeout: time.Second}

	orig := queryEvents
	defer func() {
		queryEvents = orig
	}()

	queryEvents = func(ctx context.Context, gotClient *Client, _ types.EventFilterSpec) ([]types.BaseEvent, error) {
		if gotClient != client {
			t.Fatalf("queryEvents() client = %p, want %p", gotClient, client)
		}
		deadline, ok := ctx.Deadline()
		if !ok {
			t.Fatal("QueryEvents() context has no deadline")
		}
		remaining := time.Until(deadline)
		if remaining <= 0 || remaining > client.Timeout {
			t.Fatalf("QueryEvents() deadline remaining = %s, want within %s", remaining, client.Timeout)
		}
		return nil, nil
	}

	if _, err := client.QueryEvents(context.Background(), types.EventFilterSpec{}); err != nil {
		t.Fatalf("QueryEvents() error = %v, want nil", err)
	}
}

func TestCreateVSphereClientRejectsInvalidURL(t *testing.T) {
	ipt := &Input{}

	if _, err := ipt.createVSphereClient("://bad-url"); err == nil {
		t.Fatal("createVSphereClient() error = nil, want error")
	}
	if _, err := ipt.createVSphereClient2("://bad-url"); err == nil {
		t.Fatal("createVSphereClient2() error = nil, want error")
	}
}

func TestCreateVSphereClientRejectsInvalidTLSConfig(t *testing.T) {
	ipt := &Input{
		TLSClientConfig: &dknet.TLSClientConfig{
			CaCerts: []string{"/path/to/missing-ca.pem"},
		},
	}

	if _, err := ipt.createVSphereClient("https://vcenter.local/sdk"); err == nil {
		t.Fatal("createVSphereClient() error = nil, want TLS config error")
	}
}

func TestDiscoverCanceledContext(t *testing.T) {
	client := &Client{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := client.discover(ctx); err == nil {
		t.Fatal("discover() error = nil, want canceled context error")
	}
}

func TestScaleMetricValue(t *testing.T) {
	metricInfo := map[string]*types.PerfCounterInfo{
		"cpu.usage.average": {
			UnitInfo: &types.ElementDescription{Key: "percent"},
		},
		"cpu.usagemhz.average": {
			UnitInfo: &types.ElementDescription{Key: "megaHertz"},
		},
	}

	t.Run("missing metric info skips value", func(t *testing.T) {
		if _, ok := scaleMetricValue(metricInfo, "unknown.counter", 42); ok {
			t.Fatal("scaleMetricValue() ok = true, want false")
		}
	})

	t.Run("nil metric info skips value", func(t *testing.T) {
		metricInfo["nil.counter"] = nil
		if _, ok := scaleMetricValue(metricInfo, "nil.counter", 42); ok {
			t.Fatal("scaleMetricValue() ok = true, want false")
		}
	})

	t.Run("nil metric unit skips value", func(t *testing.T) {
		metricInfo["nil.unit.counter"] = &types.PerfCounterInfo{}
		if _, ok := scaleMetricValue(metricInfo, "nil.unit.counter", 42); ok {
			t.Fatal("scaleMetricValue() ok = true, want false")
		}
	})

	t.Run("percent is scaled down", func(t *testing.T) {
		got, ok := scaleMetricValue(metricInfo, "cpu.usage.average", 1234)
		if !ok {
			t.Fatal("scaleMetricValue() ok = false, want true")
		}
		if got != 12.34 {
			t.Fatalf("scaleMetricValue() = %v, want 12.34", got)
		}
	})

	t.Run("non percent keeps original value", func(t *testing.T) {
		got, ok := scaleMetricValue(metricInfo, "cpu.usagemhz.average", 1234)
		if !ok {
			t.Fatal("scaleMetricValue() ok = false, want true")
		}
		if got != 1234 {
			t.Fatalf("scaleMetricValue() = %v, want 1234", got)
		}
	})
}

func TestGetVMGuestInfoNilGuest(t *testing.T) {
	tests := []struct {
		name      string
		vm        *mo.VirtualMachine
		wantGuest string
		wantUUID  string
	}{
		{name: "nil vm", wantGuest: "unknown"},
		{name: "nil config", vm: &mo.VirtualMachine{}, wantGuest: "unknown"},
		{
			name: "nil guest",
			vm: &mo.VirtualMachine{
				Config: &types.VirtualMachineConfigInfo{
					GuestId: "ubuntu64Guest",
					Uuid:    "vm-uuid",
				},
			},
			wantGuest: "ubuntu64",
			wantUUID:  "vm-uuid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guest, uuid, lookup := getVMGuestInfo(tt.vm)

			if guest != tt.wantGuest {
				t.Fatalf("guest = %q, want %q", guest, tt.wantGuest)
			}
			if uuid != tt.wantUUID {
				t.Fatalf("uuid = %q, want %q", uuid, tt.wantUUID)
			}
			if len(lookup) != 0 {
				t.Fatalf("lookup = %#v, want empty", lookup)
			}
		})
	}
}

func TestGetVMGuestInfoWithGuestNetwork(t *testing.T) {
	vm := &mo.VirtualMachine{
		Config: &types.VirtualMachineConfigInfo{
			GuestId: "ubuntu64Guest",
			Uuid:    "vm-uuid",
		},
		Guest: &types.GuestInfo{
			GuestId:  "centos64Guest",
			HostName: "vm.example",
			Net: []types.GuestNicInfo{
				{
					DeviceConfigId: 4000,
					IpConfig: &types.NetIpConfigInfo{
						IpAddress: []types.NetIpConfigInfoIpAddress{
							{IpAddress: "10.0.0.2", State: "deprecated"},
							{IpAddress: "10.0.0.1", State: "preferred"},
							{IpAddress: "2001:db8::1", State: "preferred"},
						},
					},
				},
			},
		},
	}

	guest, uuid, lookup := getVMGuestInfo(vm)

	if guest != "centos64" {
		t.Fatalf("guest = %q, want centos64", guest)
	}
	if uuid != "vm-uuid" {
		t.Fatalf("uuid = %q, want vm-uuid", uuid)
	}
	if lookup["guesthostname"] != "vm.example" {
		t.Fatalf("guesthostname = %q, want vm.example", lookup["guesthostname"])
	}
	if lookup["nic/4000/ipv4"] != "10.0.0.1,10.0.0.2" {
		t.Fatalf("ipv4 lookup = %q, want preferred first", lookup["nic/4000/ipv4"])
	}
	if lookup["nic/4000/ipv6"] != "2001:db8::1" {
		t.Fatalf("ipv6 lookup = %q, want 2001:db8::1", lookup["nic/4000/ipv6"])
	}
}
