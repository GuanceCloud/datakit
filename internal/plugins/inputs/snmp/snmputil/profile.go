// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmputil

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp/snmprefiles"
	"gopkg.in/yaml.v2"
)

type ProfileDefinitionMap map[string]ProfileDefinition

// DeviceMeta holds device related static metadata.
type DeviceMeta struct {
	Vendor string `yaml:"vendor"`
}

func (dm *DeviceMeta) Copy() DeviceMeta {
	return DeviceMeta{
		Vendor: dm.Vendor,
	}
}

type ProfileDefinition struct {
	Metrics      []MetricsConfig   `yaml:"metrics"`
	Metadata     MetadataConfig    `yaml:"metadata"`
	MetricTags   []MetricTagConfig `yaml:"metric_tags"`
	StaticTags   []string          `yaml:"static_tags"`
	Extends      []string          `yaml:"extends"`
	Device       DeviceMeta        `yaml:"device"`
	SysObjectIds StringArray       `yaml:"sysobjectid"`
}

func (pd *ProfileDefinition) Copy() *ProfileDefinition {
	if pd == nil {
		return nil
	}
	return &ProfileDefinition{
		Metrics:      CopyMetricsConfigs(pd.Metrics),
		Metadata:     CopyMapStringMetadataResourceConfig(pd.Metadata),
		MetricTags:   CopyMetricTagConfigs(pd.MetricTags),
		StaticTags:   CopyStrings(pd.StaticTags),
		Extends:      CopyStrings(pd.Extends),
		Device:       pd.Device.Copy(),
		SysObjectIds: CopyStrings(pd.SysObjectIds),
	}
}

func newProfileDefinition() *ProfileDefinition {
	p := &ProfileDefinition{}
	p.Metadata = make(MetadataConfig)
	return p
}

func (pc profileConfig) Copy() profileConfig {
	return profileConfig{
		DefinitionFile: pc.DefinitionFile,
		Definition:     *pc.Definition.Copy(),
	}
}

var defaultProfilesMu = &sync.Mutex{}

var globalProfileConfigMap ProfileDefinitionMap

// LoadDefaultProfiles will load the profiles from disk only once and store it
// in globalProfileConfigMap. The subsequent call to it will return profiles stored in
// globalProfileConfigMap. The mutex will help loading once when `loadDefaultProfiles`
// is called by multiple check instances.
func LoadDefaultProfiles() (ProfileDefinitionMap, error) {
	defaultProfilesMu.Lock()
	defer defaultProfilesMu.Unlock()

	if globalProfileConfigMap != nil {
		l.Debugf("loader default profiles from cache")
		return globalProfileConfigMap, nil
	}

	extraProfiles := getYamlExtraProfiles()
	defaultProfiles := getYamlDefaultProfiles()
	profiles := resolveProfiles(extraProfiles, defaultProfiles)
	if len(profiles) == 0 {
		return nil, fmt.Errorf("failed to load default profiles")
	}
	globalProfileConfigMap = profiles
	return profiles, nil
}

func getProfileDefinitions(profilesRoot string, ignoreMissing bool) (ProfileConfigMap, error) {
	files, err := os.ReadDir(profilesRoot)
	if err != nil {
		if ignoreMissing && os.IsNotExist(err) {
			return ProfileConfigMap{}, nil
		}
		return nil, fmt.Errorf("failed to read dir `%s`: %w", profilesRoot, err)
	}

	profiles := make(ProfileConfigMap)
	for _, f := range files {
		fName := f.Name()
		// Skip non yaml profiles
		if !strings.HasSuffix(fName, ".yaml") {
			continue
		}
		profileName := fName[:len(fName)-len(".yaml")]
		absPath := filepath.Join(profilesRoot, fName)
		definition, err := readProfileDefinition(absPath)
		if err != nil {
			l.Warnf("failed to read profile definition `%s`: %s", profileName, err)
			continue
		}
		profiles[profileName] = profileConfig{Definition: *definition}
	}
	return profiles, nil
}

func getYamlDefaultProfiles() ProfileConfigMap {
	profiles, err := getProfileDefinitions(snmprefiles.GetProfilesRoot(), false)
	if err != nil {
		l.Warnf("failed to load default profile definitions: %s", err)
		return ProfileConfigMap{}
	}
	return profiles
}

func getYamlExtraProfiles() ProfileConfigMap {
	profiles, err := getProfileDefinitions(snmprefiles.GetExtraProfilesRoot(), true)
	if err != nil {
		l.Warnf("failed to load extra profile definitions: %s", err)
		return ProfileConfigMap{}
	}
	return profiles
}

