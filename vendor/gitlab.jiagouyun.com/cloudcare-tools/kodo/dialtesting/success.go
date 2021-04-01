package dialtesting

import (
	"fmt"
	"regexp"
	"strings"
)

type DialResult struct {
	Success bool
	Reasons []string
}

type SuccessOption struct {
	Is    string `json:"is,omitempty"`
	IsNot string `json:"is_not,omitempty"`

	MatchRegex          string `json:"match_regex,omitempty"`
	NotMatchRegex       string `json:"not_match_regex,omitempty"`
	matchRe, notMatchRe *regexp.Regexp

	Contains    string `json:"contains,omitempty"`
	NotContains string `json:"not_contains,omitempty"`
}

func (s *SuccessOption) check(val, prompt string) error {
	if s.Is != "" {
		if s.Is != val {
			return fmt.Errorf("%s: expect to be `%s', got `%s'", prompt, s.Is, val)
		}
		return nil
	}

	if s.IsNot != "" {
		if s.IsNot == val {
			return fmt.Errorf("%s: shoud not be %s", prompt, s.IsNot)
		}
		return nil
	}

	if s.matchRe != nil {
		if !s.matchRe.MatchString(val) {
			return fmt.Errorf("%s: regex `%s` match `%s' failed", prompt, s.MatchRegex, val)
		}
	}

	if s.notMatchRe != nil {
		if s.notMatchRe.MatchString(val) {
			return fmt.Errorf("%s: regex `%s' should not match `%s'", prompt, s.NotMatchRegex, val)
		}
	}

	if s.Contains != "" {
		if !strings.Contains(val, s.Contains) {
			return fmt.Errorf("%s: do not contains `%s', got `%s'", prompt, s.Contains, val)
		}
	}

	if s.NotContains != "" {
		if strings.Contains(val, s.NotContains) {
			return fmt.Errorf("%s: should not contains `%s', got `%s'", prompt, s.NotContains, val)
		}
	}
	return nil
}
