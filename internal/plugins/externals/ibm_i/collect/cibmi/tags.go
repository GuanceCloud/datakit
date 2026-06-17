// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cibmi

func inputsMergeTags(global, extra map[string]string, remote string) map[string]string {
	out := make(map[string]string, len(global)+len(extra)+1)
	for key, value := range global {
		out[key] = value
	}
	for key, value := range extra {
		out[key] = value
	}
	if _, ok := out["host"]; !ok && remote != "" {
		out["host"] = remote
	}
	return out
}
