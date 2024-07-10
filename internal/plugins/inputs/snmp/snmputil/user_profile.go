// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmputil

import (
	"encoding/xml"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp/snmprefiles"
	"gopkg.in/yaml.v2"
)

type Profile interface {
	ReadProfile() *ProfileStore
	Preprocessing()
}

type ZabbixProfile struct {
	ProfileName string   `toml:"profile_name"`
	IPList      []string `toml:"ip_list"`
	Class       string   `toml:"class"`
}

func (p *ZabbixProfile) ReadProfile() (*ProfileStore, error) {
	profilePath := getProfilePath(p.ProfileName)
	store, err := readUserProfileZabbix(profilePath)
	if err != nil {
		l.Errorf("failed to read profile store `%s`: %w", profilePath, err)
		return nil, fmt.Errorf("failed to read profile store `%s`: %w", profilePath, err)
	}

	store.Definition.ZabbixExport.formatOIDs()
	store.IPList = p.IPList
	store.Definition.ProfileName = p.ProfileName
	store.Definition.Class = p.Class
	store.Definition.DeviceType = "unknown"

	return store, nil
}
func (p *ZabbixProfile) Preprocessing() {}

type PromProfile struct {
	ProfileName string   `toml:"profile_name"`
	IPList      []string `toml:"ip_list"`
	Class       string   `toml:"class"`
}

func (p *PromProfile) ReadProfile() ([]*ProfileStore, error) {
	profilePath := getProfilePath(p.ProfileName)
	stores, err := readPromProfile(profilePath)
	if err != nil {
		l.Errorf("failed to read profile store `%s`: %w", profilePath, err)
		return nil, fmt.Errorf("failed to read profile store `%s`: %w", profilePath, err)
	}

	if len(stores) == 1 {
		stores[0].IPList = p.IPList
	}

	for _, store := range stores {
		store.Definition.ProfileName = p.ProfileName
		store.Definition.Class = p.Class
	}

	return stores, nil
}

type DatadogProfile struct {
	ProfileName string   `toml:"profile_name"`
	IPList      []string `toml:"ip_list"`
	Class       string   `toml:"class"`
}

func (p *DatadogProfile) ReadProfile() (*ProfileStore, error) { return nil, fmt.Errorf("") }
func (p *DatadogProfile) Preprocessing()                      {}

type UserProfileDefinition struct {
	ZabbixExport   *ZabbixExport `yaml:"zabbix_export" xml:"zabbix_export"`
	StringItems    []*Item       // Result is string, example 1.3.6.1.2.1.1.4.0 -> system.contact
	Items          []*Item
	DiscoveryItems []*Item
	ProfileName    string            `yaml:"profile_name"`
	Class          string            `yaml:"class"`         // "sever" "printer" "unknown" ...
	DeviceType     string            `yaml:"device_type"`   // for prom. vpn4, apcups, cisco_wlc ...
	Community      string            `yaml:"community"`     // for prom.
	Name           string            `yaml:"name"`          // ip_"SysName" ...
	SysName        string            `yaml:"sys_name"`      // "iDRAC-4XF4BX2" "unknown"
	SysObjectID    string            `yaml:"sys_object_id"` // "1.3.6.1.4.1.2011.2.240.12"
	InputTags      map[string]string `yaml:"input_tags"`
}

type ZabbixExport struct {
	XMLName   xml.Name    `xml:"zabbix_export"`
	Version   string      `yaml:"version" xml:"version"`
	Date      time.Time   `yaml:"date" xml:"date"`
	Groups    []*Group    `yaml:"groups" xml:"groups>group"`
	Templates []*Template `yaml:"templates" xml:"templates>template"`
	Hosts     []*Host     `yaml:"hosts" xml:"hosts>host"`
}

type Group struct {
	UUID string `yaml:"uuid" xml:"uuid"`
	Name string `yaml:"name" xml:"name"`
}

