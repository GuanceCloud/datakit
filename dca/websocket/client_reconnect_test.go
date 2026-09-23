// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package websocket

import (
	"io"
	"net"
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

// switchWriter is a proxy direction that can be muted: while muted the bytes are
// swallowed, so the peer keeps its socket open but never sees them (the client
// side of a black holed connection, e.g. a proxy in front of DCA during a
// rollout).
type switchWriter struct {
	dst net.Conn

	mu    sync.Mutex
	muted bool
}

func (w *switchWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	muted := w.muted
	w.mu.Unlock()

	if muted {
		return len(p), nil
	}

	return w.dst.Write(p)
}

func (w *switchWriter) setMuted(v bool) {
	w.mu.Lock()
	w.muted = v
	w.mu.Unlock()
}

// A session can die while the TCP connection stays open: the peer stops
// answering (black holed path, proxy that keeps the client side). The datakit
// must notice and reconnect, otherwise it keeps talking to a session that does
// not exist anymore and DCA shows the host offline while its process is alive.
func TestClientReconnectsAfterBlackHole(t *testing.T) {
	var dials int32

	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		go func() { // read: gorilla answers the keepalive pings by default
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					return
				}
			}
		}()
	}))
	defer server.Close()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close() //nolint:errcheck

	var (
		writersMu sync.Mutex
		writers   []*switchWriter
	)

	go func() {
		for {
			clientConn, err := ln.Accept()
			if err != nil {
				return
			}
			atomic.AddInt32(&dials, 1)

			backendConn, err := net.Dial("tcp", strings.TrimPrefix(server.URL, "http://"))
			if err != nil {
				_ = clientConn.Close()
				continue
			}

			toBackend := &switchWriter{dst: backendConn}
			toClient := &switchWriter{dst: clientConn}

			writersMu.Lock()
			writers = append(writers, toBackend, toClient)
			writersMu.Unlock()

			go func() { _, _ = io.Copy(toBackend, clientConn); _ = backendConn.Close() }()
			go func() { _, _ = io.Copy(toClient, backendConn); _ = clientConn.Close() }()
		}
	}()

	go func() { // the connection works for 2s, then it is black holed
		time.Sleep(2 * time.Second)

		writersMu.Lock()
		for _, w := range writers {
			w.setMuted(true)
		}
		writersMu.Unlock()
	}()

	dk := &DataKit{
		RunTimeID:     "runtime-blackhole",
		HostName:      "host-blackhole",
		OS:            "linux",
		Arch:          "amd64",
		WorkspaceUUID: "workspace-blackhole",
		ConnID:        "conn-blackhole",
	}

	client, err := NewClient(
		WithWebsocketAddress("ws://"+ln.Addr().String()),
		WithDataKit(dk),
		WithHeartbeatInterval(500*time.Millisecond),
	)
	require.NoError(t, err)
	client.Start()
	t.Cleanup(client.Stop)

	require.Eventually(t, func() bool { return atomic.LoadInt32(&dials) >= 2 },
		10*time.Second, 100*time.Millisecond,
		"the client must reconnect after the connection is black holed")

	time.Sleep(2 * time.Second) // settle: the new connection keeps working

	// no reconnect storm: one connection plus the one after the black hole
	require.LessOrEqual(t, atomic.LoadInt32(&dials), int32(4),
		"the client must not reconnect in a loop")
	require.NotZero(t, client.lastPong.Load(), "the new connection must receive pongs")
}
