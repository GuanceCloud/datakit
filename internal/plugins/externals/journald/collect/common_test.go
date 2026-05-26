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

func TestOption_ParseTags(t *testing.T) {
	tests := []struct {
		name     string
		tags     string
		expected map[string]string
	}{
		{
			name:     "empty tags",
			tags:     "",
			expected: map[string]string{},
		},
		{
			name: "single tag",
			tags: "env=production",
			expected: map[string]string{
				"env": "production",
			},
		},
		{
			name: "multiple tags",
			tags: "env=production,region=us-west,app=nginx",
			expected: map[string]string{
				"env":    "production",
				"region": "us-west",
				"app":    "nginx",
			},
		},
		{
			name: "tags with spaces",
			tags: "env = production , region = us-west , app = nginx",
			expected: map[string]string{
				"env":    "production",
				"region": "us-west",
				"app":    "nginx",
			},
		},
		{
			name: "tags with empty value",
			tags: "env=production,region=,app=nginx",
			expected: map[string]string{
				"env":    "production",
				"region": "", // Empty values are included
				"app":    "nginx",
			},
		},
		{
			name: "tags with malformed pair",
			tags: "env=production,malformed,app=nginx",
			expected: map[string]string{
				"env": "production",
				"app": "nginx",
			},
		},
		{
			name: "tags with duplicate keys",
			tags: "env=production,env=staging",
			expected: map[string]string{
				"env": "staging", // Last occurrence wins
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := &Option{Tags: tt.tags}
			got := opt.parseTags()

			if len(got) != len(tt.expected) {
				t.Errorf("ParseTags() returned %d tags, expected %d", len(got), len(tt.expected))
			}

			for key, expectedValue := range tt.expected {
				if gotValue, ok := got[key]; !ok {
					t.Errorf("ParseTags() missing expected key %s", key)
				} else if gotValue != expectedValue {
					t.Errorf("ParseTags() for key %s = %s, expected %s", key, gotValue, expectedValue)
				}
			}

			// Check for unexpected keys
			for key := range got {
				if _, ok := tt.expected[key]; !ok {
					t.Errorf("ParseTags() returned unexpected key %s", key)
				}
			}
		})
	}
}

func TestOption_ParseTagsEdgeCases(t *testing.T) {
	tests := []struct {
		name string
		tags string
	}{
		{
			name: "only equals sign",
			tags: "=",
		},
		{
			name: "multiple equals signs",
			tags: "key===value",
		},
		{
			name: "empty key",
			tags: "=value",
		},
		{
			name: "trailing comma",
			tags: "env=production,",
		},
		{
			name: "leading comma",
			tags: ",env=production",
		},
		{
			name: "multiple commas",
			tags: "env=production,,,region=us-west",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := &Option{Tags: tt.tags}
			got := opt.parseTags()

			// Just ensure it doesn't panic and returns something
			if got == nil {
				t.Error("ParseTags() returned nil map")
			}
		})
	}
}
