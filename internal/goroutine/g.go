// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package goroutine

import (
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

// goroutines caches  goroutine.
var goroutines = []*Group{}

// G create a goroutine group, with namespace datakit.
func G(name string) *Group {
	panicCb := func(b []byte) bool {
		l.Errorf("recover panic: %s", string(b))
		select {
		case <-datakit.Exit.Wait(): // don't continue when exit
			return false
		default:
			return true
		}
	}

	g := NewGroup(Option{
		Name:         name,
		PanicTimes:   6,
		PanicCb:      panicCb,
		PanicTimeout: 10 * time.Millisecond,
	})
	var mu sync.Mutex
	mu.Lock()
	goroutines = append(goroutines, g)
	mu.Unlock()
	return g
}

// GWait wait all goroutine group exit.
func GWait() {
	for _, g := range goroutines {
		if err := g.Wait(); err != nil {
			l.Warnf("wait %q failed: %s, ignored", g.Name(), err.Error())
		}

		// logging exit waiting time to find these slow-exit modules
		l.Infof("goroutine Group %q exit, wait %s", g.Name(), time.Since(datakit.GlobalExitTime))
	}

	l.Infof("all goroutine group exited, total wait %s", time.Since(datakit.GlobalExitTime))
}
