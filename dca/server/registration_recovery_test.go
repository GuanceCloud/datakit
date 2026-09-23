// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
	ws "gitlab.jiagouyun.com/cloudcare-tools/datakit/dca/websocket"
)

func TestRegistrationReclaimsRowWithoutSession(t *testing.T) {
	for _, status := range []ws.DataKitStatus{ws.StatusRunning, ws.StatusStopped, ws.StatusUpgrading, ws.StatusRestarting, ws.StatusOffline} {
		t.Run(status.String(), func(t *testing.T) {
			db := withTestDatakitDB(t)
			dk := newTestDataKit("orphaned-session")
			require.NoError(t, db.Insert(dk))
			require.NoError(t, db.UpdateStatus(dk, status))
			manager := &ClientManager{Clients: map[string]*Client{}}
			client := newTestClientForRequest()
			client.ID, client.DataKit = dk.ConnID, dk

			require.True(t, manager.registerClient(client), "a stored row must not reject the first reconnect when no session owns it")
			require.Same(t, client, manager.Clients[dk.ConnID])
			found, err := db.Find(dk)
			require.NoError(t, err)
			require.Equal(t, ws.StatusRunning, found.Status)
		})
	}
}

func TestRegistrationKeepsOwnerWhenStoredStatusIsOffline(t *testing.T) {
	db := withTestDatakitDB(t)
	dk := newTestDataKit("owned-but-offline")
	require.NoError(t, db.Insert(dk))
	require.NoError(t, db.UpdateStatus(dk, ws.StatusOffline))
	owner := newTestClientForRequest()
	owner.ID, owner.DataKit = dk.ConnID, dk
	manager := &ClientManager{Clients: map[string]*Client{dk.ConnID: owner}}
	duplicate := newTestClientForRequest()
	duplicate.ID, duplicate.DataKit = dk.ConnID, dk

	require.False(t, manager.registerClient(duplicate), "database status must not allow a duplicate to replace the owning socket")
	require.Same(t, owner, manager.Clients[dk.ConnID])
	require.False(t, manager.unregisterClient(duplicate))
	require.Same(t, owner, manager.Clients[dk.ConnID])
}

func TestKeepaliveRepliesBeforeDatabaseWrite(t *testing.T) {
	db := withTestDatakitDB(t)
	replaceManagerState(t, make(chan *Client, 10), make(chan *Client, 10),
		map[string]*Client{}, map[string]chan *websocket.Conn{})
	startTestManager(t)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ws", websocketHandler)
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)
	dk := newTestDataKit("busy-heartbeat-db")
	header := http.Header{}
	header.Set(ws.HeaderDatakit, string(dk.Bytes()))
	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http")+"/ws", header)
	require.NoError(t, err)
	t.Cleanup(func() {
		conn.Close() //nolint:errcheck,gosec
		require.Eventually(t, func() bool { return !hasLiveSession(dk.ConnID) }, time.Second, 10*time.Millisecond)
	})
	require.Eventually(t, func() bool { return hasLiveSession(dk.ConnID) }, time.Second, 10*time.Millisecond)

	// Occupy the only database connection, as a slow write or query would.
	// The websocket peer must still receive a pong before that work completes.
	blocked, err := db.db.Conn(context.Background())
	require.NoError(t, err)
	defer blocked.Close() //nolint:errcheck
	pong := make(chan string, 1)
	conn.SetPongHandler(func(data string) error { pong <- data; return nil })
	readDone := make(chan struct{})
	go func() {
		defer close(readDone)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()
	t.Cleanup(func() {
		conn.Close() //nolint:errcheck,gosec
		select {
		case <-readDone:
		case <-time.After(time.Second):
			t.Error("websocket reader did not exit")
		}
	})
	require.NoError(t, conn.WriteControl(websocket.PingMessage, []byte("busy-db"), time.Now().Add(time.Second)))
	select {
	case data := <-pong:
		require.Equal(t, "busy-db", data)
	case <-time.After(time.Second):
		t.Fatal("database contention delayed the websocket pong")
	}
}
