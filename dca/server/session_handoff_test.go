// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package server

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
	ws "gitlab.jiagouyun.com/cloudcare-tools/datakit/dca/websocket"
)

type handoffPeer struct {
	conn       *websocket.Conn
	responsive atomic.Bool
	probed     chan struct{}
}

func newHandoffServer(t *testing.T) (*DB, func(*ws.DataKit, bool) *handoffPeer) {
	t.Helper()
	db := withTestDatakitDB(t)
	replaceManagerState(t, make(chan *Client, 2048), make(chan *Client, 2048),
		map[string]*Client{}, map[string]chan *websocket.Conn{})
	startTestManager(t)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ws", websocketHandler)
	server := httptest.NewServer(router)
	var peers []*handoffPeer
	var reads sync.WaitGroup
	t.Cleanup(func() {
		for _, peer := range peers {
			if err := peer.conn.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
				t.Errorf("close test peer: %s", err)
			}
		}
		reads.Wait()
		server.Close() // includes any pending probe handlers
		require.Eventually(t, func() bool {
			Manager.RLock()
			defer Manager.RUnlock()
			return len(Manager.Clients) == 0
		}, 5*time.Second, 10*time.Millisecond)
	})
	dial := func(dk *ws.DataKit, responsive bool) *handoffPeer {
		t.Helper()
		header := http.Header{}
		header.Set(ws.HeaderDatakit, string(dk.Bytes()))
		conn, response, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http")+"/ws", header)
		require.NoError(t, err)
		require.NoError(t, response.Body.Close())
		peer := &handoffPeer{conn: conn, probed: make(chan struct{}, 1)}
		peer.responsive.Store(responsive)
		conn.SetPingHandler(func(data string) error {
			select {
			case peer.probed <- struct{}{}:
			default:
			}
			if !peer.responsive.Load() {
				// An unrelated pong must not validate this connection's fresh probe.
				data = "unrelated-pong"
			}
			return conn.WriteControl(websocket.PongMessage, []byte(data), time.Now().Add(time.Second))
		})
		peers = append(peers, peer)
		reads.Add(1)
		go func() {
			defer reads.Done()
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					return
				}
			}
		}()
		return peer
	}
	return db, dial
}

func awaitHandoffOwner(t *testing.T, id string, previous *Client) *Client {
	t.Helper()
	var found *Client
	require.Eventually(t, func() bool {
		owner, ok := Manager.getClient(id)
		if ok && owner != previous && !owner.exited.Load() {
			found = owner
			return true
		}
		return false
	}, 5*time.Second, 10*time.Millisecond)
	return found
}

func TestSessionHandoffReplacesSilentOwner(t *testing.T) {
	db, dial := newHandoffServer(t)
	dk := newTestDataKit("silent-handoff")
	oldPeer := dial(dk, false)
	old := awaitHandoffOwner(t, dk.ConnID, nil)
	newDK := *dk
	newDK.RunTimeID = "new-runtime-after-reconnect"
	newPeer := dial(&newDK, true)
	select {
	case <-oldPeer.probed:
	case <-time.After(time.Second):
		t.Fatal("the existing socket must be probed")
	}
	// Traffic from a pending socket cannot change the currently owned row.
	require.NoError(t, newPeer.conn.WriteMessage(websocket.TextMessage, (&ws.WebsocketMessage{
		Action: ws.DeleteDatakit, Data: ws.ActionData{},
	}).Bytes()))
	// The probe must not block registrations for another host.
	other := newTestDataKit("unrelated-during-probe")
	dial(other, true)
	require.Eventually(t, func() bool { return hasLiveSession(other.ConnID) }, time.Second, 10*time.Millisecond)
	row, err := db.Find(dk)
	require.NoError(t, err)
	require.NotNil(t, row)
	require.Equal(t, dk.RunTimeID, row.RunTimeID)
	require.False(t, old.exited.Load(), "old socket is still present when the new connection arrives")

	current := awaitHandoffOwner(t, dk.ConnID, old)
	require.NotSame(t, old, current)
	row, err = db.Find(dk)
	require.NoError(t, err)
	require.Equal(t, newDK.RunTimeID, row.RunTimeID)
	// Delayed frames/EOF from the retired owner must not mutate its successor.
	old.receiveMessage((&ws.WebsocketMessage{Action: ws.DeleteDatakit, Data: ws.ActionData{}}).Bytes())
	old.receiveMessage((&ws.WebsocketMessage{
		Action: ws.UpdateDatakitStatus, Data: ws.ActionData{Query: url.Values{"status": {"stopped"}}},
	}).Bytes())
	old.receiveMessage((&ws.WebsocketMessage{Action: ws.UpdateDatakit, Data: ws.ActionData{Body: string(dk.Bytes())}}).Bytes())
	require.False(t, Manager.unregisterClient(old))
	row, err = db.Find(dk)
	require.NoError(t, err)
	require.NotNil(t, row)
	require.Equal(t, ws.StatusRunning, row.Status)
	require.Equal(t, newDK.RunTimeID, row.RunTimeID)
	metric := &dto.Metric{}
	require.NoError(t, datakitTotalGauge.WithLabelValues(dk.HostName, dk.OS).Write(metric))
	require.Equal(t, float64(1), metric.GetGauge().GetValue(), "replacement must not double-count the connection")
	t.Log("silent old socket replaced after bounded probe; pending/retired mutations ignored")
}

