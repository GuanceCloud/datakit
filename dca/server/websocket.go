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
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	ws "gitlab.jiagouyun.com/cloudcare-tools/datakit/dca/websocket"
)

var ErrRequestTimeout = errors.New("request_time_out")

// ErrDatakitOffline means the datakit has no live websocket session: the web UI
// should refresh the list instead of retrying the operation.
var ErrDatakitOffline = errors.New("datakit not available")

// sessionReadTimeout is how long a session may stay silent before it is dropped.
// Keepalive pings from the datakit count as traffic, otherwise an idle but
// healthy session would be dropped on every timeout.
var sessionReadTimeout = 90 * time.Second

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
	ReadTimeout       time.Duration // drop the session when nothing is received for that long

	messageNumber int64
	closeOnce     sync.Once
	// exited is set before Exit queues the unregistration, so the manager can
	// tell whether a registration is still wanted (Close is closed too late).
	exited atomic.Bool
	// Immutable registration identity, kept even if a later push changes DataKit.
	logContext           string
	registrationRejected atomic.Bool
	// replacementFor authorizes replacing only the owner that was probed.
	// It is written before this client is sent to Register.
	replacementFor *Client
	probing        atomic.Bool
	probeNumber    uint64
	probeToken     string
	probePong      chan struct{}
	retired        atomic.Bool
	counted        bool // protected by the manager lock
	metricLabels   [2]string
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
		// Pending/replaced sockets can still deliver buffered messages. Keep the
		// ownership check and DB mutation atomic with session replacement.
		Manager.Lock()
		defer Manager.Unlock()
		if Manager.Clients[c.ID] == c && !c.exited.Load() {
			doCommonAction(c, &msg)
		}
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
		if c.ReadTimeout > 0 {
			if err := c.Socket.SetReadDeadline(time.Now().Add(c.ReadTimeout)); err != nil {
				if !c.registrationRejected.Load() && !c.retired.Load() {
					l.Warnf("set read deadline failed: %s, %s", err.Error(), c.logContext)
				}
				break
			}
		}

		messageType, message, err := c.Socket.ReadMessage()
		if err != nil {
			if c.registrationRejected.Load() || c.retired.Load() {
				l.Debugf("retired or rejected session closed: %s, %s", err.Error(), c.logContext)
			} else {
				l.Warnf("read message failed: %s, %s", err.Error(), c.logContext)
			}
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

// extendReadDeadline pushes the read deadline forward: called for every
// keepalive frame so that an idle session stays alive as long as the datakit
// keeps sending pings.
func (c *Client) extendReadDeadline() {
	if c.ReadTimeout <= 0 {
		return
	}

	if err := c.Socket.SetReadDeadline(time.Now().Add(c.ReadTimeout)); err != nil {
		l.Warnf("failed to extend read deadline: %s", err.Error())
	}
}

func (c *Client) Exit() {
	c.closeOnce.Do(func() {
		// must stay before the unregistration is queued: the manager may pick a
		// queued registration up afterwards and needs to know the session is gone
		c.exited.Store(true)
		Manager.Unregister <- c
		c.Socket.Close() // nolint:errcheck,gosec
		close(c.Close)
	})
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
	case <-c.Close:
		return ErrDatakitOffline
	case <-ctx.Done():
		return ErrRequestTimeout
	}

	ctx1, cancel1 := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel1()
	select {
	case <-c.Close:
		return ErrDatakitOffline
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
				}
				c.Exit()
				return
			}

			if err := c.Socket.SetWriteDeadline(time.Now().Add(c.Timeout)); err != nil {
				l.Errorf("set write deadline failed: %s", err.Error())
				c.Exit()
				return
			}
			if err := c.Socket.WriteMessage(websocket.TextMessage, message); err != nil {
				l.Warnf("failed to write message: %s", err.Error())
				c.Exit()
				return
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
	CheckOrigin:     checkWebsocketOrigin,
}

func checkWebsocketOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}

	originURL, err := url.Parse(origin)
	if err != nil || originURL.Host == "" {
		l.Warnf("reject websocket origin %q: invalid origin", origin)
		return false
	}

	originHost := normalizeWebsocketHost(originURL.Host)
	requestHost := normalizeWebsocketHost(r.Host)
	if originHost == "" || requestHost == "" {
		l.Warnf("reject websocket origin %q for host %q", origin, r.Host)
		return false
	}

	if originHost == requestHost {
		return true
	}

	if isLocalWebsocketHost(originHost) && isLocalWebsocketHost(requestHost) {
		return true
	}

	l.Warnf("reject websocket origin %q for host %q", origin, r.Host)
	return false
}

