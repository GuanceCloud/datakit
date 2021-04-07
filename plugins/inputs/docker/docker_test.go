package docker

import (
	"fmt"
	"testing"
)

/*
 * go test -v -c && sudo ./docker.test -test.v -test.run=TestMain
 */

func TestMain(t *testing.T) {
	var err error
	var d = DockerUtil{
		Endpoint:     defaultEndpoint,
		newEnvClient: NewEnvClient,
		newClient:    NewClient,
	}

	d.client, err = d.newEnvClient()
	if err != nil {
		t.Fatal(err)
	}

	data, err := d.gather()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("%s", data)
}
