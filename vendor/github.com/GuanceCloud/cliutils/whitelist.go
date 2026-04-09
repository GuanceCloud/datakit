package cliutils

import (
	"regexp"
	"strings"
)

// WhiteListItem 白名单条目，支持普通字符串和正则表达式.
type WhiteListItem struct {
	s  string
	re *regexp.Regexp
}

// NewWhiteListItem 从字符串创建白名单条目，支持普通字符串和正则表达式.
func NewWhiteListItem(pattern string) *WhiteListItem {
	pattern = strings.TrimSpace(pattern)

	// 处理正则表达式模式
	if strings.HasPrefix(pattern, "reg:") {
		regexPattern := strings.TrimPrefix(pattern, "reg:")
		return &WhiteListItem{
			s:  regexPattern,
			re: regexp.MustCompile(regexPattern),
		}
	}

	return &WhiteListItem{
		s: pattern,
	}
}

func (item *WhiteListItem) IsRegex() bool {
	return item.re != nil
}

// Match 检查给定路径是否与白名单条目匹配.
func (item *WhiteListItem) Match(val string) bool {
	if item.re != nil {
		return item.re.MatchString(val)
	}
	return item.s == val
}

func WhiteListMatched(val string, arr []*WhiteListItem) bool {
	for _, x := range arr {
		if x.Match(val) {
			return true
		}
	}

	return false
}
