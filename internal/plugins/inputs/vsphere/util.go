// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package vsphere

import (
	"fmt"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/filter"
)

func anythingEnabled(ex []string) bool {
	for _, s := range ex {
		if s == "*" {
			return false
		}
	}
	return true
}

func isSimple(include []string, exclude []string) bool {
	if len(exclude) > 0 || len(include) == 0 {
		return false
	}
	for _, s := range include {
		if strings.Contains(s, "*") {
			return false
		}
	}
	return true
}

func newFilterOrPanic(include []string, exclude []string) filter.Filter {
	f, err := filter.NewIncludeExcludeFilter(include, exclude)
	if err != nil {
		panic(fmt.Sprintf("Include/exclude filters are invalid: %v", err))
	}
	return f
}
