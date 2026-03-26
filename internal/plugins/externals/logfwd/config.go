// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

// Package logfwd provides log collecting and forwarding capabilities.
package logfwd

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/multiline"
)

var (
	// Endpoint configuration.
	datakitHost      = os.Getenv("LOGFWD_DATAKIT_HOST")
	datakitPort      = os.Getenv("LOGFWD_DATAKIT_PORT")
	operatorEndpoint = os.Getenv("LOGFWD_DATAKIT_OPERATOR_ENDPOINT")

	// Pod metadata.
	podName      = os.Getenv("LOGFWD_POD_NAME")
	podIP        = os.Getenv("LOGFWD_POD_IP")
	podNamespace = os.Getenv("LOGFWD_POD_NAMESPACE")
	// podLabels is loaded from /etc/podinfo/labels file or environment variables.

	// Global defaults.
	globalSource       = os.Getenv("LOGFWD_GLOBAL_SOURCE")
	globalStorageIndex = os.Getenv("LOGFWD_GLOBAL_STORAGE_INDEX")
	globalService      = os.Getenv("LOGFWD_GLOBAL_SERVICE")

	// Configuration source.
	envLogConfigsStr = os.Getenv("LOGFWD_LOG_CONFIGS")
	// deprecated.
	deprecatedJSONConfig = os.Getenv("LOGFWD_JSON_CONFIG")
)

type logConfig struct {
	Disable                    bool              `json:"disable"`
	Type                       string            `json:"type"`
	Source                     string            `json:"source"`
	Path                       string            `json:"path"`
	StorageIndex               string            `json:"storage_index"`
	Service                    string            `json:"service"`
	CharacterEncoding          string            `json:"character_encoding"`
	Pipeline                   string            `json:"pipeline"`
	Multiline                  string            `json:"multiline_match"`
	RemoveAnsiEscapeCodes      bool              `json:"remove_ansi_escape_codes"`
	FromBeginning              bool              `json:"from_beginning"`
	FromBeginningThresholdSize int64             `json:"from_beginning_threshold_size"`
	Tags                       map[string]string `json:"tags"`

	multilinePatterns []string `json:"-"`
}

type deprecatedLogConfig struct {
	DataKitAddr string `json:"datakit_addr"`
	Loggings    []struct {
		LogFiles              []string          `json:"logfiles"`
		Ignore                []string          `json:"-"`
		Source                string            `json:"source"`
		StorageIndex          string            `json:"storage_index,omitempty"`
		Service               string            `json:"service"`
		Pipeline              string            `json:"pipeline"`
		CharacterEncoding     string            `json:"character_encoding"`
		MultilineMatch        string            `json:"multiline_match"`
		RemoveAnsiEscapeCodes bool              `json:"remove_ansi_escape_codes"`
		Tags                  map[string]string `json:"tags"`
	} `json:"loggings"`
}

func getEndpointConfig() (string, string, error) {
	host := datakitHost
	port := datakitPort

	if host == "" || port == "" {
		deprecatedHost, deprecatedPort, err := getDeprecatedDataKitEndpoint()
		if err != nil {
			return "", "", err
		}

		if host == "" {
			host = deprecatedHost
		}
		if port == "" {
			port = deprecatedPort
		}
	}

	if host == "" || port == "" {
		return "", "", fmt.Errorf("datakit host and port are required")
	}

	if net.ParseIP(host) == nil {
		if _, err := net.ResolveIPAddr("ip", host); err != nil {
			return "", "", fmt.Errorf("invalid datakit host: %s", host)
		}
	}

	if portNum, err := strconv.Atoi(port); err != nil || portNum < 1 || portNum > 65535 {
		return "", "", fmt.Errorf("invalid datakit port: %s", port)
	}

	operatorURL := ""
	if operatorEndpoint != "" {
		var err error
		operatorURL, err = normalizeOperatorEndpoint(operatorEndpoint)
		if err != nil {
			return "", "", fmt.Errorf("invalid operator endpoint: %s", operatorEndpoint)
		}
	}

	return fmt.Sprintf("%s:%s", host, port), operatorURL, nil
}

