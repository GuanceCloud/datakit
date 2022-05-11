package promremote

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAddTags(t *testing.T) {
	ipt := NewInput()
	ipt.Tags["c"] = "d"
	m := Measurement{
		name:   "test_add_tags",
		tags:   map[string]string{"a": "b"},
		fields: map[string]interface{}{},
		ts:     time.Time{},
	}
	ipt.addTags(&m)
	assert.Equal(t, map[string]string{"a": "b", "c": "d"}, m.tags)
}

func TestIgnoreTags(t *testing.T) {
	ipt := NewInput()
	ipt.TagsIgnore = append(ipt.TagsIgnore, "c")
	m := Measurement{
		name:   "test_add_tags",
		tags:   map[string]string{"a": "b", "c": "d"},
		fields: map[string]interface{}{},
		ts:     time.Time{},
	}
	ipt.ignoreTags(&m)
	assert.Equal(t, map[string]string{"a": "b"}, m.tags)
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
				name:   "mock_measurement",
				tags:   map[string]string{"a": "b", "c": "d"},
				fields: map[string]interface{}{},
				ts:     time.Now(),
			},
			tagsRename:   map[string]string{"a": "e"},
			expectedTags: map[string]string{"e": "b", "c": "d"},
		},
		{
			name: "don't overwrite",
			measurement: Measurement{
				name:   "mock_measurement",
				tags:   map[string]string{"a": "b", "c": "d"},
				fields: map[string]interface{}{},
				ts:     time.Now(),
			},
			overwrite:    false,
			tagsRename:   map[string]string{"a": "c"},
			expectedTags: map[string]string{"a": "b", "c": "d"},
		},
		{
			name: "overwrite",
			measurement: Measurement{
				name:   "mock_measurement",
				tags:   map[string]string{"a": "b", "c": "d"},
				fields: map[string]interface{}{},
				ts:     time.Now(),
			},
			overwrite:    true,
			tagsRename:   map[string]string{"a": "c"},
			expectedTags: map[string]string{"c": "b"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ipt := NewInput()
			ipt.Overwrite = tc.overwrite
			ipt.TagsRename = tc.tagsRename
			ipt.renameTags(&tc.measurement)
			assert.Equal(t, tc.expectedTags, tc.measurement.tags)
		})
	}
}
