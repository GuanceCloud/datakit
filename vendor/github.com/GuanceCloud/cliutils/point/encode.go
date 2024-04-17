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

type Encoding int

const (
	encProtobuf      = "protobuf"
	encProtobufAlias = "v2"
	encJSON          = "json"

	encLineprotocolAlias = "v1"
	encLineprotocol      = "line-protocol"

	contentTypeJSON      = "application/json"
	contentTypeProtobuf  = "application/protobuf; proto=com.guance.Point"
	contentTypeLineproto = "application/line-protocol"
)

// EncodingStr convert encoding-string in configure file to
// encoding enum.
//
// Here v1/v2 are alias for lineprotocol and protobuf, this makes
// people easy to switch between lineprotocol and protobuf. For
// json, you should not configure json encoding in production
// environments(json do not classify int and float).
func EncodingStr(s string) Encoding {
	switch strings.ToLower(s) {
	case encProtobuf, encProtobufAlias:
		return Protobuf
	case encJSON:
		return JSON
	case encLineprotocol,
		encLineprotocolAlias:
		return LineProtocol
	default:
		return LineProtocol
	}
}

func HTTPContentType(ct string) Encoding {
	switch ct {
	case contentTypeJSON:
		return JSON
	case contentTypeProtobuf:
		return Protobuf
	case contentTypeLineproto:
		return LineProtocol
	default: // default use line-protocol to be compatible with lagacy code
		return LineProtocol
	}
}

func (e Encoding) HTTPContentType() string {
	switch e {
	case JSON:
		return contentTypeJSON
	case Protobuf:
		return contentTypeProtobuf
	case LineProtocol:
		return contentTypeLineproto
	default: // default use line-protocol to be compatible with lagacy code
		return contentTypeLineproto
	}
}

func (e Encoding) String() string {
	switch e {
	case JSON:
		return encJSON
	case Protobuf:
		return encProtobuf
	case LineProtocol:
		return encLineprotocol
	default: // default use line-protocol to be compatible with lagacy code
		return encLineprotocol
	}
}

// EncodeFn used to iterate on []*Point payload, if error returned, the iterate terminated.
type EncodeFn func(batchSize int, payload []byte) error

type EncoderOption func(e *Encoder)

func WithEncEncoding(enc Encoding) EncoderOption {
	return func(e *Encoder) { e.enc = enc }
}

func WithEncFn(fn EncodeFn) EncoderOption {
	return func(enc *Encoder) { enc.fn = fn }
}

func WithEncBatchSize(size int) EncoderOption {
	return func(e *Encoder) { e.batchSize = size }
}

func WithEncBatchBytes(bytes int) EncoderOption {
	return func(e *Encoder) { e.bytesSize = bytes }
}

type Encoder struct {
	bytesSize,
	batchSize int

	pts []*Point
	lastPtsIdx,
	trimmed,
	parts int
	lastErr error

	lpPointBuf []byte
	pbpts      *PBPoints

	fn  EncodeFn
	enc Encoding
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
	e.trimmed = 0
	e.pbpts.Arr = e.pbpts.Arr[:0]
	e.lpPointBuf = e.lpPointBuf[:0]
}

func (e *Encoder) getPayload(pts []*Point) ([]byte, error) {
	if len(pts) == 0 {
		return nil, nil
	}

	var (
		payload []byte
		err     error
	)

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

func (e *Encoder) EncodeV2(pts []*Point) {
	e.pts = pts
}

func (e *Encoder) doEncodeProtobuf(buf []byte) ([]byte, bool) {
	var (
		curSize,
		pbptsSize int
		trimmed = 1
	)

	for _, pt := range e.pts[e.lastPtsIdx:] {
		if pt == nil {
			continue
		}

		curSize += pt.Size()

		// e.pbpts size larger than buf, we must trim some of points
		// until size fit ok or MarshalTo will panic.
		if curSize >= len(buf) {
			if len(e.pbpts.Arr) <= 1 { // nothing to trim
				e.lastErr = errTooSmallBuffer
				return nil, false
			}

			for {
				if pbptsSize = e.pbpts.Size(); pbptsSize > len(buf) {
					e.pbpts.Arr = e.pbpts.Arr[:len(e.pbpts.Arr)-trimmed]
					e.lastPtsIdx -= trimmed
					trimmed *= 2
				} else {
					goto __doEncode
				}
			}
		} else {
			e.pbpts.Arr = append(e.pbpts.Arr, pt.pt)
			e.lastPtsIdx++
		}
	}

__doEncode:
	e.trimmed = trimmed

	if len(e.pbpts.Arr) == 0 {
		return nil, false
	}

	defer func() {
		e.pbpts.Arr = e.pbpts.Arr[:0]
	}()

	if n, err := e.pbpts.MarshalTo(buf); err != nil {
		e.lastErr = err
		return nil, false
	} else {
		if e.fn != nil {
			if err := e.fn(len(e.pbpts.Arr), buf[:n]); err != nil {
				e.lastErr = err
				return nil, false
			}
		}

		e.parts++
		return buf[:n], true
	}
}

var errTooSmallBuffer = errors.New("too small buffer")

func (e *Encoder) doEncodeLineProtocol(buf []byte) ([]byte, bool) {
	curSize := 0
	npts := 0

	for _, pt := range e.pts[e.lastPtsIdx:] {
		if pt == nil {
			continue
		}

		lppt, err := pt.LPPoint()
		if err != nil {
			e.lastErr = err
			continue
		}

		ptsize := lppt.StringSize()

		if curSize+ptsize+1 > len(buf) { // extra +1 used to store the last '\n'
			if curSize == 0 { // nothing added
				e.lastErr = errTooSmallBuffer
				return nil, false
			}

			if e.fn != nil {
				if err := e.fn(npts, buf[:curSize]); err != nil {
					e.lastErr = err
					return nil, false
				}
			}
			e.parts++
			return buf[:curSize], true
		} else {
			e.lpPointBuf = lppt.AppendString(e.lpPointBuf)

			copy(buf[curSize:], e.lpPointBuf[:ptsize])

			// Always add '\n' to the end of current point, this may
			// cause a _unneeded_ '\n' to the end of buf, it's ok for
			// line-protocol parsing.
			buf[curSize+ptsize] = '\n'
			curSize += (ptsize + 1)

			// clean buffer, next time AppendString() append from byte 0
			e.lpPointBuf = e.lpPointBuf[:0]
			e.lastPtsIdx++
			npts++
		}
	}

	if curSize > 0 {
		e.parts++
		if e.fn != nil { // NOTE: encode callback error will terminate encode
			if err := e.fn(npts, buf[:curSize]); err != nil {
				e.lastErr = err
				return nil, false
			}
		}
		return buf[:curSize], true
	} else {
		return nil, false
	}
}

func (e *Encoder) Next(buf []byte) ([]byte, bool) {
	switch e.enc {
	case Protobuf:
		return e.doEncodeProtobuf(buf)
	case LineProtocol:
		return e.doEncodeLineProtocol(buf)
	case JSON:
		return nil, false
	default: // TODO: json
		return nil, false
	}
}

func (e *Encoder) LastErr() error {
	return e.lastErr
}

func (e *Encoder) String() string {
	return fmt.Sprintf("encoding: %s, parts: %d, byte size: %d, e.batchSize: %d, lastPtsIdx: %d, trimmed: %d",
		e.enc, e.parts, e.bytesSize, e.batchSize, e.lastPtsIdx, e.trimmed,
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
