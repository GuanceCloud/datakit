// +build linux

package containerd

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

func __init() {
	logger.SetGlobalRootLogger("", logger.DEBUG, logger.OPT_DEFAULT)
	l = logger.SLogger(inputName)
}

func TestMain(t *testing.T) {
	__init()
	testAssert = true

	var con = Containerd{
		HostPath:  "/run/containerd/containerd.sock",
		Namespace: "moby",
		IDList:    []string{"*"},
		Interval:  "5s",
	}

	con.Run()
}
