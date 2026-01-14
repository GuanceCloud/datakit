// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmputil

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
)

// =============================================================================

// config

// DefaultBulkMaxRepetitions is the default max rep
// Using too high max repetitions might lead to tooBig SNMP error messages.
// - Java SNMP and gosnmp (gosnmp.defaultMaxRepetitions) uses 50
// - snmp-net uses 10.
const DefaultBulkMaxRepetitions = uint32(10)

var UptimeMetricConfig = MetricsConfig{Symbol: SymbolConfig{OID: "1.3.6.1.2.1.1.3.0", Name: "sysUpTimeInstance"}}

func ParseScalarOids(metrics []MetricsConfig, metricTags []MetricTagConfig, metadataConfigs MetadataConfig, collectDeviceMetadata bool) []string {
	var oids []string
	for _, metric := range metrics {
		oids = append(oids, metric.Symbol.OID)
	}
	for _, metricTag := range metricTags {
		oids = append(oids, metricTag.Symbol.OID)
	}
	if collectDeviceMetadata {
		for resource, metadataConfig := range metadataConfigs {
			if !IsMetadataResourceWithScalarOids(resource) {
				continue
			}
			for _, field := range metadataConfig.Fields {
				oids = append(oids, field.Symbol.OID)
				for _, symbol := range field.Symbols {
					oids = append(oids, symbol.OID)
				}
			}
			// we don't support tags for now for resource (e.g. device) based on scalar OIDs
			// profile root level `metric_tags` (tags used for both metadata, metrics, service checks)
			// can be used instead
		}
	}
	newOids := make([]string, 0, len(oids))
	oidsMap := make(map[string]struct{}, len(oids))
	for _, oid := range oids {
		if oid == "" {
			continue
		}
		if _, ok := oidsMap[oid]; ok {
			continue
		}
		oidsMap[oid] = struct{}{}
		newOids = append(newOids, oid)
	}
	return newOids
}

func ParseColumnOids(metrics []MetricsConfig, metadataConfigs MetadataConfig, collectDeviceMetadata bool) []string {
	var oids []string
	for _, metric := range metrics {
		for _, symbol := range metric.Symbols {
			oids = append(oids, symbol.OID)
		}
		for _, metricTag := range metric.MetricTags {
			oids = append(oids, metricTag.Symbol.OID)
		}
	}
	if collectDeviceMetadata {
		for resource, metadataConfig := range metadataConfigs {
			if IsMetadataResourceWithScalarOids(resource) {
				continue
			}
			for _, field := range metadataConfig.Fields {
				oids = append(oids, field.Symbol.OID)
				for _, symbol := range field.Symbols {
					oids = append(oids, symbol.OID)
				}
			}
			for _, tagConfig := range metadataConfig.IDTags {
				oids = append(oids, tagConfig.Symbol.OID)
			}
		}
	}
	newOids := make([]string, 0, len(oids))
	oidsMap := make(map[string]struct{}, len(oids))
	for _, oid := range oids {
		if oid == "" {
			continue
		}
		if _, ok := oidsMap[oid]; ok {
			continue
		}
		oidsMap[oid] = struct{}{}
		newOids = append(newOids, oid)
	}
	return newOids
}

// GetProfileForSysObjectID return a profile for a sys object id.
func GetProfileForSysObjectID(profiles ProfileDefinitionMap, sysObjectID string) (string, error) {
	tmpSysOidToProfile := map[string]string{}
	var matchedOids []string

	for profile, definition := range profiles {
		for _, oidPattern := range definition.SysObjectIds {
			found, err := filepath.Match(oidPattern, sysObjectID)
			if err != nil {
				l.Debugf("pattern error: %v", err)
				continue
			}
			if !found {
				continue
			}
			if matchedProfile, ok := tmpSysOidToProfile[oidPattern]; ok {
				return "", fmt.Errorf("profile %s has the same sysObjectID (%s) as %s", profile, oidPattern, matchedProfile)
			}
			tmpSysOidToProfile[oidPattern] = profile
			matchedOids = append(matchedOids, oidPattern)
		}
	}
	oid, err := getMostSpecificOid(matchedOids)
	if err != nil {
		return "", fmt.Errorf("failed to get most specific profile for sysObjectID `%s`, for matched oids %v: %w", sysObjectID, matchedOids, err)
	}
	return tmpSysOidToProfile[oid], nil
}

