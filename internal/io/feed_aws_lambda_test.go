// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

package io

import (
	"sync"
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
)

func TestFeedAwsLambda(t *testing.T) {
	output := NewAwsLambdaOutput()
	wg := &sync.WaitGroup{}
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			err := output.Write(&feedData{
				cat: point.Metric,
				pts: make([]*point.Point, 100),
			})
			assert.NoError(t, err)
			wg.Done()
		}()
	}
	wg.Wait()
}
