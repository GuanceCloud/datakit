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
	patterns       []*scoredPattern
	noPattern      bool
	autoMatch      bool
	digitPatterns  []*scoredPattern
	letterPatterns []*scoredPattern
	symbolPatterns []*scoredPattern
	extraPatterns  []*scoredPattern
	maxScore       int
}

func NewMatcher(additionalPatterns []string) (*Matcher, error) {
	if len(additionalPatterns) == 0 {
		return &Matcher{noPattern: true}, nil
	}

	patterns, err := compilePatterns(additionalPatterns)
	if err != nil {
		return nil, err
	}

	m := &Matcher{
		patterns: patterns,
	}

	return m, nil
}

func NewAutoMatcher(extraPatterns []string) (*Matcher, error) {
	digitPatterns, err := compilePatterns(GlobalDigitPatterns)
	if err != nil {
		return nil, err
	}
	letterPatterns, err := compilePatterns(GlobalLetterPatterns)
	if err != nil {
		return nil, err
	}
	symbolPatterns, err := compilePatterns(GlobalSymbolPatterns)
	if err != nil {
		return nil, err
	}
	extraScoredPatterns, err := compilePatterns(extraPatterns)
	if err != nil {
		return nil, err
	}

	return &Matcher{
		autoMatch:      true,
		digitPatterns:  digitPatterns,
		letterPatterns: letterPatterns,
		symbolPatterns: symbolPatterns,
		extraPatterns:  extraScoredPatterns,
	}, nil
}

func compilePatterns(patterns []string) ([]*scoredPattern, error) {
	scoredPatterns := make([]*scoredPattern, len(patterns))
	for idx, pattern := range patterns {
		r, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid argument, idx:%d, pattern:'%s', error %w", idx, pattern, err)
		}

		scoredPatterns[idx] = &scoredPattern{
			score:  0,
			regexp: r,
		}
	}
	return scoredPatterns, nil
}

func (m *Matcher) MatchString(content string) bool {
	if m.noPattern {
		return !prefixIsSpace(nil, content)
	}
	if m.doMatch(nil, content) {
		return true
	}
	if m.maxScore == 0 {
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
	if m.maxScore == 0 {
		// use default pattern
		return !prefixIsSpace(content, "")
	}
	return false
}

func (m *Matcher) doMatch(b []byte, str string) bool {
	if m.autoMatch {
		return m.matchAutoPatterns(b, str)
	}

	return m.doMatchPatterns(m.patterns, b, str)
}

func (m *Matcher) doMatchPatterns(patterns []*scoredPattern, b []byte, str string) bool {
	for idx, scoredPattern := range patterns {
		match := scoredPattern.doMatch(b, str)
		if match {
			scoredPattern.score++
			if scoredPattern.score > m.maxScore {
				m.maxScore = scoredPattern.score
			}
			if idx != 0 {
				sort.Slice(patterns, func(i, j int) bool {
					return patterns[i].score > patterns[j].score
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
