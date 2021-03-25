package aliyuncost

import (
	"io/ioutil"
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func TestInput(t *testing.T) {

	data, err := ioutil.ReadFile("test.conf")
	if err != nil {
		t.Error(err)
		return
	}

	ag, err := config.LoadInputConfig(data, func() inputs.Input { return newAgent("debug") })
	if err != nil {
		t.Error(err)
		return
	}
	ag[0].Run()
}
