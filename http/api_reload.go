// +build !windows

package http

import (
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func apiReload(c *gin.Context) {

	l.Debugf("apiReload...")

	dkpid := os.Getpid()

	syscall.Kill(dkpid, syscall.SIGHUP)

	ErrOK.HttpBody(c, nil)

	c.Redirect(http.StatusFound, "/monitor")
}

func RestartHttpServer() {
	HttpStop()

	l.Info("wait HTTP server to stopping...")
	<-stopOkCh // wait HTTP server stop ok

	l.Info("reload HTTP server...")

	reload = time.Now()
	reloadCnt++

	HttpStart()
}
