// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package multiline

func (m *Matcher) matchAutoPatterns(b []byte, str string) bool {
	first := firstByte(b, str)
	if first < 0 {
		return false
	}
	if isBlank(first) {
		return m.doMatchPatterns(m.extraPatterns, b, str)
	}

	var patterns []*scoredPattern
	switch {
	case isDigit(first):
		patterns = m.digitPatterns
	case isLetter(first):
		patterns = m.letterPatterns
	default:
		patterns = m.symbolPatterns
	}

	return m.doMatchPatterns(patterns, b, str) ||
		m.doMatchPatterns(m.extraPatterns, b, str)
}

func firstByte(b []byte, str string) int {
	if len(b) != 0 {
		return int(b[0])
	}
	if len(str) != 0 {
		return int(str[0])
	}
	return -1
}

func isDigit(c int) bool {
	return '0' <= c && c <= '9'
}

func isLetter(c int) bool {
	return ('A' <= c && c <= 'Z') || ('a' <= c && c <= 'z')
}

func isBlank(c int) bool {
	return c == ' ' || c == '\t'
}
