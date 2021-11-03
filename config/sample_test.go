package config

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/influxdata/toml/ast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

// read dir samples, check if sample is marshable by current release
func TestConfSample(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)

	// jump to ../samples
	samplePath := filepath.Join(filepath.Dir(filepath.Dir(filename)), "samples")

	samples := SearchDir(samplePath, "")
	for _, s := range samples {

		t.Logf("testing %s", s)

		asttbl, err := ParseCfgFile(s)
		if err != nil {
			t.Fatalf("ParseCfgFile: %s", err.Error())
			continue
		}

		switch filepath.Base(s) {
		case "datakit.conf":
			mc := DefaultConfig()
			if err := mc.LoadMainTOML(s); err != nil {
				t.Fatalf("unmarshal main cfg failed for %s: %s", s, err.Error())
			}

		default:

			for field, node := range asttbl.Fields {
				switch field {
				case "inputs": //nolint:goconst
					stbl, ok := node.(*ast.Table)
					if !ok {
						t.Fatalf("found invalid input from %s: expect ast.Table", s)
					} else {
						for inputName, v := range stbl.Fields {
							if creator, ok := inputs.Inputs[inputName]; !ok {
								t.Logf("ignore input %s from %s", s, inputName)
								continue
							} else {
								if _, err := TryUnmarshal(v, inputName, creator); err != nil {
									t.Fatalf("unmarshal input %s failed within %s: %s", inputName, s, err.Error())
									continue
								}

								t.Logf("unmarshal input %s from %s ok", inputName, s)
							}
						}
					}

				default: // compatible with old version: no [[inputs.xxx]] header
					l.Debugf("ignore field %s in file %s", field, s)
				}
			}
		}
	}
}