type Template struct {
	UUID           string           `yaml:"uuid" xml:"uuid"`
	Template       string           `yaml:"template" xml:"template"`
	Name           string           `yaml:"name" xml:"name"`
	Description    string           `yaml:"description" xml:"description"`
	Groups         []*Group         `yaml:"groups" xml:"groups>group"`
	Items          []*Item          `yaml:"items" xml:"items>item"`
	DiscoveryRules []*DiscoveryRule `yaml:"discovery_rules" xml:"discovery_rules>discovery_rule"`
}

type Item struct {
	UUID          string        `yaml:"uuid" xml:"uuid"`
	Name          string        `yaml:"name" xml:"name"`
	Type          string        `yaml:"type" xml:"type"`
	SnmpOID       string        `yaml:"snmp_oid" xml:"snmp_oid"`
	Key           string        `yaml:"key" xml:"key"`
	History       string        `yaml:"history" xml:"history"`
	ValueType     string        `yaml:"value_type" xml:"value_type"`
	Units         string        `yaml:"units" xml:"units"`
	Preprocessing Preprocessing `yaml:"clone_preprocessing" xml:"preprocessing"`
	Steps         []*Step       `yaml:"preprocessing"` // only for zabbix6.0.yaml unmarshal, will clone to Preprocessing
	Tags          []*Tag        `yaml:"tags" xml:"tags>tag"`
	Triggers      []*Trigger    `yaml:"triggers" xml:"triggers>trigger"`

	// Only for DiscoveryRule item.
	// If nil in zabbix.yaml, will fill all odis through walk SnmpOID.
	// If not nil in zabbix.yaml, only get metric that in odis.
	// example SnmpOID+".1" , SnmpOID+".2" ...
	OIDs []string `yaml:"oids" xml:"oids>oid"`

	Macros map[string]Macro // example "SNMPVALUE" -> {"1" : "CPU1 Status","2" : "CPU2 Status"}
}

type Macro map[string]string

type Preprocessing struct {
	Steps []*Step `yaml:"steps" xml:"step"`
}

type Step struct {
	Type       string   `yaml:"type" xml:"type"`                       // example "MULTIPLIER"/"DISCARD_UNCHANGED_HEARTBEAT"
	Parameters []string `yaml:"parameters" xml:"parameters>parameter"` // example "0.1"/6h
	Params     []string `yaml:"_" xml:"params"`                        // only for zabbix5.0.xlm unmarshal, will clone to Parameters
	Parameter  float64  `yaml:"-" xml:"-"`                             // float of Parameters[0], example 0.1/1048576
}

type Tag struct {
	Tag   string `yaml:"tag" xml:"tag"`
	Value string `yaml:"value" xml:"value"`
}

type Trigger struct {
	UUID        string `yaml:"uuid" xml:"uuid"`
	Expression  string `yaml:"expression" xml:"expression"`
	Name        string `yaml:"name" xml:"name"`
	Priority    string `yaml:"priority" xml:"priority"`
	Description string `yaml:"description" xml:"description"`
	Tags        []*Tag `yaml:"tags" xml:"tags>tag"`
}

type Filter struct {
	EvalType   string      `yaml:"evaltype" xml:"evaltype"`
	Conditions []Condition `xml:"conditions>condition"`
}

type Condition struct {
	Macro     string `yaml:"macro" xml:"macro"`
	Value     string `yaml:"value" xml:"value"` //
	Formulaid string `yaml:"formulaid" xml:"formulaid"`
	Operator  string `yaml:"operator" xml:"operator"` // NOT_MATCHES/REGEX REGEXP/nil
}

type DiscoveryRule struct {
	UUID           string  `yaml:"uuid" xml:"uuid"`
	Name           string  `yaml:"name" xml:"name"`
	Type           string  `yaml:"type" xml:"type"`
	SnmpOID        string  `yaml:"snmp_oid" xml:"snmp_oid"`
	Key            string  `yaml:"key" xml:"key"`
	Delay          string  `yaml:"delay" xml:"delay"`
	Filter         Filter  `yaml:"filter" xml:"filter"`
	ItemPrototypes []*Item `yaml:"item_prototypes" xml:"item_prototypes>item_prototype"`
}

