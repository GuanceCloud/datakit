// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package websocket

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func reviewClient(t *testing.T, url string, opts ...func(*Client)) *Client {
	t.Helper()
	options := []func(*Client){WithWebsocketAddress(url), WithDataKit(&DataKit{WorkspaceUUID: "workspace", ConnID: "conn"}), WithHeartbeatInterval(40 * time.Millisecond)}
	client, err := NewClient(append(options, opts...)...)
	require.NoError(t, err)
	t.Cleanup(func() { client.Stop(); require.NoError(t, client.g.Wait()) })
	return client
}

func TestReviewHealthyConnectionAfterLongBackoff(t *testing.T) {
	var attempts atomic.Int32
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if attempts.Add(1) <= 14 {
			http.Error(w, "unavailable", http.StatusServiceUnavailable)
			return
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close() //nolint:errcheck
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}))
	defer server.Close()
	client := reviewClient(t, "ws"+strings.TrimPrefix(server.URL, "http"))
	client.Start()
	require.Eventually(t, func() bool { return client.lastPong.Load() != 0 }, 40*time.Second, 10*time.Millisecond)
	require.Equal(t, int32(15), attempts.Load())
	require.Never(t, func() bool { return attempts.Load() > 15 }, 120*client.heartbeatInterval, 10*time.Millisecond,
		"successful reconnect must restore normal heartbeat timing")
}

func TestReviewHealthyConnectionDuringSlowAction(t *testing.T) {
	var attempts atomic.Int32
	connections := make(chan *websocket.Conn, 16)
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close() //nolint:errcheck
		connections <- conn
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}))
	defer server.Close()
	release := make(chan struct{})
	var once sync.Once
	finished := make(chan int64, 2)
	started := make(chan struct{})
	client := reviewClient(t, "ws"+strings.TrimPrefix(server.URL, "http"), WithActionHandlers(map[string]ActionHandler{
		"slow": func(_ *Client, id int64, _ any) error {
			if id == 1 {
				close(started)
				<-release
			}
			finished <- id
			return nil
		},
	}))
	// Unblock the in-flight action before waiting for the client to stop.
	defer once.Do(func() { close(release) })
	client.Start()
	require.Eventually(t, func() bool { return client.lastPong.Load() != 0 }, time.Second, 10*time.Millisecond)
	conn := <-connections
	require.NoError(t, conn.WriteJSON(WebsocketMessage{ID: 1, Action: "slow"}))
	require.NoError(t, conn.WriteJSON(WebsocketMessage{ID: 2, Action: "slow"}))
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("action did not start")
	}
	require.Never(t, func() bool { return attempts.Load() != 1 }, 8*client.heartbeatInterval, 10*time.Millisecond,
		"a slow action must not block pong processing")
	require.Empty(t, finished, "actions must execute serially")
	once.Do(func() { close(release) })
	for _, id := range []int64{1, 2} {
		select {
		case got := <-finished:
			require.Equal(t, id, got)
		case <-time.After(time.Second):
			t.Fatal("action did not finish")
		}
	}
}
