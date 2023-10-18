// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package pinpoint handle Pinpoint APM traces.
package pinpoint

import (
	"testing"
	"time"
)

func TestNextID(t *testing.T) {
	uaid := emptyTransaction{now: time.Now().UnixMilli()}
	saw := make(map[string]bool)
	idchan := make(chan string, 10)
	closer := make(chan int)
	threads := 10
	count := 100
	for i := 0; i < threads; i++ {
		go func() {
			for j := 0; j < count; j++ {
				idchan <- uaid.NextID()
			}
			closer <- 1
		}()
	}
	go func() {
		c := 0
		for i := range closer {
			if c += i; c == threads {
				close(idchan)
			}
		}
	}()
	for id := range idchan {
		if _, ok := saw[id]; ok {
			t.Fatalf("duplicated id: %s", id)
		} else {
			saw[id] = true
		}
	}
}

var nextid string

func BenchmarkNextID(b *testing.B) {
	uaid := emptyTransaction{}
	for i := 0; i < b.N; i++ {
		nextid = uaid.NextID()
	}
}
