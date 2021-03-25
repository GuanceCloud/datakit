package tracerouter

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

func TestHandle(t *testing.T) {
	logger.SetGlobalRootLogger("",
		"debug",
		logger.OPT_DEFAULT)
	l = logger.SLogger("tracerouter")

	t.Run("case-tracerouter", func(t *testing.T) {
		trace := TraceRouter{}
		trace.Addr = "www.baidu.com"

		trace.handle()
		t.Log("ok")
	})
}
