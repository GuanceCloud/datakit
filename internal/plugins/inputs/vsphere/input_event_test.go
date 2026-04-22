// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package vsphere

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/vmware/govmomi/vim25/types"
)

func TestMakeEventPointEventExOverridesDefaults(t *testing.T) {
	created := time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC)
	event := &types.EventEx{
		Event: types.Event{
			Key:                  101,
			ChainId:              202,
			CreatedTime:          created,
			UserName:             "system",
			FullFormattedMessage: "hardware status changed",
			ChangeTag:            "change-tag",
		},
		EventTypeId: "com.vmware.hardware.warning",
		ObjectName:  "event-object",
		Severity:    Warning,
	}

	pt := makeEventPoint(
		event,
		"cache-object",
		map[string]string{
			"host":       "vcenter.local",
			ResourceType: "host",
		},
		map[string]string{
			"env": "prod",
		},
		false,
	)

	if got := pt.GetTag(Status); got != Warning {
		t.Fatalf("status = %q, want %q", got, Warning)
	}
	if got := pt.GetTag(EventTypeID); got != "com.vmware.hardware.warning" {
		t.Fatalf("event_type_id = %q, want EventEx event type id", got)
	}
	if got := pt.GetTag(ObjectName); got != "event-object" {
		t.Fatalf("object_name = %q, want EventEx object name", got)
	}
	if got := pt.GetTag(UserName); got != "system" {
		t.Fatalf("user_name = %q, want system", got)
	}
	if got := pt.GetTag("env"); got != "prod" {
		t.Fatalf("env = %q, want prod", got)
	}
}

func TestMakeEventPointBaseEventUsesDefaults(t *testing.T) {
	event := &types.Event{
		UserName:    "admin",
		CreatedTime: time.Date(2026, 4, 17, 11, 0, 0, 0, time.UTC),
	}

	pt := makeEventPoint(
		event,
		"cache-object",
		map[string]string{
			"host":       "vcenter.local",
			ResourceType: "vm",
		},
		nil,
		false,
	)

	if got := pt.GetTag(Status); got != Info {
		t.Fatalf("status = %q, want %q", got, Info)
	}
	if got := pt.GetTag(EventTypeID); got != "*types.Event" {
		t.Fatalf("event_type_id = %q, want *types.Event", got)
	}
	if got := pt.GetTag(ObjectName); got != "cache-object" {
		t.Fatalf("object_name = %q, want cache-object", got)
	}
}

func TestMakeEventPointEventExErrorSeverity(t *testing.T) {
	event := &types.EventEx{
		Event: types.Event{
			CreatedTime: time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC),
		},
		Severity: Error,
	}

	pt := makeEventPoint(event, "cache-object", map[string]string{}, nil, false)

	if got := pt.GetTag(Status); got != Error {
		t.Fatalf("status = %q, want %q", got, Error)
	}
}

func TestCollectResourceEventInitialLookbackAdvancesCursorWithoutEvents(t *testing.T) {
	base := time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC)
	vmRef := types.ManagedObjectReference{Type: "VirtualMachine", Value: "vm-1"}
	obj := &objectRef{name: "vm-1", ref: vmRef}
	ipt := newEventTestInput(base, obj)

	orig := queryEvents
	defer func() {
		queryEvents = orig
	}()

	var gotFilter types.EventFilterSpec
	queryEvents = func(_ context.Context, _ *Client, filter types.EventFilterSpec) ([]types.BaseEvent, error) {
		gotFilter = filter
		return nil, nil
	}

	ipt.collectResourceEvent("vm")

	wantBegin := base.Add(-eventInitLookback)
	if gotFilter.Time == nil || gotFilter.Time.BeginTime == nil || !gotFilter.Time.BeginTime.Equal(wantBegin) {
		t.Fatalf("begin time = %v, want %v", gotFilter.Time, wantBegin)
	}
	if gotFilter.Entity == nil || gotFilter.Entity.Entity.Value != vmRef.Value {
		t.Fatalf("entity filter = %#v, want %s", gotFilter.Entity, vmRef.Value)
	}
	if gotFilter.Entity.Recursion != types.EventFilterSpecRecursionOptionAll {
		t.Fatalf("recursion = %q, want %q", gotFilter.Entity.Recursion, types.EventFilterSpecRecursionOptionAll)
	}
	if obj.lastLogTime == nil {
		t.Fatal("lastLogTime is nil, want initialized cursor")
	}
	gotCursor := obj.lastLogTime["vm"]
	if gotCursor.Before(wantBegin) {
		t.Fatalf("cursor = %s, want not before %s", gotCursor, wantBegin)
	}
	if len(ipt.collectLogs) != 0 {
		t.Fatalf("collectLogs length = %d, want 0", len(ipt.collectLogs))
	}
}

func TestCollectResourceEventUsesExistingCursorAndCollectsLogs(t *testing.T) {
	base := time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC)
	cursor := base.Add(-30 * time.Minute)
	created := time.Now().Add(time.Hour).UTC()
	vmRef := types.ManagedObjectReference{Type: "VirtualMachine", Value: "vm-1"}
	obj := &objectRef{
		name:        "vm-1",
		ref:         vmRef,
		lastLogTime: map[string]time.Time{"vm": cursor},
	}
	ipt := newEventTestInput(base, obj)
	ipt.Tags = map[string]string{"env": "test"}

	orig := queryEvents
	defer func() {
		queryEvents = orig
	}()

	var gotBegin time.Time
	queryEvents = func(_ context.Context, _ *Client, filter types.EventFilterSpec) ([]types.BaseEvent, error) {
		if filter.Time == nil || filter.Time.BeginTime == nil {
			t.Fatal("QueryEvents() filter has no begin time")
		}
		gotBegin = *filter.Time.BeginTime
		return []types.BaseEvent{
			&types.Event{
				CreatedTime:          created,
				UserName:             "admin",
				FullFormattedMessage: "created vm",
			},
		}, nil
	}

	ipt.collectResourceEvent("vm")

	if !gotBegin.Equal(cursor) {
		t.Fatalf("begin time = %s, want existing cursor %s", gotBegin, cursor)
	}
	if got := obj.lastLogTime["vm"]; !got.Equal(created) {
		t.Fatalf("cursor = %s, want event created time %s", got, created)
	}
	if len(ipt.collectLogs) != 1 {
		t.Fatalf("collectLogs length = %d, want 1", len(ipt.collectLogs))
	}
	pt := ipt.collectLogs[0]
	if got := pt.GetTag("host"); got != "vcenter.local" {
		t.Fatalf("host = %q, want vcenter.local", got)
	}
	if got := pt.GetTag(ResourceType); got != "vm" {
		t.Fatalf("resource_type = %q, want vm", got)
	}
	if got := pt.GetTag("vm_name"); got != "vm-1" {
		t.Fatalf("vm_name = %q, want vm-1", got)
	}
	if got := pt.GetTag("env"); got != "test" {
		t.Fatalf("env = %q, want test", got)
	}
}

func newEventTestInput(base time.Time, obj *objectRef) *Input {
	return &Input{
		client: &Client{
			Timeout: time.Second,
			resourceKinds: map[string]*resourceKind{
				"vm": {
					name: "vm",
					pKey: "vm_name",
					objects: objectMap{
						obj.ref.Value: obj,
					},
				},
			},
		},
		vcenter: &url.URL{Host: "vcenter.local"},
		ptsTime: base,
	}
}
