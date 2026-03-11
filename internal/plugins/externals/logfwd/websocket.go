// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package logfwd

import (
	"encoding/json"
	"net/url"
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

func (m *message) json() ([]byte, error) {
	return json.Marshal(m)
}

type websocketClient struct {
	u      *url.URL
	conn   *websocket.Conn
	dataCh chan []byte
}

func newWebsocketClient(u *url.URL) *websocketClient {
	return &websocketClient{
		u:      u,
		dataCh: make(chan []byte, 64),
	}
}

func (w *websocketClient) start() {
	for data := range w.dataCh {
		if err := w.conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Errorf("client write failed: %s", err.Error())
			w.tryConnectWebsocketServer()
		}
	}
}

func (w *websocketClient) close() error {
	if w.conn == nil {
		return nil
	}
	return w.conn.Close()
}

func (w *websocketClient) tryConnectWebsocketServer() {
	for {
		wscli, _, err := websocket.DefaultDialer.Dial(w.u.String(), nil)
		if err != nil {
			log.Errorf("websocket connection failed: %s", err.Error())
			time.Sleep(time.Second)
			continue
		}
		w.conn = wscli
		log.Info("websocket connected")
		return
	}
}

func (w *websocketClient) writeMessage(data []byte) error {
	w.dataCh <- data
	return nil
}
