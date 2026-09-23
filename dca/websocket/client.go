// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package websocket implements dca websocket client and server.
package websocket

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/gorilla/websocket"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
)

type DataKitRuntimeInfo struct {
	GlobalHostTags map[string]string `json:"global_host_tags,omitempty"`
	DataDir        string            `json:"data_dir,omitempty"`
	ConfdDir       string            `json:"confd_dir,omitempty"`
	PipelineDir    string            `json:"pipeline_dir,omitempty"`
	InstallDir     string            `json:"install_dir,omitempty"`
	CPUUsage       string            `json:"cpu_usage,omitempty"`
	Log            string            `json:"log,omitempty"`
	GinLog         string            `json:"gin_log,omitempty"`
}

type DataKit struct {
	ID                   string        `json:"id" db:"id"`
	RunTimeID            string        `json:"runtime_id" db:"runtime_id"`
	Arch                 string        `json:"arch" db:"arch"`
	HostName             string        `json:"host_name" db:"host_name"`
	OS                   string        `json:"os" db:"os"`
	Version              string        `json:"version" db:"version"`
	IP                   string        `json:"ip" db:"ip"`
	StartTime            int64         `json:"start_time" db:"start_time"`
	UpdatedAt            int64         `json:"updated_at" db:"updated_at"` // last updated time, unix millisecond
	RunInContainer       bool          `json:"run_in_container" db:"run_in_container"`
	RunMode              string        `json:"run_mode" db:"run_mode"`
	UsageCores           int           `json:"usage_cores" db:"usage_cores"`
	WorkspaceUUID        string        `json:"workspace_uuid" db:"workspace_uuid"`
	ConnID               string        `json:"conn_id" db:"conn_id"`       // to retrieve connection
	Status               DataKitStatus `json:"status" db:"status"`         // status of datakit
	Upgradable           bool          `json:"upgradable" db:"upgradable"` // whether datakit can be upgraded
	URL                  string        `json:"url" db:"url"`               // url of datakit: http://localhost:9529
	GlobalHostTagsString string        `json:"global_host_tags_string" db:"global_host_tags"`
	Config               string        `json:"config" db:"config"` // config value, json string
	DataKitRuntimeInfo
}

func (dk *DataKit) GetConnID(wsAddress string) string {
	hasher := sha256.New()
	hasher.Write([]byte(fmt.Sprintf("conn_id#%s#%s#%s#%s#%s#%s", wsAddress, dk.IP, dk.HostName, dk.OS, dk.Arch, dk.WorkspaceUUID)))
	hasherBytes := hasher.Sum(nil)

	return hex.EncodeToString(hasherBytes)
}

func (dk *DataKit) Bytes() []byte {
	bytes, _ := json.Marshal(dk)
	return bytes
}

func (dk *DataKit) GetGlobalHostTagsString() string {
	globalHostTags := []byte{}

	if dk.GlobalHostTags != nil {
		globalHostTags, _ = json.Marshal(dk.GlobalHostTags)
	}

	return string(globalHostTags)
}

type DataKitStatus string

func (d DataKitStatus) String() string {
	return string(d)
}

