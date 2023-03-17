// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package snmp

import (
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/snmp/snmpmeasurement"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/snmp/snmputil"
)

// go test -v -timeout 30s -run ^Test_AvailableArchs$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/snmp
func Test_AvailableArchs(t *testing.T) {
	ipt := &Input{}
	out := ipt.AvailableArchs()
	assert.Equal(t, datakit.AllOS, out)
}

// go test -v -timeout 30s -run ^Test_SampleMeasurement$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/snmp
func Test_SampleMeasurement(t *testing.T) {
	ipt := &Input{}
	out := ipt.SampleMeasurement()
	assert.Equal(t, []inputs.Measurement{&snmpmeasurement.SNMPObject{}, &snmpmeasurement.SNMPMetric{}}, out)
}

// go test -v -timeout 30s -run ^Test_calcTagsHash$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/snmp
func Test_calcTagsHash(t *testing.T) {
	cases := []struct {
		name string
		in   *snmputil.MetricDatas
		out  *snmputil.MetricDatas
	}{
		{
			name: "normal",
			in: &snmputil.MetricDatas{
				Data: []*snmputil.MetricData{
					{
						Name:     "key1",
						Value:    1.0,
						Tags:     []string{"abc", "def"},
						TagsHash: "",
					},
					{
						Name:     "key2",
						Value:    2.0,
						Tags:     []string{"abc", "def"},
						TagsHash: "",
					},
					{
						Name:     "key3",
						Value:    3.0,
						Tags:     []string{"abc", "def", "apple"},
						TagsHash: "",
					},
					{
						Name:     "key4",
						Value:    4.0,
						Tags:     []string{"abc", "def", "apple"},
						TagsHash: "",
					},
					{
						Name:     "key5",
						Value:    5.0,
						Tags:     []string{"abc", "def", "apple"},
						TagsHash: "",
					},
				},
			},
			out: &snmputil.MetricDatas{
				Data: []*snmputil.MetricData{
					{
						Name:     "key1",
						Value:    1.0,
						Tags:     []string{"abc", "def"},
						TagsHash: "e80b5017098950fc58aad83c8c14978e",
					},
					{
						Name:     "key2",
						Value:    2.0,
						Tags:     []string{"abc", "def"},
						TagsHash: "e80b5017098950fc58aad83c8c14978e",
					},
					{
						Name:     "key3",
						Value:    3.0,
						Tags:     []string{"abc", "def", "apple"},
						TagsHash: "972d3055287ad2a9f007321eaa601c54",
					},
					{
						Name:     "key4",
						Value:    4.0,
						Tags:     []string{"abc", "def", "apple"},
						TagsHash: "972d3055287ad2a9f007321eaa601c54",
					},
					{
						Name:     "key5",
						Value:    5.0,
						Tags:     []string{"abc", "def", "apple"},
						TagsHash: "972d3055287ad2a9f007321eaa601c54",
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			calcTagsHash(tc.in)
			assert.Equal(t, tc.out, tc.in)
		})
	}
}

// go test -v -timeout 30s -run ^Test_aggregateHash$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/snmp
func Test_aggregateHash(t *testing.T) {
	cases := []struct {
		name       string
		metricData *snmputil.MetricDatas
		in         map[string]map[string]interface{}
		out        map[string]map[string]interface{}
	}{
		{
			name: "normal",
			metricData: &snmputil.MetricDatas{
				Data: []*snmputil.MetricData{
					{
						Name:     "key1",
						Value:    1.0,
						Tags:     []string{"abc", "def"},
						TagsHash: "fake_hash_1",
					},
					{
						Name:     "key2",
						Value:    2.0,
						Tags:     []string{"abc", "def"},
						TagsHash: "fake_hash_1",
					},
					{
						Name:     "key3",
						Value:    3.0,
						Tags:     []string{"abc", "def", "apple"},
						TagsHash: "fake_hash_2",
					},
					{
						Name:     "key4",
						Value:    4.0,
						Tags:     []string{"abc", "def", "apple"},
						TagsHash: "fake_hash_2",
					},
					{
						Name:     "key5",
						Value:    5.0,
						Tags:     []string{"abc", "def", "apple"},
						TagsHash: "fake_hash_2",
					},
				},
			},
			in: make(map[string]map[string]interface{}),
			out: map[string]map[string]interface{}{
				"fake_hash_1": {
					"key1": float64(1.0),
					"key2": float64(2.0),
				},
				"fake_hash_2": {
					"key3": float64(3.0),
					"key4": float64(4.0),
					"key5": float64(5.0),
				},
			},
		},
		{
			name: "larger",
			metricData: &snmputil.MetricDatas{
				Data: []*snmputil.MetricData{
					{
						Name:     "key1",
						Value:    1.0,
						Tags:     []string{"abc", "def"},
						TagsHash: "fake_hash_1",
					},
					{
						Name:     "key2",
						Value:    2.0,
						Tags:     []string{"abc", "def"},
						TagsHash: "fake_hash_1",
					},
					{
						Name:     "key3",
						Value:    3.0,
						Tags:     []string{"abc", "def", "apple"},
						TagsHash: "fake_hash_2",
					},
					{
						Name:     "key4",
						Value:    4.0,
						Tags:     []string{"abc", "def", "apple"},
						TagsHash: "fake_hash_2",
					},
					{
						Name:     "key4",
						Value:    5.0,
						Tags:     []string{"abc", "def", "apple"},
						TagsHash: "fake_hash_2",
					},
				},
			},
			in: make(map[string]map[string]interface{}),
			out: map[string]map[string]interface{}{
				"fake_hash_1": {
					"key1": float64(1.0),
					"key2": float64(2.0),
				},
				"fake_hash_2": {
					"key3": float64(3.0),
					"key4": float64(5.0),
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			aggregateHash(tc.metricData, tc.in)
			assert.Equal(t, tc.out, tc.in)
		})
	}
}

// go test -v -timeout 30s -count=1 -run ^Test_getFieldTagArr$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/snmp
func Test_getFieldTagArr(t *testing.T) {
	cases := []struct {
		name       string
		metricData *snmputil.MetricDatas
		mHash      map[string]map[string]interface{}
		metaData   *deviceMetaData
		origTags   []string
		customTags map[string]string
		out        tagFields
	}{
		{
			name: "empty_hash",
		},
		{
			name: "collect_meta",
			metricData: &snmputil.MetricDatas{
				Data: []*snmputil.MetricData{
					{
						Name:     "key1",
						Value:    1.0,
						Tags:     []string{"abc:value1", "def:value2"},
						TagsHash: "fake_hash_1",
					},
					{
						Name:     "key2",
						Value:    2.0,
						Tags:     []string{"abc:value1", "def:value2"},
						TagsHash: "fake_hash_1",
					},
					{
						Name:     "key3",
						Value:    3.0,
						Tags:     []string{"abc:value1", "def:value2", "apple:value3"},
						TagsHash: "fake_hash_2",
					},
					{
						Name:     "key4",
						Value:    4.0,
						Tags:     []string{"abc:value1", "def:value2", "apple:value3"},
						TagsHash: "fake_hash_2",
					},
					{
						Name:     "key5",
						Value:    5.0,
						Tags:     []string{"abc:value1", "def:value2", "apple:value3"},
						TagsHash: "fake_hash_2",
					},
				},
			},
			mHash: map[string]map[string]interface{}{
				"fake_hash_1": {
					"key1": float64(1.0),
					"key2": float64(2.0),
				},
				"fake_hash_2": {
					"key3": float64(3.0),
					"key4": float64(4.0),
					"key5": float64(5.0),
				},
			},
			metaData: &deviceMetaData{
				collectMeta: true,
				data: []string{
					"fruit1=banana",
					"fruit2=pear",
					"fruit3=tomato",
				},
			},
			origTags: []string{
				"device_namespace:default",
				"snmp_device:192.168.1.100",
			},
			out: tagFields{
				Data: []*tagField{
					{
						Tags: map[string]string{
							"abc": "value1",
							"def": "value2",
						},
						Fields: map[string]interface{}{
							"key1": float64(1.0),
							"key2": float64(2.0),
						},
					},
					{
						Tags: map[string]string{
							"abc":   "value1",
							"def":   "value2",
							"apple": "value3",
						},
						Fields: map[string]interface{}{
							"key3": float64(3.0),
							"key4": float64(4.0),
							"key5": float64(5.0),
						},
					},
					{
						Tags: map[string]string{
							"device_namespace": "default",
							"snmp_device":      "192.168.1.100",
						},
						Fields: map[string]interface{}{
							deviceMetaKey: "fruit1=banana, fruit2=pear, fruit3=tomato",
						},
					},
				},
			},
		},
		{
			name: "not_collect_meta",
			metricData: &snmputil.MetricDatas{
				Data: []*snmputil.MetricData{
					{
						Name:     "key1",
						Value:    1.0,
						Tags:     []string{"abc:value1", "def:value2"},
						TagsHash: "fake_hash_1",
					},
					{
						Name:     "key2",
						Value:    2.0,
						Tags:     []string{"abc:value1", "def:value2"},
						TagsHash: "fake_hash_1",
					},
					{
						Name:     "key3",
						Value:    3.0,
						Tags:     []string{"abc:value1", "def:value2", "apple:value3"},
						TagsHash: "fake_hash_2",
					},
					{
						Name:     "key4",
						Value:    4.0,
						Tags:     []string{"abc:value1", "def:value2", "apple:value3"},
						TagsHash: "fake_hash_2",
					},
					{
						Name:     "key5",
						Value:    5.0,
						Tags:     []string{"abc:value1", "def:value2", "apple:value3"},
						TagsHash: "fake_hash_2",
					},
				},
			},
			mHash: map[string]map[string]interface{}{
				"fake_hash_1": {
					"key1": float64(1.0),
					"key2": float64(2.0),
				},
				"fake_hash_2": {
					"key3": float64(3.0),
					"key4": float64(4.0),
					"key5": float64(5.0),
				},
			},
			metaData: &deviceMetaData{
				collectMeta: false,
				data: []string{
					"fruit1=banana",
					"fruit2=pear",
					"fruit3=tomato",
				},
			},
			out: tagFields{
				Data: []*tagField{
					{
						Tags: map[string]string{
							"abc": "value1",
							"def": "value2",
						},
						Fields: map[string]interface{}{
							"key1": float64(1.0),
							"key2": float64(2.0),
						},
					},
					{
						Tags: map[string]string{
							"abc":   "value1",
							"def":   "value2",
							"apple": "value3",
						},
						Fields: map[string]interface{}{
							"key3": float64(3.0),
							"key4": float64(4.0),
							"key5": float64(5.0),
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fts := tagFields{}
			getFieldTagArr(tc.metricData, tc.mHash, &fts, tc.metaData, tc.origTags, tc.customTags)
			for _, v := range tc.out.Data {
				foundIdx := -1
				for kk, vv := range fts.Data {
					resF := reflect.DeepEqual(v.Fields, vv.Fields)
					resT := reflect.DeepEqual(v.Tags, vv.Tags)
					if resF && resT {
						foundIdx = kk
						break
					}
				} // for
				assert.NotEqual(t, -1, foundIdx) // must found, so cannot be -1.
			}
		})
	}
}

// go test -v -timeout 30s -run ^Test_getDatakitStyleTags$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/snmp
func Test_getDatakitStyleTags(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		out  map[string]string
	}{
		{
			name: agentHostKey,
			in: []string{
				agentHostKey + ":apple",
			},
			out: map[string]string{},
		},
		{
			name: agentVersionKey,
			in: []string{
				agentVersionKey + ":apple",
			},
			out: map[string]string{},
		},
		{
			name: defaultSNMPHostKey,
			in: []string{
				defaultSNMPHostKey + ":apple",
			},
			out: map[string]string{
				defaultDatakitHostKey: "apple",
			},
		},
		{
			name: "other",
			in: []string{
				"whatever:apple",
			},
			out: map[string]string{
				"whatever": "apple",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := make(map[string]string)
			getDatakitStyleTags(tc.in, out)
			assert.Equal(t, tc.out, out)
		})
	}
}

// go test -v -timeout 30s -run ^Test_validateConfig$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/snmp
func Test_validateConfig(t *testing.T) {
	cases := []struct {
		name                string
		version             uint8
		autoDiscovery       []string
		discoveryIgnoredIPs []string
		specificDevices     []string
		err                 error
	}{
		{
			name: "snmp version error",
			err:  fmt.Errorf("`snmp_version` must be 1 or 2 or 3"),
		},
		{
			name:    "version only",
			version: 2,
		},
		{
			name:          "CIDR error",
			version:       2,
			autoDiscovery: []string{"", "192.168.1.100"},
			err:           &net.ParseError{Type: "CIDR address", Text: "192.168.1.100"},
		},
		{
			name:            "invalid IP address",
			version:         2,
			specificDevices: []string{"invalid_ip"},
			err:             fmt.Errorf("invalid IP address"),
		},
		{
			name:                "normal",
			version:             2,
			autoDiscovery:       []string{"", "192.168.1.0/24"},
			discoveryIgnoredIPs: []string{"", "192.168.1.101"},
			specificDevices:     []string{"", "192.168.1.100"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ipt := &Input{SNMPVersion: tc.version, AutoDiscovery: tc.autoDiscovery, DiscoveryIgnoredIPs: tc.discoveryIgnoredIPs, SpecificDevices: tc.specificDevices} //nolint:lll
			err := ipt.ValidateConfig()
			assert.Equal(t, tc.err, err)
		})
	}
}

// go test -v -timeout 30s -run ^Test_checkIPWorking_checkIPDone$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/snmp
func Test_checkIPWorking_checkIPDone(t *testing.T) {
	deviceIP1 := "1.2.3.4"
	deviceIP2 := "2.3.4.5"
	mWorkingIP.Store(deviceIP1, struct{}{})

	ipt := &Input{semStop: cliutils.NewSem()}
	ipt.checkIPWorking(deviceIP2)

	go func() {
		time.Sleep(time.Second)
		checkIPDone(deviceIP1)
	}()

	ipt.checkIPWorking(deviceIP1)
}

// go test -v -timeout 30s -run ^Test_normalizeFieldTags$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/snmp
func Test_normalizeFieldTags(t *testing.T) {
	cases := []struct {
		name string
		in   *tagField
		out  *tagField
	}{
		{
			name: "normal",
			in: &tagField{
				Tags: map[string]string{
					"aaa_a.a": "not_used",
				},
				Fields: map[string]interface{}{
					"aaa_a.a": "not_used",
				},
			},
			out: &tagField{
				Tags: map[string]string{
					"aaa_a_a": "not_used",
				},
				Fields: map[string]interface{}{
					"aaa_a_a": "not_used",
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			normalizeFieldTags(tc.in)
			assert.Equal(t, tc.out, tc.in)
		})
	}
}

// go test -v -timeout 30s -run ^Test_replaceMetricsName$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/snmp
func Test_replaceMetricsName(t *testing.T) {
	cases := []struct {
		name string
		in   string
		out  string
	}{
		{
			name: "underline_with_point",
			in:   "aaa_a.a",
			out:  "aaa_a_a",
		},
		{
			name: "underline_without_point",
			in:   "aaa_aa",
			out:  "",
		},
		{
			name: "underline_with_point_last",
			in:   "aaa_a.",
			out:  "aaa_a_",
		},
		{
			name: "CamelCase_with_point",
			in:   "Aaa.a",
			out:  "AaaA",
		},
		{
			name: "CamelCase_with_point_repeat",
			in:   "Aaa.aaa.aa.a",
			out:  "AaaAaaAaA",
		},
		{
			name: "CamelCase_without_point",
			in:   "Aaa",
			out:  "",
		},
		{
			name: "CamelCase_with_point_last",
			in:   "Aaa.",
			out:  "Aaa",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := replaceMetricsName(tc.in)
			assert.Equal(t, tc.out, out)
		})
	}
}
