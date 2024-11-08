// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmputil

import (
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/stretchr/testify/assert"
)

type MyStringArray struct {
	SomeIds StringArray `yaml:"my_field"`
}
type MyNumber struct {
	SomeNum Number `yaml:"my_field"`
}

type MyBoolean struct {
	SomeBool Boolean `yaml:"my_field"`
}

func TestStringArray_UnmarshalYAML_array(t *testing.T) {
	myStruct := MyStringArray{}
	expected := MyStringArray{SomeIds: StringArray{"aaa", "bbb"}}

	yaml.Unmarshal([]byte(`
my_field:
 - aaa
 - bbb
`), &myStruct)

	assert.Equal(t, expected, myStruct)
}

func TestStringArray_UnmarshalYAML_string(t *testing.T) {
	myStruct := MyStringArray{}
	expected := MyStringArray{SomeIds: StringArray{"aaa"}}

	yaml.Unmarshal([]byte(`
my_field: aaa
`), &myStruct)

	assert.Equal(t, expected, myStruct)
}

func Test_metricTagConfig_UnmarshalYAML(t *testing.T) {
	myStruct := MetricsConfig{}
	expected := MetricsConfig{MetricTags: []MetricTagConfig{{Index: 3}}}

	yaml.Unmarshal([]byte(`
metric_tags:
- index: 3
`), &myStruct)

	assert.Equal(t, expected, myStruct)
}

func Test_metricTagConfig_onlyTags(t *testing.T) {
	myStruct := MetricsConfig{}
	expected := MetricsConfig{MetricTags: []MetricTagConfig{{symbolTag: "aaa"}}}

	yaml.Unmarshal([]byte(`
metric_tags:
- aaa
`), &myStruct)

	assert.Equal(t, expected, myStruct)
}

func Test_Number_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name   string
		data   []byte
		result MyNumber
	}{
		{
			name: "integer number",
			data: []byte(`
my_field: 99
`),
			result: MyNumber{SomeNum: 99},
		},
		{
			name: "string number",
			data: []byte(`
my_field: "88"
`),
			result: MyNumber{SomeNum: 88},
		},
		{
			name: "empty string",
			data: []byte(`
my_field: ""
`),
			result: MyNumber{SomeNum: 0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			myStruct := MyNumber{}
			yaml.Unmarshal(tt.data, &myStruct)
			assert.Equal(t, tt.result, myStruct)
		})
	}
}

func Test_Boolean_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name   string
		data   []byte
		result MyBoolean
	}{
		{
			name: "boolean true",
			data: []byte(`
my_field: true
`),
			result: MyBoolean{SomeBool: true},
		},
		{
			name: "string boolean true",
			data: []byte(`
my_field: "true"
`),
			result: MyBoolean{SomeBool: true},
		},
		{
			name: "boolean false",
			data: []byte(`
my_field: false
`),
			result: MyBoolean{SomeBool: false},
		},
		{
			name: "string boolean false",
			data: []byte(`
my_field: "false"
`),
			result: MyBoolean{SomeBool: false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			myStruct := MyBoolean{}
			yaml.Unmarshal(tt.data, &myStruct)
			assert.Equal(t, tt.result, myStruct)
		})
	}
}

func Test_Boolean_UnmarshalYAML_invalid(t *testing.T) {
	myStruct := MyBoolean{}
	data := []byte(`
my_field: "foo"
`)
	err := yaml.Unmarshal(data, &myStruct)
	assert.EqualError(t, err, "cannot convert `foo` to boolean")
}
