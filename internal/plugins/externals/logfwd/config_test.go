// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package logfwd

import (
	"reflect"
	"testing"
)

func TestParseLabels(t *testing.T) {
	cases := []struct {
		in  string
		out map[string]string
	}{
		{
			in: "app=\"logging\"\nkind_name=\"testing-crd-pod-target-labels\"",
			out: map[string]string{
				"app":       "logging",
				"kind_name": "testing-crd-pod-target-labels",
			},
		},
	}

	for _, tc := range cases {
		res := parseLabels(tc.in)
		if !reflect.DeepEqual(res, tc.out) {
			t.Errorf("parseLabels(%q) = %#v, want %#v", tc.in, res, tc.out)
		}
	}
}
