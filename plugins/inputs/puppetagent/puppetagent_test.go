package puppetagent

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

const location = "/opt/puppetlabs/puppet/cache/state/last_run_summary.yaml"

func TestMain(t *testing.T) {
	io.TestOutput()

	var puppet = PuppetAgent{
		Location: location,
		Tags: map[string]string{
			"tags1": "value1",
		},
	}

	puppet.Run()
}
