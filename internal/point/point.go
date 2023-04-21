// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package point used convert point from old format to new format.
package point

import (
	"github.com/GuanceCloud/cliutils/point"
	influxdb "github.com/influxdata/influxdb1-client/v2"

	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
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
