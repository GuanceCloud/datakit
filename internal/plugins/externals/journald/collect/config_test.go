// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package collect

import (
	"testing"
)

func TestConfigParsing(t *testing.T) {
	tests := []struct {
		name          string
		opt           *Option
		expectedPaths []string
		expectedUnits []string
	}{
		{
			name: "default paths when empty",
			opt: &Option{
				Paths: "",
			},
			expectedPaths: []string{"/var/log/journal", "/run/log/journal"},
			expectedUnits: []string{},
		},
		{
			name: "single path",
			opt: &Option{
				Paths: "/var/log/journal",
			},
			expectedPaths: []string{"/var/log/journal"},
			expectedUnits: []string{},
		},
		{
			name: "multiple paths",
			opt: &Option{
				Paths: "/var/log/journal,/run/log/journal,/custom/path",
			},
			expectedPaths: []string{"/var/log/journal", "/run/log/journal", "/custom/path"},
			expectedUnits: []string{},
		},
		{
			name: "single unit",
			opt: &Option{
				Units: "nginx.service",
			},
			expectedPaths: []string{"/var/log/journal", "/run/log/journal"},
			expectedUnits: []string{"nginx.service"},
		},
		{
			name: "multiple units",
			opt: &Option{
				Units: "nginx.service,docker.service,kubelet.service",
			},
			expectedPaths: []string{"/var/log/journal", "/run/log/journal"},
			expectedUnits: []string{"nginx.service", "docker.service", "kubelet.service"},
		},
		{
			name: "paths and units",
			opt: &Option{
				Paths:      "/custom/journal",
				Units:      "nginx.service,docker.service",
				Priorities: "err,warning",
			},
			expectedPaths: []string{"/custom/journal"},
			expectedUnits: []string{"nginx.service", "docker.service"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create input to trigger config parsing
			ipt := NewInput(tt.opt).(*Input)

			// Verify paths
			if len(ipt.config.Paths) != len(tt.expectedPaths) {
				t.Errorf("config.Paths length = %d, expected %d", len(ipt.config.Paths), len(tt.expectedPaths))
			} else {
				for i, path := range ipt.config.Paths {
					if path != tt.expectedPaths[i] {
						t.Errorf("config.Paths[%d] = %s, expected %s", i, path, tt.expectedPaths[i])
					}
				}
			}
		})
	}
}

func TestConfig_ExcludeFields(t *testing.T) {
	tests := []struct {
		name           string
		excludeFields  string
		expectedCount  int
		expectedFields []string
	}{
		{
			name:           "empty exclude fields",
			excludeFields:  "",
			expectedCount:  0,
			expectedFields: []string{},
		},
		{
			name:           "single field",
			excludeFields:  "_BOOT_ID",
			expectedCount:  1,
			expectedFields: []string{"_BOOT_ID"},
		},
		{
			name:           "multiple fields",
			excludeFields:  "_BOOT_ID,_MACHINE_ID,__MONOTONIC_TIMESTAMP",
			expectedCount:  3,
			expectedFields: []string{"_BOOT_ID", "_MACHINE_ID", "__MONOTONIC_TIMESTAMP"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := &Option{
				ExcludeFields: tt.excludeFields,
			}
			ipt := NewInput(opt).(*Input)

			if len(ipt.config.ExcludeFields) != tt.expectedCount {
				t.Errorf("config.ExcludeFields length = %d, expected %d", len(ipt.config.ExcludeFields), tt.expectedCount)
			}

			for i, field := range tt.expectedFields {
				if i < len(ipt.config.ExcludeFields) && ipt.config.ExcludeFields[i] != field {
					t.Errorf("config.ExcludeFields[%d] = %s, expected %s", i, ipt.config.ExcludeFields[i], field)
				}
			}
		})
	}
}

func TestConfig_BatchSize(t *testing.T) {
	tests := []struct {
		name         string
		maxEntries   int
		expectedSize int
	}{
		{
			name:         "default batch size",
			maxEntries:   0,
			expectedSize: 1000,
		},
		{
			name:         "custom batch size",
			maxEntries:   500,
			expectedSize: 500,
		},
		{
			name:         "large batch size",
			maxEntries:   5000,
			expectedSize: 5000,
		},
		{
			name:         "small batch size",
			maxEntries:   100,
			expectedSize: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := &Option{
				MaxEntries: tt.maxEntries,
			}
			ipt := NewInput(opt).(*Input)

			if ipt.config.MaxEntriesPerBatch != tt.expectedSize {
				t.Errorf("config.MaxEntriesPerBatch = %d, expected %d", ipt.config.MaxEntriesPerBatch, tt.expectedSize)
			}
		})
	}
}