func normalizeWebsocketHost(hostport string) string {
	hostport = strings.TrimSpace(hostport)
	host := hostport
	if h, _, err := net.SplitHostPort(hostport); err == nil {
		host = h
	} else if idx := strings.LastIndex(hostport, ":"); idx > -1 && !strings.Contains(hostport[:idx], ":") {
		if _, err := strconv.Atoi(hostport[idx+1:]); err == nil {
			host = hostport[:idx]
		}
	}

	return strings.ToLower(strings.Trim(host, "[]"))
}

func isLocalWebsocketHost(host string) bool {
	host = normalizeWebsocketHost(host)
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}

// Message is return msg.
type Message struct {
	Sender    string `json:"sender,omitempty"`
	Recipient string `json:"recipient,omitempty"`
	Content   string `json:"content,omitempty"`
}

func (manager *ClientManager) addWebsocketConnChan(connID string, conn *websocket.Conn) error {
	manager.Lock()
	defer manager.Unlock()
	if ch, ok := manager.WebsocketConns[connID]; ok {
		select {
		case ch <- conn:
			return nil
		default:
			return fmt.Errorf("websocket conn chan is full, connID: %s", connID)
		}
	}

	return fmt.Errorf("websocket conn chan not existed, connID: %s", connID)
}

func (manager *ClientManager) initWebsocketConnChan(connID string) chan *websocket.Conn {
	manager.Lock()
	defer manager.Unlock()
	manager.WebsocketConns[connID] = make(chan *websocket.Conn, 1)
	return manager.WebsocketConns[connID]
}

func (manager *ClientManager) deleteWebsocketConnChan(connID string) {
	manager.Lock()
	defer manager.Unlock()
	if ch, ok := manager.WebsocketConns[connID]; ok {
		close(ch)
		delete(manager.WebsocketConns, connID)
	} else {
		l.Warnf("websocket conn chan not existed, connID: %s", connID)
	}
}

func (manager *ClientManager) Start() {
	manager.run(context.Background())
}

func (manager *ClientManager) run(ctx context.Context) {
	l.Infof("websocket manager started")
	for {
		select {
		case <-ctx.Done():
			return
		case conn := <-manager.Register: // new connection
			if !manager.registerClient(conn) {
				conn.registrationRejected.Store(true)
				conn.Socket.Close() //nolint:errcheck,gosec
				continue
			}

		case conn := <-manager.Unregister:
			manager.unregisterClient(conn)
			conn.Socket.Close() //nolint:errcheck,gosec
		}
	}
}

// registerClient checks for duplicated connections, persists the datakit and
// publishes the session, all under the manager lock: the duplicate check and
// the registration must be atomic, and Clients must not be touched without the
// lock (the HTTP handlers read it concurrently).
func (manager *ClientManager) registerClient(conn *Client) bool {
	manager.Lock()
	defer manager.Unlock()

	// Register and unregister are two independent channels: the unregistration
	// of a client may be processed before its registration. A client that
	// already exited must not be published as a live session, otherwise its row
	// would stay "running" without any session to route the actions to.
	if conn.exited.Load() {
		return false
	}

	if conn.DataKit == nil {
		l.Errorf("refuse to register a client without datakit")
		return false
	}

	// The socket owner, not a persisted status, decides whether this is a
	// duplicate. An orphaned running/stopped row must not reject a reconnect;
	// conversely an offline row must not let a second socket steal a session.
	owner := manager.Clients[conn.ID]
	if owner != nil && owner != conn.replacementFor && !owner.exited.Load() {
		l.Infof("datakit connection already exists: owner_exited=%t incoming={%s} owner={%s}",
			owner.exited.Load(), conn.logContext, owner.logContext)
		return false
	}

	if err := datakitDB.ForceUpdate(conn.DataKit); err != nil { // update datakit
		l.Errorf("failed to insert datakit: %s", err.Error())
		return false
	}

	manager.Clients[conn.ID] = conn
	if owner != nil {
		owner.retired.Store(true)
		owner.exited.Store(true)
		owner.updateSessionCount(false)
		// Do not call Exit under the manager lock: it queues an unregister.
		// Closing the socket wakes its reader; late events are fenced by owner.
		if owner.Socket != nil {
			owner.Socket.Close() //nolint:errcheck,gosec
		}
		l.Infof("replaced unresponsive session: incoming={%s} previous={%s}", conn.logContext, owner.logContext)
	}
	conn.updateSessionCount(true)
	l.Infof("new connection registered: %s, %s", conn.DataKit.HostName, conn.logContext)
	return true
}

