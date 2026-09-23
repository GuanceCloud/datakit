// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
	ws "gitlab.jiagouyun.com/cloudcare-tools/datakit/dca/websocket"
)

func TestReviewLogActionSelfSignedConnection(t *testing.T) {
	accepted := make(chan string, 1)
	upgrader := websocket.Upgrader{}
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close() //nolint:errcheck
		accepted <- r.Header.Get(ws.HeaderNewWebSocketConnectionID)
		// Complete the secondary connection's first read without starting a log stream.
		if err := conn.WriteJSON(ws.WebsocketMessage{Action: "review-unknown-action"}); err != nil {
			return
		}
	}))
	defer server.Close()
	client, err := ws.NewClient(ws.WithWebsocketAddress("wss"+strings.TrimPrefix(server.URL, "https")),
		ws.WithDataKit(&ws.DataKit{WorkspaceUUID: "workspace", ConnID: "conn"}))
	require.NoError(t, err)
	defer client.Stop()
	msg := ws.WebsocketMessage{Data: &ws.ActionData{Query: map[string][]string{ws.HeaderNewWebSocketConnectionID: {"log-subconnection"}}}}
	require.NoError(t, newWebsocketConnectionAction(client, 1, msg.Bytes()))
	select {
	case id := <-accepted:
		require.Equal(t, "log-subconnection", id)
	case <-time.After(time.Second):
		t.Fatal("log action did not connect to the self-signed DCA endpoint")
	}
}