const (
	HeaderNewWebSocketConnectionID = "Header_New_WebSocket_Connection_ID"
	HeaderWebsocketAction          = "Header_Websocket_Action"
	HeaderIsUpgraderService        = "dca-is-upgrader-service"
	HeaderDatakit                  = "dca-datakit"

	GetDatakitStatsAction          = "get_datakit_stats_action"
	GetDatakitConfigAction         = "get_datakit_config_action"
	ReloadDatakitAction            = "reload_datakit_action"
	UpgradeDatakitAction           = "upgrade_datakit_action"
	StopDatakitAction              = "stop_datakit_action"
	RestartDatakitAction           = "restart_datakit_action"
	UpdateDatakitStatus            = "update_datakit_status"
	UpdateDatakit                  = "update_datakit"
	DeleteDatakit                  = "delete_datakit"
	SaveDatakitConfigAction        = "save_datakit_config_action"
	DeleteDatakitConfigAction      = "delete_datakit_config_action"
	GetDatakitPipelineAction       = "get_datakit_pipeline_action"
	PatchDatakitPipelineAction     = "patch_datakit_pipeline_action"
	CreateDatakitPipelineAction    = "create_datakit_pipeline_action"
	DeleteDatakitPipelineAction    = "delete_datakit_pipeline_action"
	TestDatakitPipelineAction      = "test_datakit_pipeline_action"
	GetDatakitPipelineDetailAction = "get_datakit_pipeline_detail_action"
	GetDatakitFilterAction         = "get_datakit_filter_action"
	GetDatakitLogTailAction        = "get_datakit_log_tail_action"
	GetDatakitLogDownloadAction    = "get_datakit_log_download_action"

	// create new connection.
	NewWebsocketConnectionAction = "new_websocket_connection_action"

	// datakit status.
	StatusUpgrading  DataKitStatus = "upgrading"
	StatusOffline    DataKitStatus = "offline"
	StatusStopped    DataKitStatus = "stopped"
	StatusRunning    DataKitStatus = "running"
	StatusRestarting DataKitStatus = "restarting"
)

var l = logger.DefaultSLogger("dca_websocket")

// WebsocketMessage is the message format.
type WebsocketMessage struct {
	MessageType int    // message type
	ID          int64  `json:"id"`     // message id
	Action      string `json:"action"` // action name
	Data        any    `json:"data"`   // data
}

func (m *WebsocketMessage) Bytes() []byte {
	bytes, _ := json.Marshal(m)
	return bytes
}

type ActionGetConfigData struct {
	Path string `json:"path"`
}

type ActionData struct {
	Body  string     `json:"body"`
	Query url.Values `json:"query"`
}

// DCAResponse is the response of dca api.
type DCAResponse struct {
	Success   bool        `json:"success"`
	Content   interface{} `json:"content"`
	ErrorCode string      `json:"errorCode"`
	Code      int         `json:"code"`
	Message   string      `json:"message"`
}

// SetSuccess sets the success response.
func (r *DCAResponse) SetSuccess(datas ...interface{}) {
	var data interface{}

	if len(datas) > 0 {
		data = datas[0]
	}

	r.Code = 200
	r.Content = data
	r.Success = true
}

type ResponseError struct {
	Code      int
	ErrorCode string
	ErrorMsg  string
}

func (r *DCAResponse) SetResponse(response *DCAResponse) {
	if response != nil {
		r.Code = response.Code
		r.Content = response.Content
		r.ErrorCode = response.ErrorCode
		r.Message = response.Message
		r.Success = response.Success
	}
}

// SetError sets the error response.
func (r *DCAResponse) SetError(errors ...*ResponseError) {
	var e *ResponseError
	if len(errors) > 0 {
		e = errors[0]
	} else {
		e = &ResponseError{
			Code:      http.StatusInternalServerError,
			ErrorCode: "server.error",
			ErrorMsg:  "",
		}
	}

	code := e.Code
	errorCode := e.ErrorCode
	errorMsg := e.ErrorMsg

	if code == 0 {
		code = http.StatusInternalServerError
	}

	if errorCode == "" {
		errorCode = "server.error"
	}

	if errorMsg == "" {
		errorMsg = "server error"
	}

	r.Code = code
	r.ErrorCode = errorCode
	r.Message = errorMsg
	r.Success = false
}

type ActionHandler func(client *Client, id int64, data any) error

const actionQueueSize = 64

type queuedAction struct {
	conn *websocket.Conn
	data []byte
}

