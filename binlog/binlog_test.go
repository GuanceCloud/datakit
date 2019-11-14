package binlog

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/influxdata/toml"
)

func TestConfig(t *testing.T) {

	data, err := ioutil.ReadFile("./test.toml")
	if err != nil {
		log.Fatalln(err)
	}
	var cfg BinlogConfig
	if err = toml.Unmarshal(data, &cfg); err != nil {
		log.Fatalln(err)
	} else {
		log.Printf("%#v\n", len(cfg.Datasources))

		for _, s := range cfg.Datasources {
			log.Printf("%s", s.Addr)
		}
	}
}
