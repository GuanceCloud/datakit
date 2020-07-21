package puppetagent

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

const location = "/opt/puppetlabs/puppet/cache/state/last_run_summary.yaml"

func __init() {
	logger.SetGlobalRootLogger("", logger.DEBUG, logger.OPT_DEFAULT)
	l = logger.SLogger(inputName)
}

func TestMain(t *testing.T) {
	__init()
	testAssert = true

	var puppet = PuppetAgent{
		Location: location,
		Tags: map[string]string{
			"tags1": "value1",
		},
	}

	puppet.Run()
}
