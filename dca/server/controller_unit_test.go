// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
	ws "gitlab.jiagouyun.com/cloudcare-tools/datakit/dca/websocket"
)

func withTestDatakitDB(t *testing.T) *DB {
	t.Helper()
	oldDB := datakitDB
	db := newTestDB(t)
	datakitDB = db
	t.Cleanup(func() { datakitDB = oldDB })
	return db
}

func addWorkspaceCookie(ctx *gin.Context) {
	ctx.Request.AddCookie(&http.Cookie{Name: cookieWorkspaceUUID, Value: "workspace-1"})
}

func TestDatakitListSearchAndByIDHandlers(t *testing.T) {
	db := withTestDatakitDB(t)
	dk1 := newTestDataKit("list-1")
	dk1.HostName = "alpha"
	dk1.IP = "10.0.0.1"
	dk1.Version = "1.0.0"
	require.NoError(t, db.Insert(dk1))

	dk2 := newTestDataKit("list-2")
	dk2.HostName = "beta"
	dk2.IP = "10.0.0.2"
	dk2.Version = "2.0.0"
	require.NoError(t, db.Insert(dk2))

	ctx, rec := newGinTestContext(http.MethodGet, "/api/datakit/list?pageIndex=1&pageSize=1&search=alpha", nil)
	addWorkspaceCookie(ctx)
	datakitListHandler(ctx)
	resp := decodeDCAResponse(t, rec.Body.String())
	require.True(t, resp.Success)
	content, ok := resp.Content.(map[string]interface{})
	require.True(t, ok)
	pageInfo := content["pageInfo"].(map[string]interface{})
	require.EqualValues(t, 1, pageInfo["count"])
	require.EqualValues(t, 1, pageInfo["totalCount"])

	ctx, rec = newGinTestContext(http.MethodGet, "/api/datakit/searchValue", nil)
	addWorkspaceCookie(ctx)
	datakitSearchValueHandler(ctx)
	resp = decodeDCAResponse(t, rec.Body.String())
	require.True(t, resp.Success)
	values := resp.Content.(map[string]interface{})
	require.Contains(t, values, "host_name")
	require.Contains(t, values, "env")

	rows := []ws.DataKit{}
	require.NoError(t, db.Select("select * from datakit where conn_id=? order by id", &rows, dk1.ConnID))
	require.Len(t, rows, 1)

	ctx, rec = newGinTestContext(http.MethodGet, "/api/datakit/listByID?ids="+url.QueryEscape(rows[0].ID), nil)
	addWorkspaceCookie(ctx)
	datakitByIDHandler(ctx)
	resp = decodeDCAResponse(t, rec.Body.String())
	require.True(t, resp.Success)
	require.Len(t, resp.Content.([]interface{}), 1)

	res, err := getDatakits("all", "workspace-1")
	require.NoError(t, err)
	require.Len(t, res, 2)

	filter := `{"relation":"and","items":[{"field":"env","operator":"in","value":["missing"]}]}`
	ctx, rec = newGinTestContext(http.MethodGet, "/api/datakit/list?filter="+url.QueryEscape(filter), nil)
	addWorkspaceCookie(ctx)
	datakitListHandler(ctx)
	resp = decodeDCAResponse(t, rec.Body.String())
	require.True(t, resp.Success)
	content, ok = resp.Content.(map[string]interface{})
	require.True(t, ok)
	pageInfo = content["pageInfo"].(map[string]interface{})
	require.EqualValues(t, 0, pageInfo["totalCount"])
}

func TestDatakitHandlersMissingWorkspace(t *testing.T) {
	withTestDatakitDB(t)

	ctx, rec := newGinTestContext(http.MethodGet, "/api/datakit/list", nil)
	datakitListHandler(ctx)
	require.False(t, decodeDCAResponse(t, rec.Body.String()).Success)

	ctx, rec = newGinTestContext(http.MethodGet, "/api/datakit/listByID?ids=1", nil)
	datakitByIDHandler(ctx)
	require.False(t, decodeDCAResponse(t, rec.Body.String()).Success)

	ctx, rec = newGinTestContext(http.MethodGet, "/api/datakit/searchValue", nil)
	datakitSearchValueHandler(ctx)
	require.False(t, decodeDCAResponse(t, rec.Body.String()).Success)
}

