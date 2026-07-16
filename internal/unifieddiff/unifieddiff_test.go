// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package unifieddiff

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestText(t *testing.T) {
	assert.Equal(t, `--- a/etc/example.conf
+++ b/etc/example.conf
@@ -1,2 +1,2 @@
 keep
-old
+new
`, Text("/etc/example.conf", "keep\nold\n", "keep\nnew\n"))

	assert.Empty(t, Text("/etc/example.conf", "same\n", "same\n"))
}

func TestTextFormatsEmptyRanges(t *testing.T) {
	tests := []struct {
		name     string
		oldText  string
		newText  string
		expected string
	}{
		{
			name:    "add content",
			newText: "new\n",
			expected: `--- a/etc/example.conf
+++ b/etc/example.conf
@@ -0,0 +1 @@
+new
`,
		},
		{
			name:    "clear content",
			oldText: "old\n",
			expected: `--- a/etc/example.conf
+++ b/etc/example.conf
@@ -1 +0,0 @@
-old
`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, Text("/etc/example.conf", test.oldText, test.newText))
		})
	}
}

func TestTextKeepsLongSingleLine(t *testing.T) {
	oldText := strings.Repeat("a", 8*1024) + "-old"
	newText := strings.Repeat("a", 8*1024) + "-new"
	diffText := Text("annotations", oldText, newText)

	assert.Contains(t, diffText, "-"+oldText)
	assert.Contains(t, diffText, "+"+newText)
	assert.NotContains(t, diffText, "truncated")
}

func TestTextFallsBackToFullReplacementForLargeLineMatrix(t *testing.T) {
	oldText := strings.Repeat("same\n", 1000) + "old\n" + strings.Repeat("same\n", 1000)
	newText := strings.Repeat("same\n", 1000) + "new\n" + strings.Repeat("same\n", 1000)

	diffText := Text("large.conf", oldText, newText)

	assert.Contains(t, diffText, "@@ -1,2001 +1,2001 @@\n")
}
