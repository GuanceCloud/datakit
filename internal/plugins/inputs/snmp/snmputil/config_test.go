// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmputil

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================

// config

// go test -v -timeout 30s -run ^Test_getProfileForSysObjectID$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp/snmputil
func Test_getProfileForSysObjectID(t *testing.T) {
	mockProfiles := ProfileDefinitionMap{
		"profile1": ProfileDefinition{
			Metrics: []MetricsConfig{
				{Symbol: SymbolConfig{OID: "1.2.3.4.5", Name: "someMetric"}},
			},
			SysObjectIds: StringArray{"1.3.6.1.4.1.3375.2.1.3.4.*"},
		},
		"profile2": ProfileDefinition{
			Metrics: []MetricsConfig{
				{Symbol: SymbolConfig{OID: "1.2.3.4.5", Name: "someMetric"}},
			},
			SysObjectIds: StringArray{"1.3.6.1.4.1.3375.2.1.3.4.10"},
		},
		"profile3": ProfileDefinition{
			Metrics: []MetricsConfig{
				{Symbol: SymbolConfig{OID: "1.2.3.4.5", Name: "someMetric"}},
			},
			SysObjectIds: StringArray{"1.3.6.1.4.1.3375.2.1.3.4.5.*"},
		},
	}
	mockProfilesWithPatternError := ProfileDefinitionMap{
		"profile1": ProfileDefinition{
			Metrics: []MetricsConfig{
				{Symbol: SymbolConfig{OID: "1.2.3.4.5", Name: "someMetric"}},
			},
			SysObjectIds: StringArray{"1.3.6.1.4.1.3375.2.1.3.***.*"},
		},
	}
	mockProfilesWithInvalidPatternError := ProfileDefinitionMap{
		"profile1": ProfileDefinition{
			Metrics: []MetricsConfig{
				{Symbol: SymbolConfig{OID: "1.2.3.4.5", Name: "someMetric"}},
			},
			SysObjectIds: StringArray{"1.3.6.1.4.1.3375.2.1.3.[.*"},
		},
	}
	mockProfilesWithDuplicateSysobjectid := ProfileDefinitionMap{
		"profile1": ProfileDefinition{
			Metrics: []MetricsConfig{
				{Symbol: SymbolConfig{OID: "1.2.3.4.5", Name: "someMetric"}},
			},
			SysObjectIds: StringArray{"1.3.6.1.4.1.3375.2.1.3"},
		},
		"profile2": ProfileDefinition{
			Metrics: []MetricsConfig{
				{Symbol: SymbolConfig{OID: "1.2.3.4.5", Name: "someMetric"}},
			},
			SysObjectIds: StringArray{"1.3.6.1.4.1.3375.2.1.3"},
		},
		"profile3": ProfileDefinition{
			Metrics: []MetricsConfig{
				{Symbol: SymbolConfig{OID: "1.2.3.4.5", Name: "someMetric"}},
			},
			SysObjectIds: StringArray{"1.3.6.1.4.1.3375.2.1.4"},
		},
	}
	tests := []struct {
		name            string
		profiles        ProfileDefinitionMap
		sysObjectID     string
		expectedProfile string
		expectedError   string
	}{
		{
			name:            "found matching profile",
			profiles:        mockProfiles,
			sysObjectID:     "1.3.6.1.4.1.3375.2.1.3.4.1",
			expectedProfile: "profile1",
			expectedError:   "",
		},
		{
			name:            "found more precise matching profile",
			profiles:        mockProfiles,
			sysObjectID:     "1.3.6.1.4.1.3375.2.1.3.4.10",
			expectedProfile: "profile2",
			expectedError:   "",
		},
		{
			name:            "found even more precise matching profile",
			profiles:        mockProfiles,
			sysObjectID:     "1.3.6.1.4.1.3375.2.1.3.4.5.11",
			expectedProfile: "profile3",
			expectedError:   "",
		},
		{
			name:            "failed to get most specific profile for sysObjectID",
			profiles:        mockProfilesWithPatternError,
			sysObjectID:     "1.3.6.1.4.1.3375.2.1.3.4.5.11",
			expectedProfile: "",
			expectedError:   "failed to get most specific profile for sysObjectID `1.3.6.1.4.1.3375.2.1.3.4.5.11`, for matched oids [1.3.6.1.4.1.3375.2.1.3.***.*]: error parsing part `***` for pattern `1.3.6.1.4.1.3375.2.1.3.***.*`: strconv.Atoi: parsing \"***\": invalid syntax",
		},
		{
			name:            "invalid pattern", // profiles with invalid patterns are skipped, leading to: cannot get most specific oid from empty list of oids
			profiles:        mockProfilesWithInvalidPatternError,
			sysObjectID:     "1.3.6.1.4.1.3375.2.1.3.4.5.11",
			expectedProfile: "",
			expectedError:   "failed to get most specific profile for sysObjectID `1.3.6.1.4.1.3375.2.1.3.4.5.11`, for matched oids []: cannot get most specific oid from empty list of oids",
		},
		{
			name:            "duplicate sysobjectid",
			profiles:        mockProfilesWithDuplicateSysobjectid,
			sysObjectID:     "1.3.6.1.4.1.3375.2.1.3",
			expectedProfile: "",
			expectedError:   "has the same sysObjectID (1.3.6.1.4.1.3375.2.1.3) as",
		},
		{
			name:            "unrelated duplicate sysobjectid should not raise error",
			profiles:        mockProfilesWithDuplicateSysobjectid,
			sysObjectID:     "1.3.6.1.4.1.3375.2.1.4",
			expectedProfile: "profile3",
			expectedError:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile, err := GetProfileForSysObjectID(tt.profiles, tt.sysObjectID)
			if tt.expectedError == "" {
				assert.Nil(t, err)
			} else {
				assert.Contains(t, err.Error(), tt.expectedError)
			}
			assert.Equal(t, tt.expectedProfile, profile)
		})
	}
}

