// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package server

import (
	"encoding/json"
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

func replaceManagerState(t *testing.T, register chan *Client, unregister chan *Client, clients map[string]*Client, websocketConns map[string]chan *websocket.Conn) {
	t.Helper()

	oldRegister := Manager.Register
	oldUnregister := Manager.Unregister
	oldClients := Manager.Clients
	oldWebsocketConns := Manager.WebsocketConns

	Manager.Register = register
	Manager.Unregister = unregister
	Manager.Clients = clients
	Manager.WebsocketConns = websocketConns

	t.Cleanup(func() {
		Manager.Register = oldRegister
		Manager.Unregister = oldUnregister
		Manager.Clients = oldClients
		Manager.WebsocketConns = oldWebsocketConns
	})
}

func TestClientManagerWebsocketConnChannels(t *testing.T) {
	manager := &ClientManager{WebsocketConns: map[string]chan *websocket.Conn{}}
	ch := manager.initWebsocketConnChan("conn-chan")

	serverConn, clientConn := newTestWebsocketPair(t)
	defer serverConn.Close() //nolint:errcheck
	defer clientConn.Close() //nolint:errcheck

	require.NoError(t, manager.addWebsocketConnChan("conn-chan", serverConn))
	require.Same(t, serverConn, <-ch)

	manager.deleteWebsocketConnChan("conn-chan")
	_, ok := manager.WebsocketConns["conn-chan"]
	require.False(t, ok)

	require.Error(t, manager.addWebsocketConnChan("missing", serverConn))
	manager.deleteWebsocketConnChan("missing")
}

func TestClientManagerRejectsWebsocketConnWhenChannelIsFull(t *testing.T) {
	manager := &ClientManager{WebsocketConns: map[string]chan *websocket.Conn{}}
	ch := manager.initWebsocketConnChan("conn-chan")

	firstServerConn, firstClientConn := newTestWebsocketPair(t)
	defer firstServerConn.Close() //nolint:errcheck
	defer firstClientConn.Close() //nolint:errcheck
	secondServerConn, secondClientConn := newTestWebsocketPair(t)
	defer secondServerConn.Close() //nolint:errcheck
	defer secondClientConn.Close() //nolint:errcheck

	require.NoError(t, manager.addWebsocketConnChan("conn-chan", firstServerConn))
	require.Error(t, manager.addWebsocketConnChan("conn-chan", secondServerConn))
	require.Same(t, firstServerConn, <-ch)
}

func TestClientManagerAction(t *testing.T) {
	manager := &ClientManager{Clients: map[string]*Client{}}
	dk := newTestDataKit("manager-action")

	_, err := manager.Action("action", nil, nil)
	require.Error(t, err)

	_, err = manager.Action("action", dk, nil)
	require.Error(t, err)

	client := newTestClientForRequest()
	client.DataKit = dk
	manager.Clients[dk.ConnID] = client
	ActionHandlerMap["manager-test-action"] = func(*Client, *ws.DataKit, *gin.Context) (any, error) {
		return &ws.DCAResponse{Success: true, Code: 200}, nil
	}
	t.Cleanup(func() { delete(ActionHandlerMap, "manager-test-action") })

	ctx, _ := newGinTestContext(http.MethodGet, "/api/datakit/test", nil)
	resp, err := manager.Action("manager-test-action", dk, ctx)
	require.NoError(t, err)
	require.True(t, resp.Success)
}

func TestDealNewWebsocketConnection(t *testing.T) {
	replaceManagerState(t, Manager.Register, Manager.Unregister, Manager.Clients, map[string]chan *websocket.Conn{})

	ch := Manager.initWebsocketConnChan("new-conn")
	serverConn, clientConn := newTestWebsocketPair(t)
	defer serverConn.Close() //nolint:errcheck
	defer clientConn.Close() //nolint:errcheck

	require.NoError(t, dealNewWebsocketConnection(serverConn, "new-conn"))
	require.Same(t, serverConn, <-ch)

	queuedServerConn, queuedClientConn := newTestWebsocketPair(t)
	defer queuedServerConn.Close() //nolint:errcheck
	defer queuedClientConn.Close() //nolint:errcheck
	require.NoError(t, dealNewWebsocketConnection(queuedServerConn, "new-conn"))

	fullServerConn, fullClientConn := newTestWebsocketPair(t)
	defer fullClientConn.Close() //nolint:errcheck
	require.Error(t, dealNewWebsocketConnection(fullServerConn, "new-conn"))
	require.Error(t, fullServerConn.WriteMessage(websocket.TextMessage, []byte("closed")))
	require.Same(t, queuedServerConn, <-ch)

	rejectedServerConn, rejectedClientConn := newTestWebsocketPair(t)
	defer rejectedClientConn.Close() //nolint:errcheck
	require.Error(t, dealNewWebsocketConnection(rejectedServerConn, "missing"))
	require.Error(t, rejectedServerConn.WriteMessage(websocket.TextMessage, []byte("closed")))
}

func TestGetNewWebsocketConn(t *testing.T) {
	replaceManagerState(t, Manager.Register, Manager.Unregister, map[string]*Client{}, map[string]chan *websocket.Conn{})

	dk := newTestDataKit("new-ws")
	client := newTestClientForRequest()
	client.DataKit = dk
	Manager.Clients[dk.ConnID] = client

	var clientSideConn *websocket.Conn
	go func() {
		raw := <-client.Send
		msg := ws.WebsocketMessage{Data: &ActionData{}}
		require.NoError(t, json.Unmarshal(raw, &msg))
		data := msg.Data.(*ActionData)
		require.Equal(t, ws.NewWebsocketConnectionAction, msg.Action)
		require.Equal(t, ws.GetDatakitLogTailAction, data.Query.Get(ws.HeaderWebsocketAction))

		serverConn, clientConn := newTestWebsocketPair(t)
		clientSideConn = clientConn
		Manager.addWebsocketConnChan(data.Query.Get(ws.HeaderNewWebSocketConnectionID), serverConn)
	}()

	conn, err := getNewWebsocketConn(dk, ws.GetDatakitLogTailAction)
	require.NoError(t, err)
	require.NotNil(t, conn)
	require.NoError(t, conn.Close())
	if clientSideConn != nil {
		require.NoError(t, clientSideConn.Close())
	}

	_, err = getNewWebsocketConn(newTestDataKit("missing-client"), ws.GetDatakitLogTailAction)
	require.Error(t, err)
}

func TestClientReadAndWrite(t *testing.T) {
	replaceManagerState(t, Manager.Register, make(chan *Client, 3), Manager.Clients, Manager.WebsocketConns)

	serverConn, clientConn := newTestWebsocketPair(t)
	client := &Client{
		ID:      "read-client",
		Socket:  serverConn,
		Send:    make(chan []byte, 1),
		Receive: map[int64]chan []byte{1: make(chan []byte, 1)},
		Close:   make(chan interface{}),
		DataKit: newTestDataKit("read-client"),
		Timeout: time.Second,
	}

	go client.Read()
	msg := ws.WebsocketMessage{
		ID:     1,
		Action: ws.GetDatakitStatsAction,
		Data:   &ws.DCAResponse{Success: true, Code: 200},
	}
	require.NoError(t, clientConn.WriteMessage(websocket.TextMessage, msg.Bytes()))
	require.JSONEq(t, string(msg.Bytes()), string(<-client.Receive[1]))
	require.NoError(t, clientConn.Close())
	select {
	case <-client.Close:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for client read exit")
	}

	serverConn, clientConn = newTestWebsocketPair(t)
	writeClient := &Client{
		ID:      "write-client",
		Socket:  serverConn,
		Send:    make(chan []byte, 1),
		Close:   make(chan interface{}),
		DataKit: newTestDataKit("write-client"),
		Timeout: time.Second,
	}
	writeDone := make(chan struct{})
	go func() {
		defer close(writeDone)
		writeClient.Write()
	}()
	writeClient.Send <- []byte("hello")
	_, body, err := clientConn.ReadMessage()
	require.NoError(t, err)
	require.Equal(t, []byte("hello"), body)
	close(writeClient.Close)
	select {
	case <-writeDone:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for client write exit")
	}
	require.NoError(t, clientConn.Close())
	require.NoError(t, serverConn.Close())

	serverConn, clientConn = newTestWebsocketPair(t)
	closeSendClient := &Client{
		ID:      "close-send-client",
		Socket:  serverConn,
		Send:    make(chan []byte),
		Close:   make(chan interface{}),
		DataKit: newTestDataKit("close-send-client"),
		Timeout: time.Second,
	}
	closeSendDone := make(chan struct{})
	go func() {
		defer close(closeSendDone)
		closeSendClient.Write()
	}()
	close(closeSendClient.Send)
	messageType, _, err := clientConn.ReadMessage()
	if err == nil {
		require.Equal(t, websocket.CloseMessage, messageType)
	} else {
		require.True(t, websocket.IsCloseError(err, websocket.CloseNoStatusReceived))
	}
	select {
	case <-closeSendClient.Close:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for closed send client exit")
	}
	select {
	case <-closeSendDone:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for closed send client write exit")
	}
	require.NoError(t, clientConn.Close())
}

func TestClientReceiveMessageBranches(t *testing.T) {
	db := withTestDatakitDB(t)
	dk := newTestDataKit("receive-branches")
	require.NoError(t, db.Insert(dk))

	client := newTestClientForRequest()
	client.DataKit = dk
	client.receiveMessage([]byte("{bad-json"))

	client.receiveMessage((&ws.WebsocketMessage{
		ID:     99,
		Action: ws.GetDatakitStatsAction,
		Data:   &ws.DCAResponse{},
	}).Bytes())

	client.receiveMessage((&ws.WebsocketMessage{
		ID:     0,
		Action: ws.UpdateDatakitStatus,
		Data:   &ws.ActionData{Query: mapValues("status", ws.StatusStopped.String())},
	}).Bytes())
	found, err := db.Find(dk)
	require.NoError(t, err)
	require.Equal(t, ws.StatusStopped, found.Status)
}

func TestClientManagerStartRegisterAndUnregister(t *testing.T) {
	db := withTestDatakitDB(t)
	replaceManagerState(t, make(chan *Client, 10), make(chan *Client, 10), map[string]*Client{}, map[string]chan *websocket.Conn{})

	go Manager.Start()

	serverConn, clientConn := newTestWebsocketPair(t)
	defer clientConn.Close() //nolint:errcheck
	dk := newTestDataKit("manager-start")
	client := &Client{ID: dk.ConnID, Socket: serverConn, DataKit: dk}
	Manager.Register <- client
	require.Eventually(t, func() bool {
		_, ok := Manager.Clients[dk.ConnID]
		return ok
	}, time.Second, 10*time.Millisecond)
	found, err := db.Find(dk)
	require.NoError(t, err)
	require.NotNil(t, found)

	duplicateServerConn, duplicateClientConn := newTestWebsocketPair(t)
	defer duplicateClientConn.Close() //nolint:errcheck
	Manager.Register <- &Client{ID: dk.ConnID, Socket: duplicateServerConn, DataKit: dk}
	require.Eventually(t, func() bool {
		return duplicateServerConn.WriteMessage(websocket.TextMessage, []byte("closed")) != nil
	}, time.Second, 10*time.Millisecond)

	Manager.Unregister <- client
	require.Eventually(t, func() bool {
		_, ok := Manager.Clients[dk.ConnID]
		return !ok
	}, time.Second, 10*time.Millisecond)
}

func TestWebsocketHandlerRejectsBadDatakitHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ws", websocketHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/ws", nil)
	req.Header.Set(ws.HeaderDatakit, "{bad-json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/ws", nil)
	req.Header.Set(ws.HeaderDatakit, "{}")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWebsocketHandlerRejectsWhenRegisterQueueIsFull(t *testing.T) {
	replaceManagerState(t, make(chan *Client, 1), make(chan *Client, 1), map[string]*Client{}, map[string]chan *websocket.Conn{})
	Manager.Register <- &Client{}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ws", websocketHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	router.ServeHTTP(w, req)
	resp := decodeDCAResponse(t, w.Body.String())
	require.False(t, resp.Success)
	require.EqualValues(t, http.StatusTooManyRequests, resp.Code)
}

