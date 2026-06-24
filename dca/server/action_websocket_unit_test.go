// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	ws "gitlab.jiagouyun.com/cloudcare-tools/datakit/dca/websocket"
)

func newTestClientForRequest() *Client {
	return &Client{
		ID:      "client-1",
		Send:    make(chan []byte, 1),
		Receive: map[int64]chan []byte{},
		Close:   make(chan interface{}),
		Timeout: time.Second,
		DataKit: newTestDataKit("conn-client"),

		messageNumber: 1,
	}
}

func TestClientMessageChannels(t *testing.T) {
	client := newTestClientForRequest()
	client.messageNumber = 0

	id, ch := client.getMessageNumber()
	require.Equal(t, int64(0), id)
	require.NotNil(t, ch)

	sendCh, err := client.getMessageCh(id)
	require.NoError(t, err)
	done := make(chan struct{})
	go func() {
		sendCh <- []byte("ok")
		close(done)
	}()
	require.Equal(t, []byte("ok"), <-ch)
	<-done

	client.releaseMessageCh(id)
	_, err = client.getMessageCh(id)
	require.Error(t, err)

	client.messageNumber = math.MaxInt64
	nextID, _ := client.getMessageNumber()
	require.Equal(t, int64(1), nextID)
}

func TestClientRequest(t *testing.T) {
	client := newTestClientForRequest()

	go func() {
		raw := <-client.Send
		msg := ws.WebsocketMessage{Data: &ws.ActionData{}}
		require.NoError(t, json.Unmarshal(raw, &msg))
		resp := ws.WebsocketMessage{
			ID:     msg.ID,
			Action: msg.Action,
			Data:   &ws.DCAResponse{Success: true, Code: 200},
		}
		client.receiveMessage(resp.Bytes())
	}()

	out := ws.WebsocketMessage{Data: &ws.DCAResponse{}}
	err := client.request(&ws.WebsocketMessage{Action: ws.GetDatakitStatsAction}, &out)
	require.NoError(t, err)
	require.True(t, out.Data.(*ws.DCAResponse).Success)
	require.Empty(t, client.Receive)
}

func TestClientRequestErrorBranches(t *testing.T) {
	client := newTestClientForRequest()
	client.Timeout = 10 * time.Millisecond

	require.Error(t, client.request(nil, &ws.WebsocketMessage{}))
	require.Error(t, client.request(&ws.WebsocketMessage{}, nil))

	err := client.request(&ws.WebsocketMessage{Action: "timeout"}, &ws.WebsocketMessage{Data: &ws.DCAResponse{}})
	require.ErrorIs(t, err, ErrRequestTimeout)

	client = newTestClientForRequest()
	go func() {
		raw := <-client.Send
		msg := ws.WebsocketMessage{Data: &ws.ActionData{}}
		require.NoError(t, json.Unmarshal(raw, &msg))
		client.receiveMessage([]byte(`{bad json`))
	}()
	err = client.request(&ws.WebsocketMessage{Action: "bad-json"}, &ws.WebsocketMessage{Data: &ws.DCAResponse{}})
	require.Error(t, err)

	client = newTestClientForRequest()
	go func() {
		raw := <-client.Send
		msg := ws.WebsocketMessage{Data: &ws.ActionData{}}
		require.NoError(t, json.Unmarshal(raw, &msg))
		resp := ws.WebsocketMessage{ID: msg.ID, Action: "other", Data: &ws.DCAResponse{}}
		client.receiveMessage(resp.Bytes())
	}()
	err = client.request(&ws.WebsocketMessage{Action: "want"}, &ws.WebsocketMessage{Data: &ws.DCAResponse{}})
	require.Error(t, err)
}

