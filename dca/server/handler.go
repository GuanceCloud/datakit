// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package server

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	ws "gitlab.jiagouyun.com/cloudcare-tools/datakit/dca/websocket"
)

var ErrorMessage = map[string]string{
	"server.error": "server error",
}

// handler helps to handle response.
type handler struct {
	c *gin.Context // gin context
}
type ResponseError struct {
	Code      int
	ErrorCode string
	ErrorMsg  string
}
type Response struct {
	ws.DCAResponse
	TraceID string `json:"traceId"`
}

type pageInfo struct {
	Count      int `json:"count"`
	PageSize   int `json:"pageSize"`
	PageIndex  int `json:"pageIndex"`
	TotalCount int `json:"totalCount"`
}
type pageContent struct {
	Data     any      `json:"data"`
	PageInfo pageInfo `json:"pageInfo"`
}

// newHandler creates a new handler.
func newHandler(c *gin.Context) *handler {
	return &handler{
		c: c,
	}
}

func (h *handler) getDatakit() *ws.DataKit {
	datakitID := h.c.Query("datakit_id")
	rows := []*ws.DataKit{}
	if err := datakitDB.Select("select * from datakit where id=? and workspace_uuid=?",
		&rows, datakitID, h.getCookie(cookieWorkspaceUUID)); err != nil {
		l.Errorf("failed to query datakit: %s", err.Error())
		return nil
	}

	if len(rows) == 0 {
		l.Warnf("failed to query datakit[id: %s]: datakit not found", datakitID)
		return nil
	}

	return rows[0]
}

func (h *handler) success(contents ...any) {
	var content any
	if len(contents) > 0 {
		content = contents[0]
	}

	h.c.JSON(http.StatusOK, ws.DCAResponse{
		Success:   true,
		ErrorCode: "",
		Content:   content,
		Code:      200,
	})
}

func (h *handler) fail(code int, msg string) {
	h.c.JSON(http.StatusOK, ws.DCAResponse{
		Success:   false,
		ErrorCode: "server.error",
		Message:   msg,
		Code:      code,
	})
}

func (h *handler) send(res *ws.DCAResponse) {
	if res == nil {
		res = &ws.DCAResponse{
			Success:   false,
			ErrorCode: "server.error",
			Message:   "server error",
			Code:      500,
		}
	}

	h.c.JSON(http.StatusOK, res)
}

func (h *handler) getCookie(name string) string {
	if v, err := h.c.Cookie(name); err != nil {
		l.Warnf("get cookie %s failed: %s", name, err.Error())
	} else {
		return v
	}
	return ""
}

func (h *handler) doRequest(req *http.Request) ([]byte, error) {
	req.Header.Set("X-FT-Auth-Token", h.getCookie(cookieFrontToken))
	if req.Header.Get("X-Workspace-Uuid") == "" {
		req.Header.Set("X-Workspace-Uuid", h.getCookie(cookieWorkspaceUUID))
	}
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")

	res, err := consoleClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close() //nolint:errcheck

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, err
}

func (h *handler) pipe(method string, path string, content io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, getConsoleAPIURL(path), content)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-FT-Auth-Token", h.getCookie(cookieFrontToken))
	req.Header.Set("X-Workspace-Uuid", h.getCookie(cookieWorkspaceUUID))
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")

	res, err := consoleClient.Do(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}
