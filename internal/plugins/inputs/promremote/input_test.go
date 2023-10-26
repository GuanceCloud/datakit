// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package promremote

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAddTags(t *testing.T) {
	ipt := defaultInput()
	ipt.Tags["c"] = "d"
	m := Measurement{
		Name:   "test_add_tags",
		Tags:   map[string]string{"a": "b"},
		Fields: map[string]interface{}{},
		TS:     time.Time{},
	}
	pt := m.Point()
	ipt.addTags(pt)
	assert.Equal(t, map[string]string{"a": "b", "c": "d"}, pt.MapTags())
}

func TestIgnoreTags(t *testing.T) {
	ipt := defaultInput()
	ipt.TagsIgnore = append(ipt.TagsIgnore, "c")
	m := Measurement{
		Name:   "test_add_tags",
		Tags:   map[string]string{"a": "b", "c": "d"},
		Fields: map[string]interface{}{},
		TS:     time.Time{},
	}
	pt := m.Point()
	ipt.ignoreTags(pt)
	assert.Equal(t, map[string]string{"a": "b"}, pt.MapTags())
}

func TestRenameTags(t *testing.T) {
	testCases := []struct {
		name         string
		measurement  Measurement
		overwrite    bool
		tagsRename   map[string]string
		expectedTags map[string]string
	}{
		{
			name: "no conflict",
			measurement: Measurement{
				Name:   "mock_measurement",
				Tags:   map[string]string{"a": "b", "c": "d"},
				Fields: map[string]interface{}{},
				TS:     time.Now(),
			},
			tagsRename:   map[string]string{"a": "e"},
			expectedTags: map[string]string{"e": "b", "c": "d"},
		},
		{
			name: "don't overwrite",
			measurement: Measurement{
				Name:   "mock_measurement",
				Tags:   map[string]string{"a": "b", "c": "d"},
				Fields: map[string]interface{}{},
				TS:     time.Now(),
			},
			overwrite:    false,
			tagsRename:   map[string]string{"a": "c"},
			expectedTags: map[string]string{"a": "b", "c": "d"},
		},
		{
			name: "overwrite",
			measurement: Measurement{
				Name:   "mock_measurement",
				Tags:   map[string]string{"a": "b", "c": "d"},
				Fields: map[string]interface{}{},
				TS:     time.Now(),
			},
			overwrite:    true,
			tagsRename:   map[string]string{"a": "c"},
			expectedTags: map[string]string{"c": "b"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ipt := defaultInput()
			ipt.Overwrite = tc.overwrite
			ipt.TagsRename = tc.tagsRename

			pt := tc.measurement.Point()

			ipt.renameTags(pt)
			assert.Equal(t, tc.expectedTags, pt.MapTags())
		})
	}
}
