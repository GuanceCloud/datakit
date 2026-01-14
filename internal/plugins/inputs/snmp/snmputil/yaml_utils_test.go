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

// TestSymbolConfigCompat_UnmarshalYAML tests the SymbolConfigCompat YAML unmarshaling.
func TestSymbolConfigCompat_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name     string
		yamlData string
		expected SymbolConfigCompat
	}{
		{
			name:     "string format",
			yamlData: "sysName",
			expected: SymbolConfigCompat(SymbolConfig{
				Name: "sysName",
			}),
		},
		{
			name: "object format",
			yamlData: `
OID: 1.3.6.1.2.1.1.5.0
name: sysName
`,
			expected: SymbolConfigCompat(SymbolConfig{
				OID:  "1.3.6.1.2.1.1.5.0",
				Name: "sysName",
			}),
		},
		{
			name: "object format with extract_value",
			yamlData: `
OID: 1.3.6.1.2.1.31.1.1.1.1
name: ifName
extract_value: "(Row\\d)"
`,
			expected: SymbolConfigCompat(SymbolConfig{
				OID:          "1.3.6.1.2.1.31.1.1.1.1",
				Name:         "ifName",
				ExtractValue: "(Row\\d)",
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result SymbolConfigCompat
			err := yaml.Unmarshal([]byte(tt.yamlData), &result)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected.Name, result.Name)
			if tt.expected.OID != "" {
				assert.Equal(t, tt.expected.OID, result.OID)
			}
			if tt.expected.ExtractValue != "" {
				assert.Equal(t, tt.expected.ExtractValue, result.ExtractValue)
			}
		})
	}
}

// TestMetricTagConfig_Symbol tests MetricTagConfig with Symbol field (new format).
func TestMetricTagConfig_Symbol(t *testing.T) {
	tests := []struct {
		name     string
		yamlData string
		check    func(*testing.T, MetricTagConfig)
	}{
		{
			name: "symbol as string",
			yamlData: `
- tag: snmp_host
  symbol: sysName
`,
			check: func(t *testing.T, mtc MetricTagConfig) {
				t.Helper()
				assert.Equal(t, "snmp_host", mtc.Tag)
				assert.Equal(t, "sysName", mtc.Symbol.Name)
				assert.Equal(t, "", mtc.Symbol.OID)
			},
		},
		{
			name: "symbol as object",
			yamlData: `
- tag: snmp_host
  symbol:
   OID: 1.3.6.1.2.1.1.5.0
   name: sysName
`,
			check: func(t *testing.T, mtc MetricTagConfig) {
				t.Helper()
				assert.Equal(t, "snmp_host", mtc.Tag)
				assert.Equal(t, "sysName", mtc.Symbol.Name)
				assert.Equal(t, "1.3.6.1.2.1.1.5.0", mtc.Symbol.OID)
			},
		},
		{
			name: "symbol with mapping",
			yamlData: `
- tag: if_admin_status
  symbol:
    OID: 1.3.6.1.2.1.2.2.1.7
    name: ifAdminStatus
  mapping:
    "1": "up"
    "2": "down"
`,
			check: func(t *testing.T, mtc MetricTagConfig) {
				t.Helper()
				assert.Equal(t, "if_admin_status", mtc.Tag)
				assert.Equal(t, "ifAdminStatus", mtc.Symbol.Name)
				assert.Equal(t, "1.3.6.1.2.1.2.2.1.7", mtc.Symbol.OID)
				assert.Equal(t, "up", mtc.Mapping["1"])
				assert.Equal(t, "down", mtc.Mapping["2"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config struct {
				MetricTags []MetricTagConfig `yaml:"metric_tags"`
			}
			err := yaml.Unmarshal([]byte("metric_tags:\n"+tt.yamlData), &config)
			assert.NoError(t, err)
			assert.Len(t, config.MetricTags, 1)
			tt.check(t, config.MetricTags[0])
		})
	}
}
