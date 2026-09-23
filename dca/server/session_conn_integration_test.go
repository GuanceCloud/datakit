// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package server

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
	ws "gitlab.jiagouyun.com/cloudcare-tools/datakit/dca/websocket"
)

// A datakit reports its info over a live websocket session, and its conn id may
// change while that session is alive (IP, websocket address or workspace
// changed). The DB row belongs to the session, so the bookkeeping must stick to
// the conn id of the session. Otherwise a row is left in "running" without any
// live session: the web UI enables 管理 on it, and every request fails with
// "datakit not available".
func TestSessionKeepsConnIDBookkeeping(t *testing.T) {
	db := withTestDatakitDB(t)
	replaceManagerState(t,
		make(chan *Client, 10),
		make(chan *Client, 10),
		map[string]*Client{},
		map[string]chan *websocket.Conn{},
	)

	startTestManager(t)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ws", websocketHandler)
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	// 1) the datakit registers with conn id A (host mode: this is what dk_upgrader does)
	dkSession := newTestDataKit("session-a")
	clientA := newTestWebsocketClient(t, wsURL, dkSession)
	clientA.Start()
	t.Cleanup(clientA.Stop)
	require.Eventually(t, func() bool {
		row, err := db.Find(dkSession)
		return err == nil && row != nil && row.Status == ws.StatusRunning
	}, time.Second, 10*time.Millisecond)

	// 2) the same datakit reports a new conn id over the live session
	dkReported := newTestDataKit("session-b")
	dkReported.IP = "10.20.30.99"
	require.NoError(t, clientA.SendMessage(&ws.WebsocketMessage{
		Action: ws.UpdateDatakit,
		Data:   ws.ActionData{Body: string(dkReported.Bytes())},
	}))

	// 3) it re-registers with conn id B
	clientB := newTestWebsocketClient(t, wsURL, dkReported)
	clientB.Start()
	t.Cleanup(clientB.Stop)
	require.Eventually(t, func() bool {
		row, err := db.Find(dkReported)
		return err == nil && row != nil && row.Status == ws.StatusRunning
	}, time.Second, 10*time.Millisecond)

	// 4) the old session ends
	clientA.Stop()
	require.Eventually(t, func() bool {
		row, err := db.Find(dkSession)
		return err == nil && row != nil && row.Status == ws.StatusOffline
	}, time.Second, 10*time.Millisecond)

	// the reported data landed on the row of the session (same key, fresh data)
	rowSession, err := db.Find(dkSession)
	require.NoError(t, err)
	require.NotNil(t, rowSession)
	require.Equal(t, ws.StatusOffline, rowSession.Status)
	require.Equal(t, dkReported.IP, rowSession.IP)

	// the re-registration kept its own row, still backed by a live session
	rowReported, err := db.Find(dkReported)
	require.NoError(t, err)
	require.NotNil(t, rowReported)
	require.Equal(t, ws.StatusRunning, rowReported.Status)

	requireNoRunningRowWithoutSession(t, db)

	// Stop the remaining session inside the test: the manager processes the
	// unregistration asynchronously, and a late event must not hit the DB after
	// the test closed it.
	clientB.Stop()
	require.Eventually(t, func() bool { return !hasLiveSession(dkReported.ConnID) },
		time.Second, 10*time.Millisecond)
}

func newTestWebsocketClient(t *testing.T, url string, dk *ws.DataKit) *ws.Client {
	t.Helper()

	c, err := ws.NewClient(
		ws.WithWebsocketAddress(url),
		ws.WithDataKit(dk),
		ws.WithHeartbeatInterval(time.Minute),
	)
	require.NoError(t, err)

	return c
}

// requireNoRunningRowWithoutSession is the invariant the web UI relies on: the
// management actions are enabled for rows in "running" state only.
func requireNoRunningRowWithoutSession(t *testing.T, db *DB) {
	t.Helper()

	rows := []*ws.DataKit{}
	require.NoError(t, db.Select("select * from datakit", &rows))

	for _, row := range rows {
		if row.Status != ws.StatusRunning {
			continue
		}

		require.True(t, hasLiveSession(row.ConnID),
			"row %s(conn_id=%s) is running without a live session",
			row.HostName, row.ConnID)
	}
}