// =============================================================================

// config_metric

// go test -v -timeout 30s -run ^Test_refreshWithProfile$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp/snmputil
func Test_normalizeRegexReplaceValue(t *testing.T) {
	tests := []struct {
		val                   string
		expectedReplacedValue string
	}{
		{
			"abc",
			"abc",
		},
		{
			"a\\1b",
			"a$1b",
		},
		{
			"a$1b",
			"a$1b",
		},
		{
			"\\1",
			"$1",
		},
		{
			"\\2",
			"$2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.val, func(t *testing.T) {
			assert.Equal(t, tt.expectedReplacedValue, normalizeRegexReplaceValue(tt.val))
		})
	}
}

// =============================================================================

// config_oid

// go test -v -timeout 30s -run ^Test_oidConfig_addScalarOids$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp/snmputil
func Test_oidConfig_addScalarOids(t *testing.T) {
	conf := OidConfig{}

	assert.ElementsMatch(t, []string{}, conf.ScalarOids)

	conf.AddScalarOids([]string{"1.1"})
	conf.AddScalarOids([]string{"1.1"})
	conf.AddScalarOids([]string{"1.2"})
	conf.AddScalarOids([]string{"1.3"})
	conf.AddScalarOids([]string{"1.0"})
	conf.AddScalarOids([]string{""})
	assert.ElementsMatch(t, []string{"1.1", "1.2", "1.3", "1.0"}, conf.ScalarOids)
}

// go test -v -timeout 30s -run ^Test_oidConfig_addColumnOids$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp/snmputil
func Test_oidConfig_addColumnOids(t *testing.T) {
	conf := OidConfig{}

	assert.ElementsMatch(t, []string{}, conf.ColumnOids)

	conf.AddColumnOids([]string{"1.1"})
	conf.AddColumnOids([]string{"1.1"})
	conf.AddColumnOids([]string{"1.2"})
	conf.AddColumnOids([]string{"1.3"})
	conf.AddColumnOids([]string{"1.0"})
	conf.AddColumnOids([]string{""})
	assert.ElementsMatch(t, []string{"1.1", "1.2", "1.3", "1.0"}, conf.ColumnOids)
}

// =============================================================================

// config_validate_enrich

