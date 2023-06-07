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
	influxdb "github.com/influxdata/influxdb1-client/v2"

	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
)

// Point2dkpt convert point.Point to old io/point.Point.
func Point2dkpt(pts ...*point.Point) (res []*dkpt.Point) {
	for _, pt := range pts {
		pt, err := influxdb.NewPoint(string(pt.Name()), pt.InfluxTags(), pt.InfluxFields(), pt.Time())
		if err != nil {
			continue
		}

		res = append(res, &dkpt.Point{Point: pt})
	}

	return res
}

// Dkpt2point convert old io/point.Point to point.Point. nolint: deadcode,unused.
func Dkpt2point(pts ...*dkpt.Point) (res []*point.Point) {
	for _, pt := range pts {
		fs, err := pt.Fields()
		if err != nil {
			continue
		}

		pt := point.NewPointV2([]byte(pt.Name()),
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
}

// Timeseries get consuming time-series in Guancedb.
func Timeseries(pts []*point.Point) int {
	set := map[uint64]bool{}
	for _, pt := range pts {
		set[pointHash(pt)] = true
	}

	return len(set)
}

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
}

func pointHash(pt *point.Point) uint64 {
	arr := []string{string(pt.Name())}
	for _, kv := range pt.Tags() {
		arr = append(arr, string(
			append(append(kv.Key, byte('=')), kv.GetD()...)))
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