type Client struct {
	sync.RWMutex
	actionQueue       chan queuedAction
	tlsConfig         *tls.Config
	conn              *websocket.Conn
	l                 *logger.Logger
	websocketAddress  string
	heartbeatInterval time.Duration
	close             chan interface{}
	closed            bool
	actionHandlerMap  map[string]ActionHandler
	onInitialized     func()
	sendQueue         chan *WebsocketMessage
	timeout           time.Duration
	heartbeatJob      func(*Client)
	store             map[string]any
	g                 *goroutine.Group
	failCount         int
	datakit           *DataKit
	// lastPong is the time of the last pong received on the current connection.
	// A session can die while the TCP connection stays open (a proxy keeps the
	// client side, the peer stops answering): without it the datakit would stay
	// "connected" to a session that does not exist anymore.
	lastPong atomic.Int64
}

// WithTLSConfig applies the same TLS policy to the control and log connections.
// Pass RootCAs to verify a private CA instead of using legacy insecure TLS.
func WithTLSConfig(config *tls.Config) func(*Client) {
	return func(c *Client) {
		if config != nil {
			c.tlsConfig = config.Clone()
		}
	}
}

func WithDataKit(dk *DataKit) func(*Client) {
	return func(c *Client) {
		c.datakit = dk
	}
}

func WithOnInitialized(f func()) func(*Client) {
	return func(c *Client) {
		c.onInitialized = f
	}
}

func WithHeartbeatJob(job func(c *Client)) func(*Client) {
	return func(c *Client) {
		c.heartbeatJob = job
	}
}

func WithHeartbeatInterval(interval time.Duration) func(*Client) {
	return func(c *Client) {
		c.heartbeatInterval = interval
	}
}

func WithWebsocketAddress(addr string) func(*Client) {
	return func(c *Client) {
		c.websocketAddress = addr
	}
}

func WithLogger(l *logger.Logger) func(*Client) {
	return func(c *Client) {
		c.l = l
	}
}

func WithActionHandlers(handlers map[string]ActionHandler) func(*Client) {
	return func(c *Client) {
		if handlers != nil {
			c.actionHandlerMap = handlers
		}
	}
}

func WithTimeout(timeout time.Duration) func(*Client) {
	return func(c *Client) {
		c.timeout = timeout
	}
}