// go test -v -timeout 30s -run ^Test_ValidateEnrichMetrics$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp/snmputil
func Test_ValidateEnrichMetrics(t *testing.T) {
	tests := []struct {
		name            string
		metrics         []MetricsConfig
		expectedErrors  []string
		expectedMetrics []MetricsConfig
	}{
		{
			name: "either table symbol or scalar symbol must be provided",
			metrics: []MetricsConfig{
				{},
			},
			expectedErrors: []string{
				"either a table symbol or a scalar symbol must be provided",
			},
			expectedMetrics: []MetricsConfig{
				{},
			},
		},
		{
			name: "table column symbols and scalar symbol cannot be both provided",
			metrics: []MetricsConfig{
				{
					Symbol: SymbolConfig{
						OID:  "1.2",
						Name: "abc",
					},
					Symbols: []SymbolConfig{
						{
							OID:  "1.2",
							Name: "abc",
						},
					},
					MetricTags: MetricTagConfigList{
						MetricTagConfig{},
					},
				},
			},
			expectedErrors: []string{
				"table symbol and scalar symbol cannot be both provided",
			},
		},
		{
			name: "multiple errors",
			metrics: []MetricsConfig{
				{},
				{
					Symbol: SymbolConfig{
						OID:  "1.2",
						Name: "abc",
					},
					Symbols: []SymbolConfig{
						{
							OID:  "1.2",
							Name: "abc",
						},
					},
					MetricTags: MetricTagConfigList{
						MetricTagConfig{},
					},
				},
			},
			expectedErrors: []string{
				"either a table symbol or a scalar symbol must be provided",
				"table symbol and scalar symbol cannot be both provided",
			},
		},
		{
			name: "missing symbol name",
			metrics: []MetricsConfig{
				{
					Symbol: SymbolConfig{
						OID: "1.2.3",
					},
				},
			},
			expectedErrors: []string{
				"either a table symbol or a scalar symbol must be provided",
			},
		},
		{
			name: "table column symbol name missing",
			metrics: []MetricsConfig{
				{
					Symbols: []SymbolConfig{
						{
							OID: "1.2",
						},
						{
							Name: "abc",
						},
					},
					MetricTags: MetricTagConfigList{
						MetricTagConfig{},
					},
				},
			},
			expectedErrors: []string{
				"symbol name missing: name=`` oid=`1.2`",
				"symbol oid missing: name=`abc` oid=``",
			},
		},
		{
			name: "table external metric column tag symbol error",
			metrics: []MetricsConfig{
				{
					Symbols: []SymbolConfig{
						{
							OID:  "1.2",
							Name: "abc",
						},
					},
					MetricTags: MetricTagConfigList{
						MetricTagConfig{
							Column: SymbolConfig{
								OID: "1.2.3",
							},
						},
						MetricTagConfig{
							Column: SymbolConfig{
								Name: "abc",
							},
						},
					},
				},
			},
			expectedErrors: []string{
				"symbol name missing: name=`` oid=`1.2.3`",
				"symbol oid missing: name=`abc` oid=``",
			},
		},
		{
			name: "missing MetricTags",
			metrics: []MetricsConfig{
				{
					Symbols: []SymbolConfig{
						{
							OID:  "1.2",
							Name: "abc",
						},
					},
					MetricTags: MetricTagConfigList{},
				},
			},
			expectedErrors: []string{
				"column symbols [{1.2 abc  <nil>   <nil> 0 }] doesn't have a 'metric_tags' section",
			},
		},
		{
			name: "table external metric column tag MIB error",
			metrics: []MetricsConfig{
				{
					Symbols: []SymbolConfig{
						{
							OID:  "1.2",
							Name: "abc",
						},
					},
					MetricTags: MetricTagConfigList{
						MetricTagConfig{
							Column: SymbolConfig{
								OID: "1.2.3",
							},
						},
						MetricTagConfig{
							Column: SymbolConfig{
								Name: "abc",
							},
						},
					},
				},
			},
			expectedErrors: []string{
				"symbol name missing: name=`` oid=`1.2.3`",
				"symbol oid missing: name=`abc` oid=``",
			},
		},
		{
			name: "missing match tags",
			metrics: []MetricsConfig{
				{
					Symbols: []SymbolConfig{
						{
							OID:  "1.2",
							Name: "abc",
						},
					},
					MetricTags: MetricTagConfigList{
						MetricTagConfig{
							Column: SymbolConfig{
								OID:  "1.2.3",
								Name: "abc",
							},
							Match: "([a-z])",
						},
					},
				},
			},
			expectedErrors: []string{
				"`tags` mapping must be provided if `match` (`([a-z])`) is defined",
			},
		},
		{
			name: "match cannot compile regex",
			metrics: []MetricsConfig{
				{
					Symbols: []SymbolConfig{
						{
							OID:  "1.2",
							Name: "abc",
						},
					},
					MetricTags: MetricTagConfigList{
						MetricTagConfig{
							Column: SymbolConfig{
								OID:  "1.2.3",
								Name: "abc",
							},
							Match: "([a-z)",
							Tags: map[string]string{
								"foo": "bar",
							},
						},
					},
				},
			},
			expectedErrors: []string{
				"cannot compile `match` (`([a-z)`)",
			},
		},
		{
			name: "match cannot compile regex",
			metrics: []MetricsConfig{
				{
					Symbols: []SymbolConfig{
						{
							OID:  "1.2",
							Name: "abc",
						},
					},
					MetricTags: MetricTagConfigList{
						MetricTagConfig{
							Column: SymbolConfig{
								OID:  "1.2.3",
								Name: "abc",
							},
							Tag: "hello",
							IndexTransform: []MetricIndexTransform{
								{
									Start: 2,
									End:   1,
								},
							},
						},
					},
				},
			},
			expectedErrors: []string{
				"transform rule end should be greater than start. Invalid rule",
			},
		},
		{
			name: "compiling extract_value",
			metrics: []MetricsConfig{
				{
					Symbol: SymbolConfig{
						OID:          "1.2.3",
						Name:         "myMetric",
						ExtractValue: `(\d+)C`,
					},
				},
				{
					Symbols: []SymbolConfig{
						{
							OID:          "1.2",
							Name:         "hey",
							ExtractValue: `(\d+)C`,
						},
					},
					MetricTags: MetricTagConfigList{
						MetricTagConfig{
							Column: SymbolConfig{
								OID:          "1.2.3",
								Name:         "abc",
								ExtractValue: `(\d+)C`,
							},
							Tag: "hello",
						},
					},
				},
			},
			expectedMetrics: []MetricsConfig{
				{
					Symbol: SymbolConfig{
						OID:                  "1.2.3",
						Name:                 "myMetric",
						ExtractValue:         `(\d+)C`,
						ExtractValueCompiled: regexp.MustCompile(`(\d+)C`),
					},
				},
				{
					Symbols: []SymbolConfig{
						{
							OID:                  "1.2",
							Name:                 "hey",
							ExtractValue:         `(\d+)C`,
							ExtractValueCompiled: regexp.MustCompile(`(\d+)C`),
						},
					},
					MetricTags: MetricTagConfigList{
						MetricTagConfig{
							Column: SymbolConfig{
								OID:                  "1.2.3",
								Name:                 "abc",
								ExtractValue:         `(\d+)C`,
								ExtractValueCompiled: regexp.MustCompile(`(\d+)C`),
							},
							Tag: "hello",
						},
					},
				},
			},
			expectedErrors: []string{},
		},
		{
			name: "error compiling extract_value",
			metrics: []MetricsConfig{
				{
					Symbol: SymbolConfig{
						OID:          "1.2.3",
						Name:         "myMetric",
						ExtractValue: "[{",
					},
				},
			},
			expectedErrors: []string{
				"cannot compile `extract_value`",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateEnrichMetrics(tt.metrics)
			assert.Equal(t, len(tt.expectedErrors), len(errors), fmt.Sprintf("ERRORS: %v", errors))
			for i := range errors {
				assert.Contains(t, errors[i], tt.expectedErrors[i])
			}
			if tt.expectedMetrics != nil {
				assert.Equal(t, tt.expectedMetrics, tt.metrics)
			}
		})
	}
}