// =============================================================================

// config_metadata

// LegacyMetadataConfig contains metadata config used for backward compatibility
// When users have their own copy of _base.yaml and _generic_if.yaml files
// they won't have the new profile based metadata definitions for device and interface resources
// The LegacyMetadataConfig is used as fallback to provide metadata definitions for those resources.
var LegacyMetadataConfig = MetadataConfig{
	"device": {
		Fields: map[string]MetadataField{
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
			"name": {
				Symbol: SymbolConfig{
					OID:  "1.3.6.1.2.1.31.1.1.1.1",
					Name: "ifName",
				},
			},
			"description": {
				Symbol: SymbolConfig{
					OID:  "1.3.6.1.2.1.2.2.1.2",
					Name: "ifDescr",
				},
			},
			"admin_status": {
				Symbol: SymbolConfig{
					OID:  "1.3.6.1.2.1.2.2.1.7",
					Name: "ifAdminStatus",
				},
			},
			"oper_status": {
				Symbol: SymbolConfig{
					OID:  "1.3.6.1.2.1.2.2.1.8",
					Name: "ifOperStatus",
				},
			},
			"alias": {
				Symbol: SymbolConfig{
					OID:  "1.3.6.1.2.1.31.1.1.1.18",
					Name: "ifAlias",
				},
			},
			"mac_address": {
				Symbol: SymbolConfig{
					OID:    "1.3.6.1.2.1.2.2.1.6",
					Name:   "ifPhysAddress",
					Format: "mac_address",
				},
			},
		},
		IDTags: MetricTagConfigList{
			{
				Tag: "interface",
				Symbol: SymbolConfigCompat{
					OID:  "1.3.6.1.2.1.31.1.1.1.1",
					Name: "ifName",
				},
			},
		},
	},
	"ip_addresses": {
		Fields: map[string]MetadataField{
			"if_index": {
				Symbol: SymbolConfig{
					OID:  "1.3.6.1.2.1.4.20.1.2",
					Name: "ipAdEntIfIndex",
				},
			},
			"netmask": {
				Symbol: SymbolConfig{
					OID:  "1.3.6.1.2.1.4.20.1.3",
					Name: "ipAdEntNetMask",
				},
			},
		},
	},
}

// MetadataConfig holds configs per resource type.
type MetadataConfig map[string]MetadataResourceConfig

// MetadataResourceConfig holds configs for a metadata resource.
type MetadataResourceConfig struct {
	Fields map[string]MetadataField `yaml:"fields"`
	IDTags MetricTagConfigList      `yaml:"id_tags"`
}

func (mrc *MetadataResourceConfig) Copy() MetadataResourceConfig {
	return MetadataResourceConfig{
		Fields: CopyMapStringMetadataField(mrc.Fields),
		IDTags: CopyMetricTagConfigs(mrc.IDTags),
	}
}

// MetadataField holds configs for a metadata field.
type MetadataField struct {
	Symbol  SymbolConfig   `yaml:"symbol"`
	Symbols []SymbolConfig `yaml:"symbols"`
	Value   string         `yaml:"value"`
}

func (mf *MetadataField) Copy() MetadataField {
	return MetadataField{
		Symbol:  mf.Symbol.Copy(),
		Symbols: CopySymbolConfigs(mf.Symbols),
		Value:   mf.Value,
	}
}

// newMetadataResourceConfig returns a new metadata resource config.
func newMetadataResourceConfig() MetadataResourceConfig {
	return MetadataResourceConfig{}
}

// IsMetadataResourceWithScalarOids returns true if the resource is based on scalar OIDs
// at the moment, we only expect "device" resource to be based on scalar OIDs.
func IsMetadataResourceWithScalarOids(resource string) bool {
	return resource == MetadataDeviceResource
}

// UpdateMetadataDefinitionWithLegacyFallback will add metadata config for resources
// that does not have metadata definitions.
func UpdateMetadataDefinitionWithLegacyFallback(config MetadataConfig) MetadataConfig {
	if config == nil {
		config = MetadataConfig{}
	}
	for resourceName, resourceConfig := range LegacyMetadataConfig {
		if _, ok := config[resourceName]; !ok {
			config[resourceName] = resourceConfig
		}
	}
	return config
}

