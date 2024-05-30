// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package point

import (
	"encoding/json"
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

func (d *Decoder) Decode(data []byte, opts ...Option) ([]*Point, error) {
	var (
		pts []*Point
		err error
	)

	// point options
	cfg := GetCfg(opts...)
	defer PutCfg(cfg)

	switch d.enc {
	case JSON:
		if err := json.Unmarshal(data, &pts); err != nil {
			return nil, err
		}

	case Protobuf:
		if d.easyproto || defaultPTPool != nil { // force use easyproto when point pool enabled
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
				pts = append(pts, FromPB(pbpt))
			}
		}

		var chk *checker
		if cfg.precheck {
			chk = &checker{cfg: cfg}
			for idx, pt := range pts {
				pts[idx] = chk.check(pt)
			}
		}

	case LineProtocol:

		pts, err = parseLPPoints(data, cfg)
		if err != nil {
			d.detailedError = err
			return nil, simplifyLPError(err)
		}
	}

	// adjust timestamp precision.
	for _, pt := range pts {
		switch cfg.precision {
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

		case PrecNS:
			// pass

		case PrecW, PrecD: // not used

		default:
			// pass
		}
	}

	// check point and apply callbak on each point
	if cfg.precheck || cfg.callback != nil {
		var (
			chk = &checker{cfg: cfg}
			arr []*Point
		)

		for idx, _ := range pts {
			if cfg.precheck {
				pts[idx] = chk.check(pts[idx])
				chk.reset()
			}

			if cfg.callback != nil {
				newPoint, err := cfg.callback(pts[idx])
				if err != nil {
					return nil, err
				}

				if newPoint != nil {
					arr = append(arr, newPoint)
				}
			}
		}

		// Callback may drop some point from pts, so
		// here we override it with newPoint arr.
		if cfg.callback != nil {
			pts = arr
		}
	}

	if d.fn != nil {
		return pts, d.fn(pts)
	}

	return pts, nil
}

func (d *Decoder) DetailedError() error {
	return d.detailedError
}