type Host struct {
	Host        string      `yaml:"host" xml:"host"`
	Name        string      `yaml:"name" xml:"name"`
	Templates   []*Template `yaml:"templates" xml:"templates>template"`
	Description string      `yaml:"description" xml:"description"`
	Groups      []*Group    `yaml:"groups" xml:"groups>group"`
	Items       []*Item     `yaml:"items" xml:"items>item"`
}

// prom profile

type SNMPAuth struct {
	Community string `yaml:"community"`
}

type SNMPMetrics struct {
	Name       string            `yaml:"name"`
	Oid        string            `yaml:"oid"`
	Type       string            `yaml:"type"`
	Help       string            `yaml:"help"`
	Indexes    []*Index          `yaml:"indexes"`
	Lookups    []*Lookup         `yaml:"lookups"`
	EnumValues map[string]string `yaml:"enum_values"`
}

type Index struct {
	Labelname string `yaml:"labelname"`
	Type      string `yaml:"type"`
}

type Lookup struct {
	Labels    []string `yaml:"labels"`
	Labelname string   `yaml:"labelname"`
	Oid       string   `yaml:"oid"`
	Type      string   `yaml:"type"`
}

type PromModule struct {
	Walk    []string      `yaml:"walk"`
	Get     []string      `yaml:"get"`
	Metrics []SNMPMetrics `yaml:"metrics"`
	Retries int           `yaml:"retries"`
	Timeout string        `yaml:"timeout"`
	Auth    SNMPAuth      `yaml:"auth"`
}

type PromAuth struct {
	Community     string `yaml:"community"`
	SecurityLevel string `yaml:"security_level"`
	AuthProtocol  string `yaml:"auth_protocol"`
	PrivProtocol  string `yaml:"priv_protocol"`
	Version       string `yaml:"version"`
}

type PromConfig struct {
	Auths   map[string]PromAuth   `yaml:"auths"`
	Modules map[string]PromModule `yaml:"modules"`
}

func NewPromConfig() *PromConfig {
	return &PromConfig{
		Auths:   make(map[string]PromAuth),
		Modules: make(map[string]PromModule),
	}
}

// all will clone to Item.Preprocessing.Steps.Parameters.
//
// example
//
// preprocessing:
//   - type: MULTIPLIER
//     parameters:
//   - '0.1'
//   - type: DISCARD_UNCHANGED_HEARTBEAT
//     parameters:
//   - 6h
//
// example: "MULTIPLIER" -> 0.1
//
// zabbix6.0.yaml in Item.Steps.Parameters
// zabbix6.0.xml in Item.Preprocessing.Steps.Parameters
// zabbix5.0.xml in Item.Preprocessing.Steps.Params.
func (z *ZabbixExport) cloneXML50Parameters2XML60() {
	if z == nil {
		return
	}

	for _, host := range z.Hosts {
		for _, item := range host.Items {
			for _, step := range item.Preprocessing.Steps {
				step.Parameters = append(step.Parameters, step.Params...)
			}
		}
	}

	for _, template := range z.Templates {
		for _, item := range template.Items {
			for _, step := range item.Preprocessing.Steps {
				step.Parameters = append(step.Parameters, step.Params...)
			}
		}

		for _, discovery := range template.DiscoveryRules {
			for _, item := range discovery.ItemPrototypes {
				for _, step := range item.Preprocessing.Steps {
					step.Parameters = append(step.Parameters, step.Params...)
				}
			}
		}
	}
}

// all will clone to Item.Preprocessing.Steps.Parameters.
//
// example
//
// preprocessing:
//   - type: MULTIPLIER
//     parameters:
//   - '0.1'
//   - type: DISCARD_UNCHANGED_HEARTBEAT
//     parameters:
//   - 6h
//
// example: "MULTIPLIER" -> 0.1
//
// zabbix6.0.yaml in Item.Steps.Parameters
// zabbix6.0.xml in Item.Preprocessing.Steps.Parameters
// zabbix5.0.xml in Item.Preprocessing.Steps.Params.
func (z *ZabbixExport) cloneYamlParameters2XML60() {
	if z == nil {
		return
	}

	for _, host := range z.Hosts {
		for _, item := range host.Items {
			item.Preprocessing.Steps = item.Steps
		}
	}

	for _, template := range z.Templates {
		for _, item := range template.Items {
			item.Preprocessing.Steps = item.Steps
		}

		for _, discovery := range template.DiscoveryRules {
			for _, item := range discovery.ItemPrototypes {
				item.Preprocessing.Steps = item.Steps
			}
		}
	}
}

