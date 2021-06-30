package tailer

import (
	"bytes"
	"regexp"
)

type Multiline struct {
	enabled       bool
	patternRegexp *regexp.Regexp
	buff          bytes.Buffer
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
	if m.IsEnabled() {
		return text
	}

	if m.notMatchString(text) {
		m.buff.WriteString("\n")
		m.buff.WriteString(text)
		return ""
	}

	previousText := m.buff.String()
	m.buff.Reset()
	m.buff.WriteString(text)
	text = previousText

	return text
}

func (m *Multiline) Flush() string {
	if m.buff.Len() == 0 {
		return ""
	}
	text := m.buff.String()
	m.buff.Reset()
	return text
}

func (m *Multiline) notMatchString(text string) bool {
	return !m.patternRegexp.MatchString(text)
}
