// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package vsphere

import (
	"context"
	"testing"

	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

func TestMatchName(t *testing.T) {
	props := []types.DynamicProperty{
		{Name: "name", Val: "esx-1"},
	}

	if !matchName(property.Match{"name": "esx-1"}, props) {
		t.Fatal("matchName() = false, want true for exact match")
	}
	if !matchName(property.Match{"name": "*"}, props) {
		t.Fatal("matchName() = false, want true for wildcard")
	}
	if matchName(property.Match{"name": "esx-2"}, props) {
		t.Fatal("matchName() = true, want false for different name")
	}
	if matchName(property.Match{"name": "esx-1"}, nil) {
		t.Fatal("matchName() = true, want false when name property is absent")
	}
}

func TestObjectContentToTypedArrayConcreteType(t *testing.T) {
	objs := map[string]types.ObjectContent{
		"host-1": {
			Obj: types.ManagedObjectReference{Type: "HostSystem", Value: "host-1"},
			PropSet: []types.DynamicProperty{
				{Name: "name", Val: "esx-1"},
			},
		},
	}

	var hosts []mo.HostSystem
	if err := objectContentToTypedArray(objs, &hosts); err != nil {
		t.Fatalf("objectContentToTypedArray() error = %v", err)
	}

	if len(hosts) != 1 {
		t.Fatalf("len(hosts) = %d, want 1", len(hosts))
	}
	if hosts[0].Name != "esx-1" || hosts[0].Reference().Value != "host-1" {
		t.Fatalf("unexpected host: %#v", hosts[0])
	}
}

func TestObjectContentToTypedArrayEmbeddedType(t *testing.T) {
	objs := map[string]types.ObjectContent{
		"host-1": {
			Obj: types.ManagedObjectReference{Type: "HostSystem", Value: "host-1"},
			PropSet: []types.DynamicProperty{
				{Name: "name", Val: "esx-1"},
			},
		},
	}

	var entities []mo.ManagedEntity
	if err := objectContentToTypedArray(objs, &entities); err != nil {
		t.Fatalf("objectContentToTypedArray() error = %v", err)
	}

	if len(entities) != 1 {
		t.Fatalf("len(entities) = %d, want 1", len(entities))
	}
	if entities[0].Name != "esx-1" || entities[0].Reference().Value != "host-1" {
		t.Fatalf("unexpected entity: %#v", entities[0])
	}
}

func TestObjectContentToTypedArrayPanicsOnNonPointer(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("objectContentToTypedArray() did not panic for non-pointer dst")
		}
	}()

	var hosts []mo.HostSystem
	_ = objectContentToTypedArray(nil, hosts)
}

func TestObjectContentToTypedArrayPanicsOnUnknownType(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("objectContentToTypedArray() did not panic for unknown type")
		}
	}()

	objs := map[string]types.ObjectContent{
		"unknown-1": {
			Obj: types.ManagedObjectReference{Type: "UnknownType", Value: "unknown-1"},
		},
	}

	var hosts []mo.HostSystem
	_ = objectContentToTypedArray(objs, &hosts)
}

func TestFinderDescendUnknownRootTypeReturnsNil(t *testing.T) {
	finder := &Finder{}
	objs := map[string]types.ObjectContent{}
	root := types.ManagedObjectReference{Type: "UnknownType", Value: "unknown-1"}

	if err := finder.descend(context.Background(), root, "VirtualMachine", []property.Match{{"name": "*"}}, 0, objs); err != nil {
		t.Fatalf("descend() error = %v, want nil", err)
	}
	if len(objs) != 0 {
		t.Fatalf("descend() objects length = %d, want 0", len(objs))
	}
}

func TestFinderFindAllWithNoPathsReturnsEmpty(t *testing.T) {
	finder := &Finder{}
	var hosts []mo.HostSystem

	if err := finder.FindAll(context.Background(), "HostSystem", nil, nil, &hosts); err != nil {
		t.Fatalf("FindAll() error = %v, want nil", err)
	}
	if len(hosts) != 0 {
		t.Fatalf("FindAll() hosts length = %d, want 0", len(hosts))
	}
}
