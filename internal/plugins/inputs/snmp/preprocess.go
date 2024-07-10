// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmp

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/gosnmp/gosnmp"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp/snmputil"
)

const (
	sysNameOID   = "1.3.6.1.2.1.1.5.0"
	sysObjectOID = "1.3.6.1.2.1.1.2.0"
)

func (di *deviceInfo) preProcess() {
	var z snmputil.Session
	if di.Session == z {
		return
	}

	err := di.Session.Connect()
	if err != nil {
		l.Errorf("Connect failed: err = (%v), ip = (%s)", err, di.IP)
		return
	}
	defer func() {
		err := di.Session.Close()
		if err != nil {
			l.Warnf("failed to close session: err = (%v), ip = (%s)", err, di.IP)
		}
	}()

	// Check if the device is reachable
	_, err = di.Session.GetNext([]string{snmputil.DeviceReachableGetNextOid})
	if err != nil {
		l.Errorf("check %s device reachable: failed: %v", di.IP, err)
		return
	}

	// fill Name SysName SysObjectID
	di.UserProfileDefinition.Name, di.UserProfileDefinition.SysName, di.UserProfileDefinition.SysObjectID = di.getSysInfo()

	// fill items discoveryItems and macros
	if di.UserProfileDefinition.ZabbixExport != nil {
		di.preProcessZabbix()
	}
}

func (di *deviceInfo) getSysInfo() (string, string, string) {
	sysName := di.getStringValue(sysNameOID)
	sysObjectID := di.getStringValue(sysObjectOID)
	name := di.IP + "_" + sysName

	return name, sysName, sysObjectID
}

func (di *deviceInfo) getStringValue(oid string) string {
	snmpPacket, err := di.Session.Get([]string{oid})
	if err != nil {
		l.Debugf("get snmp: failed: %v", err)
		return "unknown"
	}

	if s, err := assertSNMPPacketString(snmpPacket); err == nil {
		return s
	}

	return "unknown"
}

func (di *deviceInfo) preProcessZabbix() {
	for _, host := range di.UserProfileDefinition.ZabbixExport.Hosts {
		items, stringItems := di.fillItems(host.Items)
		di.UserProfileDefinition.Items = append(di.UserProfileDefinition.Items, items...)
		di.UserProfileDefinition.StringItems = append(di.UserProfileDefinition.StringItems, stringItems...)
	}

	for _, template := range di.UserProfileDefinition.ZabbixExport.Templates {
		items, stringItems := di.fillItems(template.Items)
		di.UserProfileDefinition.Items = append(di.UserProfileDefinition.Items, items...)
		di.UserProfileDefinition.StringItems = append(di.UserProfileDefinition.StringItems, stringItems...)

		for _, discoveryRule := range template.DiscoveryRules {
			// fill macro
			m := formatMacroNames(discoveryRule.SnmpOID)
			macros := di.getMacros(m)

			if len(discoveryRule.Filter.Conditions) > 0 {
				macros = macrosFilter(macros, discoveryRule.Filter)
			}

			items := di.fillDiscoveryItems(discoveryRule.ItemPrototypes, macros)
			di.UserProfileDefinition.DiscoveryItems = append(di.UserProfileDefinition.DiscoveryItems, items...)
		}
	}
}

func (di *deviceInfo) getMacros(m map[string]string) map[string]snmputil.Macro {
	result := make(map[string]snmputil.Macro)

	for k, v := range m {
		if macro := di.getMacro(v); len(macro) != 0 {
			result[k] = macro
		}
	}

	return result
}

func (di *deviceInfo) getMacro(oid string) map[string]string {
	result := make(map[string]string)

	snmpPDUs, err := di.Session.GetWalkAll(oid)
	if err != nil {
		l.Debugf("get snmp: failed: %v", err)
		return result
	}

	if snmpPDUs == nil {
		return result
	}

	var idx int
	for _, snmpPDU := range snmpPDUs {
		if idx = strings.LastIndex(snmpPDU.Name, "."); idx == -1 {
			continue
		}
		snmpIdx := snmpPDU.Name[idx+1:]

		if value, err := assertString(snmpPDU.Value); err != nil {
			l.Debugf("assert snmpPDU.Value: %v to string : failed: %v", snmpPDU.Value, err)
		} else {
			result[snmpIdx] = value
		}
	}

	return result
}

