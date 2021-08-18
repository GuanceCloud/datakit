package http

import (
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/kardianos/service"

	dkservice "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/service"
)

func apiRestart(c *gin.Context) {

	ErrOK.HttpBody(c, nil)

	svc, err := dkservice.NewService()
	if err != nil {
		l.Errorf("new %s service failed: %s", runtime.GOOS, err.Error())
		return
	}

	l.Info("stoping datakit...")
	if err := service.Control(svc, "restart"); err != nil {
		l.Warnf("stop service: %s, ignored", err.Error())
	}
}
