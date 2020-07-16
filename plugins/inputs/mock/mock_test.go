package mock

import (
	"testing"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"
)

func tryParseCfg(tst *testing.T, val interface{}) {
	switch val.(type) {
	case []*ast.Table:
		arr := val.([]*ast.Table)
		for _, t := range arr {
			var m Mock
			if err := toml.UnmarshalTable(t, &m); err != nil {
				tst.Fatal(err)
			}

			tst.Logf("array: %+#v", m)
		}
	case *ast.Table:
		t := val.(*ast.Table)
		var m Mock
		if err := toml.UnmarshalTable(t, &m); err != nil {
			tst.Fatal(err)
		}

		tst.Logf("elem: %+#v", m)
	}
}

func TestParse(tst *testing.T) {

	data := [][]byte{

		[]byte(`
[[inputs.mock]]
interval = '3s'
metric = 'mock-testing'`),

		[]byte(`
[inputs.mock]
interval = '3s'
metric = 'mock-testing'`),

		[]byte(`
[mock]
interval = '3s'
metric = 'mock-testing'`),

		[]byte(`
[[mock]]
interval = '3s'
metric = 'mock-testing'`),
	}

	for _, x := range data {
		tbl, err := toml.Parse(x)
		if err != nil {
			tst.Fatal(err)
		}

		for f, val := range tbl.Fields {

			tst.Logf("field: %s", f)
			if f == "inputs" {
				val_ := val.(*ast.Table)
				for k, v := range val_.Fields {
					tst.Logf("field: %s", k)
					tryParseCfg(tst, v)
				}
			} else {
				tryParseCfg(tst, val)
			}
		}
	}
}
