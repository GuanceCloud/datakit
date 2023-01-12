// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package point

import "google.golang.org/protobuf/proto"

// DecodeFn used to iterate on []*Point payload, if error returned, the iterate terminated.
type DecodeFn func([]*Point) error

type Decoder struct {
	*cfg

	// For line-protocol parsing, keep original error.
	detailedError error
}

func NewDecoder(opts ...Option) *Decoder {
	dec := &Decoder{
		cfg: defaultCfg(),
	}
	for _, opt := range opts {
		opt(dec.cfg)
	}
	return dec
}

func (d *Decoder) Decode(data []byte) ([]*Point, error) {
	var (
		pts []*Point
		err error
	)

	switch d.cfg.enc {
	case Protobuf:
		var pbpts PBPoints
		if err = proto.Unmarshal(data, &pbpts); err != nil {
			return nil, err
		}

		for _, pbpt := range pbpts.Points {
			pt := &Point{
				pbPoint: pbpt,
			}
			pt.SetFlag(Ppb)

			pts = append(pts, pt)
		}

	case LineProtocol:
		pts, err = parseLPPoints(data, d.cfg)
		if err != nil {
			d.detailedError = err
			return nil, simplifyLPError(err)
		}
	}

	if d.cfg.decodeFn != nil {
		return pts, d.cfg.decodeFn(pts)
	}

	return pts, nil
}

func (d *Decoder) DetailedError() error {
	return d.detailedError
}
