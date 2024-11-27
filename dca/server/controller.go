// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	ws "gitlab.jiagouyun.com/cloudcare-tools/datakit/dca/websocket"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/version"
)

type getTokenByCodeContent struct {
	Token         string `json:"token"`
	WorkspaceUUID string `json:"workspaceUUID"`
}

func ssoLoginHandler(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.String(400, "code is empty")
		return
	}

	api := consoleAPIInfo[getTokenByCodePath]
	apiURL, err := url.Parse(getConsoleAPIURL(api[1]))
	if err != nil {
		l.Errorf("failed to parse url: %s", err.Error())
		c.String(500, "failed to validate")
		return
	}
	query := url.Values{
		"auth_code": []string{code},
	}
	apiURL.RawQuery = query.Encode()

	resp, err := consoleClient.Get(apiURL.String())
	if err != nil {
		l.Errorf("failed to get token from console: %s", err.Error())
		c.String(500, "failed to get token from console")
		return
	}
	defer resp.Body.Close() // nolint:errcheck

	content := &getTokenByCodeContent{}
	bodyRes := ws.DCAResponse{
		Content: content,
	}

	if err := json.NewDecoder(resp.Body).Decode(&bodyRes); err != nil {
		l.Errorf("failed to decode response from console: %s", err.Error())
		c.String(500, "failed to decode response from console")
		return
	}

	if !bodyRes.Success {
		c.String(500, "failed to get token from console, err code is: %s", bodyRes.ErrorCode)
		return
	}

	c.SetCookie(cookieFrontToken, content.Token, 3600, "/", "", false, true)
	c.SetCookie(cookieWorkspaceUUID, content.WorkspaceUUID, 3600, "/", "", false, true)

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Redirect(http.StatusTemporaryRedirect, "/dashboard")
}

func consoleHandler(name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := newHandler(c)

		if api, ok := consoleAPIInfo[name]; !ok {
			h.fail(500, fmt.Sprintf("unknown api: %s", name))
			return
		} else {
			res, err := h.pipe(api[0], api[1], h.c.Request.Body)
			if err != nil {
				h.fail(500, err.Error())
				return
			}

			reader := res.Body
			contentLength := res.ContentLength

			h.c.DataFromReader(http.StatusOK, contentLength, "application/json", reader, nil)
		}
	}
}

func logoutHandler(c *gin.Context) {
	h := newHandler(c)
	// clear cookie
	h.c.SetCookie(cookieFrontToken, "", -1, "/", "", false, true)
	h.c.SetCookie(cookieWorkspaceUUID, "", -1, "/", "", false, true)

	h.success()
}

func consoleChangeWorkspaceHandler(c *gin.Context) {
	h := newHandler(c)
	api := consoleAPIInfo[changeWorkspacePath]

	req, err := http.NewRequest(http.MethodPost, getConsoleAPIURL(api[1]), c.Request.Body)
	if err != nil {
		h.fail(500, err.Error())
		return
	}

	workspaceUUID := c.GetHeader("X-Workspace-Uuid")
	req.Header.Set("X-Workspace-Uuid", workspaceUUID)

	respbody, err := h.doRequest(req)
	if err != nil {
		h.fail(500, err.Error())
		return
	}

	resp := Response{}
	if err := json.Unmarshal(respbody, &resp); err != nil {
		h.fail(500, err.Error())
		return
	}

	if !resp.Success {
		h.fail(500, resp.Message)
		return
	}

	c.SetCookie(cookieWorkspaceUUID, workspaceUUID, 3600, "/", "", false, true)

	h.success()
}

func consoleRedirectHandler(c *gin.Context) {
	c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/integration/dca", consoleWebURL))
}

func datakitLogDownladHandler(ctx *gin.Context) {
	l.Debug("datakit log download handler")
	h := newHandler(ctx)
	datakit := h.getDatakit()
	if datakit == nil {
		h.fail(400, "datakit not found")
		return
	}

	logType := "log"
	if h.c.Query("type") == "gin.log" {
		logType = "gin.log"
	}

	conn, err := getNewWebsocketConn(datakit, ws.GetDatakitLogDownloadAction)
	if err != nil {
		h.fail(500, "server error")
		return
	}

	defer conn.Close() //nolint:errcheck

	msg := ws.WebsocketMessage{
		Action: ws.GetDatakitLogDownloadAction,
		Data:   ActionData{Query: ctx.Request.URL.Query()},
	}

	if err := conn.WriteMessage(websocket.TextMessage, msg.Bytes()); err != nil {
		h.fail(500, "server error")
		l.Warnf("failed to write message to websocket: %s", err.Error())
		return
	}

	for {
		if messageType, bytes, err := conn.ReadMessage(); err != nil {
			h.fail(500, "server error")
			l.Warnf("failed to read message from websocket: %s", err.Error())
			return
		} else {
			l.Info("received messages, length: %d, messageType: %s", len(bytes), messageType)
			switch messageType {
			case websocket.TextMessage:
				dest := ws.WebsocketMessage{}
				if err := json.Unmarshal(bytes, &dest); err != nil {
					l.Errorf("failed to unmarshal message: %s", err.Error())
					h.fail(500, "server error")
					return
				} else if msg.Action != dest.Action {
					l.Errorf("message action not match: %s, %s", msg.Action, dest.Action)
					h.fail(500, "server error")
					return
				}

				if res, ok := dest.Data.(*ws.DCAResponse); ok {
					if !res.Success {
						h.send(res)
						return
					}
				}
			case websocket.BinaryMessage:
				h.c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", logType))
				h.c.Header("Content-Type", "application/octet-stream")
				// h.c.Data(200, "application/octet-stream", bytes)
				if _, err := h.c.Writer.Write(bytes); err != nil {
					l.Warnf("write response failed: %s", err.Error())
					return
				}
			default:
				l.Warnf("got unknow message type: %d", messageType)
				h.fail(500, "server error")
				return
			}
		}
	}
}

