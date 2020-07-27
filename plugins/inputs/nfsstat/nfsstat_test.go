// +build linux

package nfsstat

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

	var nfs = NFSstat{
		Interval: "3s",
	}

	nfs.Run()
}
