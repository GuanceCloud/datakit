// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package fsm contains formatter
package fsm

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var templateReplaceCaptureRE = regexp.MustCompile(`\$\{?([a-zA-Z0-9_\$]+)\}?`)

type TemplateFormatter struct {
	captureIndexes []int
	captureCount   int
	fmtString      string
}

// NewTemplateFormatter instantiates a TemplateFormatter
// from given template string and the maximum amount of captures.
func NewTemplateFormatter(template string, captureCount int) *TemplateFormatter {
	matches := templateReplaceCaptureRE.FindAllStringSubmatch(template, -1)
	if len(matches) == 0 {
		// if no regex reference found, keep it as it is
		return &TemplateFormatter{captureCount: 0, fmtString: template}
	}

	var indexes []int
	valueFormatter := template
	for _, match := range matches {
		idx, err := strconv.Atoi(match[len(match)-1])
		if err != nil || idx > captureCount || idx < 1 {
			// if index larger than captured count or using unsupported named capture group,
			// replace with empty string
			valueFormatter = strings.ReplaceAll(valueFormatter, match[0], "")
		} else {
			valueFormatter = strings.ReplaceAll(valueFormatter, match[0], "%s")
			// note: the regex reference variable $? starts from 1
			indexes = append(indexes, idx-1)
		}
	}
	return &TemplateFormatter{
		captureIndexes: indexes,
		captureCount:   len(indexes),
		fmtString:      valueFormatter,
	}
}

// Format accepts a list containing captured strings and returns the formatted
// string using the template stored in current TemplateFormatter.
func (formatter *TemplateFormatter) Format(captures []string) string {
	if formatter.captureCount == 0 {
		// no label substitution, keep as it is
		return formatter.fmtString
	}
	indexes := formatter.captureIndexes
	vargs := make([]interface{}, formatter.captureCount)
	for i, idx := range indexes {
		vargs[i] = captures[idx]
	}
	return fmt.Sprintf(formatter.fmtString, vargs...)
}
