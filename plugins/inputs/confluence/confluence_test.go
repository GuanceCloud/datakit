package confluence

import (
	"testing"

	"github.com/influxdata/toml"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/prom"
)

func TestMain(t *testing.T) {

	var p prom.Prom

	err := toml.Unmarshal([]byte(sampleCfg), &p)
	if err != nil {
		t.Fatal(err)
	}

	prom.TestAssert = true
	p.Run()
}
