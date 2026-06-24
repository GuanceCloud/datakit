// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package demo

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func TestDemoMeasurementMetadata(t *testing.T) {
	fields := (&demoMetric{}).Info().Fields

	required := map[string]struct {
		dataType string
		mType    string
		unit     string
	}{
		"some_string": {inputs.String, inputs.String, inputs.NoUnit},
		"ok":          {inputs.Bool, inputs.Gauge, inputs.Bool},
	}

	for name, want := range required {
		field, ok := fields[name].(*inputs.FieldInfo)
		if !ok {
			t.Fatalf("%s is not an *inputs.FieldInfo", name)
		}
		if field.DataType != want.dataType || field.Type != want.mType || field.Unit != want.unit {
			t.Fatalf("%s metadata = data_type %q type %q unit %q, want data_type %q type %q unit %q",
				name, field.DataType, field.Type, field.Unit, want.dataType, want.mType, want.unit)
		}
		if field.Desc == "" {
			t.Fatalf("%s has an empty description", name)
		}
	}
}
