// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package logfwd

import (
	"net/url"
	"strings"
	"testing"
)

func TestWebsocketWriteMessageQueueFull(t *testing.T) {
	wsClient := newWebsocketClient(&url.URL{})
	for i := 0; i < cap(wsClient.dataCh); i++ {
		wsClient.dataCh <- []byte("x")
	}

	err := wsClient.writeMessage([]byte("overflow"))
	if err == nil {
		t.Fatal("writeMessage() error = nil, want queue full error")
	}

	if !strings.Contains(err.Error(), "queue full") {
		t.Fatalf("writeMessage() error = %v, want queue full", err)
	}
}

func TestWebsocketWriteMessageAfterClose(t *testing.T) {
	wsClient := newWebsocketClient(&url.URL{})

	if err := wsClient.close(); err != nil {
		t.Fatalf("close() error = %v", err)
	}

	err := wsClient.writeMessage([]byte("closed"))
	if err == nil {
		t.Fatal("writeMessage() error = nil, want closed error")
	}

	if !strings.Contains(err.Error(), "closed") {
		t.Fatalf("writeMessage() error = %v, want closed", err)
	}
}
