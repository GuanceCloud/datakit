// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package fileprovider

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilter(t *testing.T) {
	testcases := []struct {
		includePatterns []string
		excludePatterns []string
		in, out         []string
	}{
		{
			includePatterns: []string{"/tmp/123"},
			in:              []string{"/tmp/123"},
			out:             []string{"/tmp/123"},
		},
		{
			includePatterns: []string{"/tmp/*"},
			in:              []string{"/tmp/123"},
			out:             []string{"/tmp/123"},
		},
		{
			includePatterns: []string{"/tmp/**/*"},
			in:              []string{"/tmp/abc/123"},
			out:             []string{"/tmp/abc/123"},
		},
		{
			includePatterns: []string{"/tmp/**/*.log"},
			in:              []string{"/tmp/123.log"},
			out:             []string{"/tmp/123.log"},
		},
		{
			includePatterns: []string{"/tmp/**/*.log"},
			in:              []string{"/tmp/abc/123.log"},
			out:             []string{"/tmp/abc/123.log"},
		},
		{
			includePatterns: []string{"/tmp/**/hjk/*.log"},
			in:              []string{"/tmp/abc/def/hjk/123.log"},
			out:             []string{"/tmp/abc/def/hjk/123.log"},
		},
		{
			includePatterns: []string{"/tmp/**/hjk/*.log"},
			in:              []string{"/tmp/abc/def/123.log"},
			out:             nil,
		},
		{
			excludePatterns: []string{"/tmp/abc"},
			in:              []string{"/tmp/123"},
			out:             []string{"/tmp/123"},
		},
		{
			excludePatterns: []string{"/tmp/*"},
			in:              []string{"/tmp/123"},
			out:             nil,
		},
		{
			excludePatterns: []string{"/tmp/**/*"},
			in:              []string{"/tmp/abc/123"},
			out:             nil,
		},
		{
			excludePatterns: []string{"/tmp/**/*.log"},
			in:              []string{"/tmp/abc/123.log"},
			out:             nil,
		},
		{
			excludePatterns: []string{"/tmp/**/*.txt"},
			in:              []string{"/tmp/abc/123.log"},
			out:             []string{"/tmp/abc/123.log"},
		},
		{
			excludePatterns: []string{"C:/Users/admin/Desktop/tmp/*"},
			in:              []string{"C:/Users/admin/Desktop/tmp/123"},
			out:             nil,
		},
	}

	for _, tc := range testcases {
		ex, err := NewGlobFilter(tc.includePatterns, tc.excludePatterns)
		assert.NoError(t, err)

		res := ex.IncludeFilterFiles(tc.in)
		res = ex.ExcludeFilterFiles(res)
		assert.Equal(t, tc.out, res)
	}
}
