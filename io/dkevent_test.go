// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

func TestDKEvent(t *testing.T) {
	e := DKEvent{
		Status:  "info",
		Message: "demo",
	}

	tags := e.Tags()
	fields := e.Fields()

	assert.Equal(t, tags["status"], "info")
	assert.Equal(t, fields["message"], "demo")

	assert.Equal(t, e.escape("http://url?token=token_sdsf8sfsfsk"), "http://url?token=xxxxxx")

	// injection feed function
	e.feed = func(s1, s2 string, p []*Point, o *Option) error {
		assert.Equal(t, s1, "datakit")
		assert.Equal(t, s2, datakit.Logging)
		assert.Equal(t, len(p), 1)
		return nil
	}

	FeedEventLog(&e)
}
