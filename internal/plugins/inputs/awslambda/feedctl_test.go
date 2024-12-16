// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

package awslambda

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestShouldFeed(t *testing.T) {
	feedControl := NewFeedControl(20)
	for i := 0; i < 20; i++ {
		assert.True(t, feedControl.ShouldFeed())
	}
	assert.False(t, feedControl.ShouldFeed())
	time.Sleep(20 * time.Second)
	assert.True(t, feedControl.ShouldFeed())
}