// unregisterClient only updates the row owned by this exact socket. Rejected
// or replaced sockets must not touch it, even if the successor has since exited.
// Rows without a session are reclaimed by registration, not by late unregisters.
func (manager *ClientManager) unregisterClient(conn *Client) bool {
	manager.Lock()
	defer manager.Unlock()

	live, ok := manager.Clients[conn.ID]
	if !ok || live != conn {
		return false
	}

	if err := datakitDB.DeleteByConnID(conn.ID, conn.DataKit.RunInContainer); err != nil {
		l.Errorf("failed to delete datakit: %s", err.Error())
	}

	delete(manager.Clients, conn.ID)
	conn.updateSessionCount(false)
	l.Infof("connection unregistered: %s, %s", conn.DataKit.HostName, conn.logContext)

	return true
}

// updateSessionCount is called while holding the manager lock. Keep the labels
// from registration so a metadata update cannot leave a stale gauge behind.
func (c *Client) updateSessionCount(active bool) {
	if active && !c.counted {
		c.metricLabels = [2]string{c.DataKit.HostName, c.DataKit.OS}
		datakitTotalGauge.WithLabelValues(c.metricLabels[:]...).Inc()
	} else if !active && c.counted {
		datakitTotalGauge.WithLabelValues(c.metricLabels[:]...).Dec()
	}
	c.counted = active
}

// getClient returns the live session of a datakit.
func (manager *ClientManager) getClient(connID string) (*Client, bool) {
	manager.RLock()
	defer manager.RUnlock()

	client, ok := manager.Clients[connID]
	return client, ok
}

func (manager *ClientManager) Action(action string, datakit *ws.DataKit, ctx *gin.Context) (*ws.DCAResponse, error) {
	if datakit == nil {
		return nil, fmt.Errorf("datakit is required")
	}

	client, ok := manager.getClient(datakit.ConnID)
	if !ok {
		return nil, fmt.Errorf("%w: no live session, please refresh the datakit list", ErrDatakitOffline)
	}

	start := time.Now()
	res, err := client.doAction(action, datakit, ctx)
	if v, ok := res.(*ws.DCAResponse); ok {
		websocketElapsedVec.WithLabelValues(
			datakit.HostName,
			action,
			fmt.Sprintf("%d", v.Code),
		).Observe(time.Since(start).Seconds())
		return v, err
	}

	return nil, fmt.Errorf("operation failed: %w", err)
}

func dealNewWebsocketConnection(conn *websocket.Conn, websocketConnID string) error {
	if err := Manager.addWebsocketConnChan(websocketConnID, conn); err != nil {
		conn.Close() //nolint:errcheck,gosec
		return err
	}
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
		if conn != nil {
			conn.Close() //nolint:errcheck,gosec
		}
		return
	}

	if len(websocketConnID) > 0 { // new connection
		l.Infof("new websocket connection")
		if err := dealNewWebsocketConnection(conn, websocketConnID); err != nil {
			l.Errorf("failed to deal new websocket connection: %s", err.Error())
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
		ReadTimeout:       sessionReadTimeout,
		messageNumber:     1,
	}

	datakit.Status = ws.StatusRunning

	client.ID = datakit.ConnID
	client.DataKit = datakit
	client.logContext = fmt.Sprintf("host=%q conn_id=%s runtime_id=%s workspace=%s peer=%s",
		datakit.HostName, client.ID, datakit.RunTimeID, datakit.WorkspaceUUID, conn.RemoteAddr())
	client.Socket.SetPingHandler(func(appData string) error {
		// A datakit that has nothing to report only sends keepalives: they must
		// extend the read deadline, otherwise an idle (but healthy) session is
		// dropped on every timeout.
		client.extendReadDeadline()

		// A busy SQLite writer must not delay the pong and make the peer
		// consider this working connection dead before its heartbeat is saved.
		if err := client.Socket.WriteControl(
			websocket.PongMessage, []byte(appData), time.Now().Add(client.Timeout)); err != nil {
			return err
		}
		Manager.heartbeatClient(client)
		return nil
	})

	client.Socket.SetPongHandler(func(appData string) error {
		client.receiveProbePong(appData)
		client.extendReadDeadline()
		return nil
	})

	g.Go(func(ctx context.Context) error {
		client.Read()
		return nil
	})
	g.Go(func(ctx context.Context) error {
		client.Write()
		return nil
	})

	// Probe in this request's goroutine, outside the global registration loop.
	// Other hosts can register or unregister while a previous owner is checked.
	if !Manager.prepareRegistration(client) {
		client.registrationRejected.Store(true)
		client.Socket.Close() //nolint:errcheck,gosec
		return
	}
	Manager.Register <- client
}
