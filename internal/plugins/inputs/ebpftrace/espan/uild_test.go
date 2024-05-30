// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package espan

import (
	"testing"
)

func BenchmarkULID(b *testing.B) {
	ulid, err := NewRandID()
	if err != nil {
		b.Fatal(err)
	}
	b.Run("ulid_exp_rand_lock", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = ulid.ID()
		}
	})

	b.Run("meta", func(b *testing.B) {
		m := SpanMeta{}
		for i := 0; i < b.N; i++ {
			m.Reset()
		}
	})
}

func TestULID(t *testing.T) {
	{
		ulid, err := NewRandID()
		if err != nil {
			t.Error(err)
		}

		id1, ok := ulid.ID()
		if !ok {
			t.Error("!ok")
		} else {
			t.Log(id1)
		}

		id2, ok := ulid.ID()
		if !ok {
			t.Error("!ok")
		} else {
			t.Log(id2)
		}

		if id1 == id2 {
			t.Error("id1 == id2")
		}
	}
}
