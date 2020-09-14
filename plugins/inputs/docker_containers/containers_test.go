package docker_containers

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

func TestMain(t *testing.T) {
	io.TestOutput()

	var d = DockerContainers{
		Endpoint:     defaultEndpoint,
		All:          false,
		Timeout:      "5s",
		Interval:     "5s",
		newEnvClient: NewEnvClient,
		newClient:    NewClient,
	}

	d.Run()
}
