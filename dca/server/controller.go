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

const (
	errorCodeServerError = "server.error"
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
			h.fatal(500, fmt.Sprintf("unknown api: %s", name))
			return
		} else {
			res, err := h.pipe(api[0], api[1], h.c.Request.Body)
			if err != nil {
				h.fatal(500, err.Error())
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
		h.fatal(500, err.Error())
		return
	}

	workspaceUUID := c.GetHeader("X-Workspace-Uuid")
	req.Header.Set("X-Workspace-Uuid", workspaceUUID)

	respbody, err := h.doRequest(req)
	if err != nil {
		h.fatal(500, err.Error())
		return
	}

	resp := Response{}
	if err := json.Unmarshal(respbody, &resp); err != nil {
		h.fatal(500, err.Error())
		return
	}

	if !resp.Success {
		h.fatal(500, resp.Message)
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
		h.fatal(400, "datakit not found")
		return
	}

	logType := "log"
	if h.c.Query("type") == "gin.log" {
		logType = "gin.log"
	}

	conn, err := getNewWebsocketConn(datakit, ws.GetDatakitLogDownloadAction)
	if err != nil {
		h.fatal(500, fmt.Sprintf("get new websocket connection error: %s", err.Error()))
		return
	}

	defer conn.Close() //nolint:errcheck

	msg := ws.WebsocketMessage{
		Action: ws.GetDatakitLogDownloadAction,
		Data:   ActionData{Query: ctx.Request.URL.Query()},
	}

	if err := conn.WriteMessage(websocket.TextMessage, msg.Bytes()); err != nil {
		h.fatal(500, fmt.Sprintf("write message to websocket error: %s", err.Error()))
		l.Warnf("failed to write message to websocket: %s", err.Error())
		return
	}

	for {
		if messageType, bytes, err := conn.ReadMessage(); err != nil {
			h.fatal(500, fmt.Sprintf("read message from websocket error: %s", err.Error()))
			l.Warnf("failed to read message from websocket: %s", err.Error())
			return
		} else {
			l.Info("received messages, length: %d, messageType: %s", len(bytes), messageType)
			switch messageType {
			case websocket.TextMessage:
				dest := ws.WebsocketMessage{}
				if err := json.Unmarshal(bytes, &dest); err != nil {
					l.Errorf("failed to unmarshal message: %s", err.Error())
					h.fatal(500, fmt.Sprintf("failed to unmarshal message: %s", err.Error()))
					return
				} else if msg.Action != dest.Action {
					l.Errorf("message action not match: %s, %s", msg.Action, dest.Action)
					h.fatal(500, "message action not match")
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
				h.fatal(500, fmt.Sprintf("unknown message type: %d", &messageType))
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
		h.fatal(500, "failed to new request")
		return
	}

	client := &http.Client{Timeout: 30 * time.Second}

	res, err := client.Do(req)
	if err != nil {
		h.fatal(500, err.Error())
		return
	}
	defer res.Body.Close() //nolint:errcheck

	if body, err := io.ReadAll(res.Body); err != nil {
		l.Warnf("failed to read response body %s", err.Error())
		h.fatal(500, fmt.Sprintf("read response body %s", err.Error()))
		return
	} else {
		var v version.VerInfo
		if err := json.Unmarshal(body, &v); err != nil {
			l.Warnf("failed to unmarshal version info %s", err.Error())
			h.fatal(500, fmt.Sprintf("unmarshal version info %s", err.Error()))
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
		h.fatal(400, "datakit not found")
		return
	}

	newConn, err := getNewWebsocketConn(datakit, ws.GetDatakitLogTailAction)
	if err != nil {
		h.fatal(500, err.Error())
		return
	}
	defer newConn.Close() //nolint:errcheck

	msg := ws.WebsocketMessage{
		Action: ws.GetDatakitLogTailAction,
		Data:   ActionData{Query: ctx.Request.URL.Query()},
	}

	if err := newConn.WriteMessage(websocket.TextMessage, msg.Bytes()); err != nil {
		l.Warnf("failed to write message: %s, exit", err.Error())
		h.fatal(500, fmt.Sprintf("failed to write message: %s", err.Error()))
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
			h.fatal(400, "datakit not found")
			return
		}
		res, err := Manager.Action(action, datakit, ctx)
		if err != nil {
			h.fatal(500, err.Error())
			return
		}
		h.send(res)
	}
}

// datakitHandler is the handler for datakit related actions.
func datakitHandler(action string) gin.HandlerFunc {
	return getHandler(action)
}

type fieldValue struct {
	Value string `db:"value"`
}

type globalHostTagValue struct {
	Key   string `db:"key"`
	Value string `db:"value"`
}

func datakitSearchValueHandler(c *gin.Context) {
	h := newHandler(c)

	workspaceUUID, err := c.Cookie(cookieWorkspaceUUID)
	if err != nil || workspaceUUID == "" {
		h.fatal(400, "workspace_uuid not found")
		return
	}

	searchValues := map[string][]string{}

	// get host and os values

	fields := []string{"host_name", "os", "version"}

	for _, field := range fields {
		res := []fieldValue{}
		if err := datakitDB.Select(
			fmt.Sprintf("select distinct %s as value from datakit where workspace_uuid=?", field),
			&res, workspaceUUID); err != nil {
			l.Errorf("failed to query %s: %s", field, err.Error())
			h.fatal(500, fmt.Sprintf("failed to query %s", field))
			return
		}

		for _, v := range res {
			searchValues[field] = append(searchValues[field], v.Value)
		}
	}

	// get global tags values
	globalHostTagsRes := []globalHostTagValue{}
	if err := datakitDB.Select("select distinct t.key,value from global_host_tags,json_each(global_host_tags.tags) as t where workspace_uuid=?",
		&globalHostTagsRes, workspaceUUID); err != nil {
		l.Errorf("failed to query tag: %s", err.Error())
		h.fatal(500, "failed to query tag")
		return
	}

	tags := map[string]map[string]string{}

	for _, v := range globalHostTagsRes {
		if _, ok := tags[v.Key]; !ok {
			tags[v.Key] = map[string]string{}
		}
		tags[v.Key][v.Value] = ""
	}

	for k, v := range tags {
		if _, ok := searchValues[k]; !ok {
			searchValues[k] = []string{}
			for k1 := range v {
				searchValues[k] = append(searchValues[k], k1)
			}
		}
	}

	h.success(searchValues)
}

func datakitByIDHandler(c *gin.Context) {
	h := newHandler(c)
	idStr := c.Query("ids")

	workspaceUUID, err := c.Cookie(cookieWorkspaceUUID)
	if err != nil || workspaceUUID == "" {
		h.fatal(400, "workspace_uuid not found")
		return
	}

	res, err := getDatakits(idStr, workspaceUUID)
	if err != nil {
		l.Errorf("failed to query datakit list: %s", err.Error())
		h.fatal(500, "failed to query datakit list")
	}

	h.success(res)
}

func getDatakits(idStr, workspaceUUID string) ([]*ws.DataKit, error) {
	query := "select * from datakit where workspace_uuid=?"

	values := []interface{}{workspaceUUID}

	if idStr != "all" {
		ids := strings.Split(idStr, ",")
		inPlaceholder := []string{}
		for _, id := range ids {
			inPlaceholder = append(inPlaceholder, "?")
			values = append(values, id)
		}

		query += fmt.Sprintf(" and id in (%s)", strings.Join(inPlaceholder, ","))
	}

	res := []*ws.DataKit{}

	if err := datakitDB.Select(
		query, &res,
		values...); err != nil {
		l.Errorf("failed to query datakit list : %s", err.Error())
		return nil, fmt.Errorf("failed to query datakit list: %w", err)
	}
	return res, nil
}

var operationMap = map[string]string{
	"reload":  ws.ReloadDatakitAction,
	"upgrade": ws.UpgradeDatakitAction,
}

func datakitOperationHandler(c *gin.Context) {
	h := newHandler(c)
	operationType := c.Params.ByName("type")

	action, ok := operationMap[operationType]

	if !ok {
		h.fatal(http.StatusBadRequest, fmt.Sprintf("invalid operation type: %s", operationType))
		return
	}

	idStr := c.Query("ids")
	workspaceUUID, err := c.Cookie(cookieWorkspaceUUID)
	if err != nil || workspaceUUID == "" {
		h.fatal(http.StatusBadRequest, "workspace_uuid not found")
		return
	}

	datakits, err := getDatakits(idStr, workspaceUUID)
	if err != nil {
		h.fatal(http.StatusInternalServerError, fmt.Sprintf("failed to query datakit list: %s", err.Error()))
		return
	}

	for _, dk := range datakits {
		if !isOperationAllowed(dk.Status) {
			l.Infof("Operation is not allowed, current status is %s", dk.Status.String())
			continue
		}
		if _, err := Manager.Action(action, dk, c); err != nil {
			h.fatal(http.StatusInternalServerError, fmt.Sprintf("failed to %s datakit: %s", action, err.Error()))
			return
		}
	}

	h.success()
}

type filterParam struct {
	Relation string       `json:"relation"`
	Items    []filterItem `json:"items"`
}

type filterItem struct {
	Field    string   `json:"field"`
	Operator string   `json:"operator"`
	Value    []string `json:"value"`
}

// datakitListHandler is the handler for datakit list.
func datakitListHandler(c *gin.Context) {
	h := newHandler(c)
	var sql, whereSQL string

	pageIndexNum := 1
	pageSizeNum := 10
	totalCount := 0
	res := []ws.DataKit{}
	totalRes := []int{}

	pageIndex := c.Query("pageIndex")
	pageSize := c.Query("pageSize")

	workspaceUUID, err := c.Cookie(cookieWorkspaceUUID)
	if err != nil || workspaceUUID == "" {
		h.fatal(400, "workspace_uuid not found")
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

	filter := c.Query("filter")
	search := c.Query("search")

	whereSQL, whereValues, err := getWhereSQL(workspaceUUID, filter, search)
	if err != nil {
		h.fatal(400, err.Error())
		return
	}

	// at least one condition, eg workspace_uuid
	if whereSQL == "" {
		goto end
	}

	// get total number
	sql = fmt.Sprintf("select count(*) total from datakit %s", whereSQL)

	if err := datakitDB.Select(sql, &totalRes, whereValues...); err != nil {
		l.Errorf("failed to query datakit list total: %s", err.Error())
		h.fatal(500, "failed to query datakit list total")
		return
	} else if len(totalRes) == 1 {
		totalCount = totalRes[0]
	}

	if totalCount != 0 {
		sql := `
			select id,runtime_id,arch,host_name,os,version,
				ip,start_time,run_in_container,run_mode,usage_cores,
				updated_at,workspace_uuid,status,url,global_host_tags 
			from datakit %s limit %d offset %d`
		sql = fmt.Sprintf(sql,
			whereSQL, pageSizeNum, (pageIndexNum-1)*pageSizeNum)

		err := datakitDB.Select(sql, &res, whereValues...)
		if err != nil {
			l.Errorf("failed to query datakit list: %s", err.Error())
			h.fatal(500, "failed to query datakit list")
			return
		}
	}

end:
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

func getWhereSQL(workspaceUUID, filter, search string) (string, []any, error) {
	var whereSQL string
	var where strings.Builder
	whereValues := []interface{}{}
	where.WriteString("workspace_uuid=?")
	whereValues = append(whereValues, workspaceUUID)

	// filter out datakits that are not updated for more than maxUpdatedAtInterval
	where.WriteString(" and updated_at > ?")
	whereValues = append(whereValues, time.Now().Add(-maxUpdatedAtInterval).UnixMilli())

	if search != "" {
		placeholder := fmt.Sprintf("%%%s%%", search)
		where.WriteString(` and (host_name like ? or ip like ?)`)
		whereValues = append(whereValues, placeholder, placeholder)
	}

	if filter != "" {
		filterParam := &filterParam{}
		if err := json.Unmarshal([]byte(filter), filterParam); err != nil {
			return "", whereValues, fmt.Errorf("invalid filed filter: %w", err)
		}

		if filterParam.Relation != "and" && filterParam.Relation != "or" {
			return "", whereValues, fmt.Errorf("invalid field relation %s", filterParam.Relation)
		}

		whereSQLs := []string{}
		globalHostTagsWhereSQL := []string{}
		whereSQLString := ""
		whereGlobalHostTagsWhereSQLString := ""
		values := []any{}
		golbalHostTagsValues := []any{workspaceUUID}
		for _, item := range filterParam.Items {
			placeHolders := []string{}
			wheres := []interface{}{}
			for i := 0; i < len(item.Value); i++ {
				placeHolders = append(placeHolders, "?")
				wheres = append(wheres, item.Value[i])
			}
			if len(wheres) == 0 {
				return "", whereValues, fmt.Errorf("invalid field value %s", item.Value)
			}
			inOperator := ""
			matchOperator := ""

			switch item.Operator {
			case "not_in":
				inOperator = "NOT IN"
			case "in":
				inOperator = "IN"
			case "match":
				matchOperator = "REGEXP"
			case "not_match":
				matchOperator = "NOT REGEXP"
			}

			if inOperator == "" && matchOperator == "" {
				return "", whereValues, fmt.Errorf("invalid operator %s", item.Operator)
			}

			fieldName := item.Field
			switch fieldName {
			case "host_name", "os", "version":
				if inOperator != "" {
					whereSQLs = append(whereSQLs, fmt.Sprintf("%s %s (%s)", fieldName, inOperator, strings.Join(placeHolders, ",")))
					values = append(values, wheres...)
				}

				if matchOperator != "" {
					whereSQLs = append(whereSQLs, fmt.Sprintf("%s %s ?", fieldName, matchOperator))
					values = append(values, item.Value[0])
				}
			default: // from global host tags
				if inOperator != "" {
					globalHostTagsWhereSQL = append(globalHostTagsWhereSQL,
						fmt.Sprintf("(json_extract(tags,'$.%s') %s (%s))",
							fieldName, inOperator, strings.Join(placeHolders, ",")))
					golbalHostTagsValues = append(golbalHostTagsValues, wheres...)
				}
				if matchOperator != "" {
					globalHostTagsWhereSQL = append(globalHostTagsWhereSQL,
						fmt.Sprintf("(json_extract(tags,'$.%s') %s ?)", fieldName,
							matchOperator))
					golbalHostTagsValues = append(golbalHostTagsValues, item.Value[0])
				}
			}
		}

		if len(globalHostTagsWhereSQL) > 0 {
			whereGlobalHostTagsWhereSQLString = strings.Join(globalHostTagsWhereSQL, fmt.Sprintf(" %s ", filterParam.Relation))
			l.Info(whereGlobalHostTagsWhereSQLString)
			globalHostTagsSQL := fmt.Sprintf("select distinct conn_id from global_host_tags where workspace_uuid=? and (%s)",
				whereGlobalHostTagsWhereSQLString)

			dkConnIDs := []string{}
			if err := datakitDB.Select(globalHostTagsSQL, &dkConnIDs, golbalHostTagsValues...); err != nil {
				l.Errorf("failed to query global host tags: %s", err.Error())
				return "", whereValues, fmt.Errorf("failed to query global host tags: %w", err)
			}

			if len(dkConnIDs) > 0 {
				whereSQLs = append(whereSQLs, fmt.Sprintf("(conn_id in ('%s'))", strings.Join((dkConnIDs), "','")))
			} else if filterParam.Relation == "and" {
				return "", nil, nil
			}
		}

		if len(whereSQLs) > 0 {
			whereSQLString = strings.Join(whereSQLs, fmt.Sprintf(" %s ", filterParam.Relation))
		}

		where.WriteString(fmt.Sprintf(" AND (%s)", whereSQLString))
		whereValues = append(whereValues, values...)
	}

	if where.Len() > 0 {
		whereSQL = fmt.Sprintf("where %s", where.String())
	}

	return whereSQL, whereValues, nil
}
