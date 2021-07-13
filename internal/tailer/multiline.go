package tailer

import (
	"bytes"
	"regexp"
)

var maxLines = 30

type Multiline struct {
	enabled       bool
	patternRegexp *regexp.Regexp
	buff          bytes.Buffer
	lines         int
}

func NewMultiline(pattern string) (*Multiline, error) {
	enabled := false
	var r *regexp.Regexp
	var err error

	if pattern != "" {
		enabled = true
		if r, err = regexp.Compile(pattern); err != nil {
			return nil, err
		}
	}

	return &Multiline{
		enabled:       enabled,
		patternRegexp: r,
	}, nil
}

func (m *Multiline) IsEnabled() bool {
	return m.enabled
}

func (m *Multiline) ProcessLine(text string) string {
	if !m.IsEnabled() {
		return text
	}

	if m.notMatchString(text) {
		if m.lines != 0 {
			m.buff.WriteString("\n")
		}
		m.buff.WriteString(text)
		m.lines++

		if m.lines >= maxLines {
			return m.Flush()
		}
		return ""
	}

	previousText := m.Flush()
	m.buff.WriteString(text)
	m.lines++

	return previousText
}

func (m *Multiline) Flush() string {
	if m.buff.Len() == 0 {
		return ""
	}
	text := m.buff.String()
	m.buff.Reset()
	m.lines = 0
	return text
}

func (m *Multiline) notMatchString(text string) bool {
	return !m.patternRegexp.MatchString(text)
}
