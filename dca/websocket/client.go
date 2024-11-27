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
	"math"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/gorilla/websocket"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
)

type DataKitRuntimeInfo struct {
	GlobalHostTags map[string]string `json:"global_host_tags"`
	DataDir        string            `json:"data_dir"`
	ConfdDir       string            `json:"confd_dir"`
	PipelineDir    string            `json:"pipeline_dir"`
	InstallDir     string            `json:"install_dir"`
	CPUUsage       string            `json:"cpu_usage"`
	Log            string            `json:"log"`
	GinLog         string            `json:"gin_log"`
}

type DataKit struct {
	ID             string        `json:"id" db:"id"`
	RunTimeID      string        `json:"runtime_id" db:"runtime_id"`
	Arch           string        `json:"arch" db:"arch"`
	HostName       string        `json:"host_name" db:"host_name"`
	OS             string        `json:"os" db:"os"`
	Version        string        `json:"version" db:"version"`
	IP             string        `json:"ip" db:"ip"`
	StartTime      int64         `json:"start_time" db:"start_time"`
	UpdatedAt      int64         `json:"updated_at" db:"updated_at"` // last updated time, unix millisecond
	RunInContainer bool          `json:"run_in_container" db:"run_in_container"`
	RunMode        string        `json:"run_mode" db:"run_mode"`
	UsageCores     int           `json:"usage_cores" db:"usage_cores"`
	WorkspaceUUID  string        `json:"workspace_uuid" db:"workspace_uuid"`
	ConnID         string        `json:"conn_id" db:"conn_id"`       // to retrieve connection
	Status         DataKitStatus `json:"status" db:"status"`         // status of datakit
	Upgradable     bool          `json:"upgradable" db:"upgradable"` // whether datakit can be upgraded
	URL            string        `json:"url" db:"url"`               // url of datakit: http://localhost:9529
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

type Client struct {
	sync.RWMutex
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

	if c.datakit == nil || c.datakit.WorkspaceUUID == "" || c.datakit.ConnID == "" {
		return nil, errors.New("datakit is nil, or workspace uuid or conn id is empty")
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

func (c *Client) doReadMessage() (messageType int, p []byte, err error) {
	conn := c.getConn()
	if conn == nil {
		return 0, nil, errors.New("connection is nil")
	}
	return conn.ReadMessage()
}

func (c *Client) read() {
	for {
		select {
		case <-c.close:
			c.l.Info("exit read")
			return
		default:
		}

		messageType, message, err := c.doReadMessage()
		if err != nil {
			c.l.Warnf("read failed: %s", err.Error())
			time.Sleep(c.heartbeatInterval)
			continue
		}

		c.l.Debugf("get message: %s", string(message))

		switch messageType {
		case websocket.TextMessage:
			c.dealMessage(message)
		default:
			c.l.Warnf("message type: %v, ignored", messageType)
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

		c.sendQueue <- &WebsocketMessage{
			ID:     message.ID,
			Action: message.Action,
			Data:   response,
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
	case <-ctx.Done():
		return fmt.Errorf("send message timeout")
	case c.sendQueue <- message:
		return nil
	}
}

// Init websocket connection. Close the old connection if it exists.
func (c *Client) init() error {
	c.Lock()
	defer c.Unlock()
	if c.conn != nil {
		c.conn.Close() // nolint:errcheck,gosec
		c.conn = nil
	}

	header := make(http.Header)
	header.Set(HeaderDatakit, string(c.datakit.Bytes()))

	dialer := websocket.DefaultDialer
	if strings.HasPrefix(c.websocketAddress, "wss://") { // tls
		dialer.TLSClientConfig = &tls.Config{RootCAs: nil, InsecureSkipVerify: true} // nolint:gosec
	}

	conn, resp, err := dialer.Dial(c.websocketAddress, header)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}

	defer resp.Body.Close() // nolint: errcheck
	if resp.StatusCode != http.StatusSwitchingProtocols {
		if body, err := io.ReadAll(resp.Body); err != nil {
			return fmt.Errorf("read body failed: %w", err)
		} else {
			return fmt.Errorf("request failed: %s", string(body))
		}
	}

	c.l.Infof("websocket connection established")
	c.conn = conn
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
			}
		}
		if c.failCount > 0 {
			ticker.Reset(c.heartbeatInterval * time.Duration(math.Log2(float64(2+c.failCount))))
		}
		// avoid too large fail count
		if c.failCount > 1000 {
			c.failCount = 10
		}
		select {
		case <-ticker.C:
		case <-c.close:
			c.l.Infof("exit heartbeat")
			return
		}
	}
}

func (c *Client) doHeartbeat() error {
	if c.heartbeatJob != nil {
		c.heartbeatJob(c)
	}

	conn := c.getConn()
	if conn == nil {
		return fmt.Errorf("conn is nil")
	}
	if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(c.timeout)); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	return nil
}

func (c *Client) Start() {
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