// =============================================================================

// config_metric

// SymbolConfig holds info for a single symbol/oid.
type SymbolConfig struct {
	OID  string `yaml:"OID"`
	Name string `yaml:"name"`

	ExtractValue         string `yaml:"extract_value"`
	ExtractValueCompiled *regexp.Regexp

	MatchPattern         string `yaml:"match_pattern"`
	MatchValue           string `yaml:"match_value"`
	MatchPatternCompiled *regexp.Regexp

	ScaleFactor      float64 `yaml:"scale_factor"`
	Format           string  `yaml:"format"`
	ConstantValueOne bool    `yaml:"constant_value_one,omitempty"`

	// `metric_type` is used for force the metric type
	//   When empty, by default, the metric type is derived from SNMP OID value type.
	//   Valid `metric_type` types: `gauge`, `rate`, `monotonic_count`, `monotonic_count_and_rate`
	//   Deprecated types: `counter` (use `rate` instead), percent (use `scale_factor` instead)
	MetricType string `yaml:"metric_type,omitempty"`
}

func (sc *SymbolConfig) Copy() SymbolConfig {
	var extractValueCompiled, matchPatternCompiled *regexp.Regexp
	if sc.ExtractValueCompiled != nil {
		extractValueCompiled = sc.ExtractValueCompiled.Copy() //nolint:staticcheck
	}
	if sc.MatchPatternCompiled != nil {
		matchPatternCompiled = sc.MatchPatternCompiled.Copy() //nolint:staticcheck
	}

	return SymbolConfig{
		OID:                  sc.OID,
		Name:                 sc.Name,
		ExtractValue:         sc.ExtractValue,
		ExtractValueCompiled: extractValueCompiled,
		MatchPattern:         sc.MatchPattern,
		MatchValue:           sc.MatchValue,
		MatchPatternCompiled: matchPatternCompiled,
		ScaleFactor:          sc.ScaleFactor,
		Format:               sc.Format,
		ConstantValueOne:     sc.ConstantValueOne,
		MetricType:           sc.MetricType,
	}
}

// SymbolConfigCompat is used to deserialize string field or SymbolConfig.
// For OID/Name to Symbol harmonization:
// When users declare metric tag like:
//
//	metric_tags:
//	  - OID: 1.2.3
//	    symbol: aSymbol
//
// this will lead to OID stored as MetricTagConfig.OID and name stored as MetricTagConfig.Symbol.Name
// When this happens, in ValidateEnrichMetricTags we harmonize by moving MetricTagConfig.OID to MetricTagConfig.Symbol.OID.
type SymbolConfigCompat SymbolConfig

// Copy creates a duplicate of this SymbolConfigCompat.
func (s SymbolConfigCompat) Copy() SymbolConfigCompat {
	sc := SymbolConfig(s)
	return SymbolConfigCompat((&sc).Copy())
}

// MetricTagConfig holds metric tag info.
type MetricTagConfig struct {
	Tag string `yaml:"tag"`

	// Table config
	Index uint `yaml:"index"`

	// Deprecated: Use .Symbol instead.
	Column SymbolConfig       `yaml:"column"`
	Symbol SymbolConfigCompat `yaml:"symbol,omitempty"`

	// Symbol config
	OID  string `yaml:"OID"`
	Name string `yaml:"name"`

	IndexTransform []MetricIndexTransform `yaml:"index_transform"`

	Mapping map[string]string `yaml:"mapping"`

	// Regex
	Match string            `yaml:"match"`
	Tags  map[string]string `yaml:"tags"`

	symbolTag string
	pattern   *regexp.Regexp
}

func (mtc *MetricTagConfig) Copy() MetricTagConfig {
	var pattern *regexp.Regexp
	if mtc.pattern != nil {
		pattern = mtc.pattern.Copy() //nolint:staticcheck
	}

	return MetricTagConfig{
		Tag:            mtc.Tag,
		Index:          mtc.Index,
		Column:         mtc.Column.Copy(),
		Symbol:         mtc.Symbol.Copy(),
		OID:            mtc.OID,
		Name:           mtc.Name,
		IndexTransform: CopyMetricIndexTransforms(mtc.IndexTransform),
		Mapping:        CopyMapStringString(mtc.Mapping),
		Match:          mtc.Match,
		Tags:           CopyMapStringString(mtc.Tags),
		symbolTag:      mtc.symbolTag,
		pattern:        pattern,
	}
}

