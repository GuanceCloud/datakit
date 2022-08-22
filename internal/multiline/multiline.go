// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package multiline wrap regexp/match functions
package multiline

import (
	"bytes"
	"strings"
	"unicode"
	"unicode/utf8"
)

const maxMutilineBytes = 32 * 1024 * 1024

type Multiline struct {
	automult *AutoMultiline
	buff     bytes.Buffer

	// prefixSpace 用以标记 pattern 为空的情况，属于默认行为，
	// 如果一行数据，它的首字符是 WhiteSpace，那它就是多行
	// WhiteSpace 定义为 '\t', '\n', '\v', '\f', '\r', ' ', 0x85, 0xA0
	prefixSpace bool
}

func New(patterns []string) (*Multiline, error) {
	m := &Multiline{}
	var err error

	if len(patterns) == 0 {
		m.prefixSpace = true
	} else {
		m.automult, err = NewAutoMultiline(patterns)
	}

	return m, err
}

var newline = []byte{'\n'}

func (m *Multiline) ProcessLineString(text string) string {
	if !m.matchString(text) {
		if m.buff.Len() > 0 {
			m.buff.WriteString("\n")
		}
		m.buff.WriteString(text)
		if m.buff.Len() >= maxMutilineBytes {
			return m.FlushString()
		}
		return ""
	}

	previousText := m.FlushString()
	m.buff.WriteString(text)

	return previousText
}

func (m *Multiline) ProcessLine(text []byte) []byte {
	if !m.match(text) {
		if m.buff.Len() > 0 {
			m.buff.Write(newline)
		}
		m.buff.Write(text)
		if m.buff.Len() >= maxMutilineBytes {
			return m.Flush()
		}
		return nil
	}

	previousText := m.Flush()
	m.buff.Write(text)

	return previousText
}

func (m *Multiline) BuffLength() int {
	return m.buff.Len()
}

func (m *Multiline) Flush() []byte {
	if m.buff.Len() == 0 {
		return nil
	}

	text := make([]byte, m.buff.Len())
	copy(text, m.buff.Bytes())

	m.buff.Reset()
	return text
}

func (m *Multiline) match(text []byte) bool {
	if m.prefixSpace {
		return m.matchOfPrefixSpace(text)
	}
	return m.automult.Match(text)
}

func (m *Multiline) matchString(text string) bool {
	return m.match([]byte(text))
}

func (m *Multiline) matchOfPrefixSpace(text []byte) bool {
	if len(text) == 0 {
		return true
	}
	return !unicode.IsSpace(rune(text[0]))
}

func (m *Multiline) FlushString() string {
	if m.buff.Len() == 0 {
		return ""
	}
	text := m.buff.String()
	m.buff.Reset()
	return text
}

func (m *Multiline) BuffString() string {
	return m.buff.String()
}

var asciiSpace = [256]uint8{'\t': 1, '\n': 1, '\v': 1, '\f': 1, '\r': 1, ' ': 1}

func TrimRightSpace(s string) string {
	end := len(s)
	for ; end > 0; end-- {
		c := s[end-1]
		if c >= utf8.RuneSelf {
			return strings.TrimFunc(s[:end], unicode.IsSpace)
		}
		if asciiSpace[c] == 0 {
			break
		}
	}
	return s[:end]
}
