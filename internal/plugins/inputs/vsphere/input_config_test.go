// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package vsphere

import (
	"testing"
	"time"

	dknet "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
)

func TestDefaultInputInitializesOperationalDefaults(t *testing.T) {
	ipt := defaultInput()

	if ipt.Timeout.Duration != time.Minute {
		t.Fatalf("timeout = %s, want 1m", ipt.Timeout.Duration)
	}
	if ipt.ObjectDiscoveryInterval.Duration != 5*time.Minute {
		t.Fatalf("object discovery interval = %s, want 5m", ipt.ObjectDiscoveryInterval.Duration)
	}
	if ipt.MaxQueryObjects != 256 || ipt.MaxQueryMetrics != 256 {
		t.Fatalf("max query objects/metrics = %d/%d, want 256/256", ipt.MaxQueryObjects, ipt.MaxQueryMetrics)
	}
	if !ipt.HostInstances || !ipt.VMInstances {
		t.Fatalf("host/vm instances = %v/%v, want true/true", ipt.HostInstances, ipt.VMInstances)
	}
	if ipt.semStop == nil || ipt.g == nil || ipt.feeder == nil {
		t.Fatalf("runtime dependencies are not initialized: semStop=%v g=%v feeder=%v", ipt.semStop, ipt.g, ipt.feeder)
	}
}

func TestSetupResourceKinds(t *testing.T) {
	ipt := defaultInput()
	client := &Client{}

	ipt.setupResource(client)

	tests := map[string]struct {
		parent    string
		parentTag string
		pKey      string
		realTime  bool
		sampling  int32
	}{
		"datacenter": {pKey: "dcname", sampling: int32(ipt.HistoricalInterval.Duration.Seconds())},
		"cluster":    {parent: "datacenter", parentTag: "dcname", pKey: "cluster_name", sampling: int32(ipt.HistoricalInterval.Duration.Seconds())},
		"host":       {parent: "cluster", parentTag: "cluster_name", pKey: "esx_hostname", realTime: true, sampling: 20},
		"vm":         {parent: "host", parentTag: "esx_hostname", pKey: "vm_name", realTime: true, sampling: 20},
		"datastore":  {pKey: "dsname", sampling: int32(ipt.HistoricalInterval.Duration.Seconds())},
	}

	if len(client.resourceKinds) != len(tests) {
		t.Fatalf("resourceKinds length = %d, want %d", len(client.resourceKinds), len(tests))
	}

	for name, want := range tests {
		res, ok := client.resourceKinds[name]
		if !ok {
			t.Fatalf("resource kind %q missing", name)
		}
		if res.name != name || res.parent != want.parent || res.parentTag != want.parentTag || res.pKey != want.pKey {
			t.Fatalf("resource kind %q metadata = %#v, want parent=%q parentTag=%q pKey=%q", name, res, want.parent, want.parentTag, want.pKey)
		}
		if res.realTime != want.realTime || res.sampling != want.sampling {
			t.Fatalf("resource kind %q timing = realtime %v sampling %d, want realtime %v sampling %d", name, res.realTime, res.sampling, want.realTime, want.sampling)
		}
		if res.objects == nil || res.filters == nil || res.getObjects == nil {
			t.Fatalf("resource kind %q not fully initialized: %#v", name, res)
		}
	}
}

func TestInputLifecycleMethods(t *testing.T) {
	ipt := &Input{}

	ipt.setup()
	if ipt.semStop == nil {
		t.Fatal("setup() did not initialize semStop")
	}

	if err := ipt.Pause(); err != nil {
		t.Fatalf("Pause() error = %v, want nil", err)
	}
	if !ipt.pause.Load() {
		t.Fatal("Pause() did not set pause flag")
	}

	if err := ipt.Resume(); err != nil {
		t.Fatalf("Resume() error = %v, want nil", err)
	}
	if ipt.pause.Load() {
		t.Fatal("Resume() did not clear pause flag")
	}

	ipt.Terminate()
}

func TestInputInitReturnsGetClientError(t *testing.T) {
	ipt := &Input{
		Vcenter: "https://vcenter.local",
		TLSClientConfig: &dknet.TLSClientConfig{
			CaCerts: []string{"/path/to/missing-ca.pem"},
		},
	}

	if err := ipt.init(); err == nil {
		t.Fatal("init() error = nil, want TLS config error")
	}
}

func TestInputInitWithExistingClient(t *testing.T) {
	ipt := &Input{client: &Client{}}

	if err := ipt.init(); err != nil {
		t.Fatalf("init() error = %v, want nil", err)
	}
}

func TestMakePoints(t *testing.T) {
	ts := time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC)
	ipt := &Input{
		Tags: map[string]string{"env": "test"},
	}

	ipt.makePoints(map[string]metricEntry{
		"vm cpu": {
			name:   "vsphere_vm",
			ts:     ts,
			tags:   map[string]string{"host": "vcenter.local"},
			fields: map[string]interface{}{"cpu_usage_average": float64(1.2)},
		},
	})

	if len(ipt.collectCache) != 1 {
		t.Fatalf("collectCache length = %d, want 1", len(ipt.collectCache))
	}
	pt := ipt.collectCache[0]
	if pt.Name() != "vsphere_vm" {
		t.Fatalf("point name = %q, want vsphere_vm", pt.Name())
	}
	if got := pt.GetTag("host"); got != "vcenter.local" {
		t.Fatalf("host tag = %q, want vcenter.local", got)
	}
	if got := pt.GetTag("env"); got != "test" {
		t.Fatalf("env tag = %q, want test", got)
	}
}
