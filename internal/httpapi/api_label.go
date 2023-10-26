// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"fmt"
	"io"

	uhttp "github.com/GuanceCloud/cliutils/network/http"
	"github.com/gin-gonic/gin"
)

// create or update object labels.
func apiCreateOrUpdateObjectLabel(c *gin.Context) {
	// 1) get request body
	// 2) reroute to dataway
	body, err := uhttp.GinRead(c)
	if err != nil {
		l.Errorf("GinRead: %s", err.Error())
		uhttp.HttpErr(c, err)
		return
	}

	if apiServer.dw == nil {
		uhttp.HttpErr(c, fmt.Errorf("dataway not set"))
		return
	}

	tkns := apiServer.dw.GetTokens()
	if len(tkns) == 0 {
		uhttp.HttpErr(c, fmt.Errorf("dataway token missing"))
		return
	}

	resp, err := apiServer.dw.UpsertObjectLabels(tkns[0], body)
	if err != nil {
		l.Errorf("create or update object labels: %s", err)
		uhttp.HttpErr(c, err)
		return
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Errorf("read response body %s", err)
		uhttp.HttpErr(c, uhttp.Error(ErrBadReq, err.Error()))
		return
	}
	defer resp.Body.Close() //nolint:errcheck

	c.Data(resp.StatusCode, "application/json", respBody)
}

// delete object label.
func apiDeleteObjectLabel(c *gin.Context) {
	// 1) get request body
	// 2) reroute to dataway
	body, err := uhttp.GinRead(c)
	if err != nil {
		l.Errorf("GinRead: %s", err.Error())
		uhttp.HttpErr(c, err)
		return
	}

	if apiServer.dw == nil {
		uhttp.HttpErr(c, fmt.Errorf("dataway not set"))
		return
	}

	tkns := apiServer.dw.GetTokens()
	if len(tkns) == 0 {
		uhttp.HttpErr(c, fmt.Errorf("dataway token missing"))
		return
	}

	resp, err := apiServer.dw.DeleteObjectLabels(tkns[0], body)
	if err != nil {
		l.Errorf("create or update object labels: %s", err)
		uhttp.HttpErr(c, err)
		return
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Errorf("read response body %s", err)
		uhttp.HttpErr(c, uhttp.Error(ErrBadReq, err.Error()))
		return
	}
	defer resp.Body.Close() //nolint:errcheck

	c.Data(resp.StatusCode, "application/json", respBody)
}
