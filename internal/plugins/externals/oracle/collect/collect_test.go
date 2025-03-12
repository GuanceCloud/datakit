// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package collect

import (
	"testing"
	"time"
)

func TestChannel(t *testing.T) {
	t.Run("%/", func(t *testing.T) {
		arr := []int{1, 2, 3, 4, 5}
		for _, v := range arr {
			t.Log(v, "% 2 = ", v%2)
		}
	})

	t.Run("wait_chan", func(t *testing.T) {
		arr := []int{1, 2, 3, 4, 5}

		ch := make(chan struct{}, 5)

		start := time.Now()
		for _, v := range arr {
			if v%2 == 0 {
				go func() {
					time.Sleep(3 * time.Second)
					ch <- struct{}{}
				}()
			} else {
				ch <- struct{}{}
			}
		}

		t.Log("out")
		count := 0
		for range ch {
			count++
			if count == len(arr) {
				close(ch)
			}
		}

		t.Logf("duration = %s, exited", time.Since(start))
	})
}
