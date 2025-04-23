// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package pointutil implements some basic functions for handling points.
package pointutil

import (
	"encoding/json"
	"strings"

	"github.com/GuanceCloud/cliutils/point"
)

func LabelsToPointKVs(labels map[string]string, all bool, keys []string) point.KVs {
	var kvs point.KVs

	if all {
		for k, v := range labels {
			kvs = kvs.AddTag(ReplaceLabelKey(k), v)
		}
	} else {
		for _, key := range keys {
			v, ok := labels[key]
			if !ok {
				continue
			}
			kvs = kvs.AddTag(ReplaceLabelKey(key), v)
		}
	}

	return kvs
}

func ReplaceLabelKey(s string) string {
	return strings.ReplaceAll(s, ".", "_")
}

func ConvertDFLabels(labels map[string]string) point.KVs {
	// empty array
	labelsString := "[]"

	if len(labels) != 0 {
		var lb []string
		for k, v := range labels {
			lb = append(lb, k+":"+v)
		}

		b, err := json.Marshal(lb)
		if err == nil {
			labelsString = string(b)
		}
	}

	// http://gitlab.jiagouyun.com/cloudcare-tools/kodo/-/issues/61#note_11580
	var kvs point.KVs
	kvs = kvs.AddV2("df_label", labelsString, false)
	kvs = kvs.AddV2("df_label_permission", "read_only", false)
	kvs = kvs.AddV2("df_label_source", "datakit", false)
	return kvs
}
