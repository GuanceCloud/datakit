// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package inputs

import (
	T "testing"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
)

type testMeasurement struct{}

func (t *testMeasurement) Info() *MeasurementInfo {
	return &MeasurementInfo{
		Name: "test-measurement",
		Desc: "for testing",
		Type: "metric",
		Tags: map[string]any{
			"t1":       &TagInfo{},
			"t2":       &TagInfo{},
			"optional": &TagInfo{},
		},
		Fields: map[string]any{
			"f1": &FieldInfo{DataType: Int},
			"f2": &FieldInfo{DataType: Float},
			"f3": &FieldInfo{DataType: String},
			"f4": &FieldInfo{DataType: Bool},

			"optional": &FieldInfo{DataType: Int},
		},
	}
}

func TestPointChecker(t *T.T) {
	t.Run("base", func(t *T.T) {
		pt := point.NewPointV2(`test-measurement-not-defined`,
			append(point.NewTags(map[string]string{"t1": "some", "t2": "some"}),
				point.NewKVs(map[string]any{`f1`: 123, `f2`: 3.14, `f3`: "hello", `f4`: false})...))

		msg := CheckPoint(pt, WithDoc(&testMeasurement{}), WithOptionalTags("optional"))
		assert.Lenf(t, msg, 3, "got %+#v", msg)

		for _, m := range msg {
			t.Logf("%s", m)
		}
	})

	t.Run("invalid-field-type", func(t *T.T) {
		pt := point.NewPointV2(`test-measurement`,
			append(point.NewTags(map[string]string{"t1": "some", "t2": "some"}),
				point.NewKVs(map[string]any{`f1`: 1.414, `f2`: 3.14, `f3`: "hello", `f4`: false})...))

		msg := CheckPoint(pt, WithDoc(&testMeasurement{}), WithOptionalTags("optional"))
		assert.Lenf(t, msg, 3, "got %+#v", msg)

		for _, m := range msg {
			t.Logf("%s", m)
		}
	})

	t.Run("field-missing", func(t *T.T) {
		pt := point.NewPointV2(`test-measurement`,
			append(point.NewTags(map[string]string{"t1": "some", "t2": "some"}),
				point.NewKVs(map[string]any{`unknown`: 123, `f2`: 3.14, `f3`: "hello", `f4`: false})...))

		msg := CheckPoint(pt, WithDoc(&testMeasurement{}), WithOptionalTags("optional"))
		assert.Lenf(t, msg, 3, "got %+#v", msg)

		for _, m := range msg {
			t.Logf("%s", m)
		}
	})

	t.Run("field-missing-and-field-count-not-equal", func(t *T.T) {
		pt := point.NewPointV2(`test-measurement`,
			append(point.NewTags(map[string]string{"t1": "some", "t2": "some"}),
				point.NewKVs(map[string]any{`f2`: 3.14, `f3`: "hello", `f4`: false})...))

		msg := CheckPoint(pt, WithDoc(&testMeasurement{}), WithOptionalTags("optional"))
		assert.Lenf(t, msg, 3, "got %+#v", msg)

		for _, m := range msg {
			t.Logf("%s", m)
		}
	})

	t.Run("tag-missing", func(t *T.T) {
		pt := point.NewPointV2(`test-measurement`,
			append(point.NewTags(map[string]string{
				"t1": "some",
				"t3": "some", // not expect
			}),
				point.NewKVs(map[string]any{`f1`: 123, `f2`: 3.14, `f3`: "hello", `f4`: false})...))

		msg := CheckPoint(pt, WithDoc(&testMeasurement{}), WithOptionalTags("optional"))
		assert.Lenf(t, msg, 4, "got %+#v", msg)

		for _, m := range msg {
			t.Logf("%s", m)
		}
	})

	t.Run("tag-count-not-equal", func(t *T.T) {
		pt := point.NewPointV2(`test-measurement`,
			append(point.NewTags(map[string]string{
				"t1": "some",
				"t2": "some",
				"t3": "other", // not expect
			}),
				point.NewKVs(map[string]any{
					`f1`: 123,
					`f2`: 3.14,
					`f3`: "hello",
					`f4`: false,
				})...))

		msg := CheckPoint(pt, WithDoc(&testMeasurement{}), WithOptionalTags("optional"))
		assert.Lenf(t, msg, 4, "got %+#v", msg)

		for _, m := range msg {
			t.Logf("%s", m)
		}
	})

	t.Run("tag-count-not-equal-and-tag-missing", func(t *T.T) {
		pt := point.NewPointV2(`test-measurement`,
			append(point.NewTags(map[string]string{"t1": "some"}),
				point.NewKVs(map[string]any{`f1`: 123, `f2`: 3.14, `f3`: "hello", `f4`: false})...))

		msg := CheckPoint(pt, WithDoc(&testMeasurement{}), WithOptionalTags("optional"))
		assert.Lenf(t, msg, 4, "got %+#v", msg)

		for _, m := range msg {
			t.Logf("%s", m)
		}
	})

	t.Run("extra-tags", func(t *T.T) {
		pt := point.NewPointV2(`test-measurement`,
			append(point.NewTags(map[string]string{"t1": "some", "t2": "", "tx": "xt"}),
				point.NewKVs(map[string]any{`f1`: 123, `f2`: 3.14, `f3`: "hello", `f4`: false})...))

		msg := CheckPoint(pt, WithDoc(&testMeasurement{}), WithExtraTags(map[string]string{"tx": "xt"}), WithOptionalTags("optional"))
		assert.Lenf(t, msg, 2, "got %+#v", msg)

		for _, m := range msg {
			t.Logf("%s", m)
		}
	})

	t.Run("optional-fields", func(t *T.T) {
		pt := point.NewPointV2(`test-measurement`,
			append(point.NewTags(map[string]string{"t1": "some", "t2": ""}),
				point.NewKVs(map[string]any{`f1`: 123, `f2`: 3.14, `f3`: "hello", `f4`: false})...))

		msg := CheckPoint(pt, WithDoc(&testMeasurement{}), WithOptionalFields("optional"), WithOptionalTags("optional"))
		assert.Lenf(t, msg, 0, "got %+#v", msg)
	})

	t.Run("with-expect-point", func(t *T.T) {
		pt := point.NewPointV2(`test-measurement`,
			append(point.NewTags(map[string]string{"t1": "some", "t2": ""}),
				point.NewKVs(map[string]any{`f1`: 123, `f2`: 3.14, `f3`: "hello", `f4`: false})...))

		exp := point.NewPointV2(`test-measurement`,
			append(point.NewTags(map[string]string{"t1": "some", "t2": ""}),
				point.NewKVs(map[string]any{`f1`: 123, `f2`: 3.14, `f3`: "hello", `f4`: false})...))

		msg := CheckPoint(pt, WithExpectPoint(exp), WithOptionalFields("optional"), WithOptionalTags("optional"))
		assert.Lenf(t, msg, 0, "got %+#v", msg)

		for _, m := range msg {
			t.Logf("%s", m)
		}
	})

	t.Run("with-expect-point-invalid-tag-field-key", func(t *T.T) {
		pt := point.NewPointV2(`test-measurement`,
			append(point.NewTags(map[string]string{"t1": "some", "t2": ""}),
				point.NewKVs(map[string]any{`f1`: 123, `f2`: 3.14, `f3`: "hello", `f4`: false})...))

		exp := point.NewPointV2(`test-measurement-invalid`, // bad measurement name
			append(point.NewTags(map[string]string{
				"t1":          "some",
				"t2":          "some",
				"invalid-tag": "", // invalid tag: tag not found/tag count not match
			}),
				point.NewKVs(map[string]any{
					`f1`: 123,
					`f2`: 3.14,
					`f3`: "hello",
					// `f4`: false,// missing
				})...))

		msg := CheckPoint(pt, WithExpectPoint(exp), WithOptionalFields("optional"), WithOptionalTags("optional"))
		assert.Lenf(t, msg, 4, "got %+#v", msg)

		for _, m := range msg {
			t.Logf("%s", m)
		}
	})

	t.Run("with-expect-point-invalid-field-value", func(t *T.T) {
		pt := point.NewPointV2(`test-measurement`,
			append(point.NewTags(map[string]string{"t1": "some", "t2": ""}),
				point.NewKVs(map[string]any{
					`f1`: 123,
					`f2`: 3.14,
					`f3`: "hello",
					// `f4`: false, // +2: missing/field-count-not-match
				})...))

		exp := point.NewPointV2(`test-measurement`,
			append(point.NewTags(map[string]string{
				"t1": "some",
				"t2": "some",
			}),
				point.NewKVs(map[string]any{
					`f1`: "123",
					`f2`: "3.14",
					`f3`: 1.414,
					`f4`: "some-bool",
				})...))

		msg := CheckPoint(pt, WithExpectPoint(exp), WithOptionalFields("optional"), WithOptionalTags("optional"))
		assert.Lenf(t, msg, 5, "got %+#v", msg)

		for _, m := range msg {
			t.Logf("%s", m)
		}
	})

	t.Run("with-value-and-type-check-off", func(t *T.T) {
		pt := point.NewPointV2(`test-measurement`,
			append(point.NewTags(map[string]string{"t1": "some", "t2": ""}),
				point.NewKVs(map[string]any{
					`f1`: 123,
					`f2`: 3.14,
					`f3`: "hello",
					`f4`: false,
				})...))

		exp := point.NewPointV2(`test-measurement`,
			append(point.NewTags(map[string]string{
				"t1": "some",
				"t2": "some",
			}),
				point.NewKVs(map[string]any{
					`f1`: "123",
					`f2`: "3.14",
					`f3`: 1.414,
					`f4`: "some-bool",
				})...))

		msg := CheckPoint(pt, WithExpectPoint(exp),
			WithOptionalFields("optional"),
			WithOptionalTags("optional"),
			WithValueChecking(false),
			WithTypeChecking(false))
		assert.Lenf(t, msg, 0, "got %+#v", msg)

		for _, m := range msg {
			t.Logf("%s", m)
		}
	})

	t.Run(`WithMeasurementCheckIgnored`, func(t *T.T) {
		pt := point.NewPointV2(`test-measurement`,
			append(point.NewTags(map[string]string{"t1": "some", "t2": ""}),
				point.NewKVs(map[string]any{
					`f1`: 123,
					`f2`: 3.14,
					`f3`: "hello",
					`f4`: false,
				})...))

		exp := point.NewPointV2(`another-measurement`,
			append(point.NewTags(map[string]string{"t1": "some", "t2": ""}),
				point.NewKVs(map[string]any{
					`f1`: 123,
					`f2`: 3.14,
					`f3`: "hello",
					`f4`: false,
				})...))

		msg := CheckPoint(pt, WithExpectPoint(exp), WithMeasurementCheckIgnored(true))
		assert.Lenf(t, msg, 0, "got %+#v", msg)
	})

	t.Run("optional-fields", func(t *T.T) {
		pt := point.NewPointV2(`test-measurement-not-defined`,
			append(point.NewTags(map[string]string{"t1": "some", "t2": ""}),
				point.NewKVs(map[string]any{`f1`: 123, `f2`: 3.14, `f3`: "hello", `f4`: false})...))

		msg := CheckPoint(pt, WithDoc(&testMeasurement{}),
			WithOptionalFields("optional"),
			WithOptionalTags("optional"),
			WithMeasurementCheckIgnored(true))
		assert.Lenf(t, msg, 0, "got %+#v", msg)
	})
}
