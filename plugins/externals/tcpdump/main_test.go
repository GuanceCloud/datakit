package main

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

func TestHandle(t *testing.T) {
	logger.SetGlobalRootLogger("",
		"debug",
		logger.OPT_ENC_CONSOLE|logger.OPT_SHORT_CALLER)
	l := logger.SLogger("tcpdump")

	l.Info("start....")

	t.Run("case-tracerouter", func(t *testing.T) {
		dump := NetPacket{}
		// scan.Targets = []string{"127.0.0.1"}
		dump.exec()
		t.Log("ok")
	})
}
