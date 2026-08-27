// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package winnetflow

import (
	"testing"

	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

// TestMeasurementMetadata audits the registered input's measurement metadata:
// names, category, and complete field type/unit declarations. Doc generation
// and downstream metadata processing depend on this.
func TestMeasurementMetadata(t *testing.T) {
	creator, ok := inputs.AllInputs[inputName]
	if !ok {
		t.Fatalf("input %s not registered", inputName)
	}
	ipt, ok := creator().(inputs.InputV2)
	if !ok {
		t.Fatalf("input %s does not implement InputV2", inputName)
	}

	ms := ipt.SampleMeasurement()
	if len(ms) == 0 {
		t.Fatal("no measurements declared")
	}
	for _, m := range ms {
		info := m.Info()
		if info == nil {
			t.Fatal("nil measurement info")
		}
		if info.Name == "" {
			t.Fatal("empty measurement name")
		}
		if info.Cat != point.Network {
			t.Fatalf("measurement %s category = %v, want network", info.Name, info.Cat)
		}
		if len(info.Fields) == 0 {
			t.Fatalf("measurement %s has no fields", info.Name)
		}
		for name, f := range info.Fields {
			fi, ok := f.(*inputs.FieldInfo)
			if !ok {
				t.Fatalf("field %s is not *inputs.FieldInfo", name)
			}
			if fi.DataType == "" || fi.Unit == "" {
				t.Fatalf("field %s: missing DataType/Unit (dt=%q unit=%q)", name, fi.DataType, fi.Unit)
			}
			if fi.Desc == "" {
				t.Fatalf("field %s: missing description", name)
			}
		}
	}
}
