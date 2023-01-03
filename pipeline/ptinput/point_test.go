// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ptinput

import (
	"testing"
	"time"

	"github.com/GuanceCloud/platypus/pkg/ast"
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
	assert.Equal(t, &TFMeta{DType: ast.Int, PtFlag: PtField}, pt2.Meta["a"])
	assert.Equal(t, int64(1), pt2.Fields["a"])
	assert.Equal(t, &TFMeta{DType: ast.Int, PtFlag: PtField}, pt2.Meta["b"])
	assert.Equal(t, float64(3), pt2.Fields["c"])
	assert.Equal(t, &TFMeta{DType: ast.Float, PtFlag: PtField}, pt2.Meta["c"])
	assert.Equal(t, &TFMeta{DType: ast.Float, PtFlag: PtField}, pt2.Meta["d"])
	assert.Equal(t, &TFMeta{DType: ast.String, PtFlag: PtField}, pt2.Meta["e"])
	assert.Equal(t, &TFMeta{DType: ast.Bool, PtFlag: PtField}, pt2.Meta["f"])
	assert.Equal(t, &TFMeta{DType: ast.Nil, PtFlag: PtField}, pt2.Meta["g"])
	assert.Equal(t, &TFMeta{DType: ast.String, PtFlag: PtTag}, pt2.Meta["a1"])
}
