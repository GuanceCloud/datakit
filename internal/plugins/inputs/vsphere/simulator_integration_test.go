// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build vsphere_simulator
// +build vsphere_simulator

package vsphere

import (
	"context"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/vmware/govmomi/simulator"
	"github.com/vmware/govmomi/vim25/mo"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	dknet "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
)

func newSimulatorInput(t *testing.T) (*Input, func()) {
	t.Helper()

	model := simulator.VPX()
	if err := model.Create(); err != nil {
		t.Fatalf("create simulator model: %v", err)
	}

	server := model.Service.NewServer()

	username := ""
	password := ""
	if server.URL.User != nil {
		username = server.URL.User.Username()
		password, _ = server.URL.User.Password()
	}

	ipt := defaultInput()
	ipt.Vcenter = server.URL.String()
	ipt.Username = username
	ipt.Password = password
	ipt.ObjectDiscoveryInterval.Duration = 0
	ipt.timeout = 10 * time.Second

	cleanup := func() {
		if ipt.client != nil && ipt.client.Client != nil {
			_ = ipt.client.Client.Logout(context.Background())
		}
		server.Close()
		model.Remove()
	}

	return ipt, cleanup
}

func TestSimulatorDiscover(t *testing.T) {
	ipt, cleanup := newSimulatorInput(t)
	defer cleanup()

	if err := ipt.init(); err != nil {
		t.Fatalf("init simulator input: %v", err)
	}

	for _, resourceType := range []string{"datacenter", "cluster", "host", "vm", "datastore"} {
		res := ipt.client.resourceKinds[resourceType]
		if res == nil {
			t.Fatalf("resource kind %q is nil", resourceType)
		}
		if len(res.objects) == 0 {
			t.Fatalf("resource kind %q discovered no objects", resourceType)
		}
	}
}

func TestSimulatorCollectResourceObject(t *testing.T) {
	ipt, cleanup := newSimulatorInput(t)
	defer cleanup()

	if err := ipt.init(); err != nil {
		t.Fatalf("init simulator input: %v", err)
	}

	ipt.collectResourceObject("host")
	ipt.collectResourceObject("vm")
	ipt.collectResourceObject("datastore")

	if len(ipt.collectObjects) == 0 {
		t.Fatal("collectResourceObject() collected no object points")
	}
}

func TestSimulatorCollectResourceEvent(t *testing.T) {
	ipt, cleanup := newSimulatorInput(t)
	defer cleanup()

	if err := ipt.init(); err != nil {
		t.Fatalf("init simulator input: %v", err)
	}

	ipt.collectResourceEvent("vm")

	if len(ipt.collectLogs) == 0 {
		t.Fatal("collectResourceEvent() collected no event points")
	}
}

func TestSimulatorCollectResourceMetrics(t *testing.T) {
	ipt, cleanup := newSimulatorInput(t)
	defer cleanup()

	if err := ipt.init(); err != nil {
		t.Fatalf("init simulator input: %v", err)
	}

	for _, resourceType := range []string{"host", "vm", "cluster", "datastore"} {
		ipt.collectResource(resourceType)
	}

	if len(ipt.collectCache) == 0 {
		t.Fatal("collectResource() collected no metric points")
	}

	ipt.collectCache = nil
	ipt.MaxQueryObjects = 1
	ipt.MaxQueryMetrics = 1
	ipt.client.resourceKinds["host"].lastColl = time.Now().Add(-45 * time.Second)
	ipt.collectResource("host")
	if len(ipt.collectCache) == 0 {
		t.Fatal("collectResource() with small query limits collected no metric points")
	}
}

func TestSimulatorCollect(t *testing.T) {
	ipt, cleanup := newSimulatorInput(t)
	defer cleanup()

	if err := ipt.init(); err != nil {
		t.Fatalf("init simulator input: %v", err)
	}

	if err := ipt.Collect(); err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
	if len(ipt.collectCache) == 0 {
		t.Fatal("Collect() collected no metric points")
	}
	if len(ipt.collectObjects) == 0 {
		t.Fatal("Collect() collected no object points")
	}
}

func TestSimulatorRunFeedsAndTerminates(t *testing.T) {
	ipt, cleanup := newSimulatorInput(t)
	defer cleanup()

	feeder := dkio.NewMockedFeeder()
	ipt.feeder = feeder
	ipt.Interval.Duration = time.Second

	done := make(chan struct{})
	go func() {
		ipt.Run()
		close(done)
	}()

	if pts, err := feeder.AnyPoints(10 * time.Second); err != nil {
		ipt.Terminate()
		t.Fatalf("Run() did not feed points: %v", err)
	} else if len(pts) == 0 {
		ipt.Terminate()
		t.Fatal("Run() fed empty point batch")
	}

	ipt.Terminate()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("Run() did not terminate")
	}
}

func TestSimulatorRunReportsFeedErrors(t *testing.T) {
	ipt, cleanup := newSimulatorInput(t)
	defer cleanup()

	feeder := &failingFeeder{feedCalled: make(chan struct{}, 3)}
	ipt.feeder = feeder
	ipt.Interval.Duration = time.Second

	done := make(chan struct{})
	go func() {
		ipt.Run()
		close(done)
	}()

	select {
	case <-feeder.feedCalled:
	case <-time.After(10 * time.Second):
		ipt.Terminate()
		t.Fatal("Run() did not call feeder")
	}

	ipt.Terminate()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("Run() did not terminate")
	}
	if feeder.lastErrors == 0 {
		t.Fatal("Run() did not report feed error")
	}
}

