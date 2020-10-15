package docker_containers

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

/*
 *  go test -v -c
 *  sudo ./docker_containers.test -test.v -test.run=TestTesting
 */

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

func TestTesting(t *testing.T) {
	var d = DockerContainers{
		Endpoint:     defaultEndpoint,
		All:          false,
		Timeout:      "5s",
		Interval:     "5s",
		newEnvClient: NewEnvClient,
		newClient:    NewClient,
	}

	data, err := d.Test()
	if err != nil {
		t.Fatal(err)

	}
	t.Logf("%s", data)
}
