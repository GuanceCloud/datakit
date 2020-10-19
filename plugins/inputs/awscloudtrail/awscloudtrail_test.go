package awscloudtrail

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/influxdata/toml"
)

func TestSvr(t *testing.T) {

	ag := newInstance()
	ag.mode = "debug"

	if data, err := ioutil.ReadFile("./test.conf"); err != nil {
		log.Fatalf("%s", err)
	} else {
		if toml.Unmarshal(data, ag); err != nil {
			log.Fatalf("%s", err)
		}
	}

	ag.Run()
}
