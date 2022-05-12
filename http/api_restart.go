// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package http

import (
	"fmt"

	"github.com/gin-gonic/gin"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
)

func apiRestart(c *gin.Context) {
	if err := checkToken(c.Request); err != nil {
		uhttp.HttpErr(c, err)
		return
	}

	if err := restartDataKit(); err != nil {
		uhttp.HttpErr(c, fmt.Errorf("restart datakit failed: %w", err))
		return
	}

	OK.HttpBody(c, nil)
}