func LoadProfiles(pConfig ProfileConfigMap) (ProfileDefinitionMap, error) {
	profiles := make(map[string]ProfileDefinition, len(pConfig))

	for name, profile := range pConfig {
		if profile.DefinitionFile != "" {
			profileDefinition, err := readProfileDefinition(profile.DefinitionFile)
			if err != nil {
				l.Warnf("failed to read profile definition `%s`: %s", name, err)
				continue
			}

			err = recursivelyExpandBaseProfilesFromFiles(profileDefinition, profileDefinition.Extends, []string{})
			if err != nil {
				l.Warnf("failed to expand profile `%s`: %s", name, err)
				continue
			}
			NormalizeMetrics(profileDefinition.Metrics)
			errors := validateEnrichMetadata(profileDefinition.Metadata)
			errors = append(errors, ValidateEnrichMetrics(profileDefinition.Metrics)...)
			errors = append(errors, ValidateEnrichMetricTags(profileDefinition.MetricTags)...)
			if len(errors) > 0 {
				l.Warnf("validation errors in profile `%s`: %s", name, strings.Join(errors, "\n"))
				continue
			}
			profiles[name] = *profileDefinition
		} else {
			profiles[name] = profile.Definition
		}
	}
	return profiles, nil
}

func recursivelyExpandBaseProfilesFromFiles(definition *ProfileDefinition, extends []string, extendsHistory []string) error {
	for _, basePath := range extends {
		for _, extend := range extendsHistory {
			if extend == basePath {
				return fmt.Errorf("cyclic profile extend detected, `%s` has already been extended, extendsHistory=`%v`", basePath, extendsHistory)
			}
		}
		baseDefinition, err := readProfileDefinition(basePath)
		if err != nil {
			return err
		}

		mergeProfileDefinition(definition, baseDefinition)

		newExtendsHistory := append(CopyStrings(extendsHistory), basePath)
		err = recursivelyExpandBaseProfilesFromFiles(definition, baseDefinition.Extends, newExtendsHistory)
		if err != nil {
			return err
		}
	}
	return nil
}

func readProfileDefinition(definitionFile string) (*ProfileDefinition, error) {
	filePath := resolveProfileDefinitionPath(definitionFile)
	buf, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return nil, fmt.Errorf("failed to read file `%s`: %w", filePath, err)
	}

	profileDefinition := newProfileDefinition()
	err = yaml.Unmarshal(buf, profileDefinition)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshall %q: %w", filePath, err)
	}
	return profileDefinition, nil
}

func resolveProfileDefinitionPath(definitionFile string) string {
	if filepath.IsAbs(definitionFile) {
		return definitionFile
	}
	return filepath.Join(snmprefiles.GetProfilesRoot(), definitionFile)
}

func mergeProfiles(profilesA ProfileConfigMap, profilesB ProfileConfigMap) ProfileConfigMap {
	profiles := make(ProfileConfigMap)
	for name, profile := range profilesA {
		profiles[name] = profile.Copy()
	}
	for name, profile := range profilesB {
		profiles[name] = profile.Copy()
	}
	return profiles
}

func resolveProfiles(userProfiles, defaultProfiles ProfileConfigMap) ProfileDefinitionMap {
	rawProfiles := mergeProfiles(defaultProfiles, userProfiles)
	userExpandedProfiles := normalizeProfiles(rawProfiles, defaultProfiles)
	return userExpandedProfiles
}

func normalizeProfiles(pConfig ProfileConfigMap, defaultProfiles ProfileConfigMap) ProfileDefinitionMap {
	profiles := make(ProfileDefinitionMap, len(pConfig))

	for name := range pConfig {
		// No need to resolve abstract profile
		if strings.HasPrefix(name, "_") {
			continue
		}

		newProfileConfig := pConfig[name].Copy()
		err := recursivelyExpandBaseProfiles(name, &newProfileConfig.Definition, newProfileConfig.Definition.Extends, []string{}, pConfig, defaultProfiles)
		if err != nil {
			l.Warnf("failed to expand profile `%s`: %s", name, err)
			continue
		}
		NormalizeMetrics(newProfileConfig.Definition.Metrics)
		errors := validateEnrichMetadata(newProfileConfig.Definition.Metadata)
		errors = append(errors, ValidateEnrichMetrics(newProfileConfig.Definition.Metrics)...)
		errors = append(errors, ValidateEnrichMetricTags(newProfileConfig.Definition.MetricTags)...)
		if len(errors) > 0 {
			l.Warnf("validation errors in profile `%s`: %s", name, strings.Join(errors, "\n"))
			continue
		}
		profiles[name] = newProfileConfig.Definition
	}
	return profiles
}

