package http

import (
	"fmt"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/kardianos/service"

	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	dkservice "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/service"
)

func apiRestart(c *gin.Context) {
	if err := checkToken(c.Request); err != nil {
		uhttp.HttpErr(c, err)
		return
	}

	svc, err := dkservice.NewService()
	if err != nil {
		uhttp.HttpErr(c,
			fmt.Errorf("new %s service failed: %s",
				runtime.GOOS, err.Error()))
		return
	}

	l.Info("new datakit servier ok...")

	ErrOK.HttpBody(c, nil)

	l.Info("stoping datakit...")
	if err := service.Control(svc, "restart"); err != nil {
		l.Warnf("stop service: %s, ignored", err.Error())
	}
}
