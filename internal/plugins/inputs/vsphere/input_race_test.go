// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package vsphere

import (
	"net/url"
	"sync"
	"testing"

	"github.com/vmware/govmomi/vim25/types"
)

func TestResourceObjectCacheConcurrentAccess(t *testing.T) {
	client := &Client{
		resourceKinds: map[string]*resourceKind{},
	}

	clusterRefA := types.ManagedObjectReference{Type: "ClusterComputeResource", Value: "domain-c1"}
	clusterRefB := types.ManagedObjectReference{Type: "ClusterComputeResource", Value: "domain-c2"}
	hostRef := types.ManagedObjectReference{Type: "HostSystem", Value: "host-1"}

	client.resourceKinds["cluster"] = &resourceKind{
		name:    "cluster",
		objects: objectMap{},
	}
	hostKind := &resourceKind{
		name:      "host",
		pKey:      "esx_hostname",
		parent:    "cluster",
		parentTag: "cluster_name",
		objects:   objectMap{},
	}
	client.resourceKinds["host"] = hostKind
	ipt := &Input{
		client:  client,
		vcenter: &url.URL{Host: "vcenter.local"},
	}

	start := make(chan struct{})
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 10000; i++ {
			parentRef := clusterRefA
			clusterRef := clusterRefA
			clusterName := "cluster-a"
			if i%2 == 1 {
				parentRef = clusterRefB
				clusterRef = clusterRefB
				clusterName = "cluster-b"
			}

			client.collectMux.Lock()
			client.resourceKinds["cluster"].objects = objectMap{
				clusterRef.Value: {
					name: clusterName,
					ref:  clusterRef,
				},
			}
			hostKind.objects = objectMap{
				hostRef.Value: {
					name:         "esx-1",
					ref:          hostRef,
					parentRef:    &parentRef,
					objectTags:   map[string]string{Name: "esx-1"},
					objectFields: map[string]interface{}{"cpu": int64(1)},
				},
			}
			client.collectMux.Unlock()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 10000; i++ {
			ipt.collectResourceObject("host")
		}
	}()

	close(start)
	wg.Wait()
}

func TestCollectResourceObjectDoesNotMutateCachedTags(t *testing.T) {
	clusterRef := types.ManagedObjectReference{Type: "ClusterComputeResource", Value: "domain-c1"}
	hostRef := types.ManagedObjectReference{Type: "HostSystem", Value: "host-1"}
	cachedTags := map[string]string{Name: "esx-1"}
	cachedFields := map[string]interface{}{"cpu": int64(1)}

	client := &Client{
		resourceKinds: map[string]*resourceKind{
			"cluster": {
				name: "cluster",
				objects: objectMap{
					clusterRef.Value: {
						name: "cluster-a",
						ref:  clusterRef,
					},
				},
			},
			"host": {
				name:      "host",
				pKey:      "esx_hostname",
				parent:    "cluster",
				parentTag: "cluster_name",
				objects: objectMap{
					hostRef.Value: {
						name:         "esx-1",
						ref:          hostRef,
						parentRef:    &clusterRef,
						objectTags:   cachedTags,
						objectFields: cachedFields,
					},
				},
			},
		},
	}
	ipt := &Input{
		client:  client,
		vcenter: &url.URL{Host: "vcenter.local"},
	}

	ipt.collectResourceObject("host")

	if len(cachedTags) != 1 || cachedTags[Name] != "esx-1" {
		t.Fatalf("cached tags mutated: %#v", cachedTags)
	}
	if _, ok := cachedTags["host"]; ok {
		t.Fatalf("cached tags contain generated host tag: %#v", cachedTags)
	}
	if _, ok := cachedTags["cluster_name"]; ok {
		t.Fatalf("cached tags contain generated parent tag: %#v", cachedTags)
	}
	if _, ok := cachedTags["esx_hostname"]; ok {
		t.Fatalf("cached tags contain generated pkey tag: %#v", cachedTags)
	}
	if len(cachedFields) != 1 || cachedFields["cpu"] != int64(1) {
		t.Fatalf("cached fields mutated: %#v", cachedFields)
	}
	if len(ipt.collectObjects) != 1 {
		t.Fatalf("collectObjects length = %d, want 1", len(ipt.collectObjects))
	}
}

func TestCollectResourceObjectAllowsNilCachedTags(t *testing.T) {
	hostRef := types.ManagedObjectReference{Type: "HostSystem", Value: "host-1"}
	client := &Client{
		resourceKinds: map[string]*resourceKind{
			"host": {
				name: "host",
				pKey: "esx_hostname",
				objects: objectMap{
					hostRef.Value: {
						name:         "esx-1",
						ref:          hostRef,
						objectFields: map[string]interface{}{"cpu": int64(1)},
					},
				},
			},
		},
	}
	ipt := &Input{
		client:  client,
		vcenter: &url.URL{Host: "vcenter.local"},
	}

	ipt.collectResourceObject("host")

	if len(ipt.collectObjects) != 1 {
		t.Fatalf("collectObjects length = %d, want 1", len(ipt.collectObjects))
	}
}

func TestCollectResourceObjectSkipsEmptyCachedObject(t *testing.T) {
	hostRef := types.ManagedObjectReference{Type: "HostSystem", Value: "host-1"}
	client := &Client{
		resourceKinds: map[string]*resourceKind{
			"host": {
				name: "host",
				pKey: "esx_hostname",
				objects: objectMap{
					hostRef.Value: {
						name: "esx-1",
						ref:  hostRef,
					},
				},
			},
		},
	}
	ipt := &Input{
		client:  client,
		vcenter: &url.URL{Host: "vcenter.local"},
	}

	ipt.collectResourceObject("host")

	if len(ipt.collectObjects) != 0 {
		t.Fatalf("collectObjects length = %d, want 0", len(ipt.collectObjects))
	}
}

func TestCloneStringMap(t *testing.T) {
	if got := cloneStringMap(nil); got == nil || len(got) != 0 {
		t.Fatalf("cloneStringMap(nil) = %#v, want empty non-nil map", got)
	}

	src := map[string]string{"a": "1"}
	got := cloneStringMap(src)
	got["a"] = "2"
	got["b"] = "3"

	if src["a"] != "1" {
		t.Fatalf("source map mutated: %#v", src)
	}
	if _, ok := src["b"]; ok {
		t.Fatalf("source map received new key: %#v", src)
	}
}

func TestCloneInterfaceMap(t *testing.T) {
	if got := cloneInterfaceMap(nil); got != nil {
		t.Fatalf("cloneInterfaceMap(nil) = %#v, want nil", got)
	}

	src := map[string]interface{}{"a": int64(1)}
	got := cloneInterfaceMap(src)
	got["a"] = int64(2)
	got["b"] = int64(3)

	if src["a"] != int64(1) {
		t.Fatalf("source map mutated: %#v", src)
	}
	if _, ok := src["b"]; ok {
		t.Fatalf("source map received new key: %#v", src)
	}
}