func recursivelyExpandBaseProfiles(parentExtend string, definition *ProfileDefinition, extends []string, extendsHistory []string, profiles ProfileConfigMap, defaultProfiles ProfileConfigMap) error {
	for _, extendEntry := range extends {
		extendEntry = strings.TrimSuffix(extendEntry, ".yaml")

		var baseDefinition *ProfileDefinition
		// User profile can extend default profile by extending the default profile.
		// If the extend entry has the same name as the profile name, we assume the extend entry is referring to a default profile.
		if extendEntry == parentExtend {
			profile, ok := defaultProfiles[extendEntry]
			if !ok {
				return fmt.Errorf("extend does not exist: `%s`", extendEntry)
			}
			baseDefinition = &profile.Definition
		} else {
			profile, ok := profiles[extendEntry]
			if !ok {
				profile, ok = defaultProfiles[extendEntry]
				if !ok {
					return fmt.Errorf("extend does not exist: `%s`", extendEntry)
				}
			}
			baseDefinition = &profile.Definition
		}
		if slices.Contains(extendsHistory, extendEntry) {
			return fmt.Errorf("cyclic profile extend detected, `%s` has already been extended, extendsHistory=`%v`", extendEntry, extendsHistory)
		}

		mergeProfileDefinition(definition, baseDefinition)

		newExtendsHistory := append(CopyStrings(extendsHistory), extendEntry)
		err := recursivelyExpandBaseProfiles(extendEntry, definition, baseDefinition.Extends, newExtendsHistory, profiles, defaultProfiles)
		if err != nil {
			return err
		}
	}
	return nil
}

func mergeProfileDefinition(targetDefinition *ProfileDefinition, baseDefinition *ProfileDefinition) {
	targetDefinition.Metrics = append(targetDefinition.Metrics, baseDefinition.Metrics...)
	targetDefinition.MetricTags = append(targetDefinition.MetricTags, baseDefinition.MetricTags...)
	targetDefinition.StaticTags = append(targetDefinition.StaticTags, baseDefinition.StaticTags...)
	if targetDefinition.Metadata == nil {
		targetDefinition.Metadata = make(MetadataConfig)
	}
	for baseResName, baseResource := range baseDefinition.Metadata {
		if _, ok := targetDefinition.Metadata[baseResName]; !ok {
			targetDefinition.Metadata[baseResName] = newMetadataResourceConfig()
		}
		if resource, ok := targetDefinition.Metadata[baseResName]; ok {
			for _, tagConfig := range baseResource.IDTags {
				resource.IDTags = append(targetDefinition.Metadata[baseResName].IDTags, tagConfig) //nolint:gocritic
			}

			if resource.Fields == nil {
				resource.Fields = make(map[string]MetadataField, len(baseResource.Fields))
			}
			for field, symbol := range baseResource.Fields {
				if _, ok := resource.Fields[field]; !ok {
					resource.Fields[field] = symbol
				}
			}

			targetDefinition.Metadata[baseResName] = resource
		}
	}
}

func getMostSpecificOid(oids []string) (string, error) {
	var mostSpecificParts []int
	var mostSpecificOid string

	if len(oids) == 0 {
		return "", fmt.Errorf("cannot get most specific oid from empty list of oids")
	}

	for _, oid := range oids {
		parts, err := getOidPatternSpecificity(oid)
		if err != nil {
			return "", err
		}
		if len(parts) > len(mostSpecificParts) {
			mostSpecificParts = parts
			mostSpecificOid = oid
			continue
		}
		if len(parts) == len(mostSpecificParts) {
			for i := range mostSpecificParts {
				if parts[i] > mostSpecificParts[i] {
					mostSpecificParts = parts
					mostSpecificOid = oid
				}
			}
		}
	}
	return mostSpecificOid, nil
}

func getOidPatternSpecificity(pattern string) ([]int, error) {
	wildcardKey := -1
	var parts []int
	for _, part := range strings.Split(strings.TrimLeft(pattern, "."), ".") {
		if part == "*" {
			parts = append(parts, wildcardKey)
		} else {
			intPart, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("error parsing part `%s` for pattern `%s`: %v", part, pattern, err) //nolint:errorlint
			}
			parts = append(parts, intPart)
		}
	}
	return parts, nil
}