func TestDoCommonActionAndPreOperation(t *testing.T) {
	oldDB := datakitDB
	db := newTestDB(t)
	datakitDB = db
	t.Cleanup(func() { datakitDB = oldDB })

	dk := newTestDataKit("common")
	require.NoError(t, db.Insert(dk))
	client := newTestClientForRequest()
	client.DataKit = dk

	doCommonAction(client, &ws.WebsocketMessage{
		Action: ws.UpdateDatakitStatus,
		Data:   &ws.ActionData{Query: mapValues("status", ws.StatusStopped.String())},
	})
	found, err := db.Find(dk)
	require.NoError(t, err)
	require.Equal(t, ws.StatusStopped, found.Status)

	doCommonAction(client, &ws.WebsocketMessage{Action: ws.UpdateDatakitStatus})
	found, err = db.Find(dk)
	require.NoError(t, err)
	require.Equal(t, ws.StatusStopped, found.Status)

	doCommonAction(client, &ws.WebsocketMessage{
		Action: ws.UpdateDatakit,
		Data:   &ws.ActionData{Body: "{bad-json"},
	})

	updated := newTestDataKit("common")
	updated.HostName = "updated-host"
	doCommonAction(client, &ws.WebsocketMessage{
		Action: ws.UpdateDatakit,
		Data:   &ws.ActionData{Body: string(updated.Bytes())},
	})
	found, err = db.Find(updated)
	require.NoError(t, err)
	require.Equal(t, "updated-host", found.HostName)
	require.Equal(t, "updated-host", client.DataKit.HostName)

	doCommonAction(client, &ws.WebsocketMessage{Action: ws.DeleteDatakit, Data: &ws.ActionData{}})
	doCommonAction(client, &ws.WebsocketMessage{Action: "unknown", Data: &ws.ActionData{}})
	found, err = db.Find(dk)
	require.NoError(t, err)
	require.Nil(t, found)

	require.False(t, isOperationAllowed(ws.StatusOffline))
	require.False(t, isOperationAllowed(ws.StatusRestarting))
	require.False(t, isOperationAllowed(ws.StatusUpgrading))
	require.True(t, isOperationAllowed(ws.StatusRunning))

	resp := preOperation(nil, ws.StatusUpgrading)
	require.False(t, resp.Success)

	resp = preOperation(dk, ws.StatusRestarting)
	require.False(t, resp.Success)

	require.NoError(t, db.Insert(dk))
	resp = preOperation(dk, ws.StatusRestarting)
	require.True(t, resp.Success)
	found, err = db.Find(dk)
	require.NoError(t, err)
	require.Equal(t, ws.StatusRestarting, found.Status)

	resp = preOperation(dk, ws.StatusUpgrading)
	require.False(t, resp.Success)
}

func TestActionHandlers(t *testing.T) {
	oldDB := datakitDB
	db := newTestDB(t)
	datakitDB = db
	t.Cleanup(func() { datakitDB = oldDB })

	dk := newTestDataKit("action")
	require.NoError(t, db.Insert(dk))
	client := newTestClientForRequest()

	ctx, _ := newGinTestContext(http.MethodPost, "/api/datakit/getConfig?name=cpu", bytes.NewBufferString(`{"hello":"world"}`))

	go func() {
		raw := <-client.Send
		msg := ws.WebsocketMessage{Data: &ws.ActionData{}}
		require.NoError(t, json.Unmarshal(raw, &msg))
		require.Equal(t, ws.GetDatakitConfigAction, msg.Action)
		resp := ws.WebsocketMessage{
			ID:     msg.ID,
			Action: msg.Action,
			Data:   &ws.DCAResponse{Success: true, Code: 200, Content: "ok"},
		}
		client.receiveMessage(resp.Bytes())
	}()

	res, err := getActionHandler(ws.GetDatakitConfigAction)(client, dk, ctx)
	require.NoError(t, err)
	require.True(t, res.(*ws.DCAResponse).Success)

	_, err = client.doAction("missing", dk, ctx)
	require.Error(t, err)

	ActionHandlerMap["custom"] = func(*Client, *ws.DataKit, *gin.Context) (any, error) {
		return "handled", nil
	}
	t.Cleanup(func() { delete(ActionHandlerMap, "custom") })
	res, err = client.doAction("custom", dk, ctx)
	require.NoError(t, err)
	require.Equal(t, "handled", res)
}

