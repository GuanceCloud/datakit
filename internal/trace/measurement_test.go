package trace

import (
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
)

func TestTraceMeasurement_Point(t *testing.T) {
	ts := time.Now()

	tests := []struct {
		name     string
		measure  *TraceMeasurement
		wantTags map[string]string
		wantFields map[string]interface{}
	}{
		{
			name: "basic point",
			measure: &TraceMeasurement{
				Name: "test_trace",
				Tags: map[string]string{
					"service": "web",
					"host": "localhost",
				},
				Fields: map[string]interface{}{
					"duration": int64(100),
					"status": "ok",
				},
				TS: ts,
			},
			wantTags: map[string]string{
				"service": "web",
				"host": "localhost",
			},
			wantFields: map[string]interface{}{
				"duration": int64(100),
				"status": "ok",
			},
		},
		{
			name: "empty tags and fields",
			measure: &TraceMeasurement{
				Name: "empty_trace",
				TS: ts,
			},
			wantTags: map[string]string{},
			wantFields: map[string]interface{}{},
		},
		{
			name: "with build options",
			measure: &TraceMeasurement{
				Name: "option_trace",
				Tags: map[string]string{"tag1": "val1"},
				Fields: map[string]interface{}{"field1": 1},
				TS: ts,
				BuildPointOptions: []point.Option{point.WithCategory("test")},
			},
			wantTags: map[string]string{"tag1": "val1"},
			wantFields: map[string]interface{}{"field1": 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pt := tt.measure.Point()

			assert.Equal(t, tt.measure.Name, pt.Name())
			assert.Equal(t, tt.measure.TS.UnixNano(), pt.Time())

			tags := pt.Tags()
			fields := pt.Fields()

			assert.Equal(t, tt.wantTags, tags)
			assert.Equal(t, tt.wantFields, fields)
		})
	}
}

func TestTraceMeasurement_Info(t *testing.T) {
	m := &TraceMeasurement{
		Name: "test_trace",
	}

	info := m.Info()

	assert.Equal(t, m.Name, info.Name)
	assert.Equal(t, "tracing", info.Type)

	// Verify tag info exists
	assert.Contains(t, info.Tags, TagHost)
	assert.Contains(t, info.Tags, TagService)
	assert.Contains(t, info.Tags, TagOperation)
	assert.Contains(t, info.Tags, TagSourceType)
	assert.Contains(t, info.Tags, TagSpanStatus)
	assert.Contains(t, info.Tags, TagSpanType)

	// Verify field info exists
	assert.Contains(t, info.Fields, FieldDuration)
	assert.Contains(t, info.Fields, FieldMessage)
	assert.Contains(t, info.Fields, FieldParentID)
	assert.Contains(t, info.Fields, FieldResource)
	assert.Contains(t, info.Fields, FieldSpanid)
	assert.Contains(t, info.Fields, FieldStart)
	assert.Contains(t, info.Fields, FieldTraceID)
}
