package envoy

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

func __init() {
	logger.SetGlobalRootLogger("", logger.DEBUG, logger.OPT_DEFAULT)
	l = logger.SLogger(inputName)
	testAssert = true
}

func TestMain(t *testing.T) {

	__init()

	var envoyer = Envoy{
		Host:     "127.0.0.1",
		Port:     9901,
		Interval: "10s",
		TLSOpen:  false,
	}

	envoyer.Run()

}
