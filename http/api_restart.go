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

	ErrOK.HttpBody(c, nil)
}
