// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package filter wraps filtering functionality for container attributes
package filter

import (
	"fmt"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/filter"
)

type FilterType int

const (
	FilterImage FilterType = iota
	FilterImageName
	FilterImageShortName
	FilterNamespace
)

func (f FilterType) String() string {
	switch f {
	case FilterImage:
		return "image"
	case FilterImageName:
		return "image_name"
	case FilterImageShortName:
		return "image_short_name"
	case FilterNamespace:
		return "namespace"
	default:
		return "unknown"
	}
}

var (
	supportedFilterTypes    = []FilterType{FilterImage, FilterImageName, FilterImageShortName, FilterNamespace}
	supportedFilterTypesNum = len(supportedFilterTypes)
)

type Filter []filter.Filter

func NewFilter(include, exclude []string) (Filter, error) {
	in := splitRules(include)
	ex := splitRules(exclude)

	if len(in) != supportedFilterTypesNum || len(ex) != supportedFilterTypesNum {
		return nil, fmt.Errorf("unreachable, invalid filter type, expect len(%d), actual include: %d exclude: %d",
			supportedFilterTypesNum, len(in), len(ex))
	}

	filters := make([]filter.Filter, supportedFilterTypesNum)

	for _, typ := range supportedFilterTypes {
		if len(in[typ]) == 0 && len(ex[typ]) == 0 {
			continue
		}
		filter, err := filter.NewIncludeExcludeFilter(in[typ], ex[typ])
		if err != nil {
			return nil, fmt.Errorf("unexpected filter %s or %s, err: %w", in[typ], ex[typ], err)
		}
		filters[typ] = filter
	}

	return filters, nil
}

func (filters Filter) Match(typ FilterType, field string) bool {
	if field == "" || len(filters) != supportedFilterTypesNum {
		return false
	}
	if filters[typ] != nil {
		return filters[typ].Match(field)
	}
	return false
}

func splitRules(arr []string) [][]string {
	rules := make([][]string, supportedFilterTypesNum)

	split := func(str, prefix string) string {
		if !strings.HasPrefix(str, prefix) {
			return ""
		}
		content := strings.TrimPrefix(str, prefix)
		rule := strings.TrimSpace(content)
		if rule == "*" {
			// trans to double star
			return "**"
		}
		return rule
	}

	for _, str := range arr {
		for _, typ := range supportedFilterTypes {
			rule := split(strings.TrimSpace(str), typ.String()+":")
			if rule != "" {
				rules[typ] = append(rules[typ], rule)
			}
		}
	}

	return rules
}
