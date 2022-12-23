// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmp

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/snmp/snmputil"
)

// go test -v -timeout 30s -run ^Test_refreshWithProfile$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/snmp
func Test_refreshWithProfile(t *testing.T) {
	metrics := []snmputil.MetricsConfig{
		{Symbol: snmputil.SymbolConfig{OID: "1.2.3.4.5", Name: "someMetric"}},
		{
			Symbols: []snmputil.SymbolConfig{
				{
					OID:  "1.2.3.4.6",
					Name: "abc",
				},
			},
			MetricTags: snmputil.MetricTagConfigList{
				snmputil.MetricTagConfig{
					Column: snmputil.SymbolConfig{
						OID: "1.2.3.4.7",
					},
				},
			},
		},
	}
	profile1 := snmputil.ProfileDefinition{
		Device: snmputil.DeviceMeta{
			Vendor: "a-vendor",
		},
		Metrics: metrics,
		MetricTags: []snmputil.MetricTagConfig{
			{Tag: "interface", Column: snmputil.SymbolConfig{OID: "1.3.6.1.2.1.31.1.1.1.1", Name: "ifName"}},
		},
		Metadata: snmputil.MetadataConfig{
			"device": {
				Fields: map[string]snmputil.MetadataField{
					"description": {
						Symbol: snmputil.SymbolConfig{
							OID:  "1.3.6.1.2.1.1.99.3.0",
							Name: "sysDescr",
						},
					},
					"name": {
						Symbols: []snmputil.SymbolConfig{
							{
								OID:  "1.3.6.1.2.1.1.99.1.0",
								Name: "symbol1",
							},
							{
								OID:  "1.3.6.1.2.1.1.99.2.0",
								Name: "symbol2",
							},
						},
					},
				},
			},
			"interface": {
				Fields: map[string]snmputil.MetadataField{
					"oper_status": {
						Symbol: snmputil.SymbolConfig{
							OID:  "1.3.6.1.2.1.2.2.1.99",
							Name: "someIfSymbol",
						},
					},
				},
				IDTags: snmputil.MetricTagConfigList{
					{
						Tag: "interface",
						Column: snmputil.SymbolConfig{
							OID:  "1.3.6.1.2.1.31.1.1.1.1",
							Name: "ifName",
						},
					},
				},
			},
		},
		SysObjectIds: snmputil.StringArray{"1.3.6.1.4.1.3375.2.1.3.4.*"},
	}
	mockProfiles := snmputil.ProfileDefinitionMap{
		"profile1": profile1,
	}
	ipt := &Input{Profiles: mockProfiles}
	c := &deviceInfo{
		Ipt: ipt,
		IP:  "1.2.3.4",
	}
	err := c.refreshWithProfile("f5")
	assert.EqualError(t, err, "unknown profile `f5`")

	err = c.refreshWithProfile("profile1")
	assert.NoError(t, err)

	assert.Equal(t, "profile1", c.Profile)
	assert.Equal(t, profile1, *c.ProfileDef)
	assert.Equal(t, metrics, c.Metrics)
	assert.Equal(t, []snmputil.MetricTagConfig{
		{Tag: "interface", Column: snmputil.SymbolConfig{OID: "1.3.6.1.2.1.31.1.1.1.1", Name: "ifName"}},
	}, c.MetricTags)
	assert.Equal(t, snmputil.OidConfig{
		ScalarOids: []string{"1.2.3.4.5"},
		ColumnOids: []string{"1.2.3.4.6", "1.2.3.4.7"},
	}, c.OidConfig)
	assert.Equal(t, []string{"snmp_profile:profile1", "device_vendor:a-vendor"}, c.ProfileTags)

	ipt = &Input{Profiles: mockProfiles}
	c = &deviceInfo{
		Ipt:                   ipt,
		IP:                    "1.2.3.4",
		CollectDeviceMetadata: true,
	}
	err = c.refreshWithProfile("profile1")
	assert.NoError(t, err)
	assert.Equal(t, snmputil.OidConfig{
		ScalarOids: []string{
			"1.2.3.4.5",
			"1.3.6.1.2.1.1.99.1.0",
			"1.3.6.1.2.1.1.99.2.0",
			"1.3.6.1.2.1.1.99.3.0",
		},
		ColumnOids: []string{
			"1.2.3.4.6",
			"1.2.3.4.7",
			"1.3.6.1.2.1.2.2.1.99",
			"1.3.6.1.2.1.31.1.1.1.1",
		},
	}, c.OidConfig)

	// With metadata disabled
	c.CollectDeviceMetadata = false
	err = c.refreshWithProfile("profile1")
	assert.NoError(t, err)
	assert.Equal(t, snmputil.OidConfig{
		ScalarOids: []string{
			"1.2.3.4.5",
		},
		ColumnOids: []string{
			"1.2.3.4.6",
			"1.2.3.4.7",
		},
	}, c.OidConfig)
}

//------------------------------------------------------------------------------

// go test -v -timeout 30s -run ^Test_ReportNetworkDeviceMetadata$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/snmp
func Test_ReportNetworkDeviceMetadata(t *testing.T) {
	cases := []struct {
		name    string
		di      *deviceInfo
		outTags []string
	}{
		{
			name:    "normal",
			di:      &deviceInfo{},
			outTags: []string{"tag1", "tag2", "device_namespace:", "snmp_device:"},
		},
	}

	type deviceUnit struct {
		Tags []string `json:"tags"`
	}

	type dataStruct struct {
		Devices []deviceUnit `json:"devices"`
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			emptyMetadataStore := &snmputil.ResultValueStore{
				ColumnValues: snmputil.ColumnResultValuesType{},
			}
			emptyMetadataConfigs := snmputil.MetadataConfig{}
			collectTime := time.Now()
			out := deviceMetaData{}

			di := tc.di
			di.ReportNetworkDeviceMetadata(emptyMetadataStore, []string{"tag1", "tag2"}, emptyMetadataConfigs, collectTime, 1, &out)

			for k, v := range out.data {
				data := dataStruct{}
				json.Unmarshal([]byte(v), &data)
				assert.Equal(t, tc.outTags, data.Devices[k].Tags)
			}
		})
	}
}

//------------------------------------------------------------------------------

// go test -v -timeout 30s -run ^Test_getDeviceID$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/snmp
func Test_getDeviceID(t *testing.T) {
	cases := []struct {
		name string
		di   *deviceInfo
		out  string
	}{
		{
			name: "normal",
			di: &deviceInfo{
				Namespace: "default",
				IP:        "192.168.1.220",
			},
			out: "default:192.168.1.220",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			di := tc.di
			out := di.getDeviceID()
			assert.Equal(t, tc.out, out)
		})
	}
}

// go test -v -timeout 30s -run ^Test_getDeviceIDTags$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/snmp
func Test_getDeviceIDTags(t *testing.T) {
	cases := []struct {
		name string
		di   *deviceInfo
		out  []string
	}{
		{
			name: "normal",
			di: &deviceInfo{
				Namespace: "default",
				IP:        "192.168.1.220",
			},
			out: []string{
				"device_namespace:default",
				"snmp_device:192.168.1.220",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			di := tc.di
			out := di.getDeviceIDTags()
			assert.Equal(t, tc.out, out)
		})
	}
}
