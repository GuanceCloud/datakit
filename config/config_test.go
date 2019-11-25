package config

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"
)

func TestCfg(t *testing.T) {

	// cfgstr := `
	// uuid = '1122'
	// ftdataway = 'http://localhost:9527'
	// log = 'aa'
	// log_level = 'info'
	// config_dir = '/ect/cfg'

	// `

	// Cfg.GlobalTags = make(map[string]string)

	// out, _ := toml.Marshal(&Cfg)
	// fmt.Println(string(out))
	if err := LoadConfig(`./cfg.toml`); err != nil {
		log.Fatalln(err)
	}

	fdata, _ := ioutil.ReadFile(`./cfg.toml`)

	tbl, err := toml.Parse(fdata)
	if err != nil {
		log.Fatalln(err)
	}

	if val, ok := tbl.Fields["global_tags"]; ok {
		subTable, ok := val.(*ast.Table)
		var tags map[string]string
		tags = map[string]string{}
		if ok {
			if err := toml.UnmarshalTable(subTable, tags); err != nil {
				log.Fatalln(err)
			} else {
				log.Println("global_tags:", tags)
			}
		}
	}

	log.Printf("%#v", Cfg)
}
