package docker

import (
	"fmt"
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

/*
 * go test -v -c && sudo ./docker.test -test.v -test.run=TestMain
 */

func TestMain(t *testing.T) {
	var err error
	var d = Inputs{
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

func TestGatherLog(t *testing.T) {
	io.SetTest()
	var err error
	var d = Inputs{
		Endpoint:        defaultEndpoint,
		newEnvClient:    NewEnvClient,
		newClient:       NewClient,
		timeoutDuration: time.Second * 10,
	}

	d.client, err = d.newEnvClient()
	if err != nil {
		t.Fatal(err)
	}

	if err = d.initLogOption(); err != nil {
		t.Fatal(err)
	}

	d.gatherLog()

	d.Stop()
}
