// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package point

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	sync "sync"
)

// EncodeFn used to iterate on []*Point payload, if error returned, the iterate terminated.
type EncodeFn func(batchSize int, payload []byte) error

type EncoderOption func(e *Encoder)

func WithEncEncoding(enc Encoding) EncoderOption { return func(e *Encoder) { e.enc = enc } }
func WithEncFn(fn EncodeFn) EncoderOption        { return func(enc *Encoder) { enc.fn = fn } }
func WithEncBatchSize(size int) EncoderOption    { return func(e *Encoder) { e.batchSize = size } }
func WithEncBatchBytes(bytes int) EncoderOption  { return func(e *Encoder) { e.bytesSize = bytes } }

// WithIgnoreLargePoint will skip too-large point that can't encoded into buffer.
func WithIgnoreLargePoint(on bool) EncoderOption { return func(e *Encoder) { e.ignoreLargePoint = on } }

// WithApproxSize used to calculate the not-precise bytes required for protobuf encoding. Default to true
// for better performance under busy encoding conditions.
func WithApproxSize(on bool) EncoderOption { return func(e *Encoder) { e.approxsize = on } }

type Encoder struct {
	pts []*Point

	bytesSize,
	batchSize,
	totalBytes, // total bytes encoded
	totalPts, // total points encoded
	trimmedPts, // total points trimmed before encoding when size exceed buffer length
	lastPtsIdx,
	skippedPts,
	parts int

	lastErr error

	lpPointBuf []byte
	pbpts      *PBPoints

	fn  EncodeFn
	enc Encoding

	// get point size on pt.Size() instead of pt.PBSize()
	// pt.Size() is faster(2X) than pt.PBSize(), but the later is more precise.
	approxsize       bool
	ignoreLargePoint bool
}

var encPool sync.Pool

func GetEncoder(opts ...EncoderOption) *Encoder {
	v := encPool.Get()
	if v == nil {
		v = newEncoder()
	}

	x := v.(*Encoder)
	for _, opt := range opts {
		if opt != nil {
			opt(x)
		}
	}

	return x
}

func PutEncoder(e *Encoder) {
	e.reset()
	encPool.Put(e)
}

func newEncoder() *Encoder {
	return &Encoder{
		enc:   DefaultEncoding,
		pbpts: &PBPoints{},
	}
}

func (e *Encoder) reset() {
	e.batchSize = 0
	e.bytesSize = 0
	e.fn = nil
	e.pts = nil
	e.enc = DefaultEncoding
	e.lastPtsIdx = 0
	e.lastErr = nil
	e.parts = 0
	e.totalPts = 0
	e.trimmedPts = 0
	e.skippedPts = 0
	e.pbpts.Arr = e.pbpts.Arr[:0]
	e.lpPointBuf = e.lpPointBuf[:0]
	e.ignoreLargePoint = false
	e.approxsize = true

	e.totalBytes = 0
}

func (e *Encoder) getPayload(pts []*Point) ([]byte, error) {
	if len(pts) == 0 {
		return nil, nil
	}

	var (
		payload []byte
		err     error
	)

	//nolint:exhaustive
	switch e.enc {
	case Protobuf:
		pbpts := e.pbpts

		defer func() {
			// Reset e.pbpts buffer: getPayload maybe called multiple times
			// during a single Encode().
			e.pbpts.Arr = e.pbpts.Arr[:0]
		}()

		for _, pt := range pts {
			pbpts.Arr = append(pbpts.Arr, pt.PBPoint())
		}

		if payload, err = pbpts.Marshal(); err != nil {
			return nil, err
		}

	case LineProtocol:
		lppart := []string{}
		for _, pt := range pts {
			if x := pt.LineProto(); x == "" {
				continue
			} else {
				lppart = append(lppart, x)
			}
		}

		payload = []byte(strings.Join(lppart, "\n"))

	case JSON:
		payload, err = json.Marshal(pts)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("not support encode %s", e.enc)
	}

	if e.fn != nil {
		return payload, e.fn(len(pts), payload)
	}
	return payload, nil
}

func (e *Encoder) doEncode(pts []*Point) ([][]byte, error) {
	if len(pts) == 0 {
		return nil, nil
	}

	var (
		batches [][]byte
		batch   []*Point
	)

	// nolint: gocritic
	if e.bytesSize > 0 { // prefer byte size
		curBytesBatchSize := 0
		for _, pt := range pts {
			batch = append(batch, pt)
			curBytesBatchSize += pt.Size()

			if curBytesBatchSize >= e.bytesSize {
				payload, err := e.getPayload(batch)
				if err != nil {
					return nil, err
				}
				batches = append(batches, payload)

				// reset
				batch = batch[:0]
				curBytesBatchSize = 0
			}
		}

		if len(batch) > 0 { // tail
			payload, err := e.getPayload(batch)
			if err != nil {
				return nil, err
			}
			batches = append(batches, payload)
		}
	} else if e.batchSize > 0 { // then point count
		for _, pt := range pts {
			batch = append(batch, pt)
			if len(batch)%e.batchSize == 0 { // switch next batch
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
	return e.doEncode(pts)
}

var errTooSmallBuffer = errors.New("too small buffer")

func (e *Encoder) LastErr() error {
	return e.lastErr
}

func (e *Encoder) SkippedPoints() int {
	return e.skippedPts
}

func (e *Encoder) TotalPoints() int {
	return e.totalPts
}

func (e *Encoder) LastTrimmed() int {
	return e.trimmedPts
}

//nolint:lll
func (e *Encoder) String() string {
	return fmt.Sprintf("encoding: %s, parts: %d, byte size: %d, e.batchSize: %d, lastPtsIdx: %d, total pts: %d, total bytes: %d, trimmed: %d, skipped: %d, lastErr: %v",
		e.enc, e.parts, e.bytesSize, e.batchSize, e.lastPtsIdx, e.totalPts, e.totalBytes, e.trimmedPts, e.skippedPts, e.lastErr,
	)
}

// PB2LP convert protobuf Point to line-protocol Point.
func PB2LP(pb []byte) (lp []byte, err error) {
	dec := GetDecoder(WithDecEncoding(Protobuf))
	defer PutDecoder(dec)

	pts, err := dec.Decode(pb)
	if err != nil {
		return nil, err
	}

	enc := GetEncoder(WithEncEncoding(LineProtocol))
	defer PutEncoder(enc)

	arr, err := enc.Encode(pts)
	if err != nil {
		return nil, err
	}
	return arr[0], nil
}