// A datakit that reconnects with a new conn id (its IP changed) must not leave
// the previous row behind, and the ending of the old session must not touch the
// row of the live session.
func TestSessionReRegistrationSupersedesOldRow(t *testing.T) {
	db := withTestDatakitDB(t)
	replaceManagerState(t,
		make(chan *Client, 10),
		make(chan *Client, 10),
		map[string]*Client{},
		map[string]chan *websocket.Conn{},
	)

	startTestManager(t)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ws", websocketHandler)
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	dkOld := newTestDataKit("conn-old")
	dkOld.RunTimeID = "runtime-same"

	clientOld := newTestWebsocketClient(t, wsURL, dkOld)
	clientOld.Start()
	t.Cleanup(clientOld.Stop)
	require.Eventually(t, func() bool {
		row, err := db.Find(dkOld)
		return err == nil && row != nil && row.Status == ws.StatusRunning
	}, time.Second, 10*time.Millisecond)

	// the same datakit process reconnects with a new conn id
	dkNew := newTestDataKit("conn-new")
	dkNew.RunTimeID = dkOld.RunTimeID

	clientNew := newTestWebsocketClient(t, wsURL, dkNew)
	clientNew.Start()
	t.Cleanup(clientNew.Stop)
	require.Eventually(t, func() bool {
		row, err := db.Find(dkNew)
		return err == nil && row != nil && row.Status == ws.StatusRunning
	}, time.Second, 10*time.Millisecond)

	rows := []*ws.DataKit{}
	require.NoError(t, db.Select("select * from datakit", &rows))
	require.Len(t, rows, 1, "the superseded row of the same datakit must be removed")

	// the old session ending must not mark the live row offline
	clientOld.Stop()
	require.Eventually(t, func() bool { return !hasLiveSession(dkOld.ConnID) },
		time.Second, 10*time.Millisecond)

	row, err := db.Find(dkNew)
	require.NoError(t, err)
	require.NotNil(t, row)
	require.Equal(t, ws.StatusRunning, row.Status)

	clientNew.Stop()
	require.Eventually(t, func() bool { return !hasLiveSession(dkNew.ConnID) },
		time.Second, 10*time.Millisecond)
}

// A persisted row without a socket owner must allow the very first reconnect:
// rejecting it and waiting for the client's next retry delays recovery, and
// old clients may keep retrying the closed socket instead of reconnecting.
func TestFirstReconnectReclaimsStaleRow(t *testing.T) {
	db := withTestDatakitDB(t)
	replaceManagerState(t,
		make(chan *Client, 10),
		make(chan *Client, 10),
		map[string]*Client{},
		map[string]chan *websocket.Conn{},
	)

	startTestManager(t)

	// the row the previous DCA process left behind
	dk := newTestDataKit("stale-after-restart")
	require.NoError(t, db.Insert(dk))

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ws", websocketHandler)
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	// The first connection must become usable, with no rejected retry first.
	first := newTestWebsocketClient(t, wsURL, dk)
	first.Start()
	t.Cleanup(first.Stop)

	require.Eventually(t, func() bool {
		row, err := db.Find(dk)
		return err == nil && row != nil && row.Status == ws.StatusRunning && hasLiveSession(dk.ConnID)
	}, 5*time.Second, 20*time.Millisecond,
		"the first reconnect must replace the orphaned row")

	first.Stop()
	require.Eventually(t, func() bool { return !hasLiveSession(dk.ConnID) },
		time.Second, 10*time.Millisecond)
}

func hasLiveSession(connID string) bool {
	// NOTE: Manager.Clients is read under the manager lock, while the manager
	// loop itself does not guard it (same as the other tests in this package).
	Manager.RLock()
	defer Manager.RUnlock()

	_, ok := Manager.Clients[connID]
	return ok
}

