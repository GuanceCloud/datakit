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

	duplicated, err := db.IsDuplicatedConn(dk)
	require.NoError(t, err)
	require.True(t, duplicated)

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

	duplicated, err = db.IsDuplicatedConn(dk)
	require.NoError(t, err)
	require.False(t, duplicated)

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

	duplicated, err := db.IsDuplicatedConn(nil)
	require.NoError(t, err)
	require.False(t, duplicated)

	_, err = db.Find(newTestDataKit("missing"))
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
