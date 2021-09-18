package solr

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

// The incoming parameters expectMeasurement and actualMeasurement need to be sorted in advance
func AssertMeasurement(t *testing.T, expectMeasurement []inputs.Measurement, actualMeasurement []inputs.Measurement, flag int) {
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
		expect, err := expectMeasurement[i].LineProto()
		if err != nil {
			t.Error(err)
			return
		}
		actual, err := actualMeasurement[i].LineProto()
		if err != nil {
			t.Error(err)
			return
		}
		// field
		if (flag & FieldCompare) == FieldCompare {
			expectFields, err := expect.Fields()
			if err != nil {
				t.Error(err)
				return
			}
			actualFields, err := actual.Fields()
			if err != nil {
				t.Error(err)
				return
			}
			for key, valueE := range expectFields {
				valueA, ok := actualFields[key]
				if !ok {
					t.Errorf("The expected field does not exist: %s", key)
					continue
				}
				assert.Equal(t, valueE, valueA, "Field: "+key)
			}
		}

		// name
		if (flag & NameCompare) == NameCompare {
			if expect.Name() != actual.Name() {
				t.Errorf("The expected measurement name is %s, the actual is %s", expect.Name(), actual.Name())
			}
		}

		// tag
		if (flag & TagCompare) == TagCompare {
			expectTags := expect.Tags()
			actualTags := actual.Tags()
			for kE, vE := range expectTags {
				vA, ok := actualTags[kE]
				if !ok {
					t.Errorf("The expected tag does not exist: %s", kE)
					continue
				}
				assert.Equal(t, vE, vA, "Tag: "+kE)
			}
		}

		// time
		if (flag & TimeCompare) == TimeCompare {
			if expect.Time() != actual.Time() {
				t.Error("The expected time is ", expect.Time().String(), ", the actual is ", actual.Time().String())
			}
		}
	}
}
