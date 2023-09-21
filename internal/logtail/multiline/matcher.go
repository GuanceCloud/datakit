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
}

type scoredPattern struct {
	score  int
	regexp *regexp.Regexp
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

func (m *Matcher) Match(content []byte) bool {
	if m.noPattern {
		// 为什么要取反？
		// 因为默认的匹配策略是行首非空白字符，函数功能是确认行首是空白字符，所以要再进行取反
		return !prefixIsSpace(content)
	}

	for idx, scoredPattern := range m.patterns {
		match := scoredPattern.regexp.Match(content)
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

	if m.patterns[0].score == 0 {
		return !prefixIsSpace(content)
	}

	return false
}

func prefixIsSpace(text []byte) bool {
	if len(text) == 0 {
		return true
	}
	// white space 定义为 '\t', '\n', '\v', '\f', '\r', ' ', 0x85, 0xA0
	return unicode.IsSpace(rune(text[0]))
}
