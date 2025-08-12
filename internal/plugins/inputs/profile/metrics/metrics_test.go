// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package metrics

import (
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/grafana/jfr-parser/common/attributes"
	"github.com/grafana/jfr-parser/common/filters"
	"github.com/grafana/jfr-parser/common/types"
	"github.com/grafana/jfr-parser/common/units"
	"github.com/stretchr/testify/assert"
)

func TestMetricKVs(t *testing.T) {
	kvs := newMetricKVs()
	kvs.AddTag("foo", "bar")
	kvs.AddTag("hello", "world")

	kvs.Add("duration", 3.1415925)
	kvs.Add("count", 789)
	kvs.Add("bytes", 100000)

	ptKVs := kvs.toPointKVs()

	for _, tag := range ptKVs.InfluxTags() {
		t.Logf("[tag] %s : %s", tag.Key, tag.Value)
	}
	for _, field := range ptKVs.Fields() {
		t.Logf("[field] %s : %v", field.Key, field.Raw())
	}

	assert.Equal(t, 2, ptKVs.TagCount())
	assert.Equal(t, 3, ptKVs.FieldCount())

	mKVs := toMetricKVs(point.NewTags(map[string]string{
		"tag1": "value1",
		"tag2": "value2",
		"tag3": "value3",
	}))
	mKVs.AddTag("language", "java")
	mKVs.Add("foobar", 3.1415)

	assert.Equal(t, 4, mKVs.toPointKVs().TagCount())
	assert.Equal(t, 1, mKVs.toPointKVs().FieldCount())
}

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