// MetricTagConfigList holds configs for a list of metric tags.
type MetricTagConfigList []MetricTagConfig

// MetricIndexTransform holds configs for metric index transform.
type MetricIndexTransform struct {
	Start uint `yaml:"start"`
	End   uint `yaml:"end"`
}

func (mit *MetricIndexTransform) Copy() MetricIndexTransform {
	return MetricIndexTransform{
		Start: mit.Start,
		End:   mit.End,
	}
}

// MetricsConfigOption holds config for metrics options.
type MetricsConfigOption struct {
	Placement    uint   `yaml:"placement"`
	MetricSuffix string `yaml:"metric_suffix"`
}

func (mco *MetricsConfigOption) Copy() MetricsConfigOption {
	return MetricsConfigOption{
		Placement:    mco.Placement,
		MetricSuffix: mco.MetricSuffix,
	}
}

// MetricsConfig holds configs for a metric.
type MetricsConfig struct {
	// Symbol configs
	Symbol SymbolConfig `yaml:"symbol"`

	// Legacy Symbol configs syntax
	OID  string `yaml:"OID"`
	Name string `yaml:"name"`

	// Table configs
	Symbols []SymbolConfig `yaml:"symbols"`

	StaticTags []string            `yaml:"static_tags"`
	MetricTags MetricTagConfigList `yaml:"metric_tags"`

	// Deprecated: use Symbol.MetricType instead
	ForcedType string `yaml:"forced_type"`
	MetricType string `yaml:"metric_type,omitempty"`

	Options MetricsConfigOption `yaml:"options"`
}

func (m *MetricsConfig) Copy() MetricsConfig {
	return MetricsConfig{
		Symbol:     m.Symbol.Copy(),
		OID:        m.OID,
		Name:       m.Name,
		Symbols:    CopySymbolConfigs(m.Symbols),
		StaticTags: CopyStrings(m.StaticTags),
		MetricTags: CopyMetricTagConfigs(m.MetricTags),
		ForcedType: m.ForcedType,
		MetricType: m.MetricType,
		Options:    m.Options.Copy(),
	}
}

// GetSymbolTags returns symbol tags.
func (m *MetricsConfig) GetSymbolTags() []string {
	var symbolTags []string
	for _, metricTag := range m.MetricTags {
		symbolTags = append(symbolTags, metricTag.symbolTag)
	}
	return symbolTags
}

// IsColumn returns true if the metrics config define columns metrics.
func (m *MetricsConfig) IsColumn() bool {
	return len(m.Symbols) > 0
}

// IsScalar returns true if the metrics config define scalar metrics.
func (m *MetricsConfig) IsScalar() bool {
	return m.Symbol.OID != "" && m.Symbol.Name != ""
}

// GetTags returns tags based on MetricTagConfig and a value.
func (mtc *MetricTagConfig) GetTags(value string) []string {
	var tags []string
	if mtc.Tag != "" {
		if len(mtc.Mapping) > 0 {
			mappedValue, err := GetMappedValue(value, mtc.Mapping)
			if err != nil {
				l.Debugf("error getting tags. mapping for `%s` does not exist. mapping=`%v`", value, mtc.Mapping)
			} else {
				tags = append(tags, mtc.Tag+":"+mappedValue)
			}
		} else {
			tags = append(tags, mtc.Tag+":"+value)
		}
	} else if mtc.Match != "" {
		if mtc.pattern == nil {
			l.Warnf("match pattern must be present: match=%s", mtc.Match)
			return tags
		}
		if mtc.pattern.MatchString(value) {
			for key, val := range mtc.Tags {
				normalizedTemplate := normalizeRegexReplaceValue(val)
				replacedVal := RegexReplaceValue(value, mtc.pattern, normalizedTemplate)
				if replacedVal == "" {
					l.Debugf("pattern `%v` failed to match `%v` with template `%v`", mtc.pattern, value, normalizedTemplate)
					continue
				}
				tags = append(tags, key+":"+replacedVal)
			}
		}
	}
	return tags
}

