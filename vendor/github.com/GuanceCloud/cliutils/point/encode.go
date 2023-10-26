// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package point

import (
	"encoding/json"
	"strings"
	sync "sync"

	"google.golang.org/protobuf/proto"
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
		enc: DefaultEncoding,
	}
}

func (e *Encoder) reset() {
	e.batchSize = 0
	e.fn = nil
	e.enc = DefaultEncoding
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
		pbpts := &PBPoints{}

		for _, pt := range pts {
			pbpts.Arr = append(pbpts.Arr, pt.PBPoint())
		}

		payload, err = proto.Marshal(pbpts)
		if err != nil {
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

// PB2LP convert protobuf Point to line-protocol Point.
func PB2LP(pb []byte) (lp []byte, err error) {
	var pbpts PBPoints
	if err := proto.Unmarshal(pb, &pbpts); err != nil {
		return nil, err
	}

	lines := []string{}
	for _, pbpt := range pbpts.Arr {
		pt := FromPB(pbpt)

		if x := pt.LineProto(); x == "" {
			continue
		} else {
			lines = append(lines, x)
		}
	}

	return []byte(strings.Join(lines, "\n")), nil
}
