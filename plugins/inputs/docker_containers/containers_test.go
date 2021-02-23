package docker_containers

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"testing"
)

/*
 * go test -v -c && sudo ./docker_containers.test -test.v -test.run=TestData
 */

func TestMain(t *testing.T) {
	io.TestOutput()
	var d = DockerContainers{
		Endpoint:     defaultEndpoint,
		All:          false,
		Interval:     "5s",
		newEnvClient: NewEnvClient,
		newClient:    NewClient,
	}

	d.Run()
}

func TestData(t *testing.T) {
	var d = DockerContainers{
		Endpoint:     defaultEndpoint,
		All:          false,
		Interval:     "5m",
		newEnvClient: NewEnvClient,
		newClient:    NewClient,
	}

	data, err := d.Test()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%s", data.Result)
}
