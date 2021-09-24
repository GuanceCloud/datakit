package diskio

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	FieldCompare = 1 << iota
	NameCompare
	TagCompare
	TimeCompare
)

// The incoming parameters expectMeasurement and actualMeasurement need to be sorted in advance
func AssertMeasurement(t *testing.T, expectMeasurement []*diskioMeasurement, actualMeasurement []inputs.Measurement, flag int) {
	// 取长度最短的
	lenE := len(expectMeasurement)
	lenA := len(actualMeasurement)
	if lenE != lenA {
		t.Errorf("The number of objects does not match. Expect:%d  Actual:%d", lenE, lenA)
	}
	count := lenE
	if count > lenA {
		count = lenA
	}

	for i := 0; i < count; i++ {
		expect := expectMeasurement[i]
		actual := actualMeasurement[i].(*diskioMeasurement)
		if (flag & FieldCompare) == FieldCompare {
			for key, valueE := range expect.fields {
				valueA, ok := actual.fields[key]
				if !ok {
					t.Errorf("The expected field does not exist: %s", key)
					continue
				}
				assert.Equal(t, valueE, valueA, "Field: "+key)
			}
		}

		if (flag & NameCompare) == NameCompare {
			if expect.name != actual.name {
				t.Errorf("The expected measurement name is %s, the actual is %s", expect.name, actual.name)
			}
		}

		if (flag & TagCompare) == TagCompare {
			for kE, vE := range expect.tags {
				vA, ok := actual.tags[kE]
				if !ok {
					t.Errorf("The expected field does not exist: %s", kE)
					continue
				}
				assert.Equal(t, vE, vA, "Tag: "+kE)
			}
		}

		if (flag & TimeCompare) == TimeCompare {
			if expect.ts != actual.ts {
				t.Error("The expected time is ", expect.ts.String(), ", the actual is ", actual.ts.String())
			}
		}
	}
}