func normalizeOperatorEndpoint(endpoint string) (string, error) {
	ep := strings.TrimSpace(endpoint)
	if ep == "" {
		return "", nil
	}

	if !strings.HasPrefix(ep, "http://") && !strings.HasPrefix(ep, "https://") {
		ep = "https://" + ep
	}

	u, err := url.Parse(ep)
	if err != nil {
		return "", err
	}

	u.Scheme = "https"

	if u.Host == "" && u.Path != "" {
		u.Host = u.Path
		u.Path = ""
	}

	return u.String(), nil
}

func parseLogConfigs(str string) ([]*logConfig, error) {
	if str == "" {
		return nil, fmt.Errorf("invalid logConfigs data")
	}

	var configs []*logConfig
	if err := json.Unmarshal([]byte(str), &configs); err != nil {
		return nil, fmt.Errorf("failed to parse log configs, err %w", err)
	}

	for _, cfg := range configs {
		processConfig(cfg)
	}

	return configs, nil
}

func getEffectiveEnvLogConfigs() (string, error) {
	if envLogConfigsStr != "" {
		return envLogConfigsStr, nil
	}

	if deprecatedJSONConfig == "" {
		return "", nil
	}

	return convertDeprecatedJSONConfig(deprecatedJSONConfig)
}

func convertDeprecatedJSONConfig(str string) (string, error) {
	deprecatedConfigs, err := parseDeprecatedLogConfigs(str)
	if err != nil {
		return "", err
	}

	configs := make([]*logConfig, 0)
	for _, deprecatedCfg := range deprecatedConfigs {
		configs = append(configs, convertDeprecatedLoggings(deprecatedCfg)...)
	}

	data, err := json.Marshal(configs)
	if err != nil {
		return "", fmt.Errorf("failed to marshal converted deprecated log configs: %w", err)
	}

	return string(data), nil
}

func parseDeprecatedLogConfigs(str string) ([]*deprecatedLogConfig, error) {
	if str == "" {
		return nil, fmt.Errorf("invalid deprecated logConfigs data")
	}

	var configs []*deprecatedLogConfig
	errArray := json.Unmarshal([]byte(str), &configs)
	if errArray == nil {
		return configs, nil
	}

	var single deprecatedLogConfig
	errSingle := json.Unmarshal([]byte(str), &single)
	if errSingle == nil {
		return []*deprecatedLogConfig{&single}, nil
	}

	return nil, fmt.Errorf("failed to parse deprecated log configs as array (%w) or object (%v)", errArray, errSingle)
}

func convertDeprecatedLoggings(cfg *deprecatedLogConfig) []*logConfig {
	if cfg == nil {
		return nil
	}

	configs := make([]*logConfig, 0)
	for _, logging := range cfg.Loggings {
		for _, path := range logging.LogFiles {
			configs = append(configs, &logConfig{
				Type:                  "file",
				Source:                logging.Source,
				Path:                  path,
				StorageIndex:          logging.StorageIndex,
				Service:               logging.Service,
				CharacterEncoding:     logging.CharacterEncoding,
				Pipeline:              logging.Pipeline,
				Multiline:             logging.MultilineMatch,
				RemoveAnsiEscapeCodes: logging.RemoveAnsiEscapeCodes,
				Tags:                  cloneTags(logging.Tags),
			})
		}
	}

	return configs
}

func cloneTags(tags map[string]string) map[string]string {
	if tags == nil {
		return nil
	}

	cloned := make(map[string]string, len(tags))
	for k, v := range tags {
		cloned[k] = v
	}

	return cloned
}