func (di *deviceInfo) fillItems(items []*snmputil.Item) (its, stringIts []*snmputil.Item) {
	for _, item := range items {
		if item.SnmpOID == "" || item.Key == "" {
			continue
		}

		item.Key = di.filterFieldName(formatKey(item.Key), item.SnmpOID)
		item.SnmpOID = FormatOID(item.SnmpOID)

		snmpPacket, err := di.Session.Get([]string{item.SnmpOID})
		if err != nil {
			l.Debugf("get snmp: failed: %v", err)
			continue
		}

		preprocessing := getPreprocessing(item)
		item.Preprocessing = preprocessing

		_, _, isFloat, err := assertSNMPPacket(snmpPacket)
		if err != nil {
			continue
		}

		if isFloat {
			its = append(its, item)
		} else {
			stringIts = append(stringIts, item)
		}
	}
	return
}

func (di *deviceInfo) fillDiscoveryItems(items []*snmputil.Item, macros map[string]snmputil.Macro) []*snmputil.Item {
	var its []*snmputil.Item

	for _, item := range items {
		its = append(its, di.formatDiscoveryItems(item, macros)...)
	}

	return its
}

func (di *deviceInfo) formatDiscoveryItems(item *snmputil.Item, macros map[string]snmputil.Macro) []*snmputil.Item {
	if item.SnmpOID == "" {
		return nil
	}

	// ignore like '{#IFALIAS}' '{#IFNAME}'
	tags := make([]*snmputil.Tag, 0)
	for _, tag := range item.Tags {
		if strings.HasPrefix(tag.Value, "{#") {
			continue
		}
		tags = append(tags, tag)
	}
	item.Tags = tags

	// the info from m will into tags
	item.Macros = make(map[string]snmputil.Macro)
	oid, _ := getOID(item.SnmpOID)
	if oid == "" {
		return nil
	}
	item.SnmpOID = oid
	item.Key = di.filterFieldName(formatKey(item.Key), item.SnmpOID)
	item.Macros = macros

	snmpPDUs, err := di.Session.GetWalkAll(item.SnmpOID)
	if err != nil {
		l.Debugf("get snmp: failed: %v", err)
		return []*snmputil.Item{}
	}
	if len(snmpPDUs) == 0 {
		return []*snmputil.Item{}
	}

	// IF no OIDs in .yaml or .xml file.
	if len(item.OIDs) == 0 {
		for _, snmpPDU := range snmpPDUs {
			newOID := FormatOID(snmpPDU.Name)
			if validOID(newOID, macros) {
				item.OIDs = append(item.OIDs, newOID)
			}
		}
	}

	// will collect no metric, so ignore this item
	if len(item.OIDs) == 0 {
		return nil
	}

	preprocessing := getPreprocessing(item)
	item.Preprocessing = preprocessing

	return []*snmputil.Item{item}
}

// example: 1.3.6.1.2.1.2.2.1.13.{#SNMPINDEX}
func getOID(str string) (string, []string) {
	strs := strings.Split(str, "{")
	oid := strings.Trim(strs[0], ".")

	var result []string
	for i := 1; i < len(strs); i++ {
		s := strs[i]
		if idx := strings.Index(s, "}"); idx > 0 {
			result = append(result, strings.TrimPrefix(s[:idx], "#"))
		}
	}

	return oid, result
}

type reg struct {
	macroKey     string
	regular      *regexp.Regexp
	matchesRegex bool
}

// Delete some macroName according discovery_rules.filter.conditions
//
// filter example:
//
//		filter:
//		  evaltype: AND
//		  conditions:
//		    - macro: '{#IFOPERSTATUS}'
//		      value: '1'
//		      formulaid: A
//		    - macro: '{#SNMPVALUE}'
//		      value: (2|3)
//		      formulaid: B
//
//		filter:
//		  conditions:
//		    - macro: '{#ENT_NAME}'
//		      value: 'MPU.*'
//		      formulaid: A
//	          operator: NOT_MATCHES_REGEX
//
// macros example:
// "ENT_NAME": {"1":"SRU Board 0","2":"GigabitEthernet0/0/1"}, "IFALIAS": {"1":"wireless-zhuyun-2F","2":"wireless-zhuyun-4F"}.
func macrosFilter(macros map[string]snmputil.Macro, filter snmputil.Filter) map[string]snmputil.Macro {
	if len(filter.Conditions) == 0 {
		return macros
	}

	regs := make([]reg, 0)
	for _, condition := range filter.Conditions {
		regular, err := getRegular(condition)
		if err != nil {
			continue
		}
		regs = append(regs, reg{
			macroKey:     macroKey(condition.Macro),
			regular:      regular,
			matchesRegex: condition.Operator != "NOT_MATCHES_REGEX",
		})
	}

	var goodKeys map[string]bool
	if filter.EvalType == "OR" {
		goodKeys = doFilterAsOr(macros, regs)
	} else {
		goodKeys = doFilterAsAnd(macros, regs)
	}

	return doFilter(macros, goodKeys)
}

