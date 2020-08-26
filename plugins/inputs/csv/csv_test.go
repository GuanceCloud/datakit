package csv

import (
	"testing"

	"github.com/influxdata/toml"
	yaml "gopkg.in/yaml.v2"
)

func TestCSVConf(t *testing.T) {
	x := &CSV{
		StartRows: 10,
		Metric:    "test_1",
		Columns: []*column{
			&column{Index: 0, Name: "t1", NaAction: "ignore", AsTag: true},
			&column{Index: 1, Name: "f1", NaAction: "ignore", AsField: true, Type: "int"},
			&column{Index: 2, Name: "f2", NaAction: "ignore", AsField: true, Type: "float"},
			&column{Index: 3, Name: "f3", NaAction: "ignore", AsField: true, Type: "str"},
			&column{Index: 4, NaAction: "ignore", AsTime: true, TimeFormat: "15/08/27 10:20:06", TimePrecision: "s"},
		},
	}

	b, err := toml.Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%s", string(b))

	b, err = yaml.Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%s", string(b))
}
