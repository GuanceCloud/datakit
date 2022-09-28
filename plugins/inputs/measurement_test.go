// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package inputs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
)

type MockMeasurement struct {
	Name   string
	Tags   map[string]string
	Fields map[string]interface{}
	Opt    *point.PointOption
}

func (m *MockMeasurement) LineProto() (*point.Point, error) {
	return point.NewPoint(m.Name, m.Tags, m.Fields, m.Opt)
}

func (m *MockMeasurement) Info() *MeasurementInfo {
	return nil
}

func TestGetPointsFromMeasurement(t *testing.T) {
	cases := []struct {
		name     string
		m        []Measurement
		expected string
		fail     bool
	}{
		{
			name: "ignore error when make point",
			m: []Measurement{
				&MockMeasurement{
					Name: "test",
					Tags: map[string]string{},
					Fields: map[string]interface{}{ // empty field
						"f1": map[string]string{},
					},
					Opt: &point.PointOption{
						Time:     time.Unix(0, 123),
						Category: datakit.Metric,
					},
				},
				&MockMeasurement{
					Name: "test",
					Tags: map[string]string{},
					Fields: map[string]interface{}{
						"f1": 1,
					},
					Opt: &point.PointOption{
						Time:     time.Unix(0, 123),
						Category: datakit.Metric,
					},
				},
			},
			expected: "test f1=1i 123",
		},
		{
			name: "field value too long",
			m: []Measurement{
				&MockMeasurement{
					Name: "test",
					Tags: map[string]string{"t1": "t1", "t2": "t2"},
					Fields: map[string]interface{}{
						"f1": 2.0,
						"f2": 1.0,
						"f3": "abc", // dropped
					},
					Opt: &point.PointOption{
						Time:             time.Unix(0, 123),
						MaxFieldValueLen: 2,
						Category:         datakit.Metric,
					},
				},
			},
			expected: `test,t1=t1,t2=t2 f1=2,f2=1 123`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			encoder := lineproto.NewLineEncoder()
			points, err := GetPointsFromMeasurement(c.m)
			if c.fail {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			t.Logf("points: %+#v", points)
			for _, pt := range points {
				err := encoder.AppendPoint(pt.Point)
				if err != nil {
					t.Logf("encoder append point err: %s ,ignored", err)
					encoder.Reset()
				}
			}
			lines, _ := encoder.UnsafeStringWithoutLn()
			t.Log(lines)
			assert.Equal(t, c.expected, lines)
		})
	}
}