func NewClient(opts ...func(*Client)) (*Client, error) {
	c := &Client{
		actionQueue:       make(chan queuedAction, actionQueueSize),
		close:             make(chan interface{}),
		heartbeatInterval: 10 * time.Second,
		actionHandlerMap:  map[string]ActionHandler{},
		sendQueue:         make(chan *WebsocketMessage),
		timeout:           30 * time.Second,
		store:             map[string]any{},
		g:                 goroutine.NewGroup(goroutine.Option{Name: "dca_websocket"}),
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.l == nil {
		c.l = l
	}

	if c.websocketAddress == "" {
		return nil, errors.New("websocket address is empty")
	}

	if c.datakit == nil {
		return nil, errors.New("datakit info is missing, cannot create the dca websocket client")
	}

	if c.datakit.WorkspaceUUID == "" {
		return nil, errors.New(
			"datakit workspace uuid is empty: the datakit cannot resolve its workspace, check its dataway token")
	}

	if c.datakit.ConnID == "" {
		return nil, errors.New(
			"datakit conn id is empty: the datakit has no dca websocket server configured")
	}

	return c, nil
}

func (c *Client) Stop() {
	c.Lock()
	defer c.Unlock()
	if c.conn != nil {
		c.conn.Close() //nolint:errcheck,gosec
		c.conn = nil
	}

	if !c.closed {
		close(c.close)
		c.closed = true
	}
}

func (c *Client) Set(key string, value any) {
	c.Lock()
	defer c.Unlock()
	c.store[key] = value
}

func (c *Client) Get(key string) (any, bool) {
	c.RLock()
	defer c.RUnlock()
	v, ok := c.store[key]

	return v, ok
}

func (c *Client) RegisterActionHandler(action string, handler ActionHandler) {
	c.Lock()
	defer c.Unlock()

	if c.actionHandlerMap == nil {
		c.actionHandlerMap = map[string]ActionHandler{
			action: handler,
		}
		c.l.Infof("register action %s", action)
		return
	}

	if _, ok := c.actionHandlerMap[action]; !ok {
		c.actionHandlerMap[action] = handler
		c.l.Infof("register action %s", action)
	} else {
		c.l.Warnf("action %s already registered, ignore", action)
	}
}

func (c *Client) getActionHandler(action string) ActionHandler {
	c.RLock()
	defer c.RUnlock()
	if c.actionHandlerMap == nil {
		return nil
	}
	return c.actionHandlerMap[action]
}

func (c *Client) GetDatakit() *DataKit {
	c.RLock()
	defer c.RUnlock()
	return c.datakit
}

func (c *Client) doAction(message WebsocketMessage, data []byte) error {
	handler := c.getActionHandler(message.Action)

	if handler == nil {
		return fmt.Errorf("unknown action: %s", message.Action)
	} else {
		return handler(c, message.ID, data)
	}
}

func (c *Client) send() {
	for {
		select {
		case <-c.close:
			c.l.Info("exit send")
			return
		case msg := <-c.sendQueue:
			if err := c.doSendMessage(msg); err != nil {
				c.l.Warnf("send message failed: %s", err.Error())
			}
		}
	}
}

func (c *Client) getConn() *websocket.Conn {
	c.RLock()
	defer c.RUnlock()
	return c.conn
}

// dropConn closes the given connection if it is still the current one, so the
// heartbeat opens a new one. Used when the read side reports a dead connection;
// the check keeps a late error of an old connection from closing the new one.
func (c *Client) dropConn(conn *websocket.Conn) {
	c.Lock()
	defer c.Unlock()

	if conn != nil && c.conn == conn {
		conn.Close() //nolint:errcheck,gosec
		c.conn = nil
	}
}

func (c *Client) read() {
	for {
		select {
		case <-c.close:
			c.l.Info("exit read")
			return
		default:
		}

		conn := c.getConn()
		var messageType int
		var message []byte
		var err error
		if conn == nil {
			err = errors.New("connection is nil")
		} else {
			messageType, message, err = conn.ReadMessage()
		}
		if err != nil {
			if conn != nil {
				c.l.Warnf("read failed: %s", err.Error())
				// the connection is gone: drop it so the heartbeat reconnects
				// instead of retrying a dead socket forever
				c.dropConn(conn)
			}
			// without a connection the heartbeat owns reconnecting: wait quietly
			// instead of logging the same failure once per interval
			select {
			case <-c.close:
				c.l.Info("exit read")
				return
			case <-time.After(c.heartbeatInterval):
			}
			continue
		}

		c.l.Debugf("get message: %s", string(message))

		switch messageType {
		case websocket.TextMessage:
			select {
			case <-c.close:
				return
			case c.actionQueue <- queuedAction{conn: conn, data: message}:
			default:
				// Never block the only reader: queued work must not prevent pong handling.
				// Close the overloaded session so pending requests fail explicitly.
				c.l.Warn("action queue is full, closing websocket connection")
				c.dropConn(conn)
			}
		default:
			c.l.Warnf("message type: %v, ignored", messageType)
		}
	}
}

// runActions preserves action order while the reader keeps processing control frames.
// On shutdown the group waits for the in-flight action; queued actions are discarded.
func (c *Client) runActions() {
	for {
		select {
		case <-c.close:
			return
		case action := <-c.actionQueue:
			select {
			case <-c.close:
				return
			default:
			}
			if c.getConn() != action.conn {
				continue
			}
			c.dealMessage(action.data)
		}
	}
}

func (c *Client) dealMessage(rawData []byte) {
	message := WebsocketMessage{}
	var err error
	err = json.Unmarshal(rawData, &message)
	if err != nil {
		c.l.Warnf("failed to unmarshal message: %s, ignore", err.Error())
		return
	}

	if err = c.doAction(message, rawData); err != nil {
		c.l.Warnf("failed to do action: %s, ignore", err.Error())
		response := DCAResponse{}
		response.SetError(&ResponseError{
			ErrorCode: "invalid.action",
			ErrorMsg:  "action failed",
		})

		if err := c.SendMessage(&WebsocketMessage{
			ID: message.ID, Action: message.Action, Data: response,
		}); err != nil {
			c.l.Warnf("send action error failed: %s", err)
		}
	}
}

func (c *Client) GetWebsocketAddress() string {
	return c.websocketAddress
}

func (c *Client) SetDatakit(dk *DataKit) {
	if dk == nil {
		c.l.Warnf("datakit is nil, not to update the old datakit")
		return
	}

	c.Lock()
	defer c.Unlock()
	c.datakit = dk
}

func (c *Client) doSendMessage(message *WebsocketMessage) error {
	if message == nil {
		return errors.New("empty message")
	}

	conn := c.getConn()
	if conn == nil {
		c.l.Warnf("websocket connection is nil, ignore message")
		return nil
	}

	if message.MessageType == 0 {
		message.MessageType = websocket.TextMessage
	}

	if err := conn.SetWriteDeadline(time.Now().Add(c.timeout)); err != nil {
		return fmt.Errorf("set write deadline failed: %w", err)
	}
	if err := conn.WriteMessage(message.MessageType, message.Bytes()); err != nil {
		c.l.Warnf("send message failed: %s, ignore", err.Error())
		return fmt.Errorf("send message failed: %w", err)
	}

	return nil
}

func (c *Client) SendMessage(message *WebsocketMessage) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()
	select {
	case <-c.close:
		return errors.New("dca websocket client stopped")
	case <-ctx.Done():
		return fmt.Errorf("send message timeout")
	case c.sendQueue <- message:
		return nil
	}
}

