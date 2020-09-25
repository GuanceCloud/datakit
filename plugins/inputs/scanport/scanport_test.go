package scanport

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"

	"testing"
)

func TestHandle(t *testing.T) {
	logger.SetGlobalRootLogger("",
		"debug",
		logger.OPT_ENC_CONSOLE|logger.OPT_SHORT_CALLER)
	l = logger.SLogger("scanport")

	t.Run("case-tracerouter", func(t *testing.T) {
		scan := Scanport{}
		scan.Interval = "10s"

		scan.Targets = "127.0.0.1, 10.100.64.225"
		scan.Port = "60-10000"
		scan.Process = 100
		scan.handle()

		t.Log("ok")
	})
}
