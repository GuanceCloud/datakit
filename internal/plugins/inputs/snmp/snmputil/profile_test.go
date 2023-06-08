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

	assert "github.com/stretchr/testify/require"
)

func fixtureProfileDefinitionMap() ProfileDefinitionMap {
	metrics := []MetricsConfig{
		{Symbol: SymbolConfig{OID: "1.3.6.1.4.1.3375.2.1.1.2.1.44.0", Name: "sysStatMemoryTotal", ScaleFactor: 2}, ForcedType: "gauge"},
		{Symbol: SymbolConfig{OID: "1.3.6.1.4.1.3375.2.1.1.2.1.44.999", Name: "oldSyntax"}},
		{
			ForcedType: "monotonic_count",
			Symbols: []SymbolConfig{
				{OID: "1.3.6.1.2.1.2.2.1.14", Name: "ifInErrors", ScaleFactor: 0.5},
				{OID: "1.3.6.1.2.1.2.2.1.13", Name: "ifInDiscards"},
			},
			MetricTags: []MetricTagConfig{
				{Tag: "interface", Column: SymbolConfig{OID: "1.3.6.1.2.1.31.1.1.1.1", Name: "ifName"}},
				{Tag: "interface_alias", Column: SymbolConfig{OID: "1.3.6.1.2.1.31.1.1.1.18", Name: "ifAlias"}},
				{Tag: "mac_address", Column: SymbolConfig{OID: "1.3.6.1.2.1.2.2.1.6", Name: "ifPhysAddress", Format: "mac_address"}},
			},
			StaticTags: []string{"table_static_tag:val"},
		},
		{Symbol: SymbolConfig{OID: "1.2.3.4.5", Name: "someMetric"}},
	}
	return ProfileDefinitionMap{"f5-big-ip": ProfileDefinition{
		Metrics:      metrics,
		Extends:      []string{"_base.yaml", "_generic-if.yaml"},
		Device:       DeviceMeta{Vendor: "f5"},
		SysObjectIds: StringArray{"1.3.6.1.4.1.3375.2.1.3.4.*"},
		StaticTags:   []string{"static_tag:from_profile_root", "static_tag:from_base_profile"},
		MetricTags: []MetricTagConfig{
			{
				OID:     "1.3.6.1.2.1.1.5.0",
				Name:    "sysName",
				Match:   "(\\w)(\\w+)",
				pattern: regexp.MustCompile("(\\w)(\\w+)"), //nolint:gosimple
				Tags: map[string]string{
					"some_tag": "some_tag_value",
					"prefix":   "\\1",
					"suffix":   "\\2",
				},
			},
			{Tag: "snmp_host", Index: 0x0, Column: SymbolConfig{OID: "", Name: ""}, OID: "1.3.6.1.2.1.1.5.0", Name: "sysName"},
		},
		Metadata: MetadataConfig{
			"device": {
				Fields: map[string]MetadataField{
					"vendor": {
						Value: "f5",
					},
					"description": {
						Symbol: SymbolConfig{
							OID:  "1.3.6.1.2.1.1.1.0",
							Name: "sysDescr",
						},
					},
					"name": {
						Symbol: SymbolConfig{
							OID:  "1.3.6.1.2.1.1.5.0",
							Name: "sysName",
						},
					},
					"serial_number": {
						Symbol: SymbolConfig{
							OID:  "1.3.6.1.4.1.3375.2.1.3.3.3.0",
							Name: "sysGeneralChassisSerialNum",
						},
					},
					"sys_object_id": {
						Symbol: SymbolConfig{
							OID:  "1.3.6.1.2.1.1.2.0",
							Name: "sysObjectID",
						},
					},
				},
			},
			"interface": {
				Fields: map[string]MetadataField{
					"admin_status": {
						Symbol: SymbolConfig{
							OID:  "1.3.6.1.2.1.2.2.1.7",
							Name: "ifAdminStatus",
						},
					},
					"alias": {
						Symbol: SymbolConfig{
							OID:  "1.3.6.1.2.1.31.1.1.1.18",
							Name: "ifAlias",
						},
					},
					"description": {
						Symbol: SymbolConfig{
							OID:                  "1.3.6.1.2.1.31.1.1.1.1",
							Name:                 "ifName",
							ExtractValue:         "(Row\\d)",
							ExtractValueCompiled: regexp.MustCompile("(Row\\d)"), //nolint:gosimple
						},
					},
					"mac_address": {
						Symbol: SymbolConfig{
							OID:    "1.3.6.1.2.1.2.2.1.6",
							Name:   "ifPhysAddress",
							Format: "mac_address",
						},
					},
					"name": {
						Symbol: SymbolConfig{
							OID:  "1.3.6.1.2.1.31.1.1.1.1",
							Name: "ifName",
						},
					},
					"oper_status": {
						Symbol: SymbolConfig{
							OID:  "1.3.6.1.2.1.2.2.1.8",
							Name: "ifOperStatus",
						},
					},
				},
				IDTags: MetricTagConfigList{
					{
						Tag: "custom-tag",
						Column: SymbolConfig{
							OID:  "1.3.6.1.2.1.31.1.1.1.1",
							Name: "ifAlias",
						},
					},
					{
						Tag: "interface",
						Column: SymbolConfig{
							OID:  "1.3.6.1.2.1.31.1.1.1.1",
							Name: "ifName",
						},
					},
				},
			},
		},
	}}
}

