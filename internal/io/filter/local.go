// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package filter

import "encoding/json"

type localFilter struct {
	filters map[string]FilterConditions
}

func NewLocalFilter(filters map[string]FilterConditions) *localFilter {
	return &localFilter{filters: filters}
}

func (f *localFilter) Pull(_ string) ([]byte, error) {
	return json.Marshal(f.filters)
}
