// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package logfwd

import (
	"reflect"
	"testing"
)

func TestParseLabels(t *testing.T) {
	cases := []struct {
		in  string
		out map[string]string
	}{
		{
			in: "app=\"logging\"\nkind_name=\"testing-crd-pod-target-labels\"",
			out: map[string]string{
				"app":       "logging",
				"kind_name": "testing-crd-pod-target-labels",
			},
		},
	}

	for _, tc := range cases {
		res := parseLabels(tc.in)
		if !reflect.DeepEqual(res, tc.out) {
			t.Errorf("parseLabels(%q) = %#v, want %#v", tc.in, res, tc.out)
		}
	}
}

func TestGetEffectiveEnvLogConfigsPreferNew(t *testing.T) {
	prevEnvLogConfigsStr := envLogConfigsStr
	prevDeprecatedJSONConfig := deprecatedJSONConfig
	defer func() {
		envLogConfigsStr = prevEnvLogConfigsStr
		deprecatedJSONConfig = prevDeprecatedJSONConfig
	}()

	envLogConfigsStr = `[{"type":"file","path":"/var/log/new.log","source":"new-source"}]`
	deprecatedJSONConfig = `{"datakit_addr":"127.0.0.1:9533","loggings":[{"logfiles":["/var/log/old.log"],"source":"old-source"}]}`

	got, err := getEffectiveEnvLogConfigs()
	if err != nil {
		t.Fatalf("getEffectiveEnvLogConfigs() error = %v", err)
	}

	if got != envLogConfigsStr {
		t.Fatalf("getEffectiveEnvLogConfigs() = %s, want %s", got, envLogConfigsStr)
	}
}