func Test_getMostSpecificOid(t *testing.T) {
	tests := []struct {
		name           string
		oids           []string
		expectedOid    string
		expectedErrror error
	}{
		{
			"one",
			[]string{"1.2.3.4"},
			"1.2.3.4",
			nil,
		},
		{
			"error on empty oids",
			[]string{},
			"",
			fmt.Errorf("cannot get most specific oid from empty list of oids"),
		},
		{
			"error on parsing",
			[]string{"a.1.2.3"},
			"",
			fmt.Errorf("error parsing part `a` for pattern `a.1.2.3`: strconv.Atoi: parsing \"a\": invalid syntax"),
		},
		{
			"most lengthy",
			[]string{"1.3.4", "1.3.4.1.2"},
			"1.3.4.1.2",
			nil,
		},
		{
			"wild card 1",
			[]string{"1.3.4.*", "1.3.4.1"},
			"1.3.4.1",
			nil,
		},
		{
			"wild card 2",
			[]string{"1.3.4.1", "1.3.4.*"},
			"1.3.4.1",
			nil,
		},
		{
			"sample oids",
			[]string{"1.3.6.1.4.1.3375.2.1.3.4.43", "1.3.6.1.4.1.8072.3.2.10"},
			"1.3.6.1.4.1.3375.2.1.3.4.43",
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oid, err := getMostSpecificOid(tt.oids)
			assert.Equal(t, tt.expectedErrror, err)
			assert.Equal(t, tt.expectedOid, oid)
		})
	}
}

