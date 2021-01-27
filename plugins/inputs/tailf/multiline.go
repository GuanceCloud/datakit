package tailf

import (
	"bytes"
	"regexp"
)

type Multiline struct {
	config        *MultilineConfig
	enabled       bool
	patternRegexp *regexp.Regexp
}

type MultilineConfig struct {
	Pattern        string `toml:"pattern"`
	MatchWhichLine string `toml:"match_which_line"`
	InvertMatch    bool   `toml:"invert_match"`
}

const (
	// Previous => Append current line to previous line
	Previous = "previous"
	// Next => Next line will be appended to current line
	Next = "next"
)

func (m *MultilineConfig) NewMultiline() (*Multiline, error) {
	enabled := false
	var r *regexp.Regexp
	var err error

	// FIXME: use constant
	// m.InvertMatch = true
	// m.MatchWhichLine = "previous"

	if m.Pattern != "" {
		enabled = true
		if r, err = regexp.Compile(m.Pattern); err != nil {
			return nil, err
		}

		if m.MatchWhichLine != Previous && m.MatchWhichLine != Next {
			m.MatchWhichLine = Previous
		}
	}

	return &Multiline{
		config:        m,
		enabled:       enabled,
		patternRegexp: r,
	}, nil
}

func (m *Multiline) IsEnabled() bool {
	return m.enabled
}

func (m *Multiline) ProcessLine(text string, buffer *bytes.Buffer) string {
	if m.matchString(text) {
		buffer.WriteString(text)
		buffer.WriteString("\n")
		return ""
	}

	if m.config.MatchWhichLine == Previous {
		previousText := buffer.String()
		buffer.Reset()
		buffer.WriteString(text)
		text = previousText
	} else {
		// Next
		if buffer.Len() > 0 {
			buffer.WriteString(text)
			text = buffer.String()
			buffer.Reset()
		}
	}

	return text
}

func (m *Multiline) Flush(buffer *bytes.Buffer) string {
	if buffer.Len() == 0 {
		return ""
	}
	text := buffer.String()
	buffer.Reset()
	return text
}

func (m *Multiline) matchString(text string) bool {
	return m.patternRegexp.MatchString(text) != m.config.InvertMatch
}
