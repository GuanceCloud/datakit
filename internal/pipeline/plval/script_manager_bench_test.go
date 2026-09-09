// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package plval

import (
	"context"
	"testing"
)

// Compare the existing and cancelable lease APIs on the same manager. No
// script execution is included; this is not an end-to-end pipeline benchmark.
func BenchmarkManagerLease(b *testing.B) {
	for _, mode := range []string{"legacy", "background", "cancelable"} {
		for _, parallel := range []bool{false, true} {
			name := mode + "/serial"
			if parallel {
				name = mode + "/parallel"
			}
			b.Run(name, func(b *testing.B) {
				m := NewScriptManager(nil, nil)
				ctx := context.Background()
				if mode == "cancelable" {
					var cancel context.CancelFunc
					ctx, cancel = context.WithCancel(ctx)
					defer cancel()
				}
				acquire := func() {
					if mode == "legacy" {
						lease, ok := m.Acquire()
						if !ok {
							panic("manager unavailable")
						}
						lease.Release()
					} else {
						lease, err := m.AcquireContext(ctx)
						if err != nil {
							panic(err)
						}
						lease.Release()
					}
				}
				b.ReportAllocs()
				b.ResetTimer()
				if parallel {
					b.RunParallel(func(pb *testing.PB) {
						for pb.Next() {
							acquire()
						}
					})
				} else {
					for i := 0; i < b.N; i++ {
						acquire()
					}
				}
			})
		}
	}
}