func Test_mergeProfileDefinition(t *testing.T) {
	okBaseDefinition := ProfileDefinition{
		Metrics: []MetricsConfig{
			{Symbol: SymbolConfig{OID: "1.1", Name: "metric1"}, ForcedType: "gauge"},
		},
		MetricTags: []MetricTagConfig{
			{
				Tag:  "tag1",
				OID:  "2.1",
				Name: "tagName1",
			},
		},
		Metadata: MetadataConfig{
			"device": {
				Fields: map[string]MetadataField{
					"vendor": {
						Value: "f5",
					},
					"description": {
						Symbol: SymbolConfig{
							OID:  "1.3.6.1.2.1.1.1.0",
							Name: "sysDescr",
						},
					},
				},
			},
			"interface": {
				Fields: map[string]MetadataField{
					"admin_status": {
						Symbol: SymbolConfig{
							OID:  "1.3.6.1.2.1.2.2.1.7",
							Name: "ifAdminStatus",
						},
					},
				},
				IDTags: MetricTagConfigList{
					{
						Tag: "alias",
						Column: SymbolConfig{
							OID:  "1.3.6.1.2.1.31.1.1.1.1",
							Name: "ifAlias",
						},
					},
				},
			},
		},
	}
	emptyBaseDefinition := ProfileDefinition{}
	okTargetDefinition := ProfileDefinition{
		Metrics: []MetricsConfig{
			{Symbol: SymbolConfig{OID: "1.2", Name: "metric2"}, ForcedType: "gauge"},
		},
		MetricTags: []MetricTagConfig{
			{
				Tag:  "tag2",
				OID:  "2.2",
				Name: "tagName2",
			},
		},
		Metadata: MetadataConfig{
			"device": {
				Fields: map[string]MetadataField{
					"name": {
						Symbol: SymbolConfig{
							OID:  "1.3.6.1.2.1.1.5.0",
							Name: "sysName",
						},
					},
				},
			},
			"interface": {
				Fields: map[string]MetadataField{
					"oper_status": {
						Symbol: SymbolConfig{
							OID:  "1.3.6.1.2.1.2.2.1.8",
							Name: "ifOperStatus",
						},
					},
				},
				IDTags: MetricTagConfigList{
					{
						Tag: "interface",
						Column: SymbolConfig{
							OID:  "1.3.6.1.2.1.31.1.1.1.1",
							Name: "ifName",
						},
					},
				},
			},
		},
	}
	tests := []struct {
		name               string
		targetDefinition   ProfileDefinition
		baseDefinition     ProfileDefinition
		expectedDefinition ProfileDefinition
	}{
		{
			name:             "merge case",
			baseDefinition:   copyProfileDefinition(okBaseDefinition),
			targetDefinition: copyProfileDefinition(okTargetDefinition),
			expectedDefinition: ProfileDefinition{
				Metrics: []MetricsConfig{
					{Symbol: SymbolConfig{OID: "1.2", Name: "metric2"}, ForcedType: "gauge"},
					{Symbol: SymbolConfig{OID: "1.1", Name: "metric1"}, ForcedType: "gauge"},
				},
				MetricTags: []MetricTagConfig{
					{
						Tag:  "tag2",
						OID:  "2.2",
						Name: "tagName2",
					},
					{
						Tag:  "tag1",
						OID:  "2.1",
						Name: "tagName1",
					},
				},
				Metadata: MetadataConfig{
					"device": {
						Fields: map[string]MetadataField{
							"vendor": {
								Value: "f5",
							},
							"name": {
								Symbol: SymbolConfig{
									OID:  "1.3.6.1.2.1.1.5.0",
									Name: "sysName",
								},
							},
							"description": {
								Symbol: SymbolConfig{
									OID:  "1.3.6.1.2.1.1.1.0",
									Name: "sysDescr",
								},
							},
						},
					},
					"interface": {
						Fields: map[string]MetadataField{
							"oper_status": {
								Symbol: SymbolConfig{
									OID:  "1.3.6.1.2.1.2.2.1.8",
									Name: "ifOperStatus",
								},
							},
							"admin_status": {
								Symbol: SymbolConfig{
									OID:  "1.3.6.1.2.1.2.2.1.7",
									Name: "ifAdminStatus",
								},
							},
						},
						IDTags: MetricTagConfigList{
							{
								Tag: "interface",
								Column: SymbolConfig{
									OID:  "1.3.6.1.2.1.31.1.1.1.1",
									Name: "ifName",
								},
							},
							{
								Tag: "alias",
								Column: SymbolConfig{
									OID:  "1.3.6.1.2.1.31.1.1.1.1",
									Name: "ifAlias",
								},
							},
						},
					},
				},
			},
		},
		{
			name:             "empty base definition",
			baseDefinition:   copyProfileDefinition(emptyBaseDefinition),
			targetDefinition: copyProfileDefinition(okTargetDefinition),
			expectedDefinition: ProfileDefinition{
				Metrics: []MetricsConfig{
					{Symbol: SymbolConfig{OID: "1.2", Name: "metric2"}, ForcedType: "gauge"},
				},
				MetricTags: []MetricTagConfig{
					{
						Tag:  "tag2",
						OID:  "2.2",
						Name: "tagName2",
					},
				},
				Metadata: MetadataConfig{
					"device": {
						Fields: map[string]MetadataField{
							"name": {
								Symbol: SymbolConfig{
									OID:  "1.3.6.1.2.1.1.5.0",
									Name: "sysName",
								},
							},
						},
					},
					"interface": {
						Fields: map[string]MetadataField{
							"oper_status": {
								Symbol: SymbolConfig{
									OID:  "1.3.6.1.2.1.2.2.1.8",
									Name: "ifOperStatus",
								},
							},
						},
						IDTags: MetricTagConfigList{
							{
								Tag: "interface",
								Column: SymbolConfig{
									OID:  "1.3.6.1.2.1.31.1.1.1.1",
									Name: "ifName",
								},
							},
						},
					},
				},
			},
		},
		{
			name:             "empty taget definition",
			baseDefinition:   copyProfileDefinition(okBaseDefinition),
			targetDefinition: copyProfileDefinition(emptyBaseDefinition),
			expectedDefinition: ProfileDefinition{
				Metrics: []MetricsConfig{
					{Symbol: SymbolConfig{OID: "1.1", Name: "metric1"}, ForcedType: "gauge"},
				},
				MetricTags: []MetricTagConfig{
					{
						Tag:  "tag1",
						OID:  "2.1",
						Name: "tagName1",
					},
				},
				Metadata: MetadataConfig{
					"device": {
						Fields: map[string]MetadataField{
							"vendor": {
								Value: "f5",
							},
							"description": {
								Symbol: SymbolConfig{
									OID:  "1.3.6.1.2.1.1.1.0",
									Name: "sysDescr",
								},
							},
						},
					},
					"interface": {
						Fields: map[string]MetadataField{
							"admin_status": {
								Symbol: SymbolConfig{
									OID:  "1.3.6.1.2.1.2.2.1.7",
									Name: "ifAdminStatus",
								},
							},
						},
						IDTags: MetricTagConfigList{
							{
								Tag: "alias",
								Column: SymbolConfig{
									OID:  "1.3.6.1.2.1.31.1.1.1.1",
									Name: "ifAlias",
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mergeProfileDefinition(&tt.targetDefinition, &tt.baseDefinition)
			assert.Equal(t, tt.expectedDefinition.Metrics, tt.targetDefinition.Metrics)
			assert.Equal(t, tt.expectedDefinition.MetricTags, tt.targetDefinition.MetricTags)
			assert.Equal(t, tt.expectedDefinition.Metadata, tt.targetDefinition.Metadata)
		})
	}
}
