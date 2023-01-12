// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package point

import (
	"google.golang.org/protobuf/encoding/protojson"
)

type JSONPoint struct {
	// Requride
	Measurement string `json:"measurement"`

	// Optional
	Tags map[string]string `json:"tags,omitempty"`

	// Requride
	Fields map[string]interface{} `json:"fields"`

	// Unix nanosecond
	Time int64 `json:"time,omitempty"`
}

func (jp *JSONPoint) Point(opts ...Option) (*Point, error) {
	return NewPoint(jp.Measurement, jp.Tags, jp.Fields, opts...)
}

// MarshalJSON to protobuf json.
func (p *Point) MarshalJSON() ([]byte, error) {
	if p.pbPoint == nil {
		if err := p.buildPBPoint(); err != nil {
			return nil, err
		}
	}

	if x, err := protojson.Marshal(p.pbPoint); err != nil {
		return nil, err
	} else {
		return x, nil
	}
}

// UnmarshalJSON unmarshal protobuf json.
func (p *Point) UnmarshalJSON(j []byte) error {
	var x PBPoint
	if err := protojson.Unmarshal(j, &x); err != nil {
		return err
	}

	pt := &Point{}
	pt.SetFlag(Ppb)
	pt.pbPoint = &x

	*p = *pt

	return nil
}