// dialTimeout bounds one dial/handshake attempt. Without it an unreachable (or
// black holed) DCA endpoint blocks the attempt for the dialer default.
const dialTimeout = 10 * time.Second

// Dial opens a connection using this client's TLS policy, without changing the
// global dialer. Log sub-connections must use this too, including private CAs.
func (c *Client) Dial(header http.Header) (*websocket.Conn, *http.Response, error) {
	config := c.tlsConfig
	if config == nil {
		// Preserve the existing DCA self-signed certificate compatibility.
		config = &tls.Config{InsecureSkipVerify: true} //nolint:gosec
	} else {
		config = config.Clone()
	}
	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: dialTimeout,
		NetDialContext:   (&net.Dialer{Timeout: dialTimeout, KeepAlive: 30 * time.Second}).DialContext,
		TLSClientConfig:  config,
	}
	return dialer.Dial(c.websocketAddress, header)
}

// Init websocket connection. Close the old connection if it exists.
func (c *Client) init() error {
	header := make(http.Header)
	header.Set(HeaderDatakit, string(c.GetDatakit().Bytes()))

	conn, resp, err := c.Dial(header)
	if resp != nil {
		defer resp.Body.Close() //nolint: errcheck
	}
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}

	if resp.StatusCode != http.StatusSwitchingProtocols {
		if body, err := io.ReadAll(resp.Body); err != nil {
			return fmt.Errorf("read body failed: %w", err)
		} else {
			return fmt.Errorf("request failed: %s", string(body))
		}
	}

	conn.SetPongHandler(func(string) error {
		c.RLock()
		defer c.RUnlock()
		if c.conn == conn {
			c.lastPong.Store(time.Now().UnixNano())
		}
		return nil
	})

	c.Lock()
	if c.closed {
		c.Unlock()
		conn.Close() //nolint:errcheck,gosec
		return errors.New("dca websocket client stopped")
	}
	old := c.conn
	c.lastPong.Store(0) // new connection: no pong seen yet
	c.conn = conn
	c.Unlock()

	if old != nil {
		old.Close() //nolint:errcheck,gosec
	}

	c.l.Infof("websocket connection established")
	return nil
}

