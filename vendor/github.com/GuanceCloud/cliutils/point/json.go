// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package point

import (
	"encoding/json"
	"log"
	"time"

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
	// NOTE: preferred in-point time
	if jp.Time != 0 {
		opts = append(opts, WithTime(time.Unix(0, jp.Time)))
	}
	return NewPoint(jp.Measurement, jp.Tags, jp.Fields, opts...)
}

// MarshalJSON to protobuf json.
func (p *Point) MarshalJSON() ([]byte, error) {
	if p.HasFlag(Ppb) {
		pb := p.PBPoint()
		if x, err := protojson.Marshal(pb); err != nil {
			return nil, err
		} else {
			return x, nil
		}
	}

	return json.Marshal(&JSONPoint{
		Measurement: string(p.Name()),
		Tags:        p.InfluxTags(),
		Fields:      p.InfluxFields(),
		Time:        p.time.UnixNano(),
		// NOTE: warns & debugs skipped.
	})
}

// UnmarshalJSON unmarshal protobuf json.
func (p *Point) UnmarshalJSON(j []byte) error {
	var x PBPoint
	var pt *Point

	// try pb unmarshal.
	if err := protojson.Unmarshal(j, &x); err == nil {
		pt = FromPB(&x)
	} else {
		log.Printf("pb json unmarshal failed: %s", err)

		// try JSONPoint unmarshal
		var y JSONPoint
		if err := json.Unmarshal(j, &y); err == nil {
			pt = FromJSONPoint(&y)
		} else {
			return err
		}
	}

	*p = *pt

	return nil
}
