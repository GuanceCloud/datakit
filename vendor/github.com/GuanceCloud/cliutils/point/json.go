// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package point

import (
	bytes "bytes"
	"encoding/json"
	"log"
	"time"

	protojson "github.com/gogo/protobuf/jsonpb"
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
		m := &protojson.Marshaler{}
		pb := p.PBPoint()
		buf := bytes.Buffer{}
		if err := m.Marshal(&buf, pb); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}

	return json.Marshal(&JSONPoint{
		Measurement: p.Name(),
		Tags:        p.MapTags(),
		Fields:      p.InfluxFields(),
		Time:        p.pt.Time,
		// NOTE: warns & debugs skipped.
	})
}

// UnmarshalJSON unmarshal protobuf json.
func (p *Point) UnmarshalJSON(j []byte) error {
	var x PBPoint
	var pt *Point

	m := &protojson.Unmarshaler{}
	buf := bytes.NewBuffer(j)

	// try pb unmarshal.
	if err := m.Unmarshal(buf, &x); err == nil {
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
