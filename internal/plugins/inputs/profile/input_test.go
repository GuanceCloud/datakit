// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package profile

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMinHeap(t *testing.T) {
	heap := newMinHeap(16)

	fmt.Println(heap.getTop())

	tm1, _ := time.Parse("2006-01-02 15:04:05Z07:00", "2022-06-08 15:04:06Z")
	tm2, _ := time.Parse("2006-01-02 15:04:05Z07:00", "2022-06-06 15:04:06Z")
	// tm3, _ := time.Parse("2006-01-02 15:04:05Z07:00", "2022-06-05 15:04:06Z")
	tm4, _ := time.Parse("2006-01-02 15:04:05Z07:00", "2022-06-07 15:04:06Z")
	tm5, _ := time.Parse("2006-01-02 15:04:05Z07:00", "2022-06-08 15:04:06Z")

	pb1 := &profileBase{
		profileID: "1111111111111111",
		birth:     tm1,
		point:     nil,
	}
	heap.push(pb1)

	pb2 := &profileBase{
		profileID: "222222222222222",
		birth:     tm2,
		point:     nil,
	}

	heap.push(pb2)

	pb3 := &profileBase{
		profileID: "3333333333333333",
		birth:     tm5,
		point:     nil,
	}

	heap.push(pb3)

	fmt.Println("top: ", heap.getTop())
	fmt.Println("heap.Len: ", heap.Len())
	fmt.Println(heap.indexes)

	pb4 := &profileBase{
		profileID: "44444444444444",
		birth:     tm4,
		point:     nil,
	}

	heap.push(pb4)

	fmt.Println("top: ", heap.getTop())
	fmt.Println("heap.Len: ", heap.Len())
	fmt.Println(heap.indexes)

	pb := heap.pop()

	fmt.Println(pb == pb2)

	fmt.Println("top: ", heap.getTop())
	fmt.Println("heap.Len: ", heap.Len())
	fmt.Println(heap.indexes)

	heap.remove(pb3)

	fmt.Println("top: ", heap.getTop())
	fmt.Println("heap.Len: ", heap.Len())
	fmt.Println(heap.indexes)

	heap.remove(pb1)

	fmt.Println("top: ", heap.getTop())
	fmt.Println("heap.Len: ", heap.Len())
	fmt.Println(heap.indexes)

	heap.push(pb2)

	fmt.Println("top: ", heap.getTop())
	fmt.Println("heap.Len: ", heap.Len())
	fmt.Println(heap.indexes)
}

// go test -v -timeout 30s -run ^Test_addTags$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/profile
func Test_addTags(t *testing.T) {
	cases := []struct {
		name         string
		inOriginTags map[string]string
		inNewKey     string
		inNewVal     string
		expect       map[string]string
	}{
		{
			name:         "add",
			inOriginTags: map[string]string{"a1": "a11", "b1": "b11"},
			inNewKey:     "c1",
			inNewVal:     "c11",
			expect:       map[string]string{"a1": "a11", "b1": "b11", "c1": "c11"},
		},
		{
			name:         "new",
			inOriginTags: map[string]string{},
			inNewKey:     "c1",
			inNewVal:     "c11",
			expect:       map[string]string{"c1": "c11"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			addTags(tc.inOriginTags, tc.inNewKey, tc.inNewVal)
			assert.Equal(t, tc.expect, tc.inOriginTags)
		})
	}
}

// go test -v -timeout 30s -run ^Test_getPyroscopeTagFromLabels$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/profile
func Test_getPyroscopeTagFromLabels(t *testing.T) {
	cases := []struct {
		name     string
		inLabels map[string]string
		expect   map[string]string
	}{
		{
			name:     "empty",
			inLabels: map[string]string{},
			expect:   map[string]string{},
		},
		{
			name:     "name",
			inLabels: map[string]string{"__name__": "server", "a1": "a11", "a2": "a22"},
			expect:   map[string]string{"a1": "a11", "a2": "a22"},
		},
		{
			name:     "no_name",
			inLabels: map[string]string{"a1": "a11", "a2": "a22"},
			expect:   map[string]string{"a1": "a11", "a2": "a22"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := getPyroscopeTagFromLabels(tc.inLabels)
			assert.Equal(t, tc.expect, out)
		})
	}
}
