// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package stats

import (
	"fmt"
	"sync"
	"testing"
)

func TestPlStats(t *testing.T) {
	var stats ScriptStats

	var g sync.WaitGroup

	g.Add(1)
	go func() {
		defer g.Done()
		for i := 0; i < 199; i++ {
			stats.WriteErr(fmt.Sprint(i))
			stats.WritePtCount(1, 1, 1, 1)
			stats.UpdateMeta("", true, true)
			stats.UpdateMeta("", true, false)
			stats.UpdateMeta("", false, true, "")
			stats.UpdateMeta("", false, false, "test")
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
	stats = ScriptStats{}
}