// GetMappedValue retrieves mapped value from a given mapping.
// If mapping is empty, it will return the index.
func GetMappedValue(index string, mapping map[string]string) (string, error) {
	if len(mapping) > 0 {
		mappedValue, ok := mapping[index]
		if !ok {
			return "", fmt.Errorf("mapping for `%s` does not exist. mapping=`%v`", index, mapping)
		}
		return mappedValue, nil
	}
	return index, nil
}

// RegexReplaceValue replaces a value using a regex and template.
func RegexReplaceValue(value string, pattern *regexp.Regexp, normalizedTemplate string) string {
	result := []byte{}
	for _, submatches := range pattern.FindAllStringSubmatchIndex(value, 1) {
		result = pattern.ExpandString(result, normalizedTemplate, value, submatches)
	}
	return string(result)
}

// normalizeRegexReplaceValue normalize regex value to keep compatibility with Python
// Converts \1 into $1, \2 into $2, etc.
func normalizeRegexReplaceValue(val string) string {
	re := regexp.MustCompile("\\\\(\\d+)") // nolint:gosimple
	return re.ReplaceAllString(val, "$$$1")
}

// NormalizeMetrics converts legacy syntax to new syntax
// 1/ converts old symbol syntax to new symbol syntax
//
//	metric.Name and metric.OID info are moved to metric.Symbol.Name and metric.Symbol.OID
func NormalizeMetrics(metrics []MetricsConfig) {
	for i := range metrics {
		metric := &metrics[i]

		// converts old symbol syntax to new symbol syntax
		if metric.Symbol.Name == "" && metric.Symbol.OID == "" && metric.Name != "" && metric.OID != "" {
			metric.Symbol.Name = metric.Name
			metric.Symbol.OID = metric.OID
			metric.Name = ""
			metric.OID = ""
		}
	}
}

// =============================================================================

// config_oid

// OidConfig holds configs for OIDs to fetch.
type OidConfig struct {
	// ScalarOids are all scalar oids to fetch
	ScalarOids []string
	// ColumnOids are all column oids to fetch
	ColumnOids []string
}

func (oc *OidConfig) AddScalarOids(oidsToAdd []string) {
	oc.ScalarOids = oc.addOidsIfNotPresent(oc.ScalarOids, oidsToAdd)
}

func (oc *OidConfig) AddColumnOids(oidsToAdd []string) {
	oc.ColumnOids = oc.addOidsIfNotPresent(oc.ColumnOids, oidsToAdd)
}

func (oc *OidConfig) addOidsIfNotPresent(configOids []string, oidsToAdd []string) []string {
	for _, oidToAdd := range oidsToAdd {
		if oidToAdd == "" {
			continue
		}
		isAlreadyPresent := false
		for _, oid := range configOids {
			if oid == oidToAdd {
				isAlreadyPresent = true
				break
			}
		}
		if isAlreadyPresent {
			continue
		}
		configOids = append(configOids, oidToAdd)
	}
	sort.Strings(configOids)
	return configOids
}

func (oc *OidConfig) Clean() {
	oc.ScalarOids = nil
	oc.ColumnOids = nil
}

func (oc *OidConfig) Copy() OidConfig {
	return OidConfig{
		ScalarOids: CopyStrings(oc.ScalarOids),
		ColumnOids: CopyStrings(oc.ColumnOids),
	}
}

// =============================================================================

// config_profile

type ProfileConfigMap map[string]profileConfig

type profileConfig struct {
	DefinitionFile string            `yaml:"definition_file"`
	Definition     ProfileDefinition `yaml:"definition"`
}

// =============================================================================

// config_validate_enrich

var validMetadataResources = map[string]map[string]bool{
	"device": {
		"name":          true,
		"description":   true,
		"sys_object_id": true,
		"location":      true,
		"serial_number": true,
		"vendor":        true,
		"version":       true,
		"product_name":  true,
		"model":         true,
		"os_name":       true,
		"os_version":    true,
		"os_hostname":   true,
		"type":          true,
	},
	"interface": {
		"name":         true,
		"alias":        true,
		"description":  true,
		"mac_address":  true,
		"admin_status": true,
		"oper_status":  true,
	},
}

