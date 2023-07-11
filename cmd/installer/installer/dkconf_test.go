// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package installer

import (
	T "testing"

	"github.com/stretchr/testify/assert"
)

func Test_mergeDefaultInputs(t *T.T) {
	t.Run("empty", func(t *T.T) {
		expect := []string{
			"cpu",
			"mem",
		}

		assert.Equal(t, expect,
			mergeDefaultInputs([]string{"cpu", "mem"}, nil, false))
	})

	t.Run("disable-all", func(t *T.T) {
		expect := []string{
			"-cpu",
			"-mem",
		}
		assert.Equal(t, expect,
			mergeDefaultInputs([]string{"cpu", "mem"}, []string{"-"}, false))
	})

	t.Run("enable-some", func(t *T.T) {
		expect := []string{
			"-mem",
			"cpu",
		}
		assert.Equal(t, expect,
			mergeDefaultInputs([]string{"cpu", "mem"}, []string{"cpu"}, false))
	})

	t.Run("disable-some", func(t *T.T) {
		expect := []string{
			"-cpu",
			"mem",
		}
		assert.Equal(t, expect,
			mergeDefaultInputs([]string{"cpu", "mem"}, []string{"-cpu"}, false))
	})

	t.Run("disable-and-enable-some", func(t *T.T) {
		defaultList := []string{
			"cpu",
			"disk",
			"mem",
			"system",
		}
		expect := []string{
			"-cpu",
			"disk",
			"mem",
			"system",
		}

		assert.Equal(t,
			expect,
			mergeDefaultInputs(defaultList, []string{"-cpu", "mem", "disk"}, false))
	})
}
