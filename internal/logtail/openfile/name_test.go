// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package openfile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitFilenameFromKey(t *testing.T) {
	testCases := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "valid-key",
			key:      "/tmp/test.log::12345",
			expected: "/tmp/test.log",
		},
		{
			name:     "empty-key",
			key:      "",
			expected: "",
		},
		{
			name:     "key-without-separator",
			key:      "/tmp/test.log",
			expected: "/tmp/test.log",
		},
		{
			name:     "key-with-multiple-separators",
			key:      "/tmp/test.log::12345::extra",
			expected: "/tmp/test.log",
		},
		{
			name:     "key-starting-with-separator",
			key:      "::12345",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := SplitFilenameFromKey(tc.key)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetInodeFromKey(t *testing.T) {
	testCases := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "valid-key",
			key:      "/tmp/test.log::12345",
			expected: "12345",
		},
		{
			name:     "empty-key",
			key:      "",
			expected: "",
		},
		{
			name:     "key-without-separator",
			key:      "/tmp/test.log",
			expected: "",
		},
		{
			name:     "key-with-multiple-separators",
			key:      "/tmp/test.log::12345::extra",
			expected: "12345",
		},
		{
			name:     "key-starting-with-separator",
			key:      "::12345",
			expected: "12345",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := InodeFromKey(tc.key)
			assert.Equal(t, tc.expected, result)
		})
	}
}
