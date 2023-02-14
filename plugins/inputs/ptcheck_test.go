// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package inputs

import (
	"fmt"
	T "testing"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
)

type testMeasurement struct{}

func (t *testMeasurement) LineProto() (*dkpt.Point, error) {
	return nil, fmt.Errorf("not implemented")
}

func (t *testMeasurement) Info() *MeasurementInfo {
	return &MeasurementInfo{
		Name: "test-measurement",
		Desc: "for testing",
		Type: "metric",
		Tags: map[string]any{
			"t1": &TagInfo{},
			"t2": &TagInfo{},
		},
		Fields: map[string]any{
			"f1": &FieldInfo{DataType: Int},
			"f2": &FieldInfo{DataType: Float},
			"f3": &FieldInfo{DataType: String},
			"f4": &FieldInfo{DataType: Bool},
		},
	}
}

func TestPointChecker(t *T.T) {
	t.Run("base", func(t *T.T) {
		pt := point.NewPointV2([]byte(`test-measurement`),
			append(point.NewTags(map[string]string{"t1": "some", "t2": "some"}),
				point.NewKVs(map[string]any{`f1`: 123, `f2`: 3.14, `f3`: "hello", `f4`: false})...))

		msg := CheckPoint(pt, &testMeasurement{})
		assert.True(t, len(msg) == 0, "got %+#v", msg)
	})

	t.Run("invalid-field-type", func(t *T.T) {
		pt := point.NewPointV2([]byte(`test-measurement`),
			append(point.NewTags(map[string]string{"t1": "some", "t2": "some"}),
				point.NewKVs(map[string]any{`f1`: 1.414, `f2`: 3.14, `f3`: "hello", `f4`: false})...))

		msg := CheckPoint(pt, &testMeasurement{})
		assert.True(t, len(msg) == 1, "got %+#v", msg)
	})

	t.Run("field-missing", func(t *T.T) {
		pt := point.NewPointV2([]byte(`test-measurement`),
			append(point.NewTags(map[string]string{"t1": "some", "t2": "some"}),
				point.NewKVs(map[string]any{`unknown`: 123, `f2`: 3.14, `f3`: "hello", `f4`: false})...))

		msg := CheckPoint(pt, &testMeasurement{})
		assert.True(t, len(msg) == 1, "got %+#v", msg)
	})

	t.Run("field-missing-and-field-count-not-equal", func(t *T.T) {
		pt := point.NewPointV2([]byte(`test-measurement`),
			append(point.NewTags(map[string]string{"t1": "some", "t2": "some"}),
				point.NewKVs(map[string]any{`f2`: 3.14, `f3`: "hello", `f4`: false})...))

		msg := CheckPoint(pt, &testMeasurement{})
		assert.True(t, len(msg) == 2, "got %+#v", msg)
	})

	t.Run("tag-missing", func(t *T.T) {
		pt := point.NewPointV2([]byte(`test-measurement`),
			append(point.NewTags(map[string]string{"t1": "some", "t3": "some"}),
				point.NewKVs(map[string]any{`f1`: 123, `f2`: 3.14, `f3`: "hello", `f4`: false})...))

		msg := CheckPoint(pt, &testMeasurement{})
		assert.True(t, len(msg) == 1, "got %+#v", msg)
	})

	t.Run("tag-count-not-equal", func(t *T.T) {
		pt := point.NewPointV2([]byte(`test-measurement`),
			append(point.NewTags(map[string]string{"t1": "some", "t2": "some", "t3": "other"}),
				point.NewKVs(map[string]any{`f1`: 123, `f2`: 3.14, `f3`: "hello", `f4`: false})...))

		msg := CheckPoint(pt, &testMeasurement{})
		assert.True(t, len(msg) == 1, "got %+#v", msg)
	})

	t.Run("tag-count-not-equal-and-tag-missing", func(t *T.T) {
		pt := point.NewPointV2([]byte(`test-measurement`),
			append(point.NewTags(map[string]string{"t1": "some"}),
				point.NewKVs(map[string]any{`f1`: 123, `f2`: 3.14, `f3`: "hello", `f4`: false})...))

		msg := CheckPoint(pt, &testMeasurement{})
		assert.True(t, len(msg) == 2, "got %+#v", msg)
	})

	t.Run("extra-tags", func(t *T.T) {
		pt := point.NewPointV2([]byte(`test-measurement`),
			append(point.NewTags(map[string]string{"t1": "some", "t2": "", "tx": ""}),
				point.NewKVs(map[string]any{`f1`: 123, `f2`: 3.14, `f3`: "hello", `f4`: false})...))

		msg := CheckPoint(pt, &testMeasurement{}, WithAllowExtraTags(true))
		assert.True(t, len(msg) == 0, "got %+#v", msg)
	})
}
