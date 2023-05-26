// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ptinput

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInitPoint(t *testing.T) {
	pt := &Point{
		Tags: map[string]string{
			"a1": "1",
		},
		Fields: map[string]any{
			"a": int32(1),
			"b": int64(2),
			"c": float32(3),
			"d": float64(4),
			"e": "5",
			"f": true,
			"g": nil,
		},
	}
	pt2 := &Point{}
	InitPt(pt2, "", pt.Tags, pt.Fields, time.Now())
	assert.Equal(t, int32(1), pt2.Fields["a"])
	assert.Equal(t, float32(3), pt2.Fields["c"])
}
