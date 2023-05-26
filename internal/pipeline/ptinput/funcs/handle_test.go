// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	T "testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimestampHandle(t *T.T) {
	t.Run("test-tz+0", func(t *T.T) {
		for i := 1970; i < 1980; i++ {
			x := fmt.Sprintf("Thu Jan 16 10:05:19 %d", i)
			ts, err := TimestampHandle(x, "+0")

			assert.NoError(t, err)

			t.Logf("%s : %v", x, time.Unix(0, ts))
		}
	})
}
