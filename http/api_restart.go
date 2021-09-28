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

	err := restartDataKit()
	if err != nil {
		uhttp.HttpErr(c,
			fmt.Errorf("restart datakit failed: %s", err.Error()))
		return
	}

	ErrOK.HttpBody(c, nil)
}