func doFilterAsOr(macros map[string]snmputil.Macro, regs []reg) map[string]bool {
	goodKeys := make(map[string]bool)
	for _, reg := range regs {
		m, ok := macros[reg.macroKey]
		if !ok {
			continue
		}

		for k, v := range m {
			if reg.regular.MatchString(v) == reg.matchesRegex {
				goodKeys[k] = true
			}
		}
	}

	return goodKeys
}

func doFilterAsAnd(macros map[string]snmputil.Macro, regs []reg) map[string]bool {
	goodKeys := make(map[string]bool)
	for _, macro := range macros {
		for k := range macro {
			goodKeys[k] = true
		}
		break
	}

	for _, reg := range regs {
		m, ok := macros[reg.macroKey]
		if !ok {
			continue
		}

		for k, v := range m {
			if reg.regular.MatchString(v) != reg.matchesRegex {
				delete(goodKeys, k)
			}
		}
	}

	return goodKeys
}

func doFilter(macros map[string]snmputil.Macro, goodKeys map[string]bool) map[string]snmputil.Macro {
	result := cloneMapFramework(macros)

	for macroName, macro := range macros {
		for k, v := range macro {
			if goodKeys[k] {
				result[macroName][k] = v
			}
		}
	}

	return result
}

func cloneMapFramework(macros map[string]snmputil.Macro) map[string]snmputil.Macro {
	maps := make(map[string]snmputil.Macro)
	for macroKey := range macros {
		m := make(map[string]string)

		maps[macroKey] = m
	}
	return maps
}

// example:  {#ENT_NAME} -> ENT_NAME
func macroKey(s string) string {
	s = strings.Trim(s, "{")
	s = strings.Trim(s, "}")
	return strings.Trim(s, "#")
}

func getRegular(condition snmputil.Condition) (*regexp.Regexp, error) {
	if err := checkRegular(condition.Value); err != nil {
		return nil, err
	}

	re, err := regexp.Compile(condition.Value)
	if err != nil {
		l.Errorf("create regular fail, value: %s, err: %w", condition.Value, err)
		return nil, fmt.Errorf("create regular fail, value: %s, err: %w", condition.Value, err)
	}
	return re, nil
}

// example: "1" (2|3) 'MPU.*' '^power (T|R)x' '@ASR9k Optical Levels'
func checkRegular(s string) error {
	// ignore like '@ASR9k Optical Levels' '{$NET.IF.IFTYPE.MATCHES}'
	if strings.HasPrefix(s, "@") || strings.HasPrefix(s, "{$") {
		l.Debugf("create regular fail, value: %s, ignored", s)
		return fmt.Errorf("create regular fail, value: %s, ignored", s)
	}
	return nil
}

// oid's suffix need in any m. as a key.
//
// example:
// 1.3.6.1.2.1.2.2.1.13.1 valid {"1":"XXX","2":"XXX"} -> true
// 1.3.6.1.2.1.2.2.1.13.3 valid {"1":"XXX","2":"XXX"} -> false.
func validOID(oid string, m map[string]snmputil.Macro) bool {
	idx := strings.LastIndex(oid, ".")
	if idx == -1 {
		return false
	}
	key := oid[idx+1:]

	for _, item := range m {
		if _, ok := item[key]; ok {
			return true
		}
	}

	return false
}

// example:
// - system.objectid[sysObjectID.0]
// - system.hw.model[entPhysicalDescr.{#SNMPINDEX}]
// - vm.memory.pused[vm.memory.pused.{#SNMPINDEX}]
// - system.objectid
// - system.uptime[sysUpTime].
func formatKey(s string) string {
	if idx := strings.Index(s, "["); idx > -1 {
		return formatName(s[:idx])
	}
	if idx := strings.Index(s, "{"); idx > -1 {
		return formatName(s[:idx])
	}
	return formatName(s)
}

