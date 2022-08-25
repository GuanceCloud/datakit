// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package bufpool

import (
	"math/rand"
	"sync"
	"testing"
)

func TestBufPool(t *testing.T) {
	l := 1000
	buf := make([]byte, l)
	rand.Read(buf)

	var (
		threads = 100
		wg      = sync.WaitGroup{}
	)
	wg.Add(threads)
	for i := 0; i < threads; i++ {
		go func() {
			body := GetBuffer()
			defer func() {
				wg.Done()
				PutBuffer(body)
			}()
			if n, err := body.Write(buf); err != nil {
				t.Error(err.Error())
			} else if n != l {
				t.Error("read wrong bytes")
			}

			buf := make([]byte, l)
			if _, err := body.Read(buf); err != nil {
				t.Error(err.Error())
			}
			t.Logf("write to buffer finished: %s", string(buf))
		}()
	}
	wg.Wait()
}
