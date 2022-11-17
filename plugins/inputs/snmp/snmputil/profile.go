// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmputil

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/snmp/snmprefiles"
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

	pConfig, err := getDefaultProfilesDefinitionFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to get default profile definitions: %w", err)
	}
	profiles, err := LoadProfiles(pConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to load default profiles: %w", err)
	}
	globalProfileConfigMap = profiles
	return profiles, nil
}

func getDefaultProfilesDefinitionFiles() (ProfileConfigMap, error) {
	profilesRoot := snmprefiles.GetProfilesRoot()
	files, err := ioutil.ReadDir(profilesRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to read dir `%s`: %w", profilesRoot, err)
	}

	profiles := make(ProfileConfigMap)
	for _, f := range files {
		fName := f.Name()
		// Skip partial profiles
		if strings.HasPrefix(fName, "_") {
			continue
		}
		// Skip non yaml profiles
		if !strings.HasSuffix(fName, ".yaml") {
			continue
		}
		profileName := fName[:len(fName)-len(".yaml")]
		profiles[profileName] = profileConfig{DefinitionFile: filepath.Join(profilesRoot, fName)}
	}
	return profiles, nil
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

			err = recursivelyExpandBaseProfiles(profileDefinition, profileDefinition.Extends, []string{})
			if err != nil {
				l.Warnf("failed to expand profile `%s`: %s", name, err)
				continue
			}
			profiles[name] = *profileDefinition
		} else {
			profiles[name] = profile.Definition
		}
	}
	return profiles, nil
}

func readProfileDefinition(definitionFile string) (*ProfileDefinition, error) {
	filePath := resolveProfileDefinitionPath(definitionFile)
	buf, err := ioutil.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return nil, fmt.Errorf("failed to read file `%s`: %w", filePath, err)
	}

	profileDefinition := newProfileDefinition()
	err = yaml.Unmarshal(buf, profileDefinition)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshall %q: %w", filePath, err)
	}
	NormalizeMetrics(profileDefinition.Metrics)
	errors := validateEnrichMetadata(profileDefinition.Metadata)
	errors = append(errors, ValidateEnrichMetrics(profileDefinition.Metrics)...)
	errors = append(errors, ValidateEnrichMetricTags(profileDefinition.MetricTags)...)
	if len(errors) > 0 {
		return nil, fmt.Errorf("validation errors: %s", strings.Join(errors, "\n"))
	}
	return profileDefinition, nil
}

func resolveProfileDefinitionPath(definitionFile string) string {
	if filepath.IsAbs(definitionFile) {
		return definitionFile
	}
	return filepath.Join(snmprefiles.GetProfilesRoot(), definitionFile)
}

func recursivelyExpandBaseProfiles(definition *ProfileDefinition, extends []string, extendsHistory []string) error {
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
		err = recursivelyExpandBaseProfiles(definition, baseDefinition.Extends, newExtendsHistory)
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
