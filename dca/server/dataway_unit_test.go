// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package server

import (
	"io"
	"net/http"
	"net/http/httptest"

	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	ws "gitlab.jiagouyun.com/cloudcare-tools/datakit/dca/websocket"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
)

func TestDoUploadHostStatusSkipsInvalidInputs(t *testing.T) {
	db := withTestDatakitDB(t)

	doUploadHostStatus(nil, 0)

	badJSON := newTestDataKit("upload-bad-json")
	badJSON.Config = "{bad-json"
	require.NoError(t, db.Insert(badJSON))

	emptyConfig := newTestDataKit("upload-empty-config")
	emptyConfig.Config = "null"
	require.NoError(t, db.Insert(emptyConfig))

	emptyDataway := newTestDataKit("upload-empty-dataway")
	emptyDataway.Config = `{"dataway_url":""}`
	require.NoError(t, db.Insert(emptyDataway))

	invalidURL := newTestDataKit("upload-invalid-url")
	invalidURL.Config = `{"dataway_url":"://bad"}`
	invalidURL.Status = ws.StatusOffline
	require.NoError(t, db.Insert(invalidURL))

	doUploadHostStatus(&dataway.DialtestingSender{}, 0)
}

func TestDoUploadHostStatusWritesMetric(t *testing.T) {
	db := withTestDatakitDB(t)
	requests := make(chan string, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests <- r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	dk := newTestDataKit("upload-ok")
	dk.Config = `{"dataway_url":"` + server.URL + `/v1/write"}`
	dk.Status = ws.StatusRunning
	require.NoError(t, db.Insert(dk))

	sender := &dataway.DialtestingSender{}
	require.NoError(t, sender.Init(&dataway.DialtestingSenderOpt{HTTPTimeout: time.Second}))
	doUploadHostStatus(sender, 0)

	select {
	case path := <-requests:
		require.Contains(t, path, "metric")
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for dataway metric upload")
	}
}

func TestUploadHostStatusCanStopBeforeTick(t *testing.T) {
	closeCh := make(chan struct{})
	close(closeCh)
	UploadHostStatus(time.Hour, closeCh)
}

// The availability metric is tagged with the host name only, so every host must
// be reported once, using its newest row: the stale rows of an older conn id
// must not report a healthy host as offline.
func TestCurrentRowsKeepsNewestRowPerHost(t *testing.T) {
	stale := newTestDataKit("conn-stale")
	stale.UpdatedAt = 1000
	stale.Status = ws.StatusOffline

	fresh := newTestDataKit("conn-fresh")
	fresh.HostName = stale.HostName
	fresh.UpdatedAt = 2000
	fresh.Status = ws.StatusRunning

	otherHost := newTestDataKit("conn-other")
	otherHost.UpdatedAt = 1500

	rows := currentRows([]*ws.DataKit{stale, fresh, otherHost, nil})
	require.Len(t, rows, 2)

	byHost := map[string]*ws.DataKit{}
	for _, row := range rows {
		byHost[row.HostName] = row
	}

	require.Equal(t, "conn-fresh", byHost[stale.HostName].ConnID)
	require.Equal(t, "conn-other", byHost[otherHost.HostName].ConnID)
}

// With an identical timestamp the running row wins, so an offline row cannot
// hide a live one.
func TestCurrentRowsPrefersRunningOnSameTimestamp(t *testing.T) {
	offline := newTestDataKit("conn-offline")
	offline.UpdatedAt = 1000
	offline.Status = ws.StatusOffline

	running := newTestDataKit("conn-running")
	running.HostName = offline.HostName
	running.UpdatedAt = 1000
	running.Status = ws.StatusRunning

	rows := currentRows([]*ws.DataKit{offline, running})
	require.Len(t, rows, 1)
	require.Equal(t, "conn-running", rows[0].ConnID)
}

// Two workspaces can run a datakit on hosts with the same name; they upload
// through different dataways, so both must be reported.
func TestCurrentRowsKeepsSameHostOfOtherWorkspace(t *testing.T) {
	first := newTestDataKit("conn-ws1")
	first.WorkspaceUUID = "workspace-1"
	first.UpdatedAt = 2000

	second := newTestDataKit("conn-ws2")
	second.HostName = first.HostName
	second.WorkspaceUUID = "workspace-2"
	second.UpdatedAt = 1000 // older, but a different workspace

	rows := currentRows([]*ws.DataKit{first, second})
	require.Len(t, rows, 2)
}

// A half open connection must not keep a host online forever: the row is only
// fresh while the datakit keeps reporting.
func TestRowIsFresh(t *testing.T) {
	fresh := newTestDataKit("conn-fresh")
	fresh.UpdatedAt = time.Now().UnixMilli()
	require.True(t, rowIsFresh(fresh, time.Now(), time.Minute))

	stale := newTestDataKit("conn-stale")
	stale.UpdatedAt = time.Now().Add(-time.Hour).UnixMilli()
	require.False(t, rowIsFresh(stale, time.Now(), time.Minute))

	never := newTestDataKit("conn-never") // UpdatedAt is not reported (also nil safety)
	require.False(t, rowIsFresh(never, time.Now(), time.Minute))
	require.False(t, rowIsFresh(nil, time.Now(), time.Minute))

	// a non-positive threshold disables the freshness check
	require.True(t, rowIsFresh(stale, time.Now(), 0))
}

func TestUploadHostStatusUsesSnapshotTime(t *testing.T) {
	db := withTestDatakitDB(t)
	const count = 3
	var uploaded atomic.Int32
	bodies := make(chan string, count)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read metric: %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		bodies <- string(body)
		if uploaded.Add(1) == 1 {
			// Let the snapshot age past the threshold before processing other hosts.
			time.Sleep(1200 * time.Millisecond)
			// All hosts continued to heartbeat while the first upload was delayed.
			if _, err := db.Exec("update datakit set updated_at=?", time.Now().UnixMilli()); err != nil {
				t.Errorf("refresh heartbeat: %s", err)
			}
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	for _, id := range []string{"slow-first", "healthy-second", "healthy-third"} {
		dk := newTestDataKit(id)
		dk.Config = `{"dataway_url":"` + server.URL + `/v1/write"}`
		require.NoError(t, db.Insert(dk))
	}
	sender := &dataway.DialtestingSender{}
	require.NoError(t, sender.Init(&dataway.DialtestingSenderOpt{HTTPTimeout: 3 * time.Second}))
	_, err := db.Exec("update datakit set updated_at=?", time.Now().UnixMilli())
	require.NoError(t, err)
	doUploadHostStatus(sender, time.Second)
	require.EqualValues(t, count, uploaded.Load())
	for i := 0; i < count; i++ {
		body := <-bodies
		require.Contains(t, body, "status=online", "healthy snapshot became offline during upload")
	}
}
