// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package point used convert point from old format to new format.
package point

import (
	"hash/fnv"
	"sort"
	"strings"

	"github.com/GuanceCloud/cliutils/point"
)

/*
// Dkpt2point convert old io/point.Point to point.Point. nolint: deadcode,unused.
func Dkpt2point(pts ...*dkpt.Point) (res []*point.Point) {
	for _, pt := range pts {
		fs, err := pt.Fields()
		if err != nil {
			continue
		}

		pt := point.NewPointV2(pt.Name(),
			append(point.NewTags(pt.Tags()), point.NewKVs(fs)...), nil)

		res = append(res, pt)
	}

	return res
}

// LineprotoTimeseries get line-protocol point will consuming time-series in Guancedb.
func LineprotoTimeseries(pts []*dkpt.Point) int {
	set := map[uint64]bool{}
	for _, pt := range pts {
		set[lineprotoHash(pt)] = true
	}

	return len(set)
}*/

// Timeseries get consuming time-series in Guancedb.
func Timeseries(pts []*point.Point) int {
	set := map[uint64]bool{}
	for _, pt := range pts {
		set[pointHash(pt)] = true
	}

	return len(set)
}

/*
func lineprotoHash(pt *dkpt.Point) uint64 {
	arr := []string{pt.Name()}
	for k, t := range pt.Tags() {
		arr = append(arr, k+"="+t)
	}

	fs, err := pt.Fields()
	if err != nil {
		return 0
	}

	for k := range fs {
		arr = append(arr, "__field__="+k)
	}

	return arrhash(arr)
} */

func pointHash(pt *point.Point) uint64 {
	arr := []string{pt.Name()}
	for _, kv := range pt.Tags() {
		arr = append(arr, kv.Key+"="+kv.GetS())
	}

	for _, kv := range pt.Fields() {
		arr = append(arr, string(append([]byte("__field__="), kv.Key...)))
	}

	return arrhash(arr)
}

func arrhash(arr []string) uint64 {
	x := fnv.New64()
	sort.Strings(arr)

	data := []byte(strings.Join(arr, ""))

	if _, err := x.Write(data); err != nil {
		return 0
	}
	return x.Sum64()
}
