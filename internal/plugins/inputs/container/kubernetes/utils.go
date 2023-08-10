// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import "encoding/json"

func transLabels(labels map[string]string) map[string]interface{} {
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
	return map[string]interface{}{
		"df_label":            labelsString,
		"df_label_permission": "read_only",
		"df_label_source":     "datakit",
	}
}
