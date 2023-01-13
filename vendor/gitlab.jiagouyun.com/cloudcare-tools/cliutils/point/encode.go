// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package point

import (
	"strings"

	"google.golang.org/protobuf/proto"
)

// EncodeFn used to iterate on []*Point payload, if error returned, the iterate terminated.
type EncodeFn func(batchSize int, payload []byte) error

type Encoder struct {
	BatchSize int

	Fn EncodeFn
}

func (e *Encoder) getPayload(pts []*Point) ([]byte, error) {
	if len(pts) == 0 {
		return nil, nil
	}

	var (
		payload []byte
		err     error
	)

	// test if first point is protobuf or line-protocol
	if pts[0].HasFlag(Ppb) {
		pbpts := &PBPoints{}

		for _, pt := range pts {
			if pt.pbPoint == nil {
				if err := pt.buildPBPoint(); err != nil {
					return nil, err
				}
			}
			pbpts.Points = append(pbpts.Points, pt.pbPoint)
		}

		payload, err = proto.Marshal(pbpts)
		if err != nil {
			return nil, err
		}
	} else { // line-protocol.
		lppart := []string{}
		for _, pt := range pts {
			if x, err := pt.LineProto(); err != nil {
				return nil, err
			} else {
				lppart = append(lppart, x)
			}
		}

		payload = []byte(strings.Join(lppart, "\n"))
	}

	if e.Fn != nil {
		return payload, e.Fn(len(pts), payload)
	}
	return payload, nil
}

func (e *Encoder) EncodeWithCallback(pts []*Point) ([][]byte, error) {
	if len(pts) == 0 {
		return nil, nil
	}

	var (
		batches [][]byte
		batch   []*Point
	)

	if e.BatchSize > 0 {
		for _, pt := range pts {
			batch = append(batch, pt)
			if len(batch)%e.BatchSize == 0 { // switch next batch
				payload, err := e.getPayload(batch)
				if err != nil {
					return nil, err
				}
				batches = append(batches, payload)
				batch = batch[:0]
			}
		}

		if len(batch) > 0 { // tail
			payload, err := e.getPayload(batch)
			if err != nil {
				return nil, err
			}
			batches = append(batches, payload)
		}
	} else {
		payload, err := e.getPayload(pts)
		if err != nil {
			return nil, err
		}
		batches = append(batches, payload)
	}

	return batches, nil
}

// Encode get bytes form of multiple Points, often used to Write to somewhere(file/network/...),
// batchSize used to split huge points into multiple part. Set batchSize to 0 to disable the split.
func (e *Encoder) Encode(pts []*Point) ([][]byte, error) {
	return e.EncodeWithCallback(pts)
}

// PB2LP convert protobuf Point to line-protocol Point.
func PB2LP(pb []byte) (lp []byte, err error) {
	var pbpts PBPoints
	if err := proto.Unmarshal(pb, &pbpts); err != nil {
		return nil, err
	}

	lines := []string{}
	for _, pbpt := range pbpts.Points {
		pt := Point{pbPoint: pbpt}
		pt.SetFlag(Ppb)
		if x, err := pt.LineProto(); err != nil {
			return nil, err
		} else {
			lines = append(lines, x)
		}
	}

	return []byte(strings.Join(lines, "\n")), nil
}

// LP2PBP convert line-protocol encoded point to protobuf encoding.
func LP2PBP(lp []byte) (pb []byte, err error) {
	// TODO
	return nil, nil
}
