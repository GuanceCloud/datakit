package dockerContainers

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

func TestMain(t *testing.T) {
	io.TestOutput()

	var d = DockerContainers{
		Endpoint:     defaultEndpoint,
		Timeout:      "5s",
		Interval:     "5s",
		newEnvClient: NewEnvClient,
		newClient:    NewClient,
	}

	d.Run()
}
