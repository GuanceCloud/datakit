package http

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
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
	tkns := dw.GetToken()
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
