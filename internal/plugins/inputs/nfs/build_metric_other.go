// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !linux
// +build !linux

package nfs

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func (ipt *Input) buildBaseMetric() ([]*point.Point, error) {
	return []*point.Point{}, nil
}

func (ipt *Input) buildNFSdMetric() ([]*point.Point, error) {
	return []*point.Point{}, nil
}

func (ipt *Input) buildMountStats() ([]*point.Point, error) {
	return []*point.Point{}, nil
}

func getPointsFromMeasurement(ms []inputs.MeasurementV2) []*point.Point {
	pts := []*point.Point{}
	for _, m := range ms {
		pts = append(pts, m.Point())
	}

	return pts
}
