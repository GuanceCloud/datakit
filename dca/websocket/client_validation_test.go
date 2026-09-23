// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package websocket

import (
	"strings"
	"testing"
)

// TestNewClientValidationErrors: the datakit info handed over by the upgrader
// is not always complete (no workspace uuid when the datakit cannot resolve its
// workspace, no conn id when DCA is not configured). Each case must be reported
// with its own message, otherwise the field operators have to guess.
func TestNewClientValidationErrors(t *testing.T) {
	const addr = "ws://127.0.0.1:1/ws"

	t.Run("empty-websocket-address", func(t *testing.T) {
		_, err := NewClient(WithDataKit(&DataKit{WorkspaceUUID: "wksp-1", ConnID: "conn-1"}))
		if err == nil || !strings.Contains(err.Error(), "websocket address") {
			t.Fatalf("got %v, want a websocket address error", err)
		}
	})

	t.Run("missing-datakit", func(t *testing.T) {
		_, err := NewClient(WithWebsocketAddress(addr))
		if err == nil || !strings.Contains(err.Error(), "datakit info is missing") {
			t.Fatalf("got %v, want a missing datakit error", err)
		}
	})

	t.Run("empty-workspace-uuid", func(t *testing.T) {
		_, err := NewClient(
			WithWebsocketAddress(addr),
			WithDataKit(&DataKit{HostName: "host-1", ConnID: "conn-1"}),
		)
		if err == nil || !strings.Contains(err.Error(), "workspace uuid is empty") {
			t.Fatalf("got %v, want an empty workspace uuid error", err)
		}
	})

	t.Run("empty-conn-id", func(t *testing.T) {
		_, err := NewClient(
			WithWebsocketAddress(addr),
			WithDataKit(&DataKit{HostName: "host-1", WorkspaceUUID: "wksp-1"}),
		)
		if err == nil || !strings.Contains(err.Error(), "conn id is empty") {
			t.Fatalf("got %v, want an empty conn id error", err)
		}
	})

	t.Run("valid", func(t *testing.T) {
		c, err := NewClient(
			WithWebsocketAddress(addr),
			WithDataKit(&DataKit{HostName: "host-1", WorkspaceUUID: "wksp-1", ConnID: "conn-1"}),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c == nil {
			t.Fatal("client must not be nil")
		}
	})
}
