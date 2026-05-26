// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package logfwd

import (
	"encoding/json"
	"errors"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type message struct {
	Type         string                 `json:"type"`
	Source       string                 `json:"source"`
	StorageIndex string                 `json:"storage_index,omitempty"`
	Pipeline     string                 `json:"pipeline,omitempty"`
	Tags         map[string]string      `json:"tags,omitempty"`
	Fields       map[string]interface{} `json:"fields,omitempty"`
	Log          string                 `json:"log"`
}

var (
	errWebsocketClientClosed = errors.New("websocket client closed")
	errWebsocketQueueFull    = errors.New("websocket message queue full")
)

func (m *message) json() ([]byte, error) {
	return json.Marshal(m)
}

type websocketClient struct {
	u       *url.URL
	conn    *websocket.Conn
	dataCh  chan []byte
	closeCh chan struct{}

	mu     sync.RWMutex
	closed bool
}

func newWebsocketClient(u *url.URL) *websocketClient {
	return &websocketClient{
		u:       u,
		dataCh:  make(chan []byte, 64),
		closeCh: make(chan struct{}),
	}
}

func (w *websocketClient) start() {
	for {
		select {
		case <-w.closeCh:
			return
		case data := <-w.dataCh:
			if err := w.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Errorf("client write failed: %s", err.Error())
				if err := w.tryConnectWebsocketServer(); err != nil {
					return
				}
			}
		}
	}
}

func (w *websocketClient) close() error {
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return nil
	}
	w.closed = true
	close(w.closeCh)
	conn := w.conn
	w.mu.Unlock()

	if conn == nil {
		return nil
	}
	return conn.Close()
}

func (w *websocketClient) tryConnectWebsocketServer() error {
	for {
		select {
		case <-w.closeCh:
			return errWebsocketClientClosed
		default:
		}

		wscli, _, err := websocket.DefaultDialer.Dial(w.u.String(), nil)
		if err != nil {
			log.Errorf("websocket connection failed: %s", err.Error())
			select {
			case <-w.closeCh:
				return errWebsocketClientClosed
			case <-time.After(time.Second):
			}
			continue
		}

		w.mu.Lock()
		if w.closed {
			w.mu.Unlock()
			_ = wscli.Close()
			return errWebsocketClientClosed
		}
		w.conn = wscli
		w.mu.Unlock()

		log.Info("websocket connected")
		return nil
	}
}

func (w *websocketClient) writeMessage(data []byte) error {
	w.mu.RLock()
	closed := w.closed
	w.mu.RUnlock()
	if closed {
		return errWebsocketClientClosed
	}

	select {
	case <-w.closeCh:
		return errWebsocketClientClosed
	case w.dataCh <- data:
		return nil
	default:
		return errWebsocketQueueFull
	}
}
