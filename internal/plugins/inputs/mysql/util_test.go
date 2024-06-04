// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mysql

import (
	"fmt"
	"math"
	T "testing"

	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func TestConv(t *T.T) {
	u, err := Conv(uint64(math.MaxInt64+1), inputs.Int)
	assert.NoError(t, err)
	assert.Equal(t, u, uint64(math.MaxInt64+1))
	t.Logf("u: %d", u)

	i, err := Conv("-123", inputs.Int)
	assert.NoError(t, err)
	assert.Equal(t, int64(-123), i)

	u, err = Conv(fmt.Sprintf("%d", uint64(math.MaxUint64)), inputs.Int)
	assert.NoError(t, err)
	assert.Equal(t, uint64(math.MaxUint64), u)
	t.Logf("u: %d", u)

	u, err = Conv(uint64(math.MaxUint64), inputs.Int)
	assert.NoError(t, err)
	assert.Equal(t, uint64(math.MaxUint64), u)

	i, err = Conv("184", inputs.Int)
	assert.NoError(t, err)
	t.Logf("u: %d", u)
	assert.Equal(t, int64(184), i)

	t.Logf("int64(math.MaxUint64): %d", int64(maxU64()))

	i, err = cast.ToInt64E(maxU64())
	assert.NoError(t, err)
	assert.Equal(t, int64(-1), i)
	t.Logf("cast i64: %d", i)

	u, err = cast.ToUint64E(maxU64())
	assert.NoError(t, err)
	assert.Equal(t, uint64(math.MaxUint64), u)
	t.Logf("cast u64: %d", u)
}

func maxU64() uint64 {
	return uint64(math.MaxUint64)
}