// go test -v -timeout 30s -run ^Test_validateEnrichMetadata$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp/snmputil
func Test_validateEnrichMetadata(t *testing.T) {
	tests := []struct {
		name             string
		metadata         MetadataConfig
		expectedErrors   []string
		expectedMetadata MetadataConfig
	}{
		{
			name: "both field symbol and value can be provided",
			metadata: MetadataConfig{
				"device": MetadataResourceConfig{
					Fields: map[string]MetadataField{
						"name": {
							Value: "hey",
							Symbol: SymbolConfig{
								OID:  "1.2.3",
								Name: "someSymbol",
							},
						},
					},
				},
			},
			expectedMetadata: MetadataConfig{
				"device": MetadataResourceConfig{
					Fields: map[string]MetadataField{
						"name": {
							Value: "hey",
							Symbol: SymbolConfig{
								OID:  "1.2.3",
								Name: "someSymbol",
							},
						},
					},
				},
			},
		},
		{
			name: "invalid regex pattern for symbol",
			metadata: MetadataConfig{
				"device": MetadataResourceConfig{
					Fields: map[string]MetadataField{
						"name": {
							Symbol: SymbolConfig{
								OID:          "1.2.3",
								Name:         "someSymbol",
								ExtractValue: "(\\w[)",
							},
						},
					},
				},
			},
			expectedErrors: []string{
				"cannot compile `extract_value`",
			},
		},
		{
			name: "invalid regex pattern for multiple symbols",
			metadata: MetadataConfig{
				"device": MetadataResourceConfig{
					Fields: map[string]MetadataField{
						"name": {
							Symbols: []SymbolConfig{
								{
									OID:          "1.2.3",
									Name:         "someSymbol",
									ExtractValue: "(\\w[)",
								},
							},
						},
					},
				},
			},
			expectedErrors: []string{
				"cannot compile `extract_value`",
			},
		},
		{
			name: "field regex pattern is compiled",
			metadata: MetadataConfig{
				"device": MetadataResourceConfig{
					Fields: map[string]MetadataField{
						"name": {
							Symbol: SymbolConfig{
								OID:          "1.2.3",
								Name:         "someSymbol",
								ExtractValue: "(\\w)",
							},
						},
					},
				},
			},
			expectedErrors: []string{},
			expectedMetadata: MetadataConfig{
				"device": MetadataResourceConfig{
					Fields: map[string]MetadataField{
						"name": {
							Symbol: SymbolConfig{
								OID:                  "1.2.3",
								Name:                 "someSymbol",
								ExtractValue:         "(\\w)",
								ExtractValueCompiled: regexp.MustCompile(`(\w)`),
							},
						},
					},
				},
			},
		},
		{
			name: "invalid resource",
			metadata: MetadataConfig{
				"invalid-res": MetadataResourceConfig{
					Fields: map[string]MetadataField{
						"name": {
							Value: "hey",
						},
					},
				},
			},
			expectedErrors: []string{
				"invalid resource: invalid-res",
			},
		},
		{
			name: "invalid field",
			metadata: MetadataConfig{
				"device": MetadataResourceConfig{
					Fields: map[string]MetadataField{
						"invalid-field": {
							Value: "hey",
						},
					},
				},
			},
			expectedErrors: []string{
				"invalid resource (device) field: invalid-field",
			},
		},
		{
			name: "invalid idtags",
			metadata: MetadataConfig{
				"interface": MetadataResourceConfig{
					Fields: map[string]MetadataField{
						"invalid-field": {
							Value: "hey",
						},
					},
					IDTags: MetricTagConfigList{
						MetricTagConfig{
							Column: SymbolConfig{
								OID:  "1.2.3",
								Name: "abc",
							},
							Match: "([a-z)",
							Tags: map[string]string{
								"foo": "bar",
							},
						},
					},
				},
			},
			expectedErrors: []string{
				"invalid resource (interface) field: invalid-field",
				"cannot compile `match` (`([a-z)`)",
			},
		},
		{
			name: "device resource does not support id_tags",
			metadata: MetadataConfig{
				"device": MetadataResourceConfig{
					Fields: map[string]MetadataField{
						"name": {
							Value: "hey",
						},
					},
					IDTags: MetricTagConfigList{
						MetricTagConfig{
							Column: SymbolConfig{
								OID:  "1.2.3",
								Name: "abc",
							},
							Tag: "abc",
						},
					},
				},
			},
			expectedErrors: []string{
				"device resource does not support custom id_tags",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validateEnrichMetadata(tt.metadata)
			assert.Equal(t, len(tt.expectedErrors), len(errors), fmt.Sprintf("ERRORS: %v", errors))
			for i := range errors {
				assert.Contains(t, errors[i], tt.expectedErrors[i])
			}
			if tt.expectedMetadata != nil {
				assert.Equal(t, tt.expectedMetadata, tt.metadata)
			}
		})
	}
}

// =============================================================================
