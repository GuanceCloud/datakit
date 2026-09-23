// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package server

import (
	"fmt"
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

func TestFleetRegistrationAndRecovery(t *testing.T) {
	db := withTestDatakitDB(t)
	replaceManagerState(t, make(chan *Client, 2048), make(chan *Client, 2048), map[string]*Client{}, map[string]chan *websocket.Conn{})
	previousTimeout := sessionReadTimeout
	sessionReadTimeout = 90 * time.Second
	defer func() { sessionReadTimeout = previousTimeout }()
	startTestManager(t)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ws", websocketHandler)
	httpServer := httptest.NewServer(router)
	defer httpServer.Close()
	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/ws"
	kits := make([]*ws.DataKit, 46)
	clients := make([]*websocket.Conn, 46)
	done := make(chan struct{})
	defer close(done)
	dial := func(dk *ws.DataKit) *websocket.Conn {
		h := http.Header{}
		h.Set(ws.HeaderDatakit, string(dk.Bytes()))
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, h)
		require.NoError(t, err)
		return conn
	}
	keep := func(conn *websocket.Conn) {
		go func() {
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					return
				}
			}
		}()
		go func() {
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(time.Second)); err != nil {
						return
					}
				}
			}
		}()
	}
	count := func() (rows, running, offline, live int) {
		found := []ws.DataKit{}
		require.NoError(t, db.Select("select * from datakit", &found))
		for _, dk := range found {
			if dk.Status == ws.StatusRunning {
				running++
			}
			if dk.Status == ws.StatusOffline {
				offline++
			}
		}
		Manager.RLock()
		live = len(Manager.Clients)
		Manager.RUnlock()
		return len(found), running, offline, live
	}
	expect := func(stage string, wantRunning, wantOffline int) {
		require.Eventually(t, func() bool {
			n, r, o, l := count()
			return n == 46 && r == wantRunning && o == wantOffline && l == wantRunning
		}, 15*time.Second, 10*time.Millisecond, stage)
		ctx, rec := newGinTestContext(http.MethodGet, "/api/datakit/list?pageIndex=1&pageSize=20", nil)
		addWorkspaceCookie(ctx)
		datakitListHandler(ctx)
		response := decodeDCAResponse(t, rec.Body.String())
		require.True(t, response.Success)
		content := response.Content.(map[string]interface{})
		page := content["pageInfo"].(map[string]interface{})
		require.EqualValues(t, 46, page["totalCount"])
		require.EqualValues(t, 20, page["count"])
		n, r, o, l := count()
		t.Logf("DIAGNOSTIC stage=%s rows=%d running=%d offline=%d live=%d api_total=%v page_rows=%v", stage, n, r, o, l, page["totalCount"], page["count"])
	}
	defer func() {
		for _, c := range clients {
			if c != nil {
				_ = c.Close()
			}
		}
		require.Eventually(t, func() bool { _, _, _, l := count(); return l == 0 }, 15*time.Second, 10*time.Millisecond)
	}()
	for i := range kits {
		kits[i] = newTestDataKit(fmt.Sprintf("fleet-%02d", i))
		kits[i].IP = fmt.Sprintf("192.0.2.%d", i+1)
		kits[i].ConnID = kits[i].GetConnID(wsURL)
		clients[i] = dial(kits[i])
		keep(clients[i])
	}
	expect("registered46", 46, 0)

	for i := 0; i < 31; i++ {
		duplicate := dial(kits[i])
		require.NoError(t, duplicate.SetReadDeadline(time.Now().Add(time.Second)))
		_, _, err := duplicate.ReadMessage()
		require.Error(t, err, "duplicate must be rejected")
		_ = duplicate.Close()
	}
	expect("rejected31duplicates", 46, 0)
	for i := 0; i < 27; i++ {
		require.NoError(t, clients[i].Close())
		clients[i] = nil
	}
	expect("abruptly_closed27", 19, 27)
	for i := 0; i < 27; i++ {
		clients[i] = dial(kits[i])
		keep(clients[i])
	}
	expect("reregistered27", 46, 0)
}
