// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package point

import (
	"encoding/json"
	"fmt"
	sync "sync"
	"time"
)

var decPool sync.Pool

// DecodeFn used to iterate on []*Point payload, if error returned, the iterate terminated.
type DecodeFn func([]*Point) error

type DecoderOption func(e *Decoder)

func WithDecEncoding(enc Encoding) DecoderOption {
	return func(d *Decoder) { d.enc = enc }
}

func WithDecFn(fn DecodeFn) DecoderOption {
	return func(d *Decoder) { d.fn = fn }
}

func WithDecEasyproto(on bool) DecoderOption {
	return func(d *Decoder) { d.easyproto = on }
}

type Decoder struct {
	enc Encoding
	fn  DecodeFn

	easyproto bool

	// For line-protocol parsing, keep original error.
	detailedError error
}

func GetDecoder(opts ...DecoderOption) *Decoder {
	v := decPool.Get()
	if v == nil {
		v = newDecoder()
	}

	x := v.(*Decoder)

	for _, opt := range opts {
		if opt != nil {
			opt(x)
		}
	}

	return x
}

func PutDecoder(d *Decoder) {
	d.reset()
	decPool.Put(d)
}

func newDecoder() *Decoder {
	return &Decoder{}
}

func (d *Decoder) reset() {
	d.enc = 0
	d.fn = nil
	d.detailedError = nil
	d.easyproto = false
}

// nolint: gocritic
func detectTimestampPrecision(ts int64) int64 {
	if ts/1e9 < 10 { // sec
		return ts * int64(time.Second)
	} else if ts/1e12 < 10 { // milli-sec
		return ts * int64(time.Millisecond)
	} else if ts/1e15 < 10 { // micro-sec
		return ts * int64(time.Microsecond)
	} else { // nano-sec
		return ts
	}
}

func (d *Decoder) doDecode(data []byte, c *cfg) ([]*Point, error) {
	var (
		pts []*Point
		err error
	)

	//nolint:exhaustive
	switch d.enc {
	case JSON:
		if err := json.Unmarshal(data, &pts); err != nil {
			return nil, err
		}

	case Protobuf:
		if d.easyproto {
			pts, err = unmarshalPoints(data)
			if err != nil {
				return nil, err
			}
		} else {
			var pbpts PBPoints
			if err = pbpts.Unmarshal(data); err != nil {
				return nil, err
			}

			for _, pbpt := range pbpts.Arr {
				// NOTE: under gogo Unmarshal, nothing comes from point pool, so
				// we create Point without NewPoint(), and make the point escaped
				// from point pool(if defaultPTPool set).
				//
				// Although put back points that not originally from the pool is
				// possible, we still distinguish this behavior for better
				// observability of actions(these escaped point counter are export
				// by metrics).
				pt := &Point{
					pt: pbpt,
				}
				pt.SetFlag(Ppb)
				pts = append(pts, pt)
			}
		}

	case LineProtocol:
		pts, err = parseLPPoints(data, c)
		if err != nil {
			d.detailedError = err
			return nil, simplifyLPError(err)
		}

	default:
		return nil, fmt.Errorf("not support encode: %s", d.enc)
	}

	return pts, err
}

func decodeAdjustPoints(pts []*Point, c *cfg) ([]*Point, error) {
	var (
		chk       *checker
		newPoints []*Point

		// set point's default timestamp
		nowNano = c.timestamp
	)

	if nowNano == 0 { // not set
		nowNano = time.Now().UnixNano()
	}

	if c.precheck {
		chk = &checker{cfg: c}
	}

	// adjust and check the point.
	for idx, pt := range pts {
		// use current time
		if pt.pt.Time == 0 {
			pt.pt.Time = nowNano
		} else { // adjust point's timestamp
			switch c.precision {
			case PrecDyn:
				pt.pt.Time = detectTimestampPrecision(pt.pt.Time)
			case PrecUS:
				pt.pt.Time *= int64(time.Microsecond)
			case PrecMS:
				pt.pt.Time *= int64(time.Millisecond)
			case PrecS:
				pt.pt.Time *= int64(time.Second)
			case PrecM:
				pt.pt.Time *= int64(time.Minute)
			case PrecH:
				pt.pt.Time *= int64(time.Hour)
			case PrecNS: // pass
			case PrecW, PrecD: // not used
			default: // pass
			}
		}

		if c.precheck {
			pts[idx] = chk.check(pts[idx])
			chk.reset()
		}

		// Applied the callback on each point, the callback used to check if the
		// point is valid for usage, for example:
		//  - Is the measurement name are expected?
		//  - Is point's key or value are expected?
		//  - Is any warning on the point?
		//  - ...
		if c.callback != nil {
			if x, err := c.callback(pts[idx]); err != nil {
				return nil, err
			} else if x != nil {
				newPoints = append(newPoints, x)
			}
		}
	}

	if len(newPoints) > 0 {
		pts = newPoints
	}

	return pts, nil
}

func (d *Decoder) Decode(data []byte, opts ...Option) ([]*Point, error) {
	// point options
	c := GetCfg(opts...)
	defer PutCfg(c)

	pts, err := d.doDecode(data, c)
	if err != nil {
		return nil, err
	}

	pts, err = decodeAdjustPoints(pts, c)
	if err != nil {
		return nil, err
	}

	if d.fn != nil {
		return pts, d.fn(pts)
	}
	return pts, nil
}

func (d *Decoder) DetailedError() error {
	return d.detailedError
}
