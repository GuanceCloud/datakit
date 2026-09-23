// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package websocket

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func TestReviewSelfSignedDCASecondaryConnection(t *testing.T) {
	upgrader := websocket.Upgrader{}
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	roots := x509.NewCertPool()
	roots.AddCert(server.Certificate())
	for _, tt := range []struct {
		name    string
		config  *tls.Config
		allowed bool
	}{
		{name: "legacy self signed", allowed: true},
		{name: "trusted private CA", config: &tls.Config{RootCAs: roots, MinVersion: tls.VersionTLS12}, allowed: true},
		{name: "untrusted CA rejected", config: &tls.Config{RootCAs: x509.NewCertPool(), MinVersion: tls.VersionTLS12}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			original := websocket.DefaultDialer.TLSClientConfig
			client := reviewClient(t, "wss"+strings.TrimPrefix(server.URL, "https"), WithTLSConfig(tt.config))
			mainErr := client.init()
			conn, resp, err := client.Dial(http.Header{HeaderNewWebSocketConnectionID: []string{"log-connection"}})
			if resp != nil {
				defer resp.Body.Close()
			} //nolint:errcheck
			if conn != nil {
				defer conn.Close()
			} //nolint:errcheck
			if tt.allowed {
				require.NoError(t, mainErr)
				require.NoError(t, err)
			} else {
				require.Error(t, mainErr)
				require.Error(t, err)
			}
			require.True(t, original == websocket.DefaultDialer.TLSClientConfig, "global dialer must remain unchanged")
		})
	}
}

func TestReviewStopDiscardsQueuedActions(t *testing.T) {
	client := reviewClient(t, "ws://127.0.0.1:1")
	var executed bool
	client.RegisterActionHandler("queued", func(*Client, int64, any) error { executed = true; return nil })
	client.actionQueue <- queuedAction{data: []byte(`{"action":"queued"}`)}
	client.Stop()
	client.runActions()
	require.False(t, executed)
	require.Error(t, client.SendMessage(&WebsocketMessage{}))
}
