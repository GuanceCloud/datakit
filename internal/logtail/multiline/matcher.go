// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package multiline

import (
	"fmt"
	"regexp"
	"sort"
	"unicode"
)

var GlobalPatterns = []string{
	// time.RFC3339, "2006-01-02T15:04:05Z07:00"
	`^\d+-\d+-\d+T\d+:\d+:\d+(\.\d+)?(Z\d*:?\d*)?`,
	// time.ANSIC, "Mon Jan _2 15:04:05 2006"
	`^[A-Za-z_]+ [A-Za-z_]+ +\d+ \d+:\d+:\d+ \d+`,
	// time.RubyDate, "Mon Jan 02 15:04:05 -0700 2006"
	`^[A-Za-z_]+ [A-Za-z_]+ \d+ \d+:\d+:\d+ [\-\+]\d+ \d+`,
	// time.UnixDate, "Mon Jan _2 15:04:05 MST 2006"
	`^[A-Za-z_]+ [A-Za-z_]+ +\d+ \d+:\d+:\d+( [A-Za-z_]+ \d+)?`,
	// time.RFC822, "02 Jan 06 15:04 MST"
	`^\d+ [A-Za-z_]+ \d+ \d+:\d+ [A-Za-z_]+`,
	// time.RFC822Z, "02 Jan 06 15:04 -0700" // RFC822 with numeric zone
	`^\d+ [A-Za-z_]+ \d+ \d+:\d+ -\d+`,
	// time.RFC850, "Monday, 02-Jan-06 15:04:05 MST"
	`^[A-Za-z_]+, \d+-[A-Za-z_]+-\d+ \d+:\d+:\d+ [A-Za-z_]+`,
	// time.RFC1123, "Mon, 02 Jan 2006 15:04:05 MST"
	`^[A-Za-z_]+, \d+ [A-Za-z_]+ \d+ \d+:\d+:\d+ [A-Za-z_]+`,
	// time.RFC1123Z, "Mon, 02 Jan 2006 15:04:05 -0700" // RFC1123 with numeric zone
	`^[A-Za-z_]+, \d+ [A-Za-z_]+ \d+ \d+:\d+:\d+ -\d+`,
	// time.RFC3339Nano, "2006-01-02T15:04:05.999999999Z07:00"
	`^\d+-\d+-\d+[A-Za-z_]+\d+:\d+:\d+\.\d+[A-Za-z_]+\d+:\d+`,
	// 2021-07-08 05:08:19,214
	`^\d+-\d+-\d+ \d+:\d+:\d+(,\d+)?`,
	// Default java logging SimpleFormatter date format
	`^[A-Za-z_]+ \d+, \d+ \d+:\d+:\d+ (AM|PM)`,
	// 2021-01-31 - with stricter matching around the months/days
	`^\d{4}-(0?[1-9]|1[012])-(0?[1-9]|[12][0-9]|3[01])`,
	// gin log, [GIN] 2006/01/02 - 08:53:39
	`^\[GIN\] \d+/\d+/\d+ - \d+:\d+:\d+`,
}

type scoredPattern struct {
	score  int
	regexp *regexp.Regexp
}

func (s *scoredPattern) doMatch(b []byte, str string) bool {
	if len(b) != 0 {
		return s.regexp.Match(b)
	}
	if len(str) != 0 {
		return s.regexp.MatchString(str)
	}
	return false
}

func (s *scoredPattern) String() string {
	return fmt.Sprintf("score:%d, regexp:%s", s.score, s.regexp)
}

type Matcher struct {
	patterns  []*scoredPattern
	noPattern bool
}

func NewMatcher(additionalPatterns []string) (*Matcher, error) {
	if len(additionalPatterns) == 0 {
		return &Matcher{noPattern: true}, nil
	}

	m := &Matcher{
		patterns: make([]*scoredPattern, len(additionalPatterns)),
	}

	for idx, pattern := range additionalPatterns {
		r, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid argument, idx:%d, pattern:'%s', error %w", idx, pattern, err)
		}

		m.patterns[idx] = &scoredPattern{
			score:  0,
			regexp: r,
		}
	}

	return m, nil
}

func (m *Matcher) MatchString(content string) bool {
	if m.noPattern {
		return !prefixIsSpace(nil, content)
	}
	if m.doMatch(nil, content) {
		return true
	}
	if m.patterns[0].score == 0 {
		// use default pattern
		return !prefixIsSpace(nil, content)
	}
	return false
}

func (m *Matcher) Match(content []byte) bool {
	if m.noPattern {
		return !prefixIsSpace(content, "")
	}
	if m.doMatch(content, "") {
		return true
	}
	if m.patterns[0].score == 0 {
		// use default pattern
		return !prefixIsSpace(content, "")
	}
	return false
}

func (m *Matcher) doMatch(b []byte, str string) bool {
	for idx, scoredPattern := range m.patterns {
		match := scoredPattern.doMatch(b, str)
		if match {
			scoredPattern.score++
			if idx != 0 {
				sort.Slice(m.patterns, func(i, j int) bool {
					return m.patterns[i].score > m.patterns[j].score
				})
			}
			return true
		}
	}
	return false
}

func prefixIsSpace(b []byte, str string) bool {
	if len(b) == 0 && len(str) == 0 {
		return true
	}
	var r rune
	if len(b) != 0 {
		r = rune(b[0])
	} else {
		r = rune(str[0])
	}
	// white space is '\t', '\n', '\v', '\f', '\r', ' ', 0x85, 0xA0
	return unicode.IsSpace(r)
}
