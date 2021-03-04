// +build expect_fail

//
// telegraf 对 name_prefix 有特殊处理，导致这里部分 toml sample （主要是 jolokia2_agent 相关）
// 不能直接 unmarshal telegraf input 对象。如要测试，请加上 -tags expect_fai
//

package telegraf_inputs

import (
	"fmt"
	"testing"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"
)

func unmarshalInput(tbl *ast.Table, inputName string, input interface{}) error {

	if input == nil {
		return nil
	}

	for field, node := range tbl.Fields {
		switch field {
		case "inputs":
			subTable, ok := node.(*ast.Table)
			if !ok {
				return fmt.Errorf("no sub-table under input")
			}

			for name, v := range subTable.Fields {
				if name != inputName {
					return fmt.Errorf("input name not match expect %s, got %s", inputName, name)
				}

				instances := []*ast.Table{}
				switch x := v.(type) {
				case []*ast.Table:
					instances = x
				case *ast.Table:
					instances = append(instances, x)
				default:
					return fmt.Errorf("invalid ast type")
				}

				for _, x := range instances {
					if err := toml.UnmarshalTable(x, input); err != nil {
						return fmt.Errorf("UnmarshalTable: %s", err)
					}
				}
			}
		default:
			return fmt.Errorf("unknown field %s", field)
		}
	}

	return nil
}

func TestSampleValidity(t *testing.T) {
	for k, v := range samples {
		ti, ok := TelegrafInputs[k]
		if !ok {
			t.Errorf("sample %s not used in TelegrafInputs", k)
			continue
		}

		tbl, err := toml.Parse([]byte(v))
		if err != nil {
			t.Errorf("parse sample %s failed: %v", k, err)
			continue
		}

		if len(tbl.Fields) == 0 {
			t.Errorf("no fields in sample %s", k)
		}

		if err := unmarshalInput(tbl, ti.name, ti.Input); err != nil {
			t.Errorf("on %s: %v", k, err)
		}
	}
}