// example: ".1.3.6.1.4.1.18334.1.1" -> "1.3.6.1.4.1.18334.1.1"
func (z *ZabbixExport) formatOIDs() {
	if z == nil {
		return
	}

	for _, host := range z.Hosts {
		for _, item := range host.Items {
			item.SnmpOID = formatOID(item.SnmpOID)
		}
	}

	for _, template := range z.Templates {
		for _, item := range template.Items {
			item.SnmpOID = formatOID(item.SnmpOID)
		}

		for _, discovery := range template.DiscoveryRules {
			for _, item := range discovery.ItemPrototypes {
				item.SnmpOID = formatOID(item.SnmpOID)
			}
		}
	}
}

func formatOID(s string) string {
	s = strings.Trim(s, ".")
	return s
}

func NewUserProfileDefinition() *UserProfileDefinition {
	return &UserProfileDefinition{
		InputTags: make(map[string]string),
	}
}

func (d *UserProfileDefinition) Clone() *UserProfileDefinition {
	userProfileDefinition := NewUserProfileDefinition()

	buf, err := yaml.Marshal(d)
	if err != nil {
		l.Errorf("clone user profile definition marshal fail : %v", err)
		return nil
	}

	if err := yaml.Unmarshal(buf, userProfileDefinition); err != nil {
		l.Errorf("clone user profile definition unmarshal fail : %v", err)
		return nil
	}

	return userProfileDefinition
}

func newZabbixExport() *ZabbixExport {
	return &ZabbixExport{}
}

type UserProfileStore struct {
	ZabbixStores  []*ProfileStore
	DatadogStores []*ProfileStore
}

func NewUserProfileStore() *UserProfileStore {
	return &UserProfileStore{}
}

type ProfileStore struct {
	FileName   string
	IPList     []string
	Definition *UserProfileDefinition
}

func NewProfileStore() *ProfileStore {
	return &ProfileStore{
		Definition: &UserProfileDefinition{},
	}
}

func LoadUserProfiles(
	zabbixProfiles []*ZabbixProfile,
	promProfiles []*PromProfile,
	datadogProfiles []*DatadogProfile,
	inputTags map[string]string,
) *UserProfileStore {
	userProfileStore := NewUserProfileStore()

	for _, profile := range zabbixProfiles {
		if store, err := profile.ReadProfile(); err == nil {
			userProfileStore.ZabbixStores = append(userProfileStore.ZabbixStores, store)
		}
	}

	for _, profile := range promProfiles {
		if stores, err := profile.ReadProfile(); err == nil {
			userProfileStore.ZabbixStores = append(userProfileStore.ZabbixStores, stores...)
		}
	}

	return userProfileStore
}

func readUserProfileZabbix(profilePath string) (*ProfileStore, error) {
	fileSuffix := path.Ext(profilePath)
	filePath := resolveProfileDefinitionPath(profilePath)
	buf, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return nil, fmt.Errorf("failed to read file `%s`: %w", filePath, err)
	}

	profileStore := NewProfileStore()
	profileStore.FileName = profilePath

	switch {
	case fileSuffix == ".yaml" || fileSuffix == ".yml":
		err = yaml.Unmarshal(buf, profileStore.Definition)
		profileStore.Definition.ZabbixExport.cloneYamlParameters2XML60()
	case fileSuffix == ".xml":
		zabbixExport := newZabbixExport()
		err = xml.Unmarshal(buf, zabbixExport)
		zabbixExport.cloneXML50Parameters2XML60()
		profileStore.Definition.ZabbixExport = zabbixExport
	default:
		return nil, fmt.Errorf("wrong file suffix: `%s`", filePath)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal %q: %w", filePath, err)
	}

	return profileStore, nil
}