func (c *Client) Restart() {
	c.l.Info("restart websocket connection")
	c.Stop()
	_ = c.g.Wait() // wait all goroutine exit
	c.Start()
}

func (c *Client) heartbeat() {
	ticker := time.NewTicker(c.heartbeatInterval)
	defer ticker.Stop()

	for {
		if err := c.doHeartbeat(); err != nil {
			c.l.Warnf("heartbeat failed: %s", err.Error())
			if err := c.init(); err != nil {
				c.failCount++
				c.l.Warnf("failed to init websocket: %s, fail %d times", err.Error(), c.failCount)
			} else {
				c.failCount = 0
				// Reconnect backoff must never become the healthy connection heartbeat period.
				ticker.Reset(c.heartbeatInterval)
			}
		}
		if c.failCount > 0 {
			ticker.Reset(c.backoffDelay())
		}
		select {
		case <-ticker.C:
		case <-c.close:
			c.l.Infof("exit heartbeat")
			return
		}
	}
}

// backoffDelay returns how long to wait before the next heartbeat/reconnect
// attempt: exponential from the heartbeat interval, capped, plus jitter. A host
// that cannot reach the DCA endpoint keeps retrying forever, so the delay must
// grow instead of hammering the entry on every heartbeat.
func (c *Client) backoffDelay() time.Duration {
	shift := c.failCount - 1
	if shift > 5 {
		shift = 5
	}

	delay := c.heartbeatInterval * time.Duration(1<<shift)
	if maxDelay := 5 * time.Minute; delay > maxDelay {
		delay = maxDelay
	}

	if jitter := int64(delay) / 5; jitter > 0 {
		delay += time.Duration(rand.Int63n(jitter)) // up to 20% jitter
	}

	return delay
}

func (c *Client) doHeartbeat() error {
	if c.heartbeatJob != nil {
		c.heartbeatJob(c)
	}

	conn := c.getConn()
	if conn == nil {
		return fmt.Errorf("conn is nil")
	}

	// The write below only fails once the TCP stack gives up: with a black holed
	// path (or a proxy that keeps the client side) it can succeed for a long
	// time while the session on the peer is already gone. Once the DCA answered
	// a ping on this connection, require the next answer within a few heartbeat
	// intervals. Servers that never answer pings (older DCA) are unaffected.
	if last := c.lastPong.Load(); last != 0 {
		if age := time.Since(time.Unix(0, last)); age > 3*c.heartbeatInterval {
			return fmt.Errorf("no pong for %s, session seems gone", age.Round(time.Second))
		}
	}

	if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(c.timeout)); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	return nil
}

func (c *Client) Start() {
	c.g.Go(func(ctx context.Context) error {
		c.runActions()
		return nil
	})

	c.g.Go(func(ctx context.Context) error {
		c.heartbeat()
		return nil
	})

	c.g.Go(func(ctx context.Context) error {
		c.read()
		return nil
	})

	c.g.Go(func(ctx context.Context) error {
		c.send()
		return nil
	})
}

// HandlerFunc is a function type, which is used to handle each action.
type HandlerFunc func(client *Client, response *DCAResponse, data *ActionData, datakit *DataKit)

// GetActionHandler returns a ActionHandler with handlerFunc.
func GetActionHandler(action string, handle HandlerFunc) ActionHandler {
	return func(client *Client, id int64, data any) error {
		response := &DCAResponse{}
		messageData := &ActionData{}
		msg := WebsocketMessage{
			Data: &messageData,
		}

		if err := json.Unmarshal(data.([]byte), &msg); err != nil {
			return fmt.Errorf("failed to unmarshal data: %w", err)
		}

		dk := client.GetDatakit()

		if dk == nil {
			return fmt.Errorf("datakit is required")
		}

		handle(client, response, messageData, dk)

		message := &WebsocketMessage{
			ID:     id,
			Action: action,
			Data:   response,
		}
		return client.SendMessage(message)
	}
}