func TestDatakitListHandlerRejectsBadFilter(t *testing.T) {
	withTestDatakitDB(t)
	ctx, rec := newGinTestContext(http.MethodGet, "/api/datakit/list?filter=%7Bbad-json", nil)
	addWorkspaceCookie(ctx)
	datakitListHandler(ctx)
	require.False(t, decodeDCAResponse(t, rec.Body.String()).Success)
}

func TestGetWhereSQL(t *testing.T) {
	db := withTestDatakitDB(t)
	dk := newTestDataKit("where-1")
	dk.HostName = "alpha"
	dk.Version = "1.2.3"
	dk.DataKitRuntimeInfo.GlobalHostTags = map[string]string{"env": "prod", "env.name-1": "prod"}
	require.NoError(t, db.Insert(dk))

	whereSQL, values, err := getWhereSQL("workspace-1", "", "alpha")
	require.NoError(t, err)
	require.Contains(t, whereSQL, "host_name like")
	require.Len(t, values, 4)

	filter := filterParam{
		Relation: "and",
		Items: []filterItem{
			{Field: "version", Operator: "in", Value: []string{"1.2.3"}},
			{Field: "env", Operator: "in", Value: []string{"prod"}},
		},
	}
	filterBytes, err := json.Marshal(filter)
	require.NoError(t, err)
	whereSQL, values, err = getWhereSQL("workspace-1", string(filterBytes), "")
	require.NoError(t, err)
	require.Contains(t, whereSQL, "version IN")
	require.Contains(t, whereSQL, "conn_id in")
	require.NotEmpty(t, values)

	_, _, err = getWhereSQL("workspace-1", "{bad-json", "")
	require.Error(t, err)

	filter.Relation = "xor"
	filterBytes, err = json.Marshal(filter)
	require.NoError(t, err)
	_, _, err = getWhereSQL("workspace-1", string(filterBytes), "")
	require.Error(t, err)

	filter.Relation = "and"
	filter.Items = []filterItem{{Field: "version", Operator: "bad", Value: []string{"1.2.3"}}}
	filterBytes, err = json.Marshal(filter)
	require.NoError(t, err)
	_, _, err = getWhereSQL("workspace-1", string(filterBytes), "")
	require.Error(t, err)

	filter.Items = []filterItem{{Field: "version", Operator: "match", Value: []string{"1\\.2"}}}
	filterBytes, err = json.Marshal(filter)
	require.NoError(t, err)
	whereSQL, values, err = getWhereSQL("workspace-1", string(filterBytes), "")
	require.NoError(t, err)
	require.Contains(t, whereSQL, "version REGEXP")
	require.NotEmpty(t, values)

	filter.Relation = "or"
	filter.Items = []filterItem{{Field: "env", Operator: "not_match", Value: []string{"dev"}}}
	filterBytes, err = json.Marshal(filter)
	require.NoError(t, err)
	whereSQL, values, err = getWhereSQL("workspace-1", string(filterBytes), "")
	require.NoError(t, err)
	require.Contains(t, whereSQL, "conn_id in")
	require.NotEmpty(t, values)

	filter.Relation = "and"
	filter.Items = []filterItem{{Field: "env.name-1", Operator: "in", Value: []string{"prod"}}}
	filterBytes, err = json.Marshal(filter)
	require.NoError(t, err)
	whereSQL, values, err = getWhereSQL("workspace-1", string(filterBytes), "")
	require.NoError(t, err)
	require.Contains(t, whereSQL, "conn_id in")
	require.NotEmpty(t, values)

	for _, field := range []string{"env') OR 1=1 --", "env name", "$.env"} {
		filter.Items = []filterItem{{Field: field, Operator: "in", Value: []string{"prod"}}}
		filterBytes, err = json.Marshal(filter)
		require.NoError(t, err)
		_, _, err = getWhereSQL("workspace-1", string(filterBytes), "")
		require.Error(t, err)
	}

	filter.Relation = "and"
	filter.Items = []filterItem{{Field: "env", Operator: "in", Value: []string{"missing"}}}
	filterBytes, err = json.Marshal(filter)
	require.NoError(t, err)
	whereSQL, values, err = getWhereSQL("workspace-1", string(filterBytes), "")
	require.NoError(t, err)
	require.Empty(t, whereSQL)
	require.Nil(t, values)

	filter.Items = []filterItem{{Field: "version", Operator: "in", Value: nil}}
	filterBytes, err = json.Marshal(filter)
	require.NoError(t, err)
	_, _, err = getWhereSQL("workspace-1", string(filterBytes), "")
	require.Error(t, err)
}

