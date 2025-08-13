// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package server is DCA's HTTP server
package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	ws "gitlab.jiagouyun.com/cloudcare-tools/datakit/dca/websocket"
)

var ErrRequestTimeout = errors.New("request_time_out")

type ActionData struct {
	Body  string     `json:"body"`
	Query url.Values `json:"query"`
}

type Client struct {
	sync.Mutex
	ID                string
	Socket            *websocket.Conn
	Send              chan []byte
	Receive           map[int64]chan []byte
	Close             chan interface{}
	DataKit           *ws.DataKit
	Timeout           time.Duration
	HeartbeatInterval time.Duration

	messageNumber int64
}

func (c *Client) getActionHandler(action string) ActionHandler {
	if handler, ok := ActionHandlerMap[action]; ok {
		return handler
	}
	return nil
}

func (c *Client) doAction(action string, datakit *ws.DataKit, ctx *gin.Context) (any, error) {
	handler := c.getActionHandler(action)
	if handler == nil {
		return nil, fmt.Errorf("unknown action: %s", action)
	}
	return handler(c, datakit, ctx)
}

func (c *Client) getMessageCh(id int64) (chan<- []byte, error) {
	c.Lock()
	defer c.Unlock()
	if ch, ok := c.Receive[id]; !ok {
		return nil, fmt.Errorf("message receive channel not existed, id: %d", id)
	} else {
		return ch, nil
	}
}

