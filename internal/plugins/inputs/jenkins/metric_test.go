// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jenkins

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func TestJenkinsMeasurementMetadata(t *testing.T) {
	pipelineFields := (&jenkinsPipelineMeasurement{}).Info().Fields
	jobFields := (&jenkinsJobMeasurement{}).Info().Fields
	metricFields := (&metricMeasurement{}).Info().Fields

	check := func(fields map[string]interface{}, name, dataType, unit string) {
		t.Helper()
		field, ok := fields[name].(*inputs.FieldInfo)
		if !ok {
			t.Fatalf("%s is not an *inputs.FieldInfo", name)
		}
		if field.DataType != dataType || field.Unit != unit {
			t.Fatalf("%s metadata = data_type %q unit %q, want data_type %q unit %q",
				name, field.DataType, field.Unit, dataType, unit)
		}
		if field.Desc == "" {
			t.Fatalf("%s has an empty description", name)
		}
	}

	check(pipelineFields, "pipeline_id", inputs.String, inputs.NoUnit)
	check(pipelineFields, "commit_message", inputs.String, inputs.NoUnit)
	check(pipelineFields, "message", inputs.String, inputs.NoUnit)

	check(jobFields, "build_id", inputs.String, inputs.NoUnit)
	check(jobFields, "pipeline_id", inputs.String, inputs.NoUnit)
	check(jobFields, "runner_id", inputs.String, inputs.NoUnit)
	check(jobFields, "build_commit_message", inputs.String, inputs.NoUnit)
	check(jobFields, "message", inputs.String, inputs.NoUnit)

	check(metricFields, "vm_memory_total_used", inputs.Float, inputs.SizeByte)
	check(metricFields, "vm_memory_total_committed", inputs.Float, inputs.SizeByte)
}