func TestDatakitOperationHandlerInvalidAndEmpty(t *testing.T) {
	withTestDatakitDB(t)

	ctx, rec := newGinTestContext(http.MethodPost, "/api/datakit/operation/bad?ids=all", nil)
	ctx.Params = gin.Params{{Key: "type", Value: "bad"}}
	addWorkspaceCookie(ctx)
	datakitOperationHandler(ctx)
	require.False(t, decodeDCAResponse(t, rec.Body.String()).Success)

	ctx, rec = newGinTestContext(http.MethodPost, "/api/datakit/operation/reload?ids=all", nil)
	ctx.Params = gin.Params{{Key: "type", Value: "reload"}}
	addWorkspaceCookie(ctx)
	datakitOperationHandler(ctx)
	require.True(t, decodeDCAResponse(t, rec.Body.String()).Success)
}

func TestDatakitOperationHandlerCallsAllowedClients(t *testing.T) {
	db := withTestDatakitDB(t)
	replaceManagerState(t, Manager.Register, Manager.Unregister, map[string]*Client{}, Manager.WebsocketConns)

	allowed := newTestDataKit("operation-allowed")
	allowed.Status = ws.StatusRunning
	blocked := newTestDataKit("operation-blocked")
	blocked.Status = ws.StatusUpgrading
	require.NoError(t, db.Insert(allowed))
	require.NoError(t, db.Insert(blocked))
	require.NoError(t, db.UpdateStatus(blocked, ws.StatusUpgrading))

	called := 0
	ActionHandlerMap[ws.ReloadDatakitAction] = func(*Client, *ws.DataKit, *gin.Context) (any, error) {
		called++
		return &ws.DCAResponse{Success: true, Code: 200}, nil
	}
	t.Cleanup(func() { delete(ActionHandlerMap, ws.ReloadDatakitAction) })
	Manager.Clients[allowed.ConnID] = newTestClientForRequest()
	Manager.Clients[allowed.ConnID].DataKit = allowed
	Manager.Clients[blocked.ConnID] = newTestClientForRequest()
	Manager.Clients[blocked.ConnID].DataKit = blocked

	ctx, rec := newGinTestContext(http.MethodPost, "/api/datakit/operation/reload?ids=all", nil)
	ctx.Params = gin.Params{{Key: "type", Value: "reload"}}
	addWorkspaceCookie(ctx)
	datakitOperationHandler(ctx)
	require.True(t, decodeDCAResponse(t, rec.Body.String()).Success)
	require.Equal(t, 1, called)

	delete(Manager.Clients, allowed.ConnID)
	ctx, rec = newGinTestContext(http.MethodPost, "/api/datakit/operation/reload?ids=all", nil)
	ctx.Params = gin.Params{{Key: "type", Value: "reload"}}
	addWorkspaceCookie(ctx)
	datakitOperationHandler(ctx)
	require.False(t, decodeDCAResponse(t, rec.Body.String()).Success)
}

