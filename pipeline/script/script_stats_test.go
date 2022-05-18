// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package script

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlChangeEvent(t *testing.T) {
	var stats scriptStats

	var g sync.WaitGroup

	g.Add(1)
	go func() {
		defer g.Done()
		for i := 0; i < 199; i++ {
			stats.WriteErr(fmt.Sprint(i))
			stats.WritePtCount(1, 1, 1)
		}
	}()
	g.Add(1)
	go func() {
		defer g.Done()
		for i := 0; i < 299; i++ {
			stats.Read()
		}
	}()
	g.Wait()

	l.Info("")
	stats = scriptStats{}

	tmp := []string{}
	for i := 0; i < 256; i++ {
		assert.Equal(t, tmp, stats.Read().Last100err)
		c := fmt.Sprint(i)
		stats.WriteErr(c)
		tmp = append(tmp, c)
		if len(tmp) > 100 {
			tmp = tmp[len(tmp)-100:]
		}
		assert.Equal(t, tmp, stats.Read().Last100err)
	}
}
