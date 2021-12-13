package http

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

type ping struct {
	Version string `json:"version"`
	Uptime  string `json:"uptime"`
	Host    string `json:"host"`
}

func apiPing(c *gin.Context) {
	OK.HttpBody(c, &ping{Version: datakit.Version, Uptime: fmt.Sprintf("%v", time.Since(uptime)), Host: datakit.DatakitHostName})
}
