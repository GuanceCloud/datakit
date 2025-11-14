// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package metrics

import (
	"testing"
	"time"

	"github.com/grafana/jfr-parser/common/attributes"
	"github.com/grafana/jfr-parser/common/filters"
	"github.com/grafana/jfr-parser/common/types"
	"github.com/grafana/jfr-parser/common/units"
)

func TestParseJFR(t *testing.T) {
	for _, chunk := range chunks {
		chunk.ShowClassMeta(types.VmInfo)
		for _, event := range chunk.Apply(filters.VmInfo) {
			value, err := attributes.JVMStartTime.GetValue(event)
			if err != nil {
				t.Fatal(err)
			}
			t.Log(value)
			tm, err := units.ToTime(value)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("jvm start at: %v, uptime: %v", tm, time.Since(tm))
		}
	}
}
