// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package promsd

/*
Copyright The Prometheus Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import (
	"crypto/md5" //nolint:gosec // Prometheus uses MD5 to keep hashmod results stable.
	"encoding/binary"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/prometheus/common/model"
)

// RelabelConfig describes one Prometheus-compatible target relabeling rule.
type RelabelConfig struct {
	SourceLabels []string `toml:"source_labels"`
	Separator    *string  `toml:"separator"`
	Regex        *string  `toml:"regex"`
	Modulus      uint64   `toml:"modulus"`
	TargetLabel  string   `toml:"target_label"`
	Replacement  *string  `toml:"replacement"`
	Action       *string  `toml:"action"`
}

type relabelAction string

const (
	actionReplace   relabelAction = "replace"
	actionKeep      relabelAction = "keep"
	actionDrop      relabelAction = "drop"
	actionKeepEqual relabelAction = "keepequal"
	actionDropEqual relabelAction = "dropequal"
	actionHashMod   relabelAction = "hashmod"
	actionLabelMap  relabelAction = "labelmap"
	actionLabelDrop relabelAction = "labeldrop"
	actionLabelKeep relabelAction = "labelkeep"
	actionLowercase relabelAction = "lowercase"
	actionUppercase relabelAction = "uppercase"
)

type compiledRelabelConfig struct {
	sourceLabels []string
	separator    string
	regex        *regexp.Regexp
	modulus      uint64
	targetLabel  string
	replacement  string
	action       relabelAction
}

// relabelLabels adapts DataKit target labels to Prometheus relabel semantics.
// Prometheus treats an empty label value as a missing label.
type relabelLabels map[string]string

func newRelabelLabels(src map[string]string) relabelLabels {
	labels := make(relabelLabels, len(src))
	for name, value := range src {
		labels.Set(name, value)
	}
	return labels
}

func (labels relabelLabels) Get(name string) string {
	return labels[name]
}

func (labels relabelLabels) Set(name, value string) {
	if value == "" {
		labels.Del(name)
		return
	}
	labels[name] = value
}

func (labels relabelLabels) Del(name string) {
	delete(labels, name)
}

func (labels relabelLabels) Range(fn func(name, value string)) {
	type label struct {
		name  string
		value string
	}

	snapshot := make([]label, 0, len(labels))
	for name, value := range labels {
		snapshot = append(snapshot, label{name: name, value: value})
	}
	// Prometheus iterates labels deterministically. Keep the same property when
	// multiple labels are mapped to the same destination by labelmap.
	sort.Slice(snapshot, func(i, j int) bool {
		return snapshot[i].name < snapshot[j].name
	})

	for _, label := range snapshot {
		fn(label.name, label.value)
	}
}

const (
	defaultRelabelSeparator   = ";"
	defaultRelabelRegex       = "(.*)"
	defaultRelabelReplacement = "$1"
	defaultRelabelAction      = actionReplace
)

var relabelTarget = regexp.MustCompile(`^(?:(?:[a-zA-Z_]|\$(?:\{\w+\}|\w+))+\w*)+$`)

func (cfg *ScrapeConfig) setupRelabelConfigs() error {
	compiled := make([]*compiledRelabelConfig, 0, len(cfg.RelabelConfigs))
	for idx, rule := range cfg.RelabelConfigs {
		if rule == nil {
			return fmt.Errorf("relabel_configs[%d]: rule is empty", idx)
		}

		got, err := compileRelabelConfig(rule)
		if err != nil {
			return fmt.Errorf("relabel_configs[%d]: %w", idx, err)
		}
		compiled = append(compiled, got)
	}
	cfg.compiledRelabelConfigs = compiled
	return nil
}

func compileRelabelConfig(rule *RelabelConfig) (*compiledRelabelConfig, error) {
	separator := stringValue(rule.Separator, defaultRelabelSeparator)
	regexText := stringValue(rule.Regex, defaultRelabelRegex)
	replacement := stringValue(rule.Replacement, defaultRelabelReplacement)
	action := relabelAction(strings.ToLower(stringValue(rule.Action, string(defaultRelabelAction))))

	re, err := regexp.Compile("^(?s:" + regexText + ")$")
	if err != nil {
		return nil, fmt.Errorf("invalid regex %q: %w", regexText, err)
	}

	switch action {
	case actionReplace, actionHashMod, actionLowercase, actionUppercase, actionKeepEqual, actionDropEqual:
		if rule.TargetLabel == "" {
			return nil, fmt.Errorf("action %q requires target_label", action)
		}
	case actionKeep, actionDrop, actionLabelMap, actionLabelDrop, actionLabelKeep:
	default:
		return nil, fmt.Errorf("unknown action %q", action)
	}

	if action == actionHashMod && rule.Modulus == 0 {
		return nil, fmt.Errorf("action %q requires a non-zero modulus", action)
	}
	for _, sourceLabel := range rule.SourceLabels {
		if !model.LabelName(sourceLabel).IsValid() {
			return nil, fmt.Errorf("invalid source_label %q", sourceLabel)
		}
	}
	if action == actionReplace && !relabelTarget.MatchString(rule.TargetLabel) {
		return nil, fmt.Errorf("invalid target_label %q for action %q", rule.TargetLabel, action)
	}
	if (action == actionHashMod || action == actionLowercase || action == actionUppercase ||
		action == actionKeepEqual || action == actionDropEqual) && !model.LabelName(rule.TargetLabel).IsValid() {
		return nil, fmt.Errorf("invalid target_label %q for action %q", rule.TargetLabel, action)
	}
	if action == actionLabelMap && !relabelTarget.MatchString(replacement) {
		return nil, fmt.Errorf("invalid replacement %q for action %q", replacement, action)
	}

	if action == actionLabelDrop || action == actionLabelKeep {
		if rule.SourceLabels != nil || rule.TargetLabel != "" || rule.Modulus != 0 ||
			separator != defaultRelabelSeparator || replacement != defaultRelabelReplacement {
			return nil, fmt.Errorf("action %q only supports regex", action)
		}
	}
	if action == actionKeepEqual || action == actionDropEqual {
		if rule.Regex != nil || rule.Modulus != 0 ||
			separator != defaultRelabelSeparator || replacement != defaultRelabelReplacement {
			return nil, fmt.Errorf("action %q only supports source_labels and target_label", action)
		}
	}
	if (action == actionLowercase || action == actionUppercase) && replacement != defaultRelabelReplacement {
		return nil, fmt.Errorf("replacement cannot be set for action %q", action)
	}

	return &compiledRelabelConfig{
		sourceLabels: rule.SourceLabels,
		separator:    separator,
		regex:        re,
		modulus:      rule.Modulus,
		targetLabel:  rule.TargetLabel,
		replacement:  replacement,
		action:       action,
	}, nil
}

func applyRelabelConfigs(labels map[string]string, configs []*compiledRelabelConfig) bool {
	for _, cfg := range configs {
		if !applyRelabelConfig(labels, cfg) {
			return false
		}
	}
	return true
}

func applyRelabelConfig(labels map[string]string, cfg *compiledRelabelConfig) bool { //nolint:cyclop // Actions intentionally share the Prometheus rule dispatcher.
	lb := relabelLabels(labels)
	var valueArray [16]string
	values := valueArray[:0]
	if len(cfg.sourceLabels) > cap(values) {
		values = make([]string, 0, len(cfg.sourceLabels))
	}
	for _, name := range cfg.sourceLabels {
		values = append(values, lb.Get(name))
	}
	value := strings.Join(values, cfg.separator)

	switch cfg.action {
	case actionDrop:
		if cfg.regex.MatchString(value) {
			return false
		}
	case actionKeep:
		if !cfg.regex.MatchString(value) {
			return false
		}
	case actionDropEqual:
		if lb.Get(cfg.targetLabel) == value {
			return false
		}
	case actionKeepEqual:
		if lb.Get(cfg.targetLabel) != value {
			return false
		}
	case actionReplace:
		indexes := cfg.regex.FindStringSubmatchIndex(value)
		if indexes == nil {
			break
		}
		targetLabel := string(cfg.regex.ExpandString(nil, cfg.targetLabel, value, indexes))
		if !model.LabelName(targetLabel).IsValid() {
			break
		}
		replacement := string(cfg.regex.ExpandString(nil, cfg.replacement, value, indexes))
		lb.Set(targetLabel, replacement)
	case actionLowercase:
		lb.Set(cfg.targetLabel, strings.ToLower(value))
	case actionUppercase:
		lb.Set(cfg.targetLabel, strings.ToUpper(value))
	case actionHashMod:
		hash := md5.Sum([]byte(value)) //nolint:gosec // Keep behavior compatible with Prometheus.
		modulus := binary.BigEndian.Uint64(hash[8:]) % cfg.modulus
		lb.Set(cfg.targetLabel, strconv.FormatUint(modulus, 10))
	case actionLabelMap:
		lb.Range(func(name, labelValue string) {
			if cfg.regex.MatchString(name) {
				lb.Set(cfg.regex.ReplaceAllString(name, cfg.replacement), labelValue)
			}
		})
	case actionLabelDrop:
		for name := range lb {
			if cfg.regex.MatchString(name) {
				lb.Del(name)
			}
		}
	case actionLabelKeep:
		for name := range lb {
			if !cfg.regex.MatchString(name) {
				lb.Del(name)
			}
		}
	}
	return true
}

func stringValue(value *string, fallback string) string {
	if value == nil {
		return fallback
	}
	return *value
}
