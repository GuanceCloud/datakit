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
		oids = append(oids, metricTag.OID)
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
	return oids
}

func ParseColumnOids(metrics []MetricsConfig, metadataConfigs MetadataConfig, collectDeviceMetadata bool) []string {
	var oids []string
	for _, metric := range metrics {
		for _, symbol := range metric.Symbols {
			oids = append(oids, symbol.OID)
		}
		for _, metricTag := range metric.MetricTags {
			oids = append(oids, metricTag.Column.OID)
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
				oids = append(oids, tagConfig.Column.OID)
			}
		}
	}
	return oids
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
				Column: SymbolConfig{
					OID:  "1.3.6.1.2.1.31.1.1.1.1",
					Name: "ifName",
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

	ScaleFactor float64 `yaml:"scale_factor"`
	Format      string  `yaml:"format"`
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
	}
}

// MetricTagConfig holds metric tag info.
type MetricTagConfig struct {
	Tag string `yaml:"tag"`

	// Table config
	Index uint `yaml:"index"`

	// TODO: refactor to rename to `symbol` instead (keep backward compat with `column`)
	Column SymbolConfig `yaml:"column"`

	// Symbol config
	OID  string `yaml:"OID"`
	Name string `yaml:"symbol"`

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

	ForcedType string              `yaml:"forced_type"`
	Options    MetricsConfigOption `yaml:"options"`
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
		tags = append(tags, mtc.Tag+":"+value)
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
			errors = append(errors, validateEnrichSymbol(&metricConfig.Symbol)...)
		}
		if metricConfig.IsColumn() {
			for j := range metricConfig.Symbols {
				errors = append(errors, validateEnrichSymbol(&metricConfig.Symbols[j])...)
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
					errors = append(errors, validateEnrichSymbol(&field.Symbols[i])...)
				}
				if field.Symbol.OID != "" {
					errors = append(errors, validateEnrichSymbol(&field.Symbol)...)
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

func validateEnrichSymbol(symbol *SymbolConfig) []string {
	var errors []string
	if symbol.Name == "" {
		errors = append(errors, fmt.Sprintf("symbol name missing: name=`%s` oid=`%s`", symbol.Name, symbol.OID))
	}
	if symbol.OID == "" {
		errors = append(errors, fmt.Sprintf("symbol oid missing: name=`%s` oid=`%s`", symbol.Name, symbol.OID))
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
	return errors
}

func validateEnrichMetricTag(metricTag *MetricTagConfig) []string {
	var errors []string
	if metricTag.Column.OID != "" || metricTag.Column.Name != "" {
		errors = append(errors, validateEnrichSymbol(&metricTag.Column)...)
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
	for _, transform := range metricTag.IndexTransform {
		if transform.Start > transform.End {
			errors = append(errors, fmt.Sprintf("transform rule end should be greater than start. Invalid rule: %#v", transform))
		}
	}
	return errors
}

// =============================================================================
