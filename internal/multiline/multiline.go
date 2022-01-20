// Package multiline wrap regexp/match functions
package multiline

import (
	"bytes"
	"regexp"
	"unicode"
)

type Multiline struct {
	patternRegexp *regexp.Regexp
	buff          bytes.Buffer
	lines         int
	maxLines      int

	// prefixSpace 用以标记 pattern 为空的情况，属于默认行为，
	// 如果一行数据，它的首字符是 WhiteSpace，那它就是多行
	// WhiteSpace 定义为 '\t', '\n', '\v', '\f', '\r', ' ', 0x85, 0xA0
	prefixSpace bool
}

func New(pattern string, maxLines int) (*Multiline, error) {
	var r *regexp.Regexp
	var err error

	if pattern == "" {
		return &Multiline{maxLines: maxLines, prefixSpace: true}, nil
	}

	if r, err = regexp.Compile(pattern); err != nil {
		return nil, err
	}
	return &Multiline{
		patternRegexp: r,
		maxLines:      maxLines,
	}, nil
}

var newline = []byte{'\n'}

func (m *Multiline) ProcessLine(text []byte) []byte {
	if !m.match(text) {
		if m.lines != 0 {
			m.buff.Write(newline)
		}
		m.buff.Write(text)
		m.lines++

		if m.lines >= m.maxLines {
			return m.Flush()
		}
		return nil
	}

	previousText := m.Flush()
	m.buff.Write(text)
	m.lines++

	return previousText
}

func (m *Multiline) ProcessLineString(text string) string {
	if !m.matchString(text) {
		if m.lines != 0 {
			m.buff.WriteString("\n")
		}
		m.buff.WriteString(text)
		m.lines++

		if m.lines >= m.maxLines {
			return m.FlushString()
		}
		return ""
	}

	previousText := m.FlushString()
	m.buff.WriteString(text)
	m.lines++

	return previousText
}

func (m *Multiline) CacheLines() int {
	return m.lines
}

func (m *Multiline) Flush() []byte {
	if m.buff.Len() == 0 {
		return nil
	}

	text := make([]byte, m.buff.Len())
	copy(text, m.buff.Bytes())

	m.buff.Reset()
	m.lines = 0

	return text
}

func (m *Multiline) match(text []byte) bool {
	if m.prefixSpace {
		return m.matchOfPrefixSpace(text)
	}
	return m.patternRegexp.Match(text)
}

func (m *Multiline) matchOfPrefixSpace(text []byte) bool {
	if len(text) == 0 {
		return true
	}
	return unicode.IsSpace(rune(text[0]))
}

func (m *Multiline) FlushString() string {
	if m.buff.Len() == 0 {
		return ""
	}
	text := m.buff.String()
	m.buff.Reset()
	m.lines = 0
	return text
}

func (m *Multiline) matchString(text string) bool {
	if m.prefixSpace {
		return m.matchStringOfPrefixSpace(text)
	}
	return m.patternRegexp.MatchString(text)
}

func (m *Multiline) matchStringOfPrefixSpace(text string) bool {
	if len(text) == 0 {
		return true
	}
	return unicode.IsSpace(rune(text[0]))
}