func getNewWebsocketConn(datakit *ws.DataKit, action string) (*websocket.Conn, error) {
	timeout := 30 * time.Second
	if client, ok := Manager.Clients[datakit.ConnID]; !ok {
		return nil, fmt.Errorf("datakit not found")
	} else {
		query := url.Values{}
		connID := cliutils.XID("connect_id_")
		query.Add(ws.HeaderNewWebSocketConnectionID, connID)
		query.Add(ws.HeaderWebsocketAction, action)

		msg := ws.WebsocketMessage{
			Action: ws.NewWebsocketConnectionAction,
			Data:   ActionData{Query: query},
		}

		timeoutCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		select {
		case client.Send <- msg.Bytes():
		case <-timeoutCtx.Done():
			return nil, ErrRequestTimeout
		}

		ch := Manager.initWebsocketConnChan(connID)
		defer Manager.deleteWebsocketConnChan(connID)

		ctx1, cancel1 := context.WithTimeout(context.Background(), timeout)
		defer cancel1()
		select {
		case <-ctx1.Done():
			return nil, ErrRequestTimeout
		case newConn := <-ch:
			return newConn, nil
		}
	}
}

func getLastDatakitVersionHandler(c *gin.Context) {
	h := newHandler(c)

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/datakit/version", staticBaseURL), nil)
	if err != nil {
		h.fail(500, "failed to new request")
		return
	}

	client := &http.Client{Timeout: 30 * time.Second}

	res, err := client.Do(req)
	if err != nil {
		h.fail(500, err.Error())
		return
	}
	defer res.Body.Close() //nolint:errcheck

	if body, err := io.ReadAll(res.Body); err != nil {
		l.Warnf("failed to read response body %s", err.Error())
		h.fail(500, fmt.Sprintf("read response body %s", err.Error()))
		return
	} else {
		var v version.VerInfo
		if err := json.Unmarshal(body, &v); err != nil {
			l.Warnf("failed to unmarshal version info %s", err.Error())
			h.fail(500, fmt.Sprintf("unmarshal version info %s", err.Error()))
			return
		}
		h.success(v)
	}
}

func websocketLogHandler(ctx *gin.Context) {
	timeout := 30 * time.Second
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		l.Errorf("failed to upgrade websocket connection: %s", err.Error())
		ctx.String(http.StatusInternalServerError, "Failed to upgrade websocket connection")
		return
	}
	closeCh := make(chan interface{})

	g.Go(func(ctx context.Context) error {
		time.Sleep(time.Second * 5)
		if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(10*time.Second)); err != nil {
			l.Warnf("failed to send ping message: %s", err.Error())
			close(closeCh)
		}
		return nil
	})

	logType := ctx.Query("type")
	l.Debugf(logType)

	h := newHandler(ctx)
	datakit := h.getDatakit()
	if datakit == nil {
		h.fail(400, "datakit not found")
		return
	}

	newConn, err := getNewWebsocketConn(datakit, ws.GetDatakitLogTailAction)
	if err != nil {
		h.fail(500, err.Error())
		return
	}
	defer newConn.Close() //nolint:errcheck

	msg := ws.WebsocketMessage{
		Action: ws.GetDatakitLogTailAction,
		Data:   ActionData{Query: ctx.Request.URL.Query()},
	}

	if err := newConn.WriteMessage(websocket.TextMessage, msg.Bytes()); err != nil {
		l.Warnf("failed to write message: %s, exit", err.Error())
		h.fail(500, "server error")
		return
	}

	data := []byte{}
	resp := ws.DCAResponse{
		Content: &data,
	}
	dest := ws.WebsocketMessage{
		Action: ws.GetDatakitLogTailAction,
		Data:   &resp,
	}

	timeoutTicker := time.NewTicker(timeout)
	defer timeoutTicker.Stop()
	var actionError error

	defer func() {
		if actionError != nil {
			if err := conn.WriteMessage(websocket.TextMessage, []byte(actionError.Error())); err != nil {
				l.Warnf("failed to write message: %s, exit", err.Error())
				return
			}
		}
	}()

	for {
		_, bytes, err := newConn.ReadMessage()
		if err != nil {
			l.Warnf("failed to read message: %s, exit", err.Error())
			actionError = fmt.Errorf("get log failed: %w", err)
			return
		}
		if err := json.Unmarshal(bytes, &dest); err != nil {
			l.Errorf("failed to unmarshal message: %s", err.Error())
			actionError = errors.New("get log failed")
			return
		}

		if msg.Action != dest.Action {
			l.Errorf("message action not match: %s, %s", msg.Action, dest.Action)
			actionError = errors.New("get log failed")
			return
		}

		if !resp.Success {
			l.Warnf("failed to get log data: %s", resp.Message)
			actionError = fmt.Errorf("get log failed: %s", resp.Message)
			return
		}

		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			l.Warnf("failed to write message: %s, exit", err.Error())
			return
		}
	}
}