func TestConfig_CursorOptions(t *testing.T) {
	tests := []struct {
		name         string
		saveCursor   bool
		cursorFile   string
		expectedSave bool
		expectedFile string
	}{
		{
			name:         "cursor disabled",
			saveCursor:   false,
			cursorFile:   "",
			expectedSave: false,
			expectedFile: "",
		},
		{
			name:         "cursor enabled with file",
			saveCursor:   true,
			cursorFile:   "/tmp/journald.cursor",
			expectedSave: true,
			expectedFile: "/tmp/journald.cursor",
		},
		{
			name:         "cursor enabled without file",
			saveCursor:   true,
			cursorFile:   "",
			expectedSave: true,
			expectedFile: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := &Option{
				SaveCursor: tt.saveCursor,
				CursorFile: tt.cursorFile,
			}
			ipt := NewInput(opt).(*Input)

			if ipt.config.SaveCursor != tt.expectedSave {
				t.Errorf("config.SaveCursor = %v, expected %v", ipt.config.SaveCursor, tt.expectedSave)
			}

			if ipt.config.CursorFile != tt.expectedFile {
				t.Errorf("config.CursorFile = %s, expected %s", ipt.config.CursorFile, tt.expectedFile)
			}
		})
	}
}

func TestConfig_TailOnly(t *testing.T) {
	tests := []struct {
		name        string
		tailOnly    bool
		expectedVal bool
	}{
		{
			name:        "tail only enabled",
			tailOnly:    true,
			expectedVal: true,
		},
		{
			name:        "tail only disabled",
			tailOnly:    false,
			expectedVal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := &Option{
				TailOnly: tt.tailOnly,
			}
			ipt := NewInput(opt).(*Input)

			if ipt.config.TailOnly != tt.expectedVal {
				t.Errorf("config.TailOnly = %v, expected %v", ipt.config.TailOnly, tt.expectedVal)
			}
		})
	}
}

func TestInput_Tags(t *testing.T) {
	tests := []struct {
		name         string
		tags         string
		expectedTags map[string]string
	}{
		{
			name:         "empty tags",
			tags:         "",
			expectedTags: map[string]string{},
		},
		{
			name:         "single tag",
			tags:         "env=production",
			expectedTags: map[string]string{"env": "production"},
		},
		{
			name:         "multiple tags",
			tags:         "env=production,region=us-west,app=nginx",
			expectedTags: map[string]string{"env": "production", "region": "us-west", "app": "nginx"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := &Option{
				Tags: tt.tags,
			}
			ipt := NewInput(opt).(*Input)

			if len(ipt.tags) != len(tt.expectedTags) {
				t.Errorf("input.tags length = %d, expected %d", len(ipt.tags), len(tt.expectedTags))
			}

			for key, expectedValue := range tt.expectedTags {
				if gotValue, ok := ipt.tags[key]; !ok {
					t.Errorf("input.tags missing expected key %s", key)
				} else if gotValue != expectedValue {
					t.Errorf("input.tags[%s] = %s, expected %s", key, gotValue, expectedValue)
				}
			}
		})
	}
}

func TestInput_DataKitEndpoint(t *testing.T) {
	tests := []struct {
		name               string
		endpoint           string
		expectedURLPath    string
		expectedHasURLPath bool
	}{
		{
			name:               "empty endpoint",
			endpoint:           "",
			expectedURLPath:    "",
			expectedHasURLPath: false,
		},
		{
			name:               "custom endpoint",
			endpoint:           "http://custom-datakit:9529",
			expectedURLPath:    "http://custom-datakit:9529/v1/write/logging?input=journald",
			expectedHasURLPath: true,
		},
		{
			name:               "default endpoint",
			endpoint:           "http://localhost:9529",
			expectedURLPath:    "http://localhost:9529/v1/write/logging?input=journald",
			expectedHasURLPath: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := &Option{
				DatakitHTTPEndpoint: tt.endpoint,
			}
			ipt := NewInput(opt).(*Input)

			if (ipt.dkURLPath != "") != tt.expectedHasURLPath {
				t.Errorf("input.dkURLPath presence = %v, expected %v", ipt.dkURLPath != "", tt.expectedHasURLPath)
			}

			if ipt.dkURLPath != tt.expectedURLPath {
				t.Errorf("input.dkURLPath = %s, expected %s", ipt.dkURLPath, tt.expectedURLPath)
			}
		})
	}
}
