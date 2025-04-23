// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package pointutil

import (
	"encoding/json"

	"github.com/GuanceCloud/cliutils/point"
)

func PointKVsToJSON(kvs point.KVs) string {
	if len(kvs) == 0 {
		return ""
	}

	tmp := make(map[string]interface{})
	for _, field := range kvs {
		tmp[field.GetKey()] = field.Raw()
	}

	b, err := json.Marshal(tmp)
	if err != nil {
		return ""
	}
	return string(b)
}

func MapToJSON(m map[string]string) string {
	if len(m) == 0 {
		// If the map is empty, return an empty string instead of 'null'.
		return ""
	}

	b, err := json.Marshal(m)
	if err != nil {
		return ""
	}

	return string(b)
}

func TrimString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength]
}
