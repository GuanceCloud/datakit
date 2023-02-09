// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import "gitlab.jiagouyun.com/cloudcare-tools/cliutils/point"

// default IO feed implements.
type ioFeeder struct{}

func (f *ioFeeder) Feed(name, category string, pts []*point.Point, opts ...*Option) error {
	if len(pts) > 0 {
		return defaultIO.doFeed(pts, category, name, opts[0])
	} else {
		return defaultIO.doFeed(pts, category, name, nil)
	}
}

func (f *ioFeeder) FeedLastError(name, category string) {
	doFeedLastError(name, category)
}
