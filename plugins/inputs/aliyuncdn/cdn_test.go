package aliyuncdn

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/naoina/toml"
)

func TestConfig(t *testing.T) {
	var cdn CDN

	data, err := ioutil.ReadFile("./cdn.toml")
	if err != nil {
		log.Fatalf("%s", err)
	}

	err = toml.Unmarshal(data, &cdn)
	if err != nil {
		log.Fatalf("%s", err)
	}
}
