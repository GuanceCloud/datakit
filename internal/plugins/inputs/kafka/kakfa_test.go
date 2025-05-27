// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kafka

import (
	T "testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConf(t *T.T) {
	t.Run(`basic`, func(t *T.T) {
		ipt := defaultInput()

		assert.True(t, ipt.Election, true)

		go ipt.Collect()

		time.Sleep(time.Second)

		ipt.Pause()

		ipt.Resume()
		time.Sleep(time.Second)

		ipt.Terminate()
	})
}