func TestWebsocketHandlerUpgradeFailureAfterValidHeader(t *testing.T) {
	replaceManagerState(t, make(chan *Client, 1), make(chan *Client, 1), map[string]*Client{}, map[string]chan *websocket.Conn{})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ws", websocketHandler)

	dk := newTestDataKit("upgrade-failure")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	req.Header.Set(ws.HeaderDatakit, string(dk.Bytes()))
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWebsocketHandlerAcceptsDatakitAndNewConnection(t *testing.T) {
	db := withTestDatakitDB(t)
	replaceManagerState(t, make(chan *Client, 1), make(chan *Client, 1), map[string]*Client{}, map[string]chan *websocket.Conn{"extra-conn": make(chan *websocket.Conn, 1)})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ws", websocketHandler)
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	dk := newTestDataKit("handler-main")
	header := http.Header{}
	header.Set(ws.HeaderDatakit, string(dk.Bytes()))
	clientConn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http")+"/ws", header)
	require.NoError(t, err)

	registered := <-Manager.Register
	require.Equal(t, dk.ConnID, registered.ID)
	require.Equal(t, ws.StatusRunning, registered.DataKit.Status)
	require.NoError(t, db.Insert(registered.DataKit))
	require.NoError(t, clientConn.Close())

	header = http.Header{}
	header.Set(ws.HeaderNewWebSocketConnectionID, "extra-conn")
	extraConn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http")+"/ws", header)
	require.NoError(t, err)
	require.NotNil(t, <-Manager.WebsocketConns["extra-conn"])
	require.NoError(t, extraConn.Close())
}