func (c *Client) receiveMessage(message []byte) {
	msg := ws.WebsocketMessage{
		Data: &ws.ActionData{},
	}
	if err := json.Unmarshal(message, &msg); err != nil {
		l.Errorf("receive message failed: %s", err.Error())
		return
	}

	l.Debugf("receive message, message action: %s", msg.Action)

	if msg.ID == 0 { // client push message
		doCommonAction(c, &msg)
		return
	}

	// request - reply message
	ch, err := c.getMessageCh(msg.ID)
	if err != nil {
		l.Warnf("receive message error: %s", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()
	select {
	case ch <- message:
	case <-ctx.Done():
		l.Warnf("receive message timeout")
	}
}

func (c *Client) Read() {
	defer func() {
		c.Exit()
	}()
	for {
		messageType, message, err := c.Socket.ReadMessage()
		if err != nil {
			l.Warnf("read message failed: %s", err.Error())
			break
		}
		l.Debugf("get message, message type: %d", messageType)
		switch messageType {
		case websocket.TextMessage:
			c.receiveMessage(message)
		case websocket.CloseMessage:
			return
		case websocket.PingMessage:
			// c.Socket.WriteMessage(websocket.PongMessage, []byte("ping"))
		default:
			l.Infof("unknown message type: %d, ignore", messageType)
		}
	}
}

func (c *Client) Exit() {
	Manager.Unregister <- c
	c.Socket.Close() // nolint:errcheck,gosec
	close(c.Close)
}

func (c *Client) getMessageNumber() (int64, <-chan []byte) {
	c.Lock()
	defer c.Unlock()
	id := c.messageNumber
	if id > math.MaxInt64-1 {
		c.messageNumber = 1
		id = 1
	}
	c.messageNumber++
	c.Receive[id] = make(chan []byte)
	return id, c.Receive[id]
}

func (c *Client) releaseMessageCh(id int64) {
	c.Lock()
	defer c.Unlock()
	delete(c.Receive, id)
}

// request sends message to client in sequence.
func (c *Client) request(msg *ws.WebsocketMessage, dest *ws.WebsocketMessage) error {
	if msg == nil || dest == nil {
		return fmt.Errorf("invalid message: nil")
	}

	l.Debugf("write message")
	var receiveCh <-chan []byte
	msg.ID, receiveCh = c.getMessageNumber()
	defer func() {
		c.releaseMessageCh(msg.ID)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	select {
	case c.Send <- msg.Bytes():
	case <-ctx.Done():
		return ErrRequestTimeout
	}

	ctx1, cancel1 := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel1()
	select {
	case receivedMessage := <-receiveCh:
		if err := json.Unmarshal(receivedMessage, dest); err != nil {
			return fmt.Errorf("failed to unmarshal message: %w", err)
		} else if msg.Action != dest.Action {
			return fmt.Errorf("message action not match: %s, %s", msg.Action, dest.Action)
		} else if _, ok := dest.Data.(*ws.DCAResponse); !ok {
			return fmt.Errorf("message data type not match: %T", dest.Data)
		} else {
			return nil
		}
	case <-ctx1.Done():
		return ErrRequestTimeout
	}
}

func (c *Client) Write() {
	for {
		select {
		case <-c.Close:
			return
		case message, ok := <-c.Send:
			if !ok {
				if err := c.Socket.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					l.Errorf("write message failed: %s", err.Error())
					c.Exit()
					return
				}
			}

			if err := c.Socket.SetWriteDeadline(time.Now().Add(c.Timeout)); err != nil {
				l.Errorf("set write deadline failed: %s", err.Error())
				c.Exit()
				return
			}
			if err := c.Socket.WriteMessage(websocket.TextMessage, message); err != nil {
				l.Warnf("failed to write message: %s", err.Error())
			}
		}
	}
}

type ClientManager struct {
	sync.RWMutex
	Clients        map[string]*Client
	WebsocketConns map[string]chan *websocket.Conn
	Register       chan *Client
	Unregister     chan *Client
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Message is return msg.
type Message struct {
	Sender    string `json:"sender,omitempty"`
	Recipient string `json:"recipient,omitempty"`
	Content   string `json:"content,omitempty"`
}

func (manager *ClientManager) addWebsocketConnChan(connID string, conn *websocket.Conn) {
	manager.Lock()
	defer manager.Unlock()
	if ch, ok := manager.WebsocketConns[connID]; ok {
		select {
		case ch <- conn:
		default:
			l.Warnf("websocket conn chan is full, connID: %s, add failed", connID)
		}
		return
	} else {
		l.Warnf("websocket conn chan not existed, connID: %s")
	}
}

func (manager *ClientManager) initWebsocketConnChan(connID string) chan *websocket.Conn {
	manager.Lock()
	defer manager.Unlock()
	manager.WebsocketConns[connID] = make(chan *websocket.Conn)
	return manager.WebsocketConns[connID]
}

func (manager *ClientManager) deleteWebsocketConnChan(connID string) {
	manager.Lock()
	defer manager.Unlock()
	if ch, ok := manager.WebsocketConns[connID]; ok {
		close(ch)
		delete(manager.WebsocketConns, connID)
	} else {
		l.Warnf("websocket conn chan not existed, connID: %s")
	}
}

func (manager *ClientManager) Start() {
	l.Infof("websocket manager started")
	for {
		select {
		case conn := <-Manager.Register: // new connection
			isDuplicated, err := datakitDB.IsDuplicatedConn(conn.DataKit)
			if err != nil {
				l.Errorf("failed to check duplicate connection: %s", err.Error())
				conn.Socket.Close() //nolint:errcheck,gosec
				continue
			}

			if isDuplicated { // conn exists
				l.Infof("datakit connection already exists")
				conn.Socket.Close() //nolint:errcheck,gosec
				continue
			}

			if err := datakitDB.ForceUpdate(conn.DataKit); err != nil { // update datakit
				l.Errorf("failed to insert datakit: %s", err.Error())
				conn.Socket.Close() //nolint:errcheck,gosec
			} else {
				l.Infof("new connection registered: %s...", conn.DataKit.HostName)
				manager.Clients[conn.ID] = conn
				connectionTotalGauge.Set(float64(len(manager.Clients)))
			}

		case conn := <-Manager.Unregister:
			// delete hardly when run in container
			if err := datakitDB.DeleteByConnID(conn.DataKit.ConnID, conn.DataKit.RunInContainer); err != nil {
				l.Errorf("failed to delete datakit: %s", err.Error())
			}
			delete(manager.Clients, conn.ID)
			l.Infof("connection unregistered: %s", conn.DataKit.HostName)
			conn.Socket.Close() //nolint:errcheck,gosec
			connectionTotalGauge.Set(float64(len(manager.Clients)))
		}
	}
}

func (manager *ClientManager) Action(action string, datakit *ws.DataKit, ctx *gin.Context) (*ws.DCAResponse, error) {
	if datakit == nil {
		return nil, fmt.Errorf("datakit is required")
	}

	if client, ok := manager.Clients[datakit.ConnID]; !ok {
		return nil, fmt.Errorf("datakit not available")
	} else {
		res, err := client.doAction(action, datakit, ctx)
		if v, ok := res.(*ws.DCAResponse); ok {
			return v, err
		} else {
			return nil, fmt.Errorf("operation failed: %w", err)
		}
	}
}

func dealNewWebsocketConnection(conn *websocket.Conn, websocketConnID string) error {
	Manager.addWebsocketConnChan(websocketConnID, conn)
	return nil
}

func websocketHandler(c *gin.Context) {
	h := newHandler(c)
	if len(Manager.Register) == cap(Manager.Register) {
		l.Warnf("register full, failed to register client from %s", c.ClientIP())
		h.fatal(http.StatusTooManyRequests, "too many requests")
		return
	}
	var datakit *ws.DataKit

	websocketConnID := c.GetHeader(ws.HeaderNewWebSocketConnectionID)
	if websocketConnID == "" { // first connection
		datakitRaw := c.GetHeader(ws.HeaderDatakit)
		if datakitRaw == "" {
			l.Errorf("invalid datakit header")
			h.c.String(http.StatusBadRequest, "invalid datakit header")
			return
		}

		if err := json.Unmarshal([]byte(datakitRaw), &datakit); err != nil {
			l.Errorf("invalid datakit header: %s", err.Error)
			h.c.String(http.StatusBadRequest, "invalid datakit header")
			return
		}

		if datakit == nil {
			l.Errorf("invalid datakit header, got 0 datakit")
			h.c.String(http.StatusBadRequest, "invalid datakit header")
			return
		}

		if datakit.ConnID == "" || datakit.WorkspaceUUID == "" {
			l.Errorf("invalid datakit workspace or connID")
			h.c.String(http.StatusBadRequest, "invalid datakit header, connID or workspaceUUID is empty")
			return
		}
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		l.Errorf("failed to upgrade websocket connection: %s", err.Error())
		conn.Close() //nolint:errcheck,gosec
		return
	}

	if len(websocketConnID) > 0 { // new connection
		l.Infof("new websocket connection")
		if err := dealNewWebsocketConnection(conn, websocketConnID); err != nil {
			l.Errorf("failed to deal new websocket connection: %s", err.Error())
			conn.Close() //nolint:errcheck,gosec
		}
		return
	}

	client := &Client{
		Socket:            conn,
		Send:              make(chan []byte),
		Receive:           make(map[int64]chan []byte),
		Close:             make(chan interface{}),
		Timeout:           30 * time.Second,
		HeartbeatInterval: time.Second * 30,
		messageNumber:     1,
	}

	datakit.Status = ws.StatusRunning

	client.ID = datakit.ConnID
	client.DataKit = datakit
	client.Socket.SetPingHandler(func(appData string) error {
		if err := datakitDB.Heartbeat(client.ID); err != nil {
			l.Warnf("failed to heartbeat(%s): %s", datakit.HostName, err.Error())
		}

		return nil
	})

	Manager.Register <- client

	g.Go(func(ctx context.Context) error {
		client.Read()
		return nil
	})
	g.Go(func(ctx context.Context) error {
		client.Write()
		return nil
	})
}
