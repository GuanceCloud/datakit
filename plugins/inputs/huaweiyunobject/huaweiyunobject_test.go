package huaweiyunobject

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/influxdata/toml"
)

func TestConfig(t *testing.T) {

	var ag objectAgent

	ag.AccessKeyID = `xxx`
	ag.AccessKeySecret = `yyy`
	ag.Tags = map[string]string{
		"key1": "val1",
	}

	load := false

	if !load {
		ag.Ecs = &Ecs{
			InstancesIDs:       []string{"id1", "id2"},
			ExcludeInstanceIDs: []string{"exid1", "exid2"},
		}

		data, err := toml.Marshal(&ag)
		if err != nil {
			t.Error(err)
		}
		ioutil.WriteFile("test.conf", data, 0777)
	} else {
		data, err := ioutil.ReadFile("test.conf")
		if err != nil {
			t.Error(err)
		}
		if err = toml.Unmarshal(data, &ag); err != nil {
			t.Error(err)
		} else {
			log.Println("ok")
		}
	}

}

func TestInput(t *testing.T) {

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
