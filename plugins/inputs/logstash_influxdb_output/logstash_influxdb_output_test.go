package logstash_influxdb_output

import (
	"testing"

	"github.com/gin-gonic/gin"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

func TestInput(t *testing.T) {

	moduleLogger = logger.DefaultSLogger("dummy")

	router := gin.New()
	router.POST("/write", WriteHandler)
	router.Run(":8080")
}
