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

func TestPriorityToStatus(t *testing.T) {
	tests := []struct {
		name     string
		priority int
		expected string
	}{
		{
			name:     "emergency to error",
			priority: 0,
			expected: "error",
		},
		{
			name:     "alert to warn",
			priority: 1,
			expected: "warn",
		},
		{
			name:     "critical to critical",
			priority: 2,
			expected: "critical",
		},
		{
			name:     "error to error",
			priority: 3,
			expected: "error",
		},
		{
			name:     "warning to warn",
			priority: 4,
			expected: "warn",
		},
		{
			name:     "notice to notice",
			priority: 5,
			expected: "notice",
		},
		{
			name:     "info to info",
			priority: 6,
			expected: "info",
		},
		{
			name:     "debug to debug",
			priority: 7,
			expected: "debug",
		},
		{
			name:     "negative priority to unknown",
			priority: -1,
			expected: "unknown",
		},
		{
			name:     "high priority to unknown",
			priority: 8,
			expected: "unknown",
		},
		{
			name:     "very high priority to unknown",
			priority: 100,
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := priorityToStatus(tt.priority)
			if got != tt.expected {
				t.Errorf("priorityToStatus(%d) = %s, expected %s", tt.priority, got, tt.expected)
			}
		})
	}
}
