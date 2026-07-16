// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package unifieddiff generates standard unified text diffs.
package unifieddiff

import (
	"strconv"
	"strings"

	"github.com/pmezard/go-difflib/difflib"
)

// SequenceMatcher has quadratic worst-case runtime. Above this line matrix
// size, emit a full replacement diff so runtime remains linear in input size.
const maxLineMatrixSize = 4_000_000

// Text returns a unified diff that transforms oldText into newText.
// Path is used only as the display name in the diff headers.
func Text(path, oldText, newText string) string {
	if oldText == newText {
		return ""
	}

	path = normalizePath(path)
	from := "a/" + path
	to := "b/" + path
	oldLines := splitLines(oldText)
	newLines := splitLines(newText)
	if exceedsLineMatrixLimit(len(oldLines), len(newLines)) {
		return fullReplacementDiff(from, to, oldLines, newLines)
	}

	diffText, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        oldLines,
		B:        newLines,
		FromFile: from,
		ToFile:   to,
		Context:  3,
	})
	if err != nil {
		return ""
	}

	return diffText
}

func exceedsLineMatrixLimit(oldCount, newCount int) bool {
	return oldCount > 0 && newCount > maxLineMatrixSize/oldCount
}

func fullReplacementDiff(from, to string, oldLines, newLines []string) string {
	var diffText strings.Builder
	diffText.WriteString("--- ")
	diffText.WriteString(from)
	diffText.WriteString("\n+++ ")
	diffText.WriteString(to)
	diffText.WriteString("\n@@ -")
	diffText.WriteString(fullRange(len(oldLines)))
	diffText.WriteString(" +")
	diffText.WriteString(fullRange(len(newLines)))
	diffText.WriteString(" @@\n")

	for _, line := range oldLines {
		diffText.WriteByte('-')
		diffText.WriteString(line)
	}
	for _, line := range newLines {
		diffText.WriteByte('+')
		diffText.WriteString(line)
	}

	return diffText.String()
}

func fullRange(lineCount int) string {
	switch lineCount {
	case 0:
		return "0,0"
	case 1:
		return "1"
	default:
		return "1," + strconv.Itoa(lineCount)
	}
}

func splitLines(text string) []string {
	if text == "" {
		return nil
	}

	lines := strings.SplitAfter(text, "\n")
	if lines[len(lines)-1] == "" {
		return lines[:len(lines)-1]
	}

	lines[len(lines)-1] += "\n\\ No newline at end of file\n"
	return lines
}

func normalizePath(path string) string {
	path = strings.ReplaceAll(path, `\`, "/")
	path = strings.TrimLeft(path, "/")
	path = strings.ReplaceAll(path, "\r", "_")
	path = strings.ReplaceAll(path, "\n", "_")
	if path == "" || path == "." {
		return "content"
	}

	return path
}
