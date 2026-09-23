// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package server

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	ws "gitlab.jiagouyun.com/cloudcare-tools/datakit/dca/websocket"
)

func newTestDB(t *testing.T) *DB {
	t.Helper()

	oldDBPath := dbPath
	dbPath = filepath.Join(t.TempDir(), "dca-test.db")
	t.Cleanup(func() {
		dbPath = oldDBPath
	})

	db := NewDB()
	require.NoError(t, db.Init())
	t.Cleanup(func() {
		if db.db != nil {
			_ = db.db.Close()
		}
	})

	return db
}

func newTestDataKit(connID string) *ws.DataKit {
	return &ws.DataKit{
		RunTimeID:      "runtime-" + connID,
		Arch:           "amd64",
		HostName:       "host-" + connID,
		OS:             "darwin",
		Version:        "1.0.0",
		IP:             "127.0.0.1",
		StartTime:      time.Now().Add(-time.Hour).UnixMilli(),
		RunInContainer: false,
		RunMode:        "host",
		UsageCores:     2,
		WorkspaceUUID:  "workspace-1",
		ConnID:         connID,
		Status:         ws.StatusRunning,
		URL:            "http://127.0.0.1:9529",
		DataKitRuntimeInfo: ws.DataKitRuntimeInfo{
			GlobalHostTags: map[string]string{
				"env":  "test",
				"role": "server",
			},
		},
		Config: `{"enabled":true}`,
	}
}

func TestDBLifecycle(t *testing.T) {
	db := newTestDB(t)
	dk := newTestDataKit("conn-1")

	require.NoError(t, db.Insert(nil))
	require.NoError(t, db.Insert(dk))

	found, err := db.Find(dk)
	require.NoError(t, err)
	require.NotNil(t, found)
	require.Equal(t, dk.ConnID, found.ConnID)
	require.Equal(t, ws.StatusRunning, found.Status)

	require.NoError(t, db.Heartbeat(dk.ConnID))
	require.NoError(t, db.UpdateStatus(dk, ws.StatusRestarting))
	require.Error(t, db.UpdateStatus(dk, ws.StatusUpgrading))
	require.NoError(t, db.UpdateStatus(dk, ws.StatusOffline))
	require.NoError(t, db.UpdateStatus(dk, ws.StatusRunning))

	dk.HostName = "host-updated"
	dk.DataKitRuntimeInfo.GlobalHostTags = map[string]string{"env": "prod"}
	require.NoError(t, db.Update(dk))
	found, err = db.Find(dk)
	require.NoError(t, err)
	require.Equal(t, "host-updated", found.HostName)

	tags := []struct {
		ConnID string `db:"conn_id"`
		Tags   string `db:"tags"`
	}{}
	require.NoError(t, db.Select("select conn_id,tags from global_host_tags where conn_id=?", &tags, dk.ConnID))
	require.Len(t, tags, 1)
	require.Contains(t, tags[0].Tags, "prod")

	require.NoError(t, db.Delete(dk, false))
	found, err = db.Find(dk)
	require.NoError(t, err)
	require.Equal(t, ws.StatusOffline, found.Status)

	require.NoError(t, db.Delete(dk, true))
	found, err = db.Find(dk)
	require.NoError(t, err)
	require.Nil(t, found)
}

func TestDBUpdateInsertForceUpdateAndDeleteByConnID(t *testing.T) {
	db := newTestDB(t)
	dk := newTestDataKit("conn-2")

	require.NoError(t, db.UpdateInsert(dk))
	found, err := db.Find(dk)
	require.NoError(t, err)
	require.NotNil(t, found)

	dk.Version = "2.0.0"
	require.NoError(t, db.UpdateInsert(dk))
	found, err = db.Find(dk)
	require.NoError(t, err)
	require.Equal(t, "2.0.0", found.Version)

	dk.Version = "3.0.0"
	require.NoError(t, db.ForceUpdate(dk))
	found, err = db.Find(dk)
	require.NoError(t, err)
	require.Equal(t, "3.0.0", found.Version)

	require.NoError(t, db.DeleteByConnID(dk.ConnID, false))
	found, err = db.Find(dk)
	require.NoError(t, err)
	require.Equal(t, ws.StatusOffline, found.Status)

	require.NoError(t, db.DeleteByConnID(dk.ConnID, true))
	found, err = db.Find(dk)
	require.NoError(t, err)
	require.Nil(t, found)
}

func TestDBNilAndErrorBranches(t *testing.T) {
	db := newTestDB(t)

	require.NoError(t, db.ForceUpdate(nil))
	require.NoError(t, db.UpdateGlobalHostTags(nil))
	require.NoError(t, db.DeleteGlobalHostTags(nil))
	require.NoError(t, db.UpdateStatus(nil, ws.StatusOffline))
	require.NoError(t, db.Delete(nil, true))

	_, err := db.Find(newTestDataKit("missing"))
	require.NoError(t, err)

	require.Error(t, db.UpdateStatus(newTestDataKit("missing"), ws.StatusOffline))
	require.Error(t, db.UpdateByConnID(newTestDataKit("missing"), "missing"))
}

func TestCheckStatus(t *testing.T) {
	require.True(t, checkStatus(ws.StatusRunning, ws.StatusUpgrading))
	require.True(t, checkStatus(ws.StatusUpgrading, ws.StatusOffline))
	require.True(t, checkStatus(ws.StatusUpgrading, ws.StatusRunning))
	require.False(t, checkStatus(ws.StatusUpgrading, ws.StatusRestarting))
}

// The same datakit process reconnects with a new conn id when its IP (or the
// websocket address / workspace) changes. The old row must be removed, otherwise
// the host is listed twice and reported as offline.
func TestDBForceUpdateSupersedesSameRuntime(t *testing.T) {
	db := newTestDB(t)

	old := newTestDataKit("conn-old")
	old.RunTimeID = "runtime-1"
	require.NoError(t, db.Insert(old))

	reconnected := newTestDataKit("conn-new")
	reconnected.RunTimeID = "runtime-1" // same datakit process, new conn id
	reconnected.IP = "10.20.30.99"
	require.NoError(t, db.ForceUpdate(reconnected))

	rows := []*ws.DataKit{}
	require.NoError(t, db.Select("select * from datakit", &rows))
	require.Len(t, rows, 1)
	require.Equal(t, "conn-new", rows[0].ConnID)
	require.Equal(t, "10.20.30.99", rows[0].IP)

	// a restarted datakit (new runtime id) keeps its own row
	restarted := newTestDataKit("conn-restart")
	restarted.RunTimeID = "runtime-2"
	require.NoError(t, db.ForceUpdate(restarted))

	rows = []*ws.DataKit{} // sqlx appends to the given slice
	require.NoError(t, db.Select("select * from datakit", &rows))
	require.Len(t, rows, 2)
}

// The periodic cleanup must not drop the row of a running container datakit: its
// session stays alive for a long time, so the row would not come back until the
// next reconnect. Container rows are removed at startup only.
func TestDBDeleteExpiredKeepsContainerRows(t *testing.T) {
	db := newTestDB(t)

	container := newTestDataKit("conn-container")
	container.RunInContainer = true
	require.NoError(t, db.Insert(container))

	require.NoError(t, db.DeleteExpired())
	found, err := db.Find(container)
	require.NoError(t, err)
	require.NotNil(t, found)

	require.NoError(t, db.DeleteContainerRows())
	found, err = db.Find(container)
	require.NoError(t, err)
	require.Nil(t, found)
}
