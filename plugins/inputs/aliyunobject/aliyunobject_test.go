package aliyunobject

import (
	"fmt"
	"io/ioutil"
	"log"

	"testing"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func TestInput(t *testing.T) {

	ag := loadCfg("test.conf").(*objectAgent)
	ag.mode = "debug"
	ag.Run()
}

func loadCfg(file string) inputs.Input {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalf("read config file failed: %s", err.Error())
		return nil
	}

	tbl, err := toml.Parse(data)
	if err != nil {
		log.Fatalf("parse toml file failed, %s", err)
		return nil
	}

	creator := func() inputs.Input {
		return newAgent()
	}

	for field, node := range tbl.Fields {

		switch field {
		case "inputs":
			stbl, ok := node.(*ast.Table)
			if !ok {
				log.Fatalf("ignore bad toml node")
			} else {
				for inputName, v := range stbl.Fields {
					input, err := tryUnmarshal(v, inputName, creator)
					if err != nil {
						log.Fatalf("unmarshal input %s failed: %s", inputName, err.Error())
						return nil
					}
					return input
				}
			}

		default: // compatible with old version: no [[inputs.xxx]] header
			input, err := tryUnmarshal(node, "aa", creator)
			if err != nil {
				log.Fatalf("unmarshal input failed: %s", err.Error())
				return nil
			}
			return input
		}
	}

	return nil
}

func tryUnmarshal(tbl interface{}, name string, creator inputs.Creator) (inputs.Input, error) {

	tbls := []*ast.Table{}

	switch t := tbl.(type) {
	case []*ast.Table:
		tbls = tbl.([]*ast.Table)
	case *ast.Table:
		tbls = append(tbls, tbl.(*ast.Table))
	default:
		err := fmt.Errorf("invalid toml format: %v", t)
		return nil, err
	}

	for _, t := range tbls {
		input := creator()

		err := toml.UnmarshalTable(t, input)
		if err != nil {
			log.Fatalf("toml unmarshal failed: %v", err)
			return nil, err
		}
		return input, nil

	}
	return nil, nil
}
