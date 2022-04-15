// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

var defLogFilterMock logFilterMock = &prodLogFilterMock{}

type logFilterMock interface {
	getLogFilter() ([]byte, error)
	preparePoints(pts []*Point) []*Point
}

type prodLogFilterMock struct{}

func (*prodLogFilterMock) getLogFilter() ([]byte, error) {
	if defaultIO.dw == nil {
		return []byte{}, nil
	}
	return defaultIO.dw.GetLogFilter()
}

func (*prodLogFilterMock) preparePoints(pts []*Point) []*Point {
	return pts
}