// ValidateEnrichMetricTags validates and enrich metric tags.
func ValidateEnrichMetricTags(metricTags []MetricTagConfig) []string {
	var errors []string
	for i := range metricTags {
		errors = append(errors, validateEnrichMetricTag(&metricTags[i])...)
	}
	return errors
}

// ValidateEnrichMetrics will validate MetricsConfig and enrich it.
// Example of enrichment:
// - storage of compiled regex pattern.
func ValidateEnrichMetrics(metrics []MetricsConfig) []string {
	var errors []string
	for i := range metrics {
		metricConfig := &metrics[i]
		if !metricConfig.IsScalar() && !metricConfig.IsColumn() {
			errors = append(errors, fmt.Sprintf("either a table symbol or a scalar symbol must be provided: %#v", metricConfig))
		}
		if metricConfig.IsScalar() && metricConfig.IsColumn() {
			errors = append(errors, fmt.Sprintf("table symbol and scalar symbol cannot be both provided: %#v", metricConfig))
		}
		if metricConfig.IsScalar() {
			errors = append(errors, validateEnrichSymbol(&metricConfig.Symbol, ScalarSymbol)...)
		}
		if metricConfig.IsColumn() {
			for j := range metricConfig.Symbols {
				errors = append(errors, validateEnrichSymbol(&metricConfig.Symbols[j], ColumnSymbol)...)
			}
			if len(metricConfig.MetricTags) == 0 {
				errors = append(errors, fmt.Sprintf("column symbols %v doesn't have a 'metric_tags' section, all its metrics will use the same tags; "+
					"if the table has multiple rows, only one row will be submitted; "+
					"please add at least one discriminating metric tag (such as a row index) "+
					"to ensure metrics of all rows are submitted", metricConfig.Symbols))
			}
			for i := range metricConfig.MetricTags {
				metricTag := &metricConfig.MetricTags[i]
				errors = append(errors, validateEnrichMetricTag(metricTag)...)
			}
		}
		if metricConfig.MetricType == "" && metricConfig.ForcedType != "" {
			metricConfig.MetricType = metricConfig.ForcedType
			metricConfig.ForcedType = ""
		}
	}
	return errors
}

// validateEnrichMetadata will validate MetadataConfig and enrich it.
func validateEnrichMetadata(metadata MetadataConfig) []string {
	var errors []string
	for resName := range metadata {
		_, isValidRes := validMetadataResources[resName]
		if !isValidRes {
			errors = append(errors, fmt.Sprintf("invalid resource: %s", resName))
		} else {
			res := metadata[resName]
			for fieldName := range res.Fields {
				_, isValidField := validMetadataResources[resName][fieldName]
				if !isValidField {
					errors = append(errors, fmt.Sprintf("invalid resource (%s) field: %s", resName, fieldName))
					continue
				}
				field := res.Fields[fieldName]
				for i := range field.Symbols {
					errors = append(errors, validateEnrichSymbol(&field.Symbols[i], MetadataSymbol)...)
				}
				if field.Symbol.OID != "" {
					errors = append(errors, validateEnrichSymbol(&field.Symbol, MetadataSymbol)...)
				}
				res.Fields[fieldName] = field
			}
			metadata[resName] = res
		}
		if resName == "device" && len(metadata[resName].IDTags) > 0 {
			errors = append(errors, "device resource does not support custom id_tags")
		}
		for i := range metadata[resName].IDTags {
			metricTag := &metadata[resName].IDTags[i]
			errors = append(errors, validateEnrichMetricTag(metricTag)...)
		}
	}
	return errors
}

// SymbolContext represent the context in which the symbol is used.
type SymbolContext int64

// ScalarSymbol enums.
const (
	ScalarSymbol SymbolContext = iota
	ColumnSymbol
	MetricTagSymbol
	MetadataSymbol
)

