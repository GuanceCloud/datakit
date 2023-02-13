// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package http

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	uhttp "github.com/GuanceCloud/cliutils/network/http"
)

type Workspace struct {
	Token []string `json:"token"`
}

func apiWorkspace(c *gin.Context) {
	var w Workspace
	if dw == nil {
		uhttp.HttpErr(c, fmt.Errorf("dataway not set"))
		return
	}
	tkns := dw.GetTokens()
	if len(tkns) == 0 {
		uhttp.HttpErr(c, fmt.Errorf("dataway token missing"))
		return
	}
	w.Token = tkns
	j, err := json.Marshal(w)
	if err != nil {
		l.Errorf("json.Marshal: %s", err.Error())
		uhttp.HttpErr(c, err)
		return
	}

	l.Debugf("query: %s", string(j))

	resp, err := dw.WorkspaceQuery(j)
	if err != nil {
		l.Errorf("DQLQuery: %s", err)
		uhttp.HttpErr(c, err)
		return
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Errorf("read response body %s", err)
		uhttp.HttpErr(c, uhttp.Error(ErrBadReq, err.Error()))
		return
	}
	l.Debugf("read response body %s", string(respBody))
	defer resp.Body.Close() //nolint:errcheck

	switch resp.StatusCode / 100 {
	case 2:
		l.Debugf("workspace ok: %s", resp.Status)
	default:
		l.Errorf("workspace fail: %s", resp.Status)
		uhttp.HttpErr(c, uhttp.Error(ErrBadReq, fmt.Sprintf("%s%d%s", "http_request_", resp.StatusCode, "_err")))
		return
	}
	c.Data(http.StatusOK, "application/json", respBody)
}
