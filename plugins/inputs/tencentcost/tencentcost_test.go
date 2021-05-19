package tencentcost

import (
	"io/ioutil"
	"testing"

	"github.com/influxdata/toml"
)

func TestServe(t *testing.T) {

	data, err := ioutil.ReadFile("test.conf")
	if err != nil {
		t.Error(err)
	}
	ag := newAgent()
	if err = toml.Unmarshal(data, &ag); err != nil {
		t.Error(err)
	}
	ag.Run()
}