func TestSessionHandoffRejectsStaleDecision(t *testing.T) {
	db := withTestDatakitDB(t)
	dk := newTestDataKit("stale-probe-decision")
	previous := newTestClientForRequest()
	current := newTestClientForRequest()
	current.ID, current.DataKit = dk.ConnID, dk
	incoming := newTestClientForRequest()
	incoming.ID, incoming.DataKit = dk.ConnID, dk
	incoming.replacementFor = previous
	manager := &ClientManager{Clients: map[string]*Client{dk.ConnID: current}}
	require.NoError(t, db.Insert(dk))
	require.False(t, manager.registerClient(incoming), "an old probe must not authorize replacing a different owner")
	require.Same(t, current, manager.Clients[dk.ConnID])

	previous.ID, previous.DataKit = dk.ConnID, &ws.DataKit{ConnID: dk.ConnID, RunInContainer: true}
	require.True(t, manager.unregisterClient(current))
	require.False(t, manager.unregisterClient(previous), "late container EOF must not delete the successor's offline host row")
	row, err := db.Find(dk)
	require.NoError(t, err)
	require.NotNil(t, row)
	require.Equal(t, ws.StatusOffline, row.Status)
	_, err = db.Exec("update datakit set updated_at=42 where conn_id=?", dk.ConnID)
	require.NoError(t, err)
	manager.heartbeatClient(previous)
	row, err = db.Find(dk)
	require.NoError(t, err)
	require.EqualValues(t, 42, row.UpdatedAt)
}

func TestSessionHandoffKeepsResponsiveOwner(t *testing.T) {
	db, dial := newHandoffServer(t)
	dk := newTestDataKit("healthy-handoff")
	oldPeer := dial(dk, true)
	old := awaitHandoffOwner(t, dk.ConnID, nil)
	duplicate := dial(dk, true)
	select {
	case <-oldPeer.probed:
	case <-time.After(time.Second):
		t.Fatal("healthy owner was not probed")
	}
	// A closed/rejected socket will no longer accept writes; keep the real owner.
	require.Eventually(t, func() bool {
		return duplicate.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(time.Second)) != nil
	}, 5*time.Second, 10*time.Millisecond)
	owner, _ := Manager.getClient(dk.ConnID)
	require.Same(t, old, owner)
	row, err := db.Find(dk)
	require.NoError(t, err)
	require.Equal(t, ws.StatusRunning, row.Status)
}

func TestSessionHandoffOwnerExitsDuringProbe(t *testing.T) {
	_, dial := newHandoffServer(t)
	dk := newTestDataKit("exit-during-probe")
	oldPeer := dial(dk, false)
	old := awaitHandoffOwner(t, dk.ConnID, nil)
	dial(dk, true)
	select {
	case <-oldPeer.probed:
	case <-time.After(time.Second):
		t.Fatal("owner was not probed")
	}
	require.NoError(t, oldPeer.conn.Close())
	require.Eventually(t, func() bool {
		current, ok := Manager.getClient(dk.ConnID)
		return ok && current != old
	}, time.Second, 10*time.Millisecond, "EOF should admit the waiting connection without another retry")
}

func TestSessionHandoffFleet(t *testing.T) {
	db, dial := newHandoffServer(t)
	const fleetSize = 46
	const disconnected = 27
	kits := make([]*ws.DataKit, fleetSize)
	owners := make([]*Client, fleetSize)
	for i := range kits {
		kits[i] = newTestDataKit(fmt.Sprintf("handoff-fleet-%02d", i))
		dial(kits[i], i >= disconnected)
		owners[i] = awaitHandoffOwner(t, kits[i].ConnID, nil)
	}
	// Each affected client reconnects once while the old socket remains open.
	started := time.Now()
	for i := 0; i < disconnected; i++ {
		dial(kits[i], true)
	}
	require.Eventually(t, func() bool {
		for i := 0; i < disconnected; i++ {
			owner, ok := Manager.getClient(kits[i].ConnID)
			if !ok || owner == owners[i] {
				return false
			}
		}
		return true
	}, 7*time.Second, 10*time.Millisecond, "probes must run concurrently, not 27 sequential waits")
	for _, old := range owners[:disconnected] {
		require.False(t, Manager.unregisterClient(old))
	}
	rows := []ws.DataKit{}
	require.NoError(t, db.Select("select * from datakit", &rows))
	require.Len(t, rows, fleetSize)
	for _, row := range rows {
		require.Equal(t, ws.StatusRunning, row.Status)
	}
	ctx, rec := newGinTestContext(http.MethodGet, "/api/datakit/list?pageIndex=1&pageSize=20", nil)
	addWorkspaceCookie(ctx)
	datakitListHandler(ctx)
	response := decodeDCAResponse(t, rec.Body.String())
	require.True(t, response.Success)
	page := response.Content.(map[string]interface{})["pageInfo"].(map[string]interface{})
	require.EqualValues(t, fleetSize, page["totalCount"])
	t.Logf("46 registered; 27 replacements without closing old sockets first; api_total=%v; recovery=%s", page["totalCount"], time.Since(started))
}
