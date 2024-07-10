// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp/snmputil"
	"gopkg.in/yaml.v2"
)

func Test_getPreprocessing(t *testing.T) {
	type args struct {
		item *snmputil.Item
	}
	tests := []struct {
		name string
		args args
		want snmputil.Preprocessing
	}{
		{
			name: "zabbix6.0",
			args: args{
				item: &snmputil.Item{
					Preprocessing: snmputil.Preprocessing{
						Steps: []*snmputil.Step{
							{
								Type:       "MULTIPLIER",
								Parameters: []string{"0.1"},
							},
							{
								Type:       "DISCARD_UNCHANGED_HEARTBEAT",
								Parameters: []string{"6h"},
							},
						},
					},
				},
			},
			want: snmputil.Preprocessing{
				Steps: []*snmputil.Step{
					{
						Parameter: 0.1,
					},
					{
						Parameter: 0,
					},
				},
			},
		},
		{
			name: "zabbix5.0.xml",
			args: args{
				item: &snmputil.Item{
					Preprocessing: snmputil.Preprocessing{
						Steps: []*snmputil.Step{
							{
								Type:   "MULTIPLIER",
								Params: []string{"0.1"},
							},
							{
								Type:   "DISCARD_UNCHANGED_HEARTBEAT",
								Params: []string{"6h"},
							},
						},
					},
				},
			},
			want: snmputil.Preprocessing{
				Steps: []*snmputil.Step{
					{
						Parameter: 0.1,
					},
					{
						Parameter: 0,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getPreprocessing(tt.args.item)

			for i, step := range got.Steps {
				assert.Equal(t, tt.want.Steps[i].Parameter, step.Parameter)
			}
		})
	}
}

func Test_formatMacroNames(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "5_macros",
			args: args{str: "discovery[{#SNMPVALUE},1.3.6.1.2.1.10.7.2.1.19,{#IFOPERSTATUS},1.3.6.1.2.1.2.2.1.8,{#IFALIAS},1.3.6.1.2.1.31.1.1.1.18,{#IFNAME},1.3.6.1.2.1.31.1.1.1.1,{#IFDESCR},1.3.6.1.2.1.2.2.1.2]"},
			want: map[string]string{
				"SNMPVALUE":    "1.3.6.1.2.1.10.7.2.1.19",
				"IFOPERSTATUS": "1.3.6.1.2.1.2.2.1.8",
				"IFALIAS":      "1.3.6.1.2.1.31.1.1.1.18",
				"IFNAME":       "1.3.6.1.2.1.31.1.1.1.1",
				"IFDESCR":      "1.3.6.1.2.1.2.2.1.2",
			},
		},
		{
			name: "1.5_macros_error",
			args: args{str: "discovery[{#SNMPVALUE},1.3.6.1.2.1.10.7.2.1.19,{#IFOPERSTATUS}"},
			want: map[string]string{
				"SNMPVALUE": "1.3.6.1.2.1.10.7.2.1.19",
			},
		},
		{
			name: "1_macros",
			args: args{str: "discovery[{#SNMPVALUE},1.3.6.1.2.1.10.7.2.1.19"},
			want: map[string]string{
				"SNMPVALUE": "1.3.6.1.2.1.10.7.2.1.19",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatMacroNames(tt.args.str)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_formatKey(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "any",
			args: args{s: "system.objectid[sysObjectID.0]"},
			want: "system_objectid",
		},
		{
			name: "any",
			args: args{s: "system.hw.model[entPhysicalDescr.{#SNMPINDEX}]"},
			want: "system_hw_model",
		},
		{
			name: "any",
			args: args{s: "vm.memory.pused[vm.memory.pused.{#SNMPINDEX}]"},
			want: "vm_memory_pused",
		},
		{
			name: "any",
			args: args{s: "system.objectid"},
			want: "system_objectid",
		},
		{
			name: "any",
			args: args{s: "system.uptime[sysUpTime]"},
			want: "system_uptime",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatKey(tt.args.s)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_validOID(t *testing.T) {
	type args struct {
		oid string
		m   map[string]snmputil.Macro
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "all_match",
			args: args{
				oid: "1.3.6.1.2.1.2.2.1.13.1",
				m: map[string]snmputil.Macro{
					"MACRO1": map[string]string{"1": "XXX", "2": "XXX"},
					"MACRO2": map[string]string{"1": "XXX", "2": "XXX"},
				},
			},
			want: true,
		},
		{
			name: "some_match",
			args: args{
				oid: "1.3.6.1.2.1.2.2.1.13.1",
				m: map[string]snmputil.Macro{
					"MACRO1": map[string]string{"3": "XXX", "2": "XXX"},
					"MACRO2": map[string]string{"1": "XXX", "2": "XXX"},
				},
			},
			want: true,
		},
		{
			name: "no_match",
			args: args{
				oid: "1.3.6.1.2.1.2.2.1.13.1",
				m: map[string]snmputil.Macro{
					"MACRO1": map[string]string{"3": "XXX", "2": "XXX"},
					"MACRO2": map[string]string{"4": "XXX", "2": "XXX"},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validOID(tt.args.oid, tt.args.m)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_macrosFilter(t *testing.T) {
	type args struct {
		macros map[string]snmputil.Macro
		filter snmputil.Filter
	}
	tests := []struct {
		name string
		args args
		want map[string]snmputil.Macro
	}{
		{
			name: "no_filter",
			args: args{
				macros: map[string]snmputil.Macro{
					"ENT_NAME": {"1": "SRU Board 0", "2": "GigabitEthernet0/0/1"},
					"IFALIAS":  {"1": "wireless-zhuyun-2F", "2": "wireless-zhuyun-4F"},
				},
			},
			want: map[string]snmputil.Macro{
				"ENT_NAME": {"1": "SRU Board 0", "2": "GigabitEthernet0/0/1"},
				"IFALIAS":  {"1": "wireless-zhuyun-2F", "2": "wireless-zhuyun-4F"},
			},
		},
		{
			name: "match_0",
			args: args{
				macros: map[string]snmputil.Macro{
					"ENT_NAME": {"1": "SRU Board 0", "2": "GigabitEthernet0/0/1"},
					"IFALIAS":  {"1": "wireless-zhuyun-2F", "2": "wireless-zhuyun-4F"},
				},
				filter: snmputil.Filter{
					Conditions: []snmputil.Condition{{
						Macro: "{#ENT_NAME}",
						Value: "MPU.*",
					}},
				},
			},
			want: map[string]snmputil.Macro{
				"ENT_NAME": {},
				"IFALIAS":  {},
			},
		},
		{
			name: "match_1",
			args: args{
				macros: map[string]snmputil.Macro{
					"ENT_NAME": {"1": "SRU Board 0", "2": "GigabitEthernet0/0/1"},
					"IFALIAS":  {"1": "wireless-zhuyun-2F", "2": "wireless-zhuyun-4F"},
				},
				filter: snmputil.Filter{
					Conditions: []snmputil.Condition{{
						Macro: "{#ENT_NAME}",
						Value: "(MPU.*|SRU.*)",
					}},
				},
			},
			want: map[string]snmputil.Macro{
				"ENT_NAME": {"1": "SRU Board 0"},
				"IFALIAS":  {"1": "wireless-zhuyun-2F"},
			},
		},
		{
			name: "as_NOT_MATCHES_REGEX_match_1",
			args: args{
				macros: map[string]snmputil.Macro{
					"ENT_NAME": {"1": "SRU Board 0", "2": "GigabitEthernet0/0/1"},
					"IFALIAS":  {"1": "wireless-zhuyun-2F", "2": "wireless-zhuyun-4F"},
				},
				filter: snmputil.Filter{
					Conditions: []snmputil.Condition{{
						Macro:    "{#ENT_NAME}",
						Value:    "(MPU.*|SRU.*)",
						Operator: "NOT_MATCHES_REGEX",
					}},
				},
			},
			want: map[string]snmputil.Macro{
				"ENT_NAME": {"2": "GigabitEthernet0/0/1"},
				"IFALIAS":  {"2": "wireless-zhuyun-4F"},
			},
		},
		{
			name: "as_NOT_MATCHES_REGEX_match_0",
			args: args{
				macros: map[string]snmputil.Macro{
					"ENT_NAME": {"1": "SRU Board 0", "2": "GigabitEthernet0/0/1"},
					"IFALIAS":  {"1": "wireless-zhuyun-2F", "2": "wireless-zhuyun-4F"},
				},
				filter: snmputil.Filter{
					Conditions: []snmputil.Condition{{
						Macro:    "{#ENT_NAME}",
						Value:    "MPU.*",
						Operator: "NOT_MATCHES_REGEX",
					}},
				},
			},
			want: map[string]snmputil.Macro{
				"ENT_NAME": {"1": "SRU Board 0", "2": "GigabitEthernet0/0/1"},
				"IFALIAS":  {"1": "wireless-zhuyun-2F", "2": "wireless-zhuyun-4F"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := macrosFilter(tt.args.macros, tt.args.filter)

			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_deviceInfo_formatDiscoveryItems(t *testing.T) {
	type args struct {
		sess   snmputil.Session
		item   *snmputil.Item
		macros map[string]snmputil.Macro
		ipt    *Input
	}
	tests := []struct {
		name string
		args args
		want []*snmputil.Item
	}{
		{
			name: "1.3.6.1.4.1.2011.5.25.31.1.1.1.1.5.9_cpuUsage",
			args: args{
				sess:   getMockSession01(),
				item:   getMockItem01(),
				macros: getMockMacros01(),
				ipt:    getMockInput01(),
			},
			want: []*snmputil.Item{{
				Key:  "cpuUsage",
				Tags: []*snmputil.Tag{{Tag: "component", Value: "cpu"}},
				OIDs: []string{"1.3.6.1.4.1.2011.5.25.31.1.1.1.1.5.9"},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			di := &deviceInfo{
				Session: tt.args.sess,
			}
			di.Ipt = tt.args.ipt

			got := di.formatDiscoveryItems(tt.args.item, tt.args.macros)

			assert.Equal(t, tt.want[0].Key, got[0].Key)
			assert.Equal(t, tt.want[0].Tags, got[0].Tags)
			assert.Equal(t, tt.want[0].OIDs, got[0].OIDs)
		})
	}
}

func Test_deviceInfo_getMacro(t *testing.T) {
	type args struct {
		sess snmputil.Session
		ipt  *Input
		oid  string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "1.3.6.1.2.1.2.2.1.2_IFDESCR",
			args: args{
				sess: getMockSession01(),
				ipt:  getMockInput01(),
				oid:  "1.3.6.1.2.1.2.2.1.2",
			},
			want: map[string]string{
				"1":   "InLoopBack0",
				"2":   "NULL0",
				"3":   "GigabitEthernet0/0/1",
				"4":   "GigabitEthernet0/0/2",
				"5":   "GigabitEthernet0/0/3",
				"6":   "GigabitEthernet0/0/4",
				"7":   "GigabitEthernet0/0/5",
				"8":   "GigabitEthernet0/0/6",
				"9":   "GigabitEthernet0/0/7",
				"10":  "GigabitEthernet0/0/8",
				"11":  "Vlanif1",
				"12":  "Vlanif305",
				"13":  "Eth-Trunk1",
				"196": "Vlanif302",
				"197": "Vlanif300",
				"198": "Vlanif304",
				"199": "Vlanif301",
				"200": "Vlanif303",
				"201": "Ethernet0/0/47",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			di := &deviceInfo{
				Session: tt.args.sess,
			}
			di.Ipt = tt.args.ipt

			got := di.getMacro(tt.args.oid)

			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_deviceInfo_preProcess(t *testing.T) {
	type args struct {
		sess snmputil.Session
		ipt  *Input
	}
	tests := []struct {
		name               string
		args               args
		wantStringItems    []snmputil.Item
		wantItems          []snmputil.Item
		wantDiscoveryItems []snmputil.Item
	}{
		{
			name: "preProcess_test",
			args: args{
				sess: getMockSession01(),
				ipt:  getMockInput01(),
			},
			wantStringItems: []snmputil.Item{
				{Key: "system_location", SnmpOID: "1.3.6.1.2.1.1.6.0"},
				{Key: "uptime", SnmpOID: "1.3.6.1.2.1.25.1.1.0"},
				{Key: "netUptime", SnmpOID: "1.3.6.1.2.1.1.3.0"},
			},
			wantItems: []snmputil.Item{{
				Key:     "netUptime",
				SnmpOID: "1.3.6.1.2.1.1.3.0",
				Steps: []*snmputil.Step{{
					Type:       "MULTIPLIER",
					Parameters: []string{"0.01"},
				}},
			}},
			wantDiscoveryItems: []snmputil.Item{
				{
					Key:    "cpuUsage",
					OIDs:   []string{"1.3.6.1.4.1.2011.5.25.31.1.1.1.1.5.9"},
					Macros: map[string]snmputil.Macro{"ENT_NAME": {"9": "SRU Board 0"}},
				},
				{
					Key:  "ifHighSpeed",
					OIDs: getOIDsFromMock(getIFSpeed03()),
					Macros: map[string]snmputil.Macro{
						"SNMPVALUE":     getMacroFromMock(getSNMPVALUE02()),
						"IFADMINSTATUS": getMacroFromMock(getIFADMINSTATUS02()),
						"IFALIAS":       getMacroFromMock(getIFALIAS02()),
						"IFNAME":        getMacroFromMock(getIFNAME02()),
						"IFDESCR":       getMacroFromMock(getIFDESCR02()),
						"IFTYPE":        getMacroFromMock(getIFTYPE02()),
					},
					Steps: []*snmputil.Step{{
						Type:       "MULTIPLIER",
						Parameters: []string{"1000000"},
					}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			di := &deviceInfo{UserProfileDefinition: snmputil.NewUserProfileDefinition()}
			err := yaml.Unmarshal([]byte(zabbixYaml), di.UserProfileDefinition)
			assert.NoError(t, err)
			di.Session = tt.args.sess
			di.Ipt = tt.args.ipt

			di.preProcess()

			for i, item := range di.UserProfileDefinition.StringItems {
				assert.Equal(t, tt.wantStringItems[i].Key, item.Key)
				assert.Equal(t, tt.wantStringItems[i].SnmpOID, item.SnmpOID)
			}

			for i, item := range di.UserProfileDefinition.Items {
				assert.Equal(t, tt.wantItems[i].Key, item.Key)
				assert.Equal(t, tt.wantItems[i].SnmpOID, item.SnmpOID)
				assert.Equal(t, tt.wantItems[i].Steps, item.Steps)
			}

			for i, item := range di.UserProfileDefinition.DiscoveryItems {
				assert.Equal(t, tt.wantDiscoveryItems[i].Key, item.Key)
				assert.Equal(t, tt.wantDiscoveryItems[i].OIDs, item.OIDs)
				assert.Equal(t, tt.wantDiscoveryItems[i].Macros, item.Macros)
				assert.Equal(t, tt.wantDiscoveryItems[i].Steps, item.Steps)
			}
		})
	}
}
