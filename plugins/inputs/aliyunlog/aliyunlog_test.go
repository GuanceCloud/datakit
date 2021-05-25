package aliyunlog

import (
	//"fmt"

	"io/ioutil"
	"log"
	"testing"

	"github.com/influxdata/toml"
)

func testAgent() (*ConsumerInstance, error) {
	ag := NewAgent()

	data, err := ioutil.ReadFile("./test.conf")
	if err != nil {
		log.Fatalf("%s", err)
		return nil, err
	}

	err = toml.Unmarshal(data, ag)
	if err != nil {
		log.Fatalf("%s", err)
		return nil, err
	}
	return ag, nil
}

func TestService(t *testing.T) {

	ag, err := testAgent()
	if err != nil {
		t.Error(err)
		return
	}
	ag.mode = "debug"

	ag.Run()

}
