// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"github.com/GuanceCloud/cliutils/point"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
)

type Feeder interface {
	Feed(name, category string, pts []*point.Point, opt ...*Option) error
	FeedLastError(inputName string, err string)
}

// default IO feed implements.
type ioFeeder struct{}

func ptConvert(pts ...*point.Point) (res []*dkpt.Point) {
	for _, pt := range pts {
		pt, err := dkpt.NewPoint(string(pt.Name()), pt.InfluxTags(), pt.InfluxFields(), nil)
		if err != nil {
			continue
		}

		res = append(res, pt)
	}

	return res
}

func (f *ioFeeder) Feed(name, category string, pts []*point.Point, opts ...*Option) error {
	// convert cliutils.Point to io.Point
	iopts := ptConvert(pts...)

	if len(pts) > 0 {
		return defaultIO.doFeed(iopts, category, name, opts[0])
	} else {
		return defaultIO.doFeed(iopts, category, name, nil)
	}
}

func (f *ioFeeder) FeedLastError(name, category string) {
	doFeedLastError(name, category)
}
