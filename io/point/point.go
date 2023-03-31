// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package point implements datakits basic data structure.
package point

import (
	"time"

	influxdb "github.com/influxdata/influxdb1-client/v2"
)

type Point struct {
	*influxdb.Point
}

func WrapPoint(pts []*influxdb.Point) (x []*Point) {
	for _, pt := range pts {
		x = append(x, &Point{Point: pt})
	}
	return
}

func (p *Point) ToPoint() *influxdb.Point {
	return p.Point
}

//// SetWritten set point already written flag. Once set, the status kept permantly.
// func (p *Point) SetWritten() {
//	if !p.bAlreadyWritten {
//		p.bAlreadyWritten = true
//	}
//}
//
//// GetWritten get point already written flag.
// func (p *Point) GetWritten() bool {
//	return p.bAlreadyWritten
//}

type JSONPoint struct {
	Measurement string                 `json:"measurement"`    // measurement name of the point.
	Tags        map[string]string      `json:"tags,omitempty"` // tags associated with the point.
	Fields      map[string]interface{} `json:"fields"`         // the fields for the point.
	Time        time.Time              `json:"time,omitempty"` // timestamp for the point.
}

func (p *Point) ToJSON() (*JSONPoint, error) {
	fields, err := p.Point.Fields()
	if err != nil {
		return nil, err
	}

	return &JSONPoint{
		Measurement: p.ToPoint().Name(),
		Tags:        p.Point.Tags(),
		Fields:      fields,
		Time:        p.Point.Time(),
	}, nil
}

type Failed struct {
	Ranges  [][2]int // failed parts
	Indexes []int    // failed index
}