func TestGetEffectiveEnvLogConfigsFallbackDeprecated(t *testing.T) {
	prevEnvLogConfigsStr := envLogConfigsStr
	prevDeprecatedJSONConfig := deprecatedJSONConfig
	prevGlobalSource := globalSource
	prevGlobalService := globalService
	prevGlobalStorageIndex := globalStorageIndex
	defer func() {
		envLogConfigsStr = prevEnvLogConfigsStr
		deprecatedJSONConfig = prevDeprecatedJSONConfig
		globalSource = prevGlobalSource
		globalService = prevGlobalService
		globalStorageIndex = prevGlobalStorageIndex
	}()

	envLogConfigsStr = ""
	deprecatedJSONConfig = `{"datakit_addr":"127.0.0.1:9533","loggings":[{"logfiles":["/var/log/a.log","/var/log/b.log"],"ignore":["*.tmp"],"source":"legacy-source","storage_index":"legacy-index","service":"legacy-service","pipeline":"legacy.p","character_encoding":"utf-8","multiline_match":"^\\d{4}","remove_ansi_escape_codes":true,"tags":{"env":"prod"}}]}`
	globalSource = ""
	globalService = ""
	globalStorageIndex = ""

	raw, err := getEffectiveEnvLogConfigs()
	if err != nil {
		t.Fatalf("getEffectiveEnvLogConfigs() error = %v", err)
	}

	got, err := parseLogConfigs(raw)
	if err != nil {
		t.Fatalf("parseLogConfigs() error = %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("len(configs) = %d, want 2", len(got))
	}

	if got[0].Type != "file" || got[1].Type != "file" {
		t.Fatalf("unexpected type, got %#v %#v", got[0].Type, got[1].Type)
	}

	if got[0].Path != "/var/log/a.log" || got[1].Path != "/var/log/b.log" {
		t.Fatalf("unexpected converted paths: %#v %#v", got[0].Path, got[1].Path)
	}

	for i := range got {
		if got[i].Source != "legacy-source" {
			t.Fatalf("config[%d].Source = %s, want legacy-source", i, got[i].Source)
		}
		if got[i].Service != "legacy-service" {
			t.Fatalf("config[%d].Service = %s, want legacy-service", i, got[i].Service)
		}
		if got[i].StorageIndex != "legacy-index" {
			t.Fatalf("config[%d].StorageIndex = %s, want legacy-index", i, got[i].StorageIndex)
		}
		if got[i].Pipeline != "legacy.p" {
			t.Fatalf("config[%d].Pipeline = %s, want legacy.p", i, got[i].Pipeline)
		}
		if got[i].CharacterEncoding != "utf-8" {
			t.Fatalf("config[%d].CharacterEncoding = %s, want utf-8", i, got[i].CharacterEncoding)
		}
		if got[i].Multiline != "^\\d{4}" {
			t.Fatalf("config[%d].Multiline = %s, want ^\\\\d{4}", i, got[i].Multiline)
		}
		if !got[i].RemoveAnsiEscapeCodes {
			t.Fatalf("config[%d].RemoveAnsiEscapeCodes = false, want true", i)
		}
		if got[i].Tags["env"] != "prod" {
			t.Fatalf("config[%d].Tags[env] = %s, want prod", i, got[i].Tags["env"])
		}
	}
}

func TestGetEffectiveEnvLogConfigsDeprecatedParseError(t *testing.T) {
	prevEnvLogConfigsStr := envLogConfigsStr
	prevDeprecatedJSONConfig := deprecatedJSONConfig
	defer func() {
		envLogConfigsStr = prevEnvLogConfigsStr
		deprecatedJSONConfig = prevDeprecatedJSONConfig
	}()

	envLogConfigsStr = ""
	deprecatedJSONConfig = `{bad-json`

	_, err := getEffectiveEnvLogConfigs()
	if err == nil {
		t.Fatal("getEffectiveEnvLogConfigs() error = nil, want non-nil")
	}
}

func TestGetEffectiveEnvLogConfigsFallbackDeprecatedArray(t *testing.T) {
	prevEnvLogConfigsStr := envLogConfigsStr
	prevDeprecatedJSONConfig := deprecatedJSONConfig
	prevGlobalSource := globalSource
	prevGlobalService := globalService
	prevGlobalStorageIndex := globalStorageIndex
	defer func() {
		envLogConfigsStr = prevEnvLogConfigsStr
		deprecatedJSONConfig = prevDeprecatedJSONConfig
		globalSource = prevGlobalSource
		globalService = prevGlobalService
		globalStorageIndex = prevGlobalStorageIndex
	}()

	envLogConfigsStr = ""
	deprecatedJSONConfig = `[{"datakit_addr":"127.0.0.1:9533","loggings":[{"logfiles":["/var/log/array.log"],"source":"legacy-array"}]}]`
	globalSource = ""
	globalService = ""
	globalStorageIndex = ""

	raw, err := getEffectiveEnvLogConfigs()
	if err != nil {
		t.Fatalf("getEffectiveEnvLogConfigs() error = %v", err)
	}

	got, err := parseLogConfigs(raw)
	if err != nil {
		t.Fatalf("parseLogConfigs() error = %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("len(configs) = %d, want 1", len(got))
	}

	if got[0].Path != "/var/log/array.log" {
		t.Fatalf("config[0].Path = %s, want /var/log/array.log", got[0].Path)
	}
}

func TestGetEndpointConfigPreferEnvOverDeprecated(t *testing.T) {
	prevDatakitHost := datakitHost
	prevDatakitPort := datakitPort
	prevDeprecatedJSONConfig := deprecatedJSONConfig
	prevOperatorEndpoint := operatorEndpoint
	defer func() {
		datakitHost = prevDatakitHost
		datakitPort = prevDatakitPort
		deprecatedJSONConfig = prevDeprecatedJSONConfig
		operatorEndpoint = prevOperatorEndpoint
	}()

	datakitHost = "127.0.0.1"
	datakitPort = "9533"
	deprecatedJSONConfig = `{"datakit_addr":"10.0.0.1:9999","loggings":[{"logfiles":["/var/log/a.log"],"source":"legacy"}]}`
	operatorEndpoint = ""

	got, _, err := getEndpointConfig()
	if err != nil {
		t.Fatalf("getEndpointConfig() error = %v", err)
	}

	if got != "127.0.0.1:9533" {
		t.Fatalf("getEndpointConfig() datakit endpoint = %s, want 127.0.0.1:9533", got)
	}
}

func TestGetEndpointConfigFallbackDeprecated(t *testing.T) {
	prevDatakitHost := datakitHost
	prevDatakitPort := datakitPort
	prevDeprecatedJSONConfig := deprecatedJSONConfig
	prevOperatorEndpoint := operatorEndpoint
	defer func() {
		datakitHost = prevDatakitHost
		datakitPort = prevDatakitPort
		deprecatedJSONConfig = prevDeprecatedJSONConfig
		operatorEndpoint = prevOperatorEndpoint
	}()

	datakitHost = ""
	datakitPort = ""
	deprecatedJSONConfig = `{"datakit_addr":"127.0.0.2:9534","loggings":[{"logfiles":["/var/log/a.log"],"source":"legacy"}]}`
	operatorEndpoint = ""

	got, _, err := getEndpointConfig()
	if err != nil {
		t.Fatalf("getEndpointConfig() error = %v", err)
	}

	if got != "127.0.0.2:9534" {
		t.Fatalf("getEndpointConfig() datakit endpoint = %s, want 127.0.0.2:9534", got)
	}
}