func formatName(s string) string {
	s = strings.ReplaceAll(s, "#", "")
	s = strings.ReplaceAll(s, ":", "")
	s = strings.ReplaceAll(s, "(", "")
	s = strings.ReplaceAll(s, ")", "")
	s = strings.ReplaceAll(s, "[", "")
	s = strings.ReplaceAll(s, "]", "")
	s = strings.ReplaceAll(s, "{", "")
	s = strings.ReplaceAll(s, "}", "")
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, ".", " ")
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "   ", " ")
	s = strings.ReplaceAll(s, "  ", " ")
	s = strings.ReplaceAll(s, " ", "_")
	return strings.ToLower(s)
}

func FormatOID(s string) string {
	s = strings.Trim(s, ".")
	return s
}

func assertSNMPPacketString(snmpPacket *gosnmp.SnmpPacket) (string, error) {
	if snmpPacket == nil {
		return "", fmt.Errorf("snmpPacket is nil")
	}

	if len(snmpPacket.Variables) < 1 {
		return "", fmt.Errorf("snmpPacket.Variables is nil")
	}

	return assertString(snmpPacket.Variables[0].Value)
}

func assertString(value any) (string, error) {
	if v, ok := value.([]byte); ok {
		return string(v), nil
	} else if v, ok := value.(string); ok {
		return v, nil
	}
	return fmt.Sprintf("%v", value), nil
}

func assertSNMPPacket(snmpPacket *gosnmp.SnmpPacket) (float64, string, bool, error) {
	if snmpPacket == nil {
		return 0, "", false, fmt.Errorf("snmpPacket is nil")
	}

	if len(snmpPacket.Variables) < 1 {
		return 0, "", false, fmt.Errorf("snmpPacket.Variables is nil")
	}

	value, err := snmputil.GetValueFromPDU(snmpPacket.Variables[0])
	if err != nil {
		l.Debugf("assertSNMPPacket fail, oid:%s, type:%s", snmpPacket.Variables[0].Name, snmpPacket.Variables[0].Type.String())
		return 0, "", false, fmt.Errorf("assertSNMPPacket fail, oid:%s, type:%s", snmpPacket.Variables[0].Name, snmpPacket.Variables[0].Type.String())
	}

	if v, ok := value.(float64); ok {
		return v, "", true, nil
	} else if v, ok := value.(string); ok {
		return 0, v, false, nil
	} else if v, ok := value.([]byte); ok {
		return 0, string(v), false, nil
	}

	l.Debugf("assertSNMPPacket fail, oid:%s, type:%s", snmpPacket.Variables[0].Name, snmpPacket.Variables[0].Type.String())
	return 0, "", false, fmt.Errorf("assertSNMPPacket fail, oid:%s, type:%s", snmpPacket.Variables[0].Name, snmpPacket.Variables[0].Type.String())
}

// nolint:lll
// example: 'discovery[{#SNMPVALUE},1.3.6.1.2.1.10.7.2.1.19,{#IFOPERSTATUS},1.3.6.1.2.1.2.2.1.8,{#IFALIAS},1.3.6.1.2.1.31.1.1.1.18,{#IFNAME},1.3.6.1.2.1.31.1.1.1.1,{#IFDESCR},1.3.6.1.2.1.2.2.1.2]'
func formatMacroNames(str string) map[string]string {
	result := make(map[string]string)

	str = strings.TrimPrefix(str, "discovery[")
	str = strings.TrimSuffix(str, "]")
	strs := strings.Split(str, ",")

	// every 2 string be a pair
	for i := 0; i < len(strs)-1; i += 2 {
		k := strings.TrimPrefix(strs[i], "{")
		k = strings.TrimSuffix(k, "}")
		k = strings.TrimPrefix(k, "#")
		v := strings.Trim(strs[i+1], ".")
		result[k] = v
	}

	return result
}

var wantPreprocessing = map[string]func(x, y float64) float64{
	"MULTIPLIER": multiplier,
}

func multiplier(x, y float64) float64 {
	return x * y
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
func getPreprocessing(item *snmputil.Item) snmputil.Preprocessing {
	result := snmputil.Preprocessing{}

	for _, process := range item.Preprocessing.Steps {
		if _, ok := wantPreprocessing[process.Type]; !ok {
			continue
		}
		if len(process.Parameters) == 0 {
			continue
		}

		value, err := strconv.ParseFloat(process.Parameters[0], 64)
		if err != nil {
			continue
		}

		result.Steps = append(result.Steps, &snmputil.Step{
			Type:      process.Type,
			Parameter: value,
		})
		return result
	}

	return result
}