func TestDatakitHandlerDispatch(t *testing.T) {
	db := withTestDatakitDB(t)
	replaceManagerState(t, Manager.Register, Manager.Unregister, map[string]*Client{}, Manager.WebsocketConns)

	dk := newTestDataKit("dispatch")
	require.NoError(t, db.Insert(dk))
	rows := []ws.DataKit{}
	require.NoError(t, db.Select("select * from datakit where conn_id=?", &rows, dk.ConnID))
	require.Len(t, rows, 1)

	ActionHandlerMap["dispatch-action"] = func(*Client, *ws.DataKit, *gin.Context) (any, error) {
		return &ws.DCAResponse{Success: true, Code: 200, Content: "done"}, nil
	}
	t.Cleanup(func() { delete(ActionHandlerMap, "dispatch-action") })
	Manager.Clients[dk.ConnID] = newTestClientForRequest()
	Manager.Clients[dk.ConnID].DataKit = dk

	ctx, rec := newGinTestContext(http.MethodGet, "/api/datakit/dispatch?datakit_id="+rows[0].ID, nil)
	addWorkspaceCookie(ctx)
	datakitHandler("dispatch-action")(ctx)
	resp := decodeDCAResponse(t, rec.Body.String())
	require.True(t, resp.Success)
	require.Equal(t, "done", resp.Content)

	ctx, rec = newGinTestContext(http.MethodGet, "/api/datakit/dispatch?datakit_id=missing", nil)
	addWorkspaceCookie(ctx)
	datakitHandler("dispatch-action")(ctx)
	require.False(t, decodeDCAResponse(t, rec.Body.String()).Success)
}

func TestDatakitHandlerNormalizesParamInvalidResponseCode(t *testing.T) {
	db := withTestDatakitDB(t)
	replaceManagerState(t, Manager.Register, Manager.Unregister, map[string]*Client{}, Manager.WebsocketConns)

	dk := newTestDataKit("dispatch-param-invalid")
	require.NoError(t, db.Insert(dk))
	rows := []ws.DataKit{}
	require.NoError(t, db.Select("select * from datakit where conn_id=?", &rows, dk.ConnID))
	require.Len(t, rows, 1)

	ActionHandlerMap["dispatch-param-invalid-action"] = func(*Client, *ws.DataKit, *gin.Context) (any, error) {
		return &ws.DCAResponse{
			Success:   false,
			Code:      http.StatusInternalServerError,
			ErrorCode: "param.invalid",
			Message:   "pipeline path is not valid",
		}, nil
	}
	t.Cleanup(func() { delete(ActionHandlerMap, "dispatch-param-invalid-action") })
	Manager.Clients[dk.ConnID] = newTestClientForRequest()
	Manager.Clients[dk.ConnID].DataKit = dk

	ctx, rec := newGinTestContext(http.MethodGet, "/api/datakit/dispatch?datakit_id="+rows[0].ID, nil)
	addWorkspaceCookie(ctx)
	datakitHandler("dispatch-param-invalid-action")(ctx)

	resp := decodeDCAResponse(t, rec.Body.String())
	require.False(t, resp.Success)
	require.Equal(t, http.StatusBadRequest, resp.Code)
	require.Equal(t, "param.invalid", resp.ErrorCode)
}

