// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package hash

import (
	"hash/fnv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T) {
	d := []string{
		`k8s_pod_namespace=d["k8s_pod_namespace"]`,
		`k8s_container_name=d["k8s_container_name"]`,
		`k8s_c	ontainer_name=1`, `k8s_node_name=1,
	fields=1`,
	}

	h1 := Fnv1aNew()
	for _, s := range d {
		h1 = Fnv1aHashAdd(h1, s)
	}

	h2 := Fnv1aHash(d)

	h3 := hash4ts(d)

	var s string
	for _, v := range d {
		s += v
	}
	h4 := Fnv1aStrHash(s)

	assert.NotEqual(t, Fnv1aNew(), h1)
	assert.Equal(t, h1, h2)
	assert.Equal(t, h1, h3)
	assert.Equal(t, h1, h4)
}

func BenchmarkIHashAdd(b *testing.B) {
	d := []string{
		`k8s_pod_namespace=d["k8s_pod_namespace"]`,
		`k8s_container_name=d["k8s_container_name"]`,
		`k8s_c	ontainer_name=1`, `k8s_node_name=1,
		fields=1`,
	}
	for n := 0; n < b.N; n++ {
		h := Fnv1aNew()
		for _, s := range d {
			h = Fnv1aHashAdd(h, s)
		}
	}
}

func BenchmarkIHash(b *testing.B) {
	d := []string{
		`k8s_pod_namespace=d["k8s_pod_namespace"]`,
		`k8s_container_name=d["k8s_container_name"]`,
		`k8s_c	ontainer_name=1`, `k8s_node_name=1,
		fields=1`,
	}
	for n := 0; n < b.N; n++ {
		Fnv1aHash(d)
		// b.Error(Fnv1aHash2(d))
	}
}

func BenchmarkHash(b *testing.B) {
	d := []string{
		`k8s_pod_namespace=d["k8s_pod_namespace"]`,
		`k8s_container_name=d["k8s_container_name"]`,
		`k8s_c	ontainer_name=1`, `k8s_node_name=1,
		fields=1`,
	}
	for n := 0; n < b.N; n++ {
		hash4ts(d)
	}
}

func BenchmarkHash2(b *testing.B) {
	d := [][]byte{
		[]byte(`k8s_pod_namespace=d["k8s_pod_namespace"]`),
		[]byte(`k8s_container_name=d["k8s_container_name"]`),
		[]byte(`k8s_c	ontainer_name=1`), []byte(`k8s_node_name=1,
		fields=1`),
	}
	for n := 0; n < b.N; n++ {
		hash4tb(d)
	}
}

func hash4ts(v []string) uint64 {
	h := fnv.New64a()
	for _, v := range v {
		h.Write([]byte(v))
	}
	return h.Sum64()
}

func hash4tb(v [][]byte) uint64 {
	h := fnv.New64a()
	for _, v := range v {
		h.Write(v)
	}
	return h.Sum64()
}
