package http

import (
	"fmt"
	"io/ioutil"

	"github.com/gin-gonic/gin"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
)

// create or update object labels
func apiCreateOrUpdateObjectLabel(c *gin.Context) {
	// 1) get request body
	// 2) reroute to dataway
	body, err := uhttp.GinRead(c)
	if err != nil {
		l.Errorf("GinRead: %s", err.Error())
		uhttp.HttpErr(c, err)
		return
	}

	if dw == nil {
		uhttp.HttpErr(c, fmt.Errorf("dataway not set"))
		return
	}

	tkns := dw.GetToken()
	if len(tkns) == 0 {
		uhttp.HttpErr(c, fmt.Errorf("dataway token missing"))
		return
	}

	resp, err := dw.UpsertObjectLabels(tkns[0], body)
	if err != nil {
		l.Errorf("create or update object labels: %s", err)
		uhttp.HttpErr(c, err)
		return
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Errorf("read response body %s", err)
		uhttp.HttpErr(c, uhttp.Error(ErrBadReq, err.Error()))
		return
	}
	defer resp.Body.Close()

	c.Data(resp.StatusCode, "application/json", respBody)
}

// delete object label
func apiDeleteObjectLabel(c *gin.Context) {
	// 1) get request body
	// 2) reroute to dataway
	body, err := uhttp.GinRead(c)
	if err != nil {
		l.Errorf("GinRead: %s", err.Error())
		uhttp.HttpErr(c, err)
		return
	}

	if dw == nil {
		uhttp.HttpErr(c, fmt.Errorf("dataway not set"))
		return
	}

	tkns := dw.GetToken()
	if len(tkns) == 0 {
		uhttp.HttpErr(c, fmt.Errorf("dataway token missing"))
		return
	}

	resp, err := dw.DeleteObjectLabels(tkns[0], body)
	if err != nil {
		l.Errorf("create or update object labels: %s", err)
		uhttp.HttpErr(c, err)
		return
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Errorf("read response body %s", err)
		uhttp.HttpErr(c, uhttp.Error(ErrBadReq, err.Error()))
		return
	}
	defer resp.Body.Close()

	c.Data(resp.StatusCode, "application/json", respBody)
}