func getHandler(action string) gin.HandlerFunc {
	// register action handler
	if ActionHandlerMap[action] == nil {
		ActionHandlerMap[action] = getActionHandler(action)
	}

	return func(ctx *gin.Context) {
		h := newHandler(ctx)
		datakit := h.getDatakit()
		if datakit == nil {
			h.fail(400, "datakit not found")
			return
		}
		res, err := Manager.Action(action, datakit, ctx)
		if err != nil {
			h.fail(500, err.Error())
			return
		}
		h.send(res)
	}
}

// datakitHandler is the handler for datakit related actions.
func datakitHandler(action string) gin.HandlerFunc {
	return getHandler(action)
}

func datakitByIDHandler(c *gin.Context) {
	h := newHandler(c)
	idStr := c.Query("ids")
	ids := strings.Split(idStr, ",")

	var values []interface{}
	inPlaceholder := []string{}
	for _, id := range ids {
		inPlaceholder = append(inPlaceholder, "?")
		values = append(values, id)
	}

	workspaceUUID, err := c.Cookie(cookieWorkspaceUUID)
	if err != nil || workspaceUUID == "" {
		h.fail(400, "workspace_uuid not found")
		return
	}

	values = append(values, workspaceUUID)

	res := []ws.DataKit{}

	if err := datakitDB.Select(
		fmt.Sprintf("select * from datakit where id in (%s) and workspace_uuid=?", strings.Join(inPlaceholder, ",")), &res,
		values...); err != nil {
		l.Errorf("failed to query datakit list : %s", err.Error())
		h.fail(500, "failed to query datakit list total")
		return
	}

	h.success(res)
}

// datakitListHandler is the handler for datakit list.
func datakitListHandler(c *gin.Context) {
	h := newHandler(c)

	pageIndexNum := 1
	pageSizeNum := 10

	pageIndex := c.Query("pageIndex")
	pageSize := c.Query("pageSize")

	workspaceUUID, err := c.Cookie(cookieWorkspaceUUID)
	if err != nil || workspaceUUID == "" {
		h.fail(400, "workspace_uuid not found")
		return
	}

	if pageIndex != "" {
		if v, err := strconv.Atoi(pageIndex); err == nil {
			pageIndexNum = v
		}
	}

	if pageSize != "" {
		if v, err := strconv.Atoi(pageSize); err == nil {
			pageSizeNum = v
		}
	}

	var where strings.Builder
	whereValues := []interface{}{}
	where.WriteString("workspace_uuid=?")
	whereValues = append(whereValues, workspaceUUID)

	// filter out datakits that are not updated for more than maxUpdatedAtInterval
	where.WriteString(" and updated_at > ?")
	whereValues = append(whereValues, time.Now().Add(-maxUpdatedAtInterval).UnixMilli())

	if search := c.Query("search"); search != "" {
		placeholder := fmt.Sprintf("%%%s%%", search)
		where.WriteString(` and (host_name like ? or ip like ?)`)
		whereValues = append(whereValues, placeholder, placeholder)
	}

	whereSQL := ""
	if where.Len() > 0 {
		whereSQL = fmt.Sprintf("where %s", where.String())
	}

	// get total number
	sql := fmt.Sprintf("select count(*) total from datakit %s", whereSQL)
	totalCount := 0
	totalRes := []int{}

	if err := datakitDB.Select(sql, &totalRes, whereValues...); err != nil {
		l.Errorf("failed to query datakit list total: %s", err.Error())
		h.fail(500, "failed to query datakit list total")
		return
	} else if len(totalRes) == 1 {
		totalCount = totalRes[0]
	}

	res := []ws.DataKit{}
	if totalCount != 0 {
		//nolint:lll
		sql = fmt.Sprintf("select id,runtime_id,arch,host_name,os,version,ip,start_time,run_in_container,run_mode,usage_cores,updated_at,workspace_uuid,status,url from datakit %s limit %d offset %d",
			whereSQL, pageSizeNum, (pageIndexNum-1)*pageSizeNum)

		err := datakitDB.Select(sql, &res, whereValues...)
		if err != nil {
			l.Errorf("failed to query datakit list: %s", err.Error())
			h.fail(500, "failed to query datakit list")
			return
		}
	}

	h.success(pageContent{
		Data: res,
		PageInfo: pageInfo{
			Count:      len(res),
			PageIndex:  pageIndexNum,
			PageSize:   pageSizeNum,
			TotalCount: totalCount,
		},
	})
}