func TestOperationActionHandlers(t *testing.T) {
	oldDB := datakitDB
	db := newTestDB(t)
	datakitDB = db
	t.Cleanup(func() { datakitDB = oldDB })

	dk := newTestDataKit("operation-action")
	require.NoError(t, db.Insert(dk))

	client := newTestClientForRequest()
	ctx, _ := newGinTestContext(http.MethodPost, "/api/datakit/operation", bytes.NewBufferString(`{}`))

	go func() {
		raw := <-client.Send
		msg := ws.WebsocketMessage{Data: &ws.ActionData{}}
		require.NoError(t, json.Unmarshal(raw, &msg))
		resp := ws.WebsocketMessage{
			ID:     msg.ID,
			Action: msg.Action,
			Data:   &ws.DCAResponse{Success: true, Code: 200},
		}
		client.receiveMessage(resp.Bytes())
	}()
	res, err := restartDatakitAction(client, dk, ctx)
	require.NoError(t, err)
	require.True(t, res.(*ws.DCAResponse).Success)

	require.NoError(t, db.UpdateStatus(dk, ws.StatusRunning))
	go func() {
		raw := <-client.Send
		msg := ws.WebsocketMessage{Data: &ws.ActionData{}}
		require.NoError(t, json.Unmarshal(raw, &msg))
		resp := ws.WebsocketMessage{
			ID:     msg.ID,
			Action: msg.Action,
			Data:   &ws.DCAResponse{Success: true, Code: 200},
		}
		client.receiveMessage(resp.Bytes())
	}()
	res, err = upgradeDatakitAction(client, dk, ctx)
	require.NoError(t, err)
	require.True(t, res.(*ws.DCAResponse).Success)

	res, err = restartDatakitAction(client, nil, ctx)
	require.NoError(t, err)
	require.False(t, res.(*ws.DCAResponse).Success)

	res, err = upgradeDatakitAction(client, nil, ctx)
	require.NoError(t, err)
	require.False(t, res.(*ws.DCAResponse).Success)
}

func TestGetDatakitStatsAction(t *testing.T) {
	client := newTestClientForRequest()
	dk := newTestDataKit("stats")

	go func() {
		raw := <-client.Send
		msg := ws.WebsocketMessage{Data: &ws.ActionData{}}
		require.NoError(t, json.Unmarshal(raw, &msg))
		require.Equal(t, ws.GetDatakitStatsAction, msg.Action)
		resp := ws.WebsocketMessage{
			ID:     msg.ID,
			Action: msg.Action,
			Data:   &ws.DCAResponse{Success: true, Code: 200, Content: map[string]any{"cpu": 1}},
		}
		client.receiveMessage(resp.Bytes())
	}()

	res, err := getDatakitStatsAction(client, dk, nil)
	require.NoError(t, err)
	require.True(t, res.(*ws.DCAResponse).Success)

	timeoutClient := newTestClientForRequest()
	timeoutClient.Timeout = time.Millisecond
	_, err = getDatakitStatsAction(timeoutClient, dk, nil)
	require.Error(t, err)
}

type errorReadCloser struct{}

func (errorReadCloser) Read([]byte) (int, error) { return 0, errors.New("read failed") }
func (errorReadCloser) Close() error             { return nil }

func TestGenericActionHandlerReadError(t *testing.T) {
	client := newTestClientForRequest()
	dk := newTestDataKit("read-error")
	ctx, _ := newGinTestContext(http.MethodPost, "/api/datakit/getConfig", nil)
	ctx.Request.Body = io.NopCloser(errorReadCloser{})

	_, err := getActionHandler(ws.GetDatakitConfigAction)(client, dk, ctx)
	require.Error(t, err)
}

func mapValues(key, value string) map[string][]string {
	return map[string][]string{key: {value}}
}
