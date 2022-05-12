// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package strarr wraps string array operations.
package strarr

func Contains(set []string, elem string) bool {
	for _, v := range set {
		if v == elem {
			return true
		}
	}

	return false
}

func Differ(source, compare []string) []string {
	m := make(map[string]struct{}, len(compare))
	for _, v := range compare {
		m[v] = struct{}{}
	}

	var diff []string
	for _, v := range source {
		if _, found := m[v]; !found {
			diff = append(diff, v)
		}
	}

	return diff
}

func Intersect(set1, set2 []string) []string {
	if len(set1) == 0 {
		return set2
	} else if len(set2) == 0 {
		return set1
	}

	m := make(map[string]struct{}, len(set1))
	for _, v := range set1 {
		m[v] = struct{}{}
	}

	var inter []string
	for _, v := range set2 {
		if _, ok := m[v]; ok {
			inter = append(inter, v)
		}
	}

	return inter
}
