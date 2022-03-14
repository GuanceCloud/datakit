// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package dkstring contains string utils
package dkstring

import "strings"

func TrimString(s string) string {
	ns := strings.ReplaceAll(s, " ", "")
	ns = strings.ReplaceAll(ns, "\t", "")
	ns = strings.ReplaceAll(ns, "\n", "")
	return strings.ReplaceAll(ns, "\r", "")
}

// GetUniqueArray []string{"abc", "abc"} returns []string{"abc"}.
func GetUniqueArray(arr []string) []string {
	m := make(map[string]struct{})
	for _, v := range arr {
		m[v] = struct{}{}
	}
	var ret []string
	for k := range m {
		ret = append(ret, k)
	}
	return ret
}
