// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"testing"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
)

func TestDoFeed(t *testing.T) {
	cases := []struct {
		name           string
		pts            []*point.Point
		opt            *Option
		category, from string
		fail           bool
	}{
		{
			name:     "nil-on-invalid-category",
			pts:      nil,
			opt:      nil,
			category: "invalid",
			fail:     true,
		},
		{
			name:     "nil",
			pts:      nil,
			opt:      nil,
			category: datakit.Metric,
		},

		{
			name:     "nil",
			pts:      nil,
			opt:      nil,
			category: datakit.Metric,
		},

		{
			name:     "n-pts-on-nil-opt",
			pts:      point.RandPoints(1000),
			opt:      nil,
			category: datakit.Metric,
		},
	}

	x := getDefault()
	x.init()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := x.DoFeed(tc.pts, tc.category, tc.from, tc.opt)
			if tc.fail {
				tu.NotOk(t, err, "")
				return
			}

			tu.Ok(t, err)
		})
	}
}
