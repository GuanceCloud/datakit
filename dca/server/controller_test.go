// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package server

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

func TestForwardDatakitLogMessagesKeepsIdleDatakitConnOpen(t *testing.T) {
	frontServerConn, frontClientConn := newTestWebsocketPair(t)
	defer frontServerConn.Close() //nolint:errcheck
	defer frontClientConn.Close() //nolint:errcheck

	datakitServerConn, datakitClientConn := newTestWebsocketPair(t)
	defer datakitServerConn.Close() //nolint:errcheck
	defer datakitClientConn.Close() //nolint:errcheck

	done := make(chan struct{})
	errCh := make(chan error, 1)
	go func() {
		errCh <- forwardDatakitLogMessages(
			frontServerConn,
			datakitServerConn,
			ws.GetDatakitLogTailAction,
			done,
		)
	}()

	select {
	case err := <-errCh:
		t.Fatalf("forwardDatakitLogMessages returned while datakit log was idle: %v", err)
	case <-time.After(50 * time.Millisecond):
	}

	close(done)

	require.NoError(t, <-errCh)
}

func TestForwardDatakitLogMessagesForwardsAndReportsErrors(t *testing.T) {
	t.Run("forward text payload", func(t *testing.T) {
		frontServerConn, frontClientConn := newTestWebsocketPair(t)
		defer frontServerConn.Close() //nolint:errcheck
		defer frontClientConn.Close() //nolint:errcheck

		datakitServerConn, datakitClientConn := newTestWebsocketPair(t)
		defer datakitServerConn.Close() //nolint:errcheck
		defer datakitClientConn.Close() //nolint:errcheck

		done := make(chan struct{})
		errCh := make(chan error, 1)
		go func() {
			errCh <- forwardDatakitLogMessages(frontServerConn, datakitServerConn, ws.GetDatakitLogTailAction, done)
		}()

		payload := []byte("line-1")
		msg := ws.WebsocketMessage{
			Action: ws.GetDatakitLogTailAction,
			Data:   &ws.DCAResponse{Success: true, Code: 200, Content: &payload},
		}
		require.NoError(t, datakitClientConn.WriteMessage(websocket.TextMessage, msg.Bytes()))
		_, body, err := frontClientConn.ReadMessage()
		require.NoError(t, err)
		require.Equal(t, payload, body)
		close(done)
		require.NoError(t, <-errCh)
	})

	cases := []struct {
		name string
		msg  []byte
	}{
		{name: "bad json", msg: []byte("{bad-json")},
		{name: "action mismatch", msg: (&ws.WebsocketMessage{
			Action: "other",
			Data:   &ws.DCAResponse{Success: true, Code: 200},
		}).Bytes()},
		{name: "response failed", msg: (&ws.WebsocketMessage{
			Action: ws.GetDatakitLogTailAction,
			Data:   &ws.DCAResponse{Success: false, Code: 500, Message: "failed"},
		}).Bytes()},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			frontServerConn, frontClientConn := newTestWebsocketPair(t)
			defer frontServerConn.Close() //nolint:errcheck
			defer frontClientConn.Close() //nolint:errcheck

			datakitServerConn, datakitClientConn := newTestWebsocketPair(t)
			defer datakitServerConn.Close() //nolint:errcheck
			defer datakitClientConn.Close() //nolint:errcheck

			errCh := make(chan error, 1)
			go func() {
				errCh <- forwardDatakitLogMessages(frontServerConn, datakitServerConn, ws.GetDatakitLogTailAction, nil)
			}()
			require.NoError(t, datakitClientConn.WriteMessage(websocket.TextMessage, tc.msg))
			require.Error(t, <-errCh)
		})
	}
}

func newTestWebsocketPair(t *testing.T) (*websocket.Conn, *websocket.Conn) {
	t.Helper()

	serverConnCh := make(chan *websocket.Conn, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
		serverConnCh <- conn
	}))
	t.Cleanup(server.Close)

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	clientConn, _, err := websocket.DefaultDialer.Dial(url, nil)
	require.NoError(t, err)

	select {
	case serverConn := <-serverConnCh:
		return serverConn, clientConn
	case <-time.After(time.Second):
		clientConn.Close() //nolint:errcheck
		t.Fatal("timeout waiting for websocket server connection")
		return nil, nil
	}
}