// The list tells the web UI whether a row has a live session (so the management
// actions can be disabled instead of failing), and acting on a session-less row
// returns a dedicated error code.
func TestListAndActionReportSessionState(t *testing.T) {
	db := withTestDatakitDB(t)
	replaceManagerState(t,
		make(chan *Client, 10),
		make(chan *Client, 10),
		map[string]*Client{},
		map[string]chan *websocket.Conn{},
	)

	dk := newTestDataKit("session-state")
	require.NoError(t, db.Insert(dk))

	rows := []ws.DataKit{}
	require.NoError(t, db.Select("select * from datakit where conn_id=?", &rows, dk.ConnID))
	require.Len(t, rows, 1)

	// no live session: alive=false, and the action fails with datakit.offline
	ctx, rec := newGinTestContext(http.MethodGet, "/api/datakit/list?pageIndex=1&pageSize=10", nil)
	addWorkspaceCookie(ctx)
	datakitListHandler(ctx)
	require.Contains(t, rec.Body.String(), `"alive":false`)

	statsCtx, statsRec := newGinTestContext(http.MethodGet,
		"/api/datakit/stats?datakit_id="+url.QueryEscape(rows[0].ID), nil)
	addWorkspaceCookie(statsCtx)
	getHandler(ws.GetDatakitStatsAction)(statsCtx)
	require.Contains(t, statsRec.Body.String(), "datakit.offline")

	// with a live session: alive=true
	Manager.Lock()
	Manager.Clients[dk.ConnID] = &Client{ID: dk.ConnID, DataKit: dk}
	Manager.Unlock()
	t.Cleanup(func() {
		Manager.Lock()
		delete(Manager.Clients, dk.ConnID)
		Manager.Unlock()
	})

	ctx, rec = newGinTestContext(http.MethodGet, "/api/datakit/list?pageIndex=1&pageSize=10", nil)
	addWorkspaceCookie(ctx)
	datakitListHandler(ctx)
	require.Contains(t, rec.Body.String(), `"alive":true`)
}

// A datakit only sends keepalive pings while nothing changes, and that must be
// enough to keep the session: dropping it would make the host flap in the list
// (and the platform would see it go offline) even though it is perfectly alive.
func TestSessionStaysAliveWithPingsOnly(t *testing.T) {
	withTestDatakitDB(t)
	replaceManagerState(t,
		make(chan *Client, 10),
		make(chan *Client, 10),
		map[string]*Client{},
		map[string]chan *websocket.Conn{},
	)

	oldTimeout := sessionReadTimeout
	sessionReadTimeout = 1500 * time.Millisecond
	t.Cleanup(func() { sessionReadTimeout = oldTimeout })

	startTestManager(t)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ws", websocketHandler)
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	dk := newTestDataKit("ping-only")
	// only keepalives: no stats/config request is ever sent, and the heartbeat
	// has to be shorter than the (test) session timeout
	client, err := ws.NewClient(
		ws.WithWebsocketAddress("ws"+strings.TrimPrefix(server.URL, "http")+"/ws"),
		ws.WithDataKit(dk),
		ws.WithHeartbeatInterval(sessionReadTimeout/3),
	)
	require.NoError(t, err)
	client.Start()
	t.Cleanup(client.Stop)

	require.Eventually(t, func() bool { return hasLiveSession(dk.ConnID) },
		time.Second, 10*time.Millisecond)

	// several read-timeouts worth of ping-only traffic
	time.Sleep(3 * sessionReadTimeout)
	require.True(t, hasLiveSession(dk.ConnID),
		"session dropped even though the datakit kept sending keepalives")

	// stop inside the test: the unregistration is processed asynchronously and
	// must not reach the DB after the test restored it
	client.Stop()
	require.Eventually(t, func() bool { return !hasLiveSession(dk.ConnID) },
		time.Second, 10*time.Millisecond)
}