func readPromProfile(profilePath string) ([]*ProfileStore, error) {
	filePath := resolveProfileDefinitionPath(profilePath)
	buf, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return nil, fmt.Errorf("failed to read file `%s`: %w", filePath, err)
	}

	promConfig := NewPromConfig()

	// try prom snmp_exporter config
	err = yaml.Unmarshal(buf, promConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal %q: %w", filePath, err)
	}

	// try prom snmp_exporter config only have modules
	if len(promConfig.Modules) == 0 {
		err = yaml.Unmarshal(buf, promConfig.Modules)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal %q: %w", filePath, err)
		}
	}

	stores := make([]*ProfileStore, 0)
	for deviceType, module := range promConfig.Modules {
		store := NewProfileStore()
		store.FileName = profilePath
		store.Definition.Community = module.Auth.Community
		store.Definition.DeviceType = deviceType

		discoveryRules := []*DiscoveryRule{}
		for idx := range module.Metrics {
			discoveryRule := getDiscoveryRule(idx, module)
			idx := getIndex(discoveryRule, discoveryRules)
			if idx == -1 {
				discoveryRules = append(discoveryRules, discoveryRule)
			} else {
				discoveryRules[idx].ItemPrototypes = append(discoveryRules[idx].ItemPrototypes, discoveryRule.ItemPrototypes...)
			}
		}

		store.Definition.ZabbixExport = &ZabbixExport{
			Templates: []*Template{{
				DiscoveryRules: discoveryRules,
			}},
		}
		store.Definition.ZabbixExport.formatOIDs()

		stores = append(stores, store)
	}

	return stores, nil
}

func getIndex(discoveryRule *DiscoveryRule, discoveryRules []*DiscoveryRule) int {
	idx := -1

	for i, rule := range discoveryRules {
		if discoveryRule.SnmpOID == rule.SnmpOID {
			return i
		}
	}

	return idx
}

func getDiscoveryRule(idx int, module PromModule) *DiscoveryRule {
	snmpMetric := module.Metrics[idx]
	discoveryRule := &DiscoveryRule{}

	// init Item
	itemPrototype := &Item{
		Key:     snmpMetric.Name, // metric field name
		SnmpOID: snmpMetric.Oid,  // walk oid
		OIDs:    getOIDs(snmpMetric.Oid, module.Get),
	}
	discoveryRule.ItemPrototypes = append(discoveryRule.ItemPrototypes, itemPrototype)

	// example: {MACRO0},1.3.6.1.2.1.31.1.1.1.18,{MACRO1},1.3.6.1.2.1.2.2.1.2
	if len(snmpMetric.Lookups) == 0 {
		for i, v := range module.Walk {
			lookup := &Lookup{
				Labelname: "MACRO" + fmt.Sprint(i),
				Oid:       v,
			}
			snmpMetric.Lookups = append(snmpMetric.Lookups, lookup)
		}
	}

	// init discoveryRule
	// example: {ifAlias},1.3.6.1.2.1.31.1.1.1.18,{ifDescr},1.3.6.1.2.1.2.2.1.2
	discoveryRuleSnmpOID := ""
	for i, lookup := range snmpMetric.Lookups {
		if i > 0 {
			discoveryRuleSnmpOID += ","
		}
		discoveryRuleSnmpOID += "{" + lookup.Labelname + "}," + lookup.Oid
	}
	discoveryRule.SnmpOID = discoveryRuleSnmpOID

	return discoveryRule
}

func getOIDs(oid string, oids []string) []string {
	result := []string{}
	for _, v := range oids {
		if strings.HasPrefix(v, oid) {
			result = append(result, v)
		}
	}

	return result
}

func getProfilePath(profile string) string {
	baseName := filepath.Base(profile)
	if baseName != profile {
		return profile
	}
	userProfilesRoot := snmprefiles.GetUserProfilesRoot()
	return filepath.Join(userProfilesRoot, profile)
}