// validateEnrichSymbol validates and enrich symbol.
func validateEnrichSymbol(symbol *SymbolConfig, symbolContext SymbolContext) []string {
	var errors []string
	if symbol.Name == "" {
		errors = append(errors, fmt.Sprintf("symbol name missing: name=`%s` oid=`%s`", symbol.Name, symbol.OID))
	}
	if symbol.OID == "" {
		if symbolContext == ColumnSymbol && !symbol.ConstantValueOne {
			errors = append(errors, fmt.Sprintf("symbol oid or send_as_one missing: name=`%s` oid=`%s`", symbol.Name, symbol.OID))
		} else if symbolContext != ColumnSymbol {
			errors = append(errors, fmt.Sprintf("symbol oid missing: name=`%s` oid=`%s`", symbol.Name, symbol.OID))
		}
	}
	if symbol.ExtractValue != "" {
		pattern, err := regexp.Compile(symbol.ExtractValue)
		if err != nil {
			errors = append(errors, fmt.Sprintf("cannot compile `extract_value` (%s): %v", symbol.ExtractValue, err))
		} else {
			symbol.ExtractValueCompiled = pattern
		}
	}
	if symbol.MatchPattern != "" {
		pattern, err := regexp.Compile(symbol.MatchPattern)
		if err != nil {
			errors = append(errors, fmt.Sprintf("cannot compile `extract_value` (%s): %v", symbol.ExtractValue, err))
		} else {
			symbol.MatchPatternCompiled = pattern
		}
	}
	if symbolContext != ColumnSymbol && symbol.ConstantValueOne {
		errors = append(errors, "`constant_value_one` cannot be used outside of tables")
	}
	if (symbolContext != ColumnSymbol && symbolContext != ScalarSymbol) && symbol.MetricType != "" {
		errors = append(errors, "`metric_type` cannot be used outside scalar/table metric symbols and metrics root")
	}
	return errors
}

func validateEnrichMetricTag(metricTag *MetricTagConfig) []string {
	var errors []string
	if (metricTag.Column.OID != "" || metricTag.Column.Name != "") && (metricTag.Symbol.OID != "" || metricTag.Symbol.Name != "") {
		//nolint:lll
		errors = append(errors, fmt.Sprintf("metric tag symbol and column cannot be both declared: symbol=%v, column=%v", metricTag.Symbol, metricTag.Column))
	}
	if metricTag.Column.OID != "" || metricTag.Column.Name != "" {
		metricTag.Symbol = SymbolConfigCompat(metricTag.Column.Copy())
		metricTag.Column = SymbolConfig{}
	}
	// OID/Name to Symbol harmonization:
	// When users declare metric tag like:
	//   metric_tags:
	//     - OID: 1.2.3
	//       symbol: aSymbol
	// this will lead to OID stored as MetricTagConfig.OID  and name stored as MetricTagConfig.Symbol.Name
	// When this happens, we harmonize by moving MetricTagConfig.OID to MetricTagConfig.Symbol.OID.
	if metricTag.OID != "" && metricTag.Symbol.OID != "" {
		//nolint:lll
		errors = append(errors, fmt.Sprintf("metric tag OID and symbol.OID cannot be both declared: OID=%s, symbol.OID=%s", metricTag.OID, metricTag.Symbol.OID))
	}
	if metricTag.OID != "" && metricTag.Symbol.OID == "" {
		metricTag.Symbol.OID = metricTag.OID
	}
	if metricTag.Name != "" && metricTag.Symbol.Name == "" {
		metricTag.Symbol.Name = metricTag.Name
	}
	if metricTag.Symbol.OID != "" || metricTag.Symbol.Name != "" {
		symbol := SymbolConfig(metricTag.Symbol)
		errors = append(errors, validateEnrichSymbol(&symbol, MetricTagSymbol)...)
		metricTag.Symbol = SymbolConfigCompat(symbol)
	}
	if metricTag.Match != "" {
		pattern, err := regexp.Compile(metricTag.Match)
		if err != nil {
			errors = append(errors, fmt.Sprintf("cannot compile `match` (`%s`): %v", metricTag.Match, err))
		} else {
			metricTag.pattern = pattern
		}
		if len(metricTag.Tags) == 0 {
			errors = append(errors, fmt.Sprintf("`tags` mapping must be provided if `match` (`%s`) is defined", metricTag.Match))
		}
	}
	if len(metricTag.Mapping) > 0 && metricTag.Tag == "" {
		errors = append(errors, fmt.Sprintf("`tag` must be provided if `mapping` (`%s`) is defined", metricTag.Mapping))
	}
	for _, transform := range metricTag.IndexTransform {
		if transform.Start > transform.End {
			errors = append(errors, fmt.Sprintf("transform rule end should be greater than start. Invalid rule: %#v", transform))
		}
	}
	return errors
}

// =============================================================================