func TestLogHandlersRejectMissingDatakit(t *testing.T) {
	withTestDatakitDB(t)

	ctx, rec := newGinTestContext(http.MethodGet, "/api/datakit/log/download?datakit_id=missing", nil)
	addWorkspaceCookie(ctx)
	datakitLogDownladHandler(ctx)
	require.False(t, decodeDCAResponse(t, rec.Body.String()).Success)

	router := gin.New()
	router.GET("/api/datakit/ws/log", websocketLogHandler)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/datakit/ws/log?datakit_id=missing", nil)
	req.AddCookie(&http.Cookie{Name: cookieWorkspaceUUID, Value: "workspace-1"})
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDatakitLogDownloadHandlerStreamsBinary(t *testing.T) {
	db := withTestDatakitDB(t)
	replaceManagerState(t, Manager.Register, Manager.Unregister, map[string]*Client{}, map[string]chan *websocket.Conn{})

	dk := newTestDataKit("download-log")
	require.NoError(t, db.Insert(dk))
	rows := []ws.DataKit{}
	require.NoError(t, db.Select("select * from datakit where conn_id=?", &rows, dk.ConnID))
	require.Len(t, rows, 1)

	client := newTestClientForRequest()
	client.DataKit = dk
	Manager.Clients[dk.ConnID] = client

	done := make(chan struct{})
	go func() {
		raw := <-client.Send
		msg := ws.WebsocketMessage{Data: &ActionData{}}
		require.NoError(t, json.Unmarshal(raw, &msg))
		data := msg.Data.(*ActionData)
		serverConn, clientConn := newTestWebsocketPair(t)
		defer clientConn.Close() //nolint:errcheck
		Manager.addWebsocketConnChan(data.Query.Get(ws.HeaderNewWebSocketConnectionID), serverConn)

		_, body, err := clientConn.ReadMessage()
		require.NoError(t, err)
		request := ws.WebsocketMessage{Data: &ActionData{}}
		require.NoError(t, json.Unmarshal(body, &request))
		require.Equal(t, ws.GetDatakitLogDownloadAction, request.Action)

		require.NoError(t, clientConn.WriteMessage(websocket.BinaryMessage, []byte(`log-bytes token=tkn_secret password = "secret-pass" kv=real-value OSS_ACCESS_KEY_SECRET=oss-secret url=http://user:secret@example.com`)))
		require.NoError(t, clientConn.WriteMessage(websocket.TextMessage, (&ws.WebsocketMessage{
			Action: ws.GetDatakitLogDownloadAction,
			Data:   &ws.DCAResponse{Success: true, Code: 200},
		}).Bytes()))
		close(done)
	}()

	ctx, rec := newGinTestContext(http.MethodGet, "/api/datakit/log/download?datakit_id="+rows[0].ID, nil)
	addWorkspaceCookie(ctx)
	datakitLogDownladHandler(ctx)
	require.Contains(t, rec.Body.String(), "log-bytes")
	require.Contains(t, rec.Body.String(), "token=******")
	require.Contains(t, rec.Body.String(), `password = "******"`)
	require.Contains(t, rec.Body.String(), "kv=******")
	require.Contains(t, rec.Body.String(), "OSS_ACCESS_KEY_SECRET=******")
	require.Contains(t, rec.Body.String(), "http://user:******@example.com")
	require.NotContains(t, rec.Body.String(), "tkn_secret")
	require.NotContains(t, rec.Body.String(), "secret-pass")
	require.NotContains(t, rec.Body.String(), "real-value")
	require.NotContains(t, rec.Body.String(), "oss-secret")
	<-done
}

func TestDatakitLogDownloadHandlerReportsDownstreamErrors(t *testing.T) {
	cases := []struct {
		name    string
		message []byte
	}{
		{
			name:    "bad json",
			message: []byte("{bad-json"),
		},
		{
			name: "action mismatch",
			message: (&ws.WebsocketMessage{
				Action: "other",
				Data:   &ws.DCAResponse{Success: true, Code: 200},
			}).Bytes(),
		},
		{
			name: "response failed",
			message: (&ws.WebsocketMessage{
				Action: ws.GetDatakitLogDownloadAction,
				Data:   &ws.DCAResponse{Success: false, Code: 500, Message: "downstream failed"},
			}).Bytes(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := withTestDatakitDB(t)
			replaceManagerState(t, Manager.Register, Manager.Unregister, map[string]*Client{}, map[string]chan *websocket.Conn{})

			dk := newTestDataKit("download-error-" + strings.ReplaceAll(tc.name, " ", "-"))
			require.NoError(t, db.Insert(dk))
			rows := []ws.DataKit{}
			require.NoError(t, db.Select("select * from datakit where conn_id=?", &rows, dk.ConnID))
			require.Len(t, rows, 1)

			client := newTestClientForRequest()
			client.DataKit = dk
			Manager.Clients[dk.ConnID] = client

			go func() {
				raw := <-client.Send
				msg := ws.WebsocketMessage{Data: &ActionData{}}
				require.NoError(t, json.Unmarshal(raw, &msg))
				data := msg.Data.(*ActionData)
				serverConn, clientConn := newTestWebsocketPair(t)
				defer clientConn.Close() //nolint:errcheck
				Manager.addWebsocketConnChan(data.Query.Get(ws.HeaderNewWebSocketConnectionID), serverConn)
				_, _, err := clientConn.ReadMessage()
				require.NoError(t, err)
				require.NoError(t, clientConn.WriteMessage(websocket.TextMessage, tc.message))
			}()

			ctx, rec := newGinTestContext(http.MethodGet, "/api/datakit/log/download?datakit_id="+rows[0].ID, nil)
			addWorkspaceCookie(ctx)
			datakitLogDownladHandler(ctx)
			require.False(t, decodeDCAResponse(t, rec.Body.String()).Success)
		})
	}
}

func TestLogDownloadRedactorRedactsSplitLine(t *testing.T) {
	redactor := &logDownloadRedactor{}

	require.Empty(t, redactor.RedactChunk([]byte(`Do: Post "http://example.com/write?token=tkn_forethought`)))
	out := redactor.RedactChunk([]byte(`local0123456789abcdef": failed` + "\nnext line\n"))

	require.Contains(t, string(out), "token=******")
	require.Contains(t, string(out), "next line")
	require.NotContains(t, string(out), "tkn_forethoughtlocal0123456789abcdef")
	require.Empty(t, redactor.Flush())
}

func TestLogDownloadRedactorFlushesPartialLine(t *testing.T) {
	redactor := &logDownloadRedactor{}

	require.Empty(t, redactor.RedactChunk([]byte(`OSS_ACCESS_KEY_ID=key OSS_ACCESS_KEY_SECRET=secret`)))
	out := redactor.Flush()

	require.Contains(t, string(out), "OSS_ACCESS_KEY_ID=******")
	require.Contains(t, string(out), "OSS_ACCESS_KEY_SECRET=******")
	require.NotContains(t, string(out), "OSS_ACCESS_KEY_ID=key")
	require.NotContains(t, string(out), "OSS_ACCESS_KEY_SECRET=secret")
}

func TestLogDownloadRedactorKeepsEmittedLinesBeforeBufferedTail(t *testing.T) {
	redactor := &logDownloadRedactor{}

	out := redactor.RedactChunk([]byte(
		`Do: Post "http://example.com/write?token=tkn_forethoughtlocal0123456789abcdef": failed` + "\n" +
			`tail without newline token=tkn_tail_should_remain_buffered`,
	))

	require.Contains(t, string(out), "token=******")
	require.NotContains(t, string(out), "tkn_forethoughtlocal0123456789abcdef")
	require.NotContains(t, string(out), "tail without newline")

	flushed := redactor.Flush()
	require.Contains(t, string(flushed), "token=******")
	require.NotContains(t, string(flushed), "tkn_tail_should_remain_buffered")
}

func TestLogDownloadRedactorDropsOversizedLine(t *testing.T) {
	redactor := &logDownloadRedactor{}
	secretTail := "secret-tail"

	out := redactor.RedactChunk([]byte("token=tkn_" + strings.Repeat("a", logRedactMaxLineBuffer) + secretTail))
	require.Equal(t, logRedactOversizedLine, string(out))
	require.NotContains(t, string(out), secretTail)

	require.Empty(t, redactor.RedactChunk([]byte("more-secret-bytes")))

	out = redactor.RedactChunk([]byte("\nnext token=tkn_next\n"))
	require.Contains(t, string(out), "next token=******")
	require.NotContains(t, string(out), "more-secret-bytes")
	require.NotContains(t, string(out), "tkn_next")
	require.Empty(t, redactor.Flush())
}

func TestLogDownloadRedactorDoesNotFlushOversizedLineAtEOF(t *testing.T) {
	redactor := &logDownloadRedactor{}
	secretTail := "secret-tail"

	out := redactor.RedactChunk([]byte("password=" + strings.Repeat("a", logRedactMaxLineBuffer) + secretTail))
	require.Equal(t, logRedactOversizedLine, string(out))

	flushed := redactor.Flush()
	require.Empty(t, flushed)
	require.NotContains(t, string(flushed), secretTail)
}

func TestDatakitLogDownloadHandlerRejectsUnavailableDatakitConnection(t *testing.T) {
	db := withTestDatakitDB(t)
	replaceManagerState(t, Manager.Register, Manager.Unregister, map[string]*Client{}, Manager.WebsocketConns)

	dk := newTestDataKit("download-offline")
	require.NoError(t, db.Insert(dk))
	rows := []ws.DataKit{}
	require.NoError(t, db.Select("select * from datakit where conn_id=?", &rows, dk.ConnID))
	require.Len(t, rows, 1)

	ctx, rec := newGinTestContext(http.MethodGet, "/api/datakit/log/download?datakit_id="+rows[0].ID, nil)
	addWorkspaceCookie(ctx)
	datakitLogDownladHandler(ctx)
	require.False(t, decodeDCAResponse(t, rec.Body.String()).Success)
}

func TestWebsocketLogHandlerForwardsTailMessage(t *testing.T) {
	db := withTestDatakitDB(t)
	replaceManagerState(t, Manager.Register, Manager.Unregister, map[string]*Client{}, map[string]chan *websocket.Conn{})

	dk := newTestDataKit("tail-log")
	require.NoError(t, db.Insert(dk))
	rows := []ws.DataKit{}
	require.NoError(t, db.Select("select * from datakit where conn_id=?", &rows, dk.ConnID))
	require.Len(t, rows, 1)

	client := newTestClientForRequest()
	client.DataKit = dk
	Manager.Clients[dk.ConnID] = client

	datakitSide := make(chan *websocket.Conn, 1)
	go func() {
		raw := <-client.Send
		msg := ws.WebsocketMessage{Data: &ActionData{}}
		require.NoError(t, json.Unmarshal(raw, &msg))
		data := msg.Data.(*ActionData)
		serverConn, clientConn := newTestWebsocketPair(t)
		Manager.addWebsocketConnChan(data.Query.Get(ws.HeaderNewWebSocketConnectionID), serverConn)
		datakitSide <- clientConn
	}()

	router := gin.New()
	router.GET("/api/datakit/ws/log", websocketLogHandler)
	server := httptest.NewServer(router)
	defer server.Close()

	header := http.Header{}
	header.Add("Cookie", cookieWorkspaceUUID+"=workspace-1")
	frontConn, _, err := websocket.DefaultDialer.Dial(
		"ws"+strings.TrimPrefix(server.URL, "http")+"/api/datakit/ws/log?datakit_id="+rows[0].ID,
		header,
	)
	require.NoError(t, err)
	defer frontConn.Close() //nolint:errcheck

	clientConn := <-datakitSide
	defer clientConn.Close() //nolint:errcheck
	_, body, err := clientConn.ReadMessage()
	require.NoError(t, err)
	request := ws.WebsocketMessage{Data: &ActionData{}}
	require.NoError(t, json.Unmarshal(body, &request))
	require.Equal(t, ws.GetDatakitLogTailAction, request.Action)

	payload := []byte("tail-line")
	require.NoError(t, clientConn.WriteMessage(websocket.TextMessage, (&ws.WebsocketMessage{
		Action: ws.GetDatakitLogTailAction,
		Data:   &ws.DCAResponse{Success: true, Code: 200, Content: &payload},
	}).Bytes()))
	_, body, err = frontConn.ReadMessage()
	require.NoError(t, err)
	require.Equal(t, payload, body)
	require.NoError(t, clientConn.Close())
	require.Eventually(t, func() bool {
		return frontConn.WriteMessage(websocket.TextMessage, []byte("close-check")) != nil
	}, time.Second, 10*time.Millisecond)
}

func TestWebsocketLogHandlerReturnsActionErrorsToFront(t *testing.T) {
	db := withTestDatakitDB(t)
	replaceManagerState(t, Manager.Register, Manager.Unregister, map[string]*Client{}, map[string]chan *websocket.Conn{})

	dk := newTestDataKit("tail-error")
	require.NoError(t, db.Insert(dk))
	rows := []ws.DataKit{}
	require.NoError(t, db.Select("select * from datakit where conn_id=?", &rows, dk.ConnID))
	require.Len(t, rows, 1)

	client := newTestClientForRequest()
	client.DataKit = dk
	Manager.Clients[dk.ConnID] = client

	datakitSide := make(chan *websocket.Conn, 1)
	go func() {
		raw := <-client.Send
		msg := ws.WebsocketMessage{Data: &ActionData{}}
		require.NoError(t, json.Unmarshal(raw, &msg))
		data := msg.Data.(*ActionData)
		serverConn, clientConn := newTestWebsocketPair(t)
		Manager.addWebsocketConnChan(data.Query.Get(ws.HeaderNewWebSocketConnectionID), serverConn)
		datakitSide <- clientConn
	}()

	router := gin.New()
	router.GET("/api/datakit/ws/log", websocketLogHandler)
	server := httptest.NewServer(router)
	defer server.Close()

	header := http.Header{}
	header.Add("Cookie", cookieWorkspaceUUID+"=workspace-1")
	frontConn, _, err := websocket.DefaultDialer.Dial(
		"ws"+strings.TrimPrefix(server.URL, "http")+"/api/datakit/ws/log?datakit_id="+rows[0].ID,
		header,
	)
	require.NoError(t, err)
	defer frontConn.Close() //nolint:errcheck

	clientConn := <-datakitSide
	defer clientConn.Close() //nolint:errcheck
	_, _, err = clientConn.ReadMessage()
	require.NoError(t, err)
	require.NoError(t, clientConn.WriteMessage(websocket.TextMessage, (&ws.WebsocketMessage{
		Action: ws.GetDatakitLogTailAction,
		Data:   &ws.DCAResponse{Success: false, Code: 500, Message: "tail failed"},
	}).Bytes()))

	_, body, err := frontConn.ReadMessage()
	require.NoError(t, err)
	require.Contains(t, string(body), "tail failed")
}

func TestWebsocketLogHandlerDownstreamConnectionFailures(t *testing.T) {
	db := withTestDatakitDB(t)
	dk := newTestDataKit("tail-no-client")
	require.NoError(t, db.Insert(dk))
	rows := []ws.DataKit{}
	require.NoError(t, db.Select("select * from datakit where conn_id=?", &rows, dk.ConnID))
	require.Len(t, rows, 1)

	t.Run("missing datakit client", func(t *testing.T) {
		replaceManagerState(t, Manager.Register, Manager.Unregister, map[string]*Client{}, map[string]chan *websocket.Conn{})

		router := gin.New()
		router.GET("/api/datakit/ws/log", websocketLogHandler)
		server := httptest.NewServer(router)
		defer server.Close()

		header := http.Header{}
		header.Add("Cookie", cookieWorkspaceUUID+"=workspace-1")
		frontConn, _, err := websocket.DefaultDialer.Dial(
			"ws"+strings.TrimPrefix(server.URL, "http")+"/api/datakit/ws/log?datakit_id="+rows[0].ID,
			header,
		)
		require.NoError(t, err)
		defer frontConn.Close() //nolint:errcheck

		require.Eventually(t, func() bool {
			return frontConn.WriteMessage(websocket.TextMessage, []byte("closed")) != nil
		}, time.Second, 10*time.Millisecond)
	})

	t.Run("downstream write failed", func(t *testing.T) {
		replaceManagerState(t, Manager.Register, Manager.Unregister, map[string]*Client{}, map[string]chan *websocket.Conn{})

		client := newTestClientForRequest()
		client.DataKit = dk
		Manager.Clients[dk.ConnID] = client

		datakitSide := make(chan *websocket.Conn, 1)
		go func() {
			raw := <-client.Send
			msg := ws.WebsocketMessage{Data: &ActionData{}}
			require.NoError(t, json.Unmarshal(raw, &msg))
			data := msg.Data.(*ActionData)
			serverConn, clientConn := newTestWebsocketPair(t)
			Manager.addWebsocketConnChan(data.Query.Get(ws.HeaderNewWebSocketConnectionID), serverConn)
			_ = clientConn.Close()
			datakitSide <- clientConn
		}()

		router := gin.New()
		router.GET("/api/datakit/ws/log", websocketLogHandler)
		server := httptest.NewServer(router)
		defer server.Close()

		header := http.Header{}
		header.Add("Cookie", cookieWorkspaceUUID+"=workspace-1")
		frontConn, _, err := websocket.DefaultDialer.Dial(
			"ws"+strings.TrimPrefix(server.URL, "http")+"/api/datakit/ws/log?datakit_id="+rows[0].ID,
			header,
		)
		require.NoError(t, err)
		defer frontConn.Close() //nolint:errcheck

		<-datakitSide
		require.Eventually(t, func() bool {
			return frontConn.WriteMessage(websocket.TextMessage, []byte("closed")) != nil
		}, time.Second, 10*time.Millisecond)
	})
}
