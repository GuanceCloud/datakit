// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package internal contains functions for internal use.
package internal

// CopyMapString returns a copy of incoming map.
func CopyMapString(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// MergeMapString returns a copy merged 2 incoming maps.
func MergeMapString(ins ...map[string]string) map[string]string {
	l := 0
	for _, in := range ins {
		l += len(in)
	}
	out := make(map[string]string, l)

	for _, in := range ins {
		for k, v := range in {
			out[k] = v
		}
	}

	return out
}

// CopyMapStringInterface returns a copy of incoming map[string]interface{}.
func CopyMapStringInterface(in map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