func TestSimulatorRunPausedTerminates(t *testing.T) {
	ipt, cleanup := newSimulatorInput(t)
	defer cleanup()

	ipt.feeder = dkio.NewMockedFeeder()
	ipt.Interval.Duration = time.Second
	if err := ipt.Pause(); err != nil {
		t.Fatalf("Pause() error = %v", err)
	}

	done := make(chan struct{})
	go func() {
		ipt.Run()
		close(done)
	}()

	time.Sleep(500 * time.Millisecond)
	ipt.Terminate()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("paused Run() did not terminate")
	}
}

func TestSimulatorStartDiscoveryPeriodic(t *testing.T) {
	ipt, cleanup := newSimulatorInput(t)
	defer cleanup()

	ipt.ObjectDiscoveryInterval.Duration = 10 * time.Millisecond
	if err := ipt.init(); err != nil {
		t.Fatalf("init simulator input: %v", err)
	}
	defer ipt.Terminate()

	time.Sleep(50 * time.Millisecond)
}

func TestSimulatorFinderFindAndExclude(t *testing.T) {
	ipt, cleanup := newSimulatorInput(t)
	defer cleanup()

	if err := ipt.init(); err != nil {
		t.Fatalf("init simulator input: %v", err)
	}

	finder := &Finder{client: ipt.client}
	var hosts []mo.HostSystem
	if err := finder.Find(context.Background(), "HostSystem", "/*/host/**", &hosts); err != nil {
		t.Fatalf("Find() error = %v", err)
	}
	if len(hosts) == 0 {
		t.Fatal("Find() returned no hosts")
	}

	var datacenters []mo.Datacenter
	if err := finder.Find(context.Background(), "Datacenter", "/*", &datacenters); err != nil {
		t.Fatalf("Find() datacenter error = %v", err)
	}
	if len(datacenters) == 0 {
		t.Fatal("Find() returned no datacenters")
	}

	var included []mo.HostSystem
	if err := finder.FindAll(context.Background(), "HostSystem", []string{"/*/host/**"}, []string{"/*/host/**"}, &included); err != nil {
		t.Fatalf("FindAll() error = %v", err)
	}
	if len(included) != 0 {
		t.Fatalf("FindAll() length = %d, want 0 after excluding same path", len(included))
	}
}

func TestSimulatorSimpleMetricDiscovery(t *testing.T) {
	ipt, cleanup := newSimulatorInput(t)
	defer cleanup()

	ipt.HostMetricInclude = []string{"cpu.usage.average", "missing.metric"}
	ipt.HostMetricExclude = nil

	if err := ipt.init(); err != nil {
		t.Fatalf("init simulator input: %v", err)
	}

	res := ipt.client.resourceKinds["host"]
	if !res.simple {
		t.Fatal("host resource simple = false, want true")
	}
	if len(res.metrics) == 0 {
		t.Fatal("simple metric discovery selected no metrics")
	}
	for _, metric := range res.metrics {
		if metric.Instance != "*" {
			t.Fatalf("metric instance = %q, want *", metric.Instance)
		}
	}

	iptNoInstances, cleanupNoInstances := newSimulatorInput(t)
	defer cleanupNoInstances()
	iptNoInstances.HostMetricInclude = []string{"cpu.usage.average"}
	iptNoInstances.HostMetricExclude = nil
	iptNoInstances.HostInstances = false

	if err := iptNoInstances.init(); err != nil {
		t.Fatalf("init simulator input without instances: %v", err)
	}
	res = iptNoInstances.client.resourceKinds["host"]
	if len(res.metrics) == 0 {
		t.Fatal("simple metric discovery without instances selected no metrics")
	}
	for _, metric := range res.metrics {
		if metric.Instance != "" {
			t.Fatalf("metric instance = %q, want empty", metric.Instance)
		}
	}
}

func TestSimulatorCreateVSphereClient2(t *testing.T) {
	ipt, cleanup := newSimulatorInput(t)
	defer cleanup()

	ipt.TLSClientConfig = &dknet.TLSClientConfig{}
	ipt.InsecureSkipVerify = true

	if err := ipt.init(); err != nil {
		t.Fatalf("init simulator input with createVSphereClient2: %v", err)
	}
	if ipt.client == nil || ipt.client.Client == nil || ipt.client.Perf == nil {
		t.Fatalf("client not fully initialized: %#v", ipt.client)
	}
}

func TestSimulatorCreateVSphereClientWithTLSConfig(t *testing.T) {
	ipt, cleanup := newSimulatorInput(t)
	defer cleanup()

	ipt.TLSClientConfig = &dknet.TLSClientConfig{InsecureSkipVerify: true}

	client, err := ipt.createVSphereClient(ipt.Vcenter)
	if err != nil {
		t.Fatalf("createVSphereClient() error = %v", err)
	}
	defer client.Client.Logout(context.Background()) //nolint:errcheck

	if client.Client == nil || client.Perf == nil || client.Timeout != ipt.timeout {
		t.Fatalf("client not fully initialized: %#v", client)
	}
}

func TestSimulatorTestClientReauthenticates(t *testing.T) {
	ipt, cleanup := newSimulatorInput(t)
	defer cleanup()

	if err := ipt.init(); err != nil {
		t.Fatalf("init simulator input: %v", err)
	}

	if err := ipt.client.Client.Logout(context.Background()); err != nil {
		t.Fatalf("logout simulator client: %v", err)
	}
	if err := ipt.testClient(context.Background()); err != nil {
		t.Fatalf("testClient() error = %v", err)
	}
}

type failingFeeder struct {
	feedCalled chan struct{}
	lastErrors int
}

func (f *failingFeeder) Feed(_ point.Category, _ []*point.Point, _ ...dkio.FeedOption) error {
	select {
	case f.feedCalled <- struct{}{}:
	default:
	}
	return dkio.ErrBusy
}

func (f *failingFeeder) FeedLastError(_ string, _ ...metrics.LastErrorOption) {
	f.lastErrors++
}