func getDeprecatedDataKitEndpoint() (string, string, error) {
	if deprecatedJSONConfig == "" {
		return "", "", nil
	}

	deprecatedConfigs, err := parseDeprecatedLogConfigs(deprecatedJSONConfig)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse deprecated datakit_addr: %w", err)
	}

	for _, cfg := range deprecatedConfigs {
		if cfg == nil || cfg.DataKitAddr == "" {
			continue
		}

		host, port, err := splitHostPort(cfg.DataKitAddr)
		if err != nil {
			return "", "", fmt.Errorf("invalid deprecated datakit_addr: %s", cfg.DataKitAddr)
		}

		return host, port, nil
	}

	return "", "", nil
}

func splitHostPort(addr string) (string, string, error) {
	host, port, err := net.SplitHostPort(addr)
	if err == nil {
		return host, port, nil
	}

	if strings.Count(addr, ":") == 1 {
		parts := strings.SplitN(addr, ":", 2)
		if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
			return parts[0], parts[1], nil
		}
	}

	return "", "", fmt.Errorf("invalid host:port: %s", addr)
}

func processConfig(cfg *logConfig) {
	setConfigDefaults(cfg)
	setPodTags(cfg)
	setMultilinePatterns(cfg)
}

func setConfigDefaults(cfg *logConfig) {
	if cfg.Tags == nil {
		cfg.Tags = make(map[string]string)
	}

	if globalSource != "" {
		cfg.Source = globalSource
	}
	if cfg.Source == "" {
		cfg.Source = "default"
	}

	if globalService != "" {
		cfg.Service = globalService
	}
	if cfg.Service == "" {
		cfg.Service = cfg.Source
	}
	cfg.Tags["service"] = cfg.Service

	if globalStorageIndex != "" {
		cfg.StorageIndex = globalStorageIndex
	}
}

func setPodTags(cfg *logConfig) {
	if _, exists := cfg.Tags["pod_name"]; !exists && podName != "" {
		cfg.Tags["pod_name"] = podName
	}
	if _, exists := cfg.Tags["namespace"]; !exists && podNamespace != "" {
		cfg.Tags["namespace"] = podNamespace
	}
	if _, exists := cfg.Tags["pod_ip"]; !exists && podIP != "" {
		cfg.Tags["pod_ip"] = podIP
	}
}

func setTagsFromPodLabels(cfg *logConfig, podLabels map[string]string, podTargetLabels []string) {
	if cfg.Tags == nil {
		cfg.Tags = make(map[string]string)
	}

	for _, key := range podTargetLabels {
		if value, exists := podLabels[key]; exists {
			if _, tagExists := cfg.Tags[key]; !tagExists {
				cfg.Tags[key] = value
			}
		}
	}
}

func setMultilinePatterns(cfg *logConfig) {
	if cfg.Multiline != "" {
		cfg.multilinePatterns = []string{cfg.Multiline}
	} else {
		cfg.multilinePatterns = multiline.GlobalPatterns
	}
}

func readPodLabels(filePath string) (map[string]string, string, error) {
	data, err := os.ReadFile(filePath) // nolint:gosec
	if err != nil {
		return nil, "", fmt.Errorf("failed to read pod labels file: %w", err)
	}

	labelsStr := strings.TrimSpace(string(data))
	labels := parseLabels(labelsStr)
	labelsStrFormatted := formatLabels(labels)
	return labels, labelsStrFormatted, nil
}

func parseLabels(labelsStr string) map[string]string {
	if labelsStr == "" {
		return map[string]string{}
	}

	labels := make(map[string]string)
	splitFn := func(r rune) bool {
		return r == ',' || r == '\n'
	}

	for _, part := range strings.FieldsFunc(labelsStr, splitFn) {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			log.Warnf("invalid label format: %s", part)
			continue
		}

		key := strings.TrimSpace(kv[0])
		if key == "" {
			log.Warnf("empty label key: %s", part)
			continue
		}

		value := strings.TrimSpace(kv[1])
		value = strings.Trim(value, `"`)

		labels[key] = value
	}

	return labels
}

func formatLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}

	// Sort keys for consistent output
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, labels[k]))
	}
	return strings.Join(parts, ",")
}
