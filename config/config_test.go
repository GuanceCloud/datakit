package config

import (
	"testing"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"
)

func TestLoadMainCfg(t *testing.T) {

	c := newDefaultCfg()
	if err := c.LoadMainConfig(); err != nil {
		t.Errorf("%s", err)
	}
}

func TestTomlParse(t *testing.T) {
	x := []byte(`
[[inputs.abc]]
	key1 = 1
	key2 = "a"
	key3 = 3.14

[[inputs.abc]]
	key1 = 11
	key2 = "aa"
	key3 = 6.28

[[inputs.abc]]
	key1 = 22
	key2 = "aaa"
	key3 = 6.18

[[inputs.def]]
	key1 = 22
	key2 = "aaa"
	key3 = 6.18

	`)

	type obj struct {
		Key1 int     `toml:"key1"`
		Key2 string  `toml:"key2"`
		Key3 float64 `toml:"key3"`
	}

	tbl, err := toml.Parse(x)
	if err != nil {
		t.Fatal(err)
	}

	if tbl.Fields == nil {
		t.Fatal("empty data")
	}

	for f, v := range tbl.Fields {
		switch f {
		case "inputs":
			tbl_ := v.(*ast.Table)
			t.Logf("tbl_: %+#v", tbl_)

			for k, v_ := range tbl_.Fields {
				// t.Logf("%s: %+#v", k, v_)

				tbls := []*ast.Table{}

				switch v_.(type) {
				case []*ast.Table:
					tbls = v_.([]*ast.Table)
				case *ast.Table:
					tbls = append(tbls, v_.(*ast.Table))
				default:
					t.Fatal("bad data")
				}

				t.Logf("elems: %d", len(tbls))

				for idx, elem := range tbls {
					var o obj
					toml.UnmarshalTable(elem, &o)
					t.Logf("[%s] %d: %+#v\n", k, idx, o)
				}
			}

		default:
			t.Fatal("bad data")
		}
	}
}

func TestInitCfg(t *testing.T) {
	// TODO
}
