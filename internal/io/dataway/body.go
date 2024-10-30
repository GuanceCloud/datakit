// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
)

type walFrom int8

const (
	walFromMem walFrom = iota
	walFromDisk
)

func (f walFrom) String() string {
	// nolint: exhaustive
	switch f {
	case walFromMem:
		return "M"
	default:
		return "D"
	}
}

type body struct {
	CacheData

	// NOTE: these 2 buffer may comes from:
	//       - reusable buffer that not allocated by body instance, or
	//       - new allocated by apply withCap() when getBody().
	// So during putBody(), do not touch these 2 buffer.
	marshalBuf []byte // buffer used for dump pb binary
	sendBuf    []byte // buffer used for encoding points to pb/line-proto

	chksum string

	selfBuffer, // buffer that belongs to itself, and we should not drop it when putback
	gzon int8
	from      walFrom
	checkSize bool
}

func (b *body) reset() {
	b.CacheData.Payload = nil
	b.CacheData.PayloadType = int32(point.Protobuf)
	b.CacheData.Category = int32(point.Protobuf)

	b.CacheData.Headers = b.CacheData.Headers[:0]
	b.CacheData.DynURL = ""
	b.CacheData.Pts = 0
	b.CacheData.RawLen = 0

	if b.selfBuffer != 1 { // buffer not managed by itself
		b.sendBuf = nil
		b.marshalBuf = nil
	}

	// NOTE: do not touch b.sendBuf and b.marshalBuf, we use the buffer for encoding
	// and WAL protobuf marshal, their len(x) is always it's capacity. If len(x) changed,
	// this will **panic** body encoding and protobuf marshal.

	b.gzon = -1
	b.from = walFromMem
}

func (b *body) buf() []byte {
	return b.CacheData.Payload
}

func (b *body) headers() []*HTTPHeader {
	return b.CacheData.Headers
}

func (b *body) url() string {
	return b.CacheData.DynURL
}

func (b *body) cat() point.Category {
	return point.Category(b.CacheData.Category)
}

func (b *body) enc() point.Encoding {
	return point.Encoding(b.CacheData.PayloadType)
}

func (b *body) npts() int32 {
	return b.CacheData.Pts
}

func (b *body) rawLen() int32 {
	return b.CacheData.RawLen
}

func (b *body) loadCache(data []byte) error {
	if err := b.CacheData.Unmarshal(data); err != nil {
		return fmt.Errorf("Unmarshal: %w", err)
	}

	return nil
}

func (b *body) dump() ([]byte, error) {
	if b.checkSize { // checkSize will not set on production, just for testing cases.
		// NOTE: check required size before marshal, extra Size() call may cause a bit CPU time.
		if s := b.CacheData.Size(); s > len(b.marshalBuf) {
			return nil, fmt.Errorf("too small(%d) marshal buffer, need %d", len(b.marshalBuf), s)
		}
	}

	// MarshalTo() all call Size() within itself.
	if n, err := b.CacheData.MarshalTo(b.marshalBuf); err != nil {
		return nil, fmt.Errorf("MarshalTo: %w", err)
	} else {
		return b.marshalBuf[:n], nil
	}
}

func (b *body) String() string {
	return fmt.Sprintf("from: %s, enc: %s, cat: %s, gzon: %v, headers: %d, pts: %d, buf bytes: %d, chksum: %s, rawLen: %d, cap: %d",
		b.from, b.enc(), b.cat(), b.gzon, len(b.headers()), b.npts(), len(b.buf()), b.chksum, b.rawLen(), cap(b.sendBuf))
}

func (b *body) pretty() string {
	var arr []string
	arr = append(arr, fmt.Sprintf("\n%p from: %s", b, b.from))
	arr = append(arr, fmt.Sprintf("enc: %s", b.enc()))
	arr = append(arr, fmt.Sprintf("cat: %s", b.cat()))
	arr = append(arr, fmt.Sprintf("gzon: %d", b.gzon))
	arr = append(arr, fmt.Sprintf("#buf: %d", len(b.buf())))
	arr = append(arr, fmt.Sprintf("#send-buf: %d", len(b.sendBuf)))
	arr = append(arr, fmt.Sprintf("#mars-buf: %d", len(b.sendBuf)))
	arr = append(arr, fmt.Sprintf("url: %s", b.url()))
	arr = append(arr, fmt.Sprintf("raw-len: %d", b.rawLen()))

	arr = append(arr, fmt.Sprintf("headers(%d):\n", len(b.headers())))

	for _, h := range b.headers() {
		arr = append(arr, fmt.Sprintf("  %s: %s", h.Key, h.Value))
	}

	return strings.Join(arr, "\n")
}

type bodyCallback func(w *writer, b *body) error

func dumpPoints(pts []*point.Point) string {
	var arr []string

	for _, pt := range pts {
		arr = append(arr, pt.Pretty())
	}
	return strings.Join(arr, "\n")
}

// buildPointsBody build points within w into line-protocol(v1) or protobuf(v2).
//
// If there too many points, it will automatically split them on multipart on dataway's MaxRawBodySize.
func (w *writer) buildPointsBody() error {
	var (
		nptsArr []int
		parts   int
	)

	// encode callback: to trace payload info.
	encFn := func(n int, _ []byte) error {
		nptsArr = append(nptsArr, n)
		return nil
	}

	encOpts := []point.EncoderOption{
		point.WithEncEncoding(w.httpEncoding),
		point.WithEncFn(encFn),
	}

	enc := point.GetEncoder(encOpts...)

	defer func() {
		point.PutEncoder(enc)
	}()

	enc.EncodeV2(w.points)

	// for panic logging, when panics, we know:
	// - what these points are
	// - how points encoded and sent
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok { // we got some panic
				buf := make([]byte, 1<<12)
				runtime.Stack(buf, false)

				l.Errorf("panic: %s\n%s", err.Error(), string(buf))

				l.Errorf("encode: %s, total points: %d, current part: %d, body cap: %d",
					err.Error(), len(w.points), parts, w.batchBytesSize)

				panic(fmt.Errorf("dump points: %s", dumpPoints(w.points)))
			}
		}
	}()

	for {
		var (
			compactStart = time.Now()
			b            = getNewBufferBody(withNewBuffer(w.batchBytesSize))
		)

		encodeBytes, ok := enc.Next(b.sendBuf)
		if !ok {
			if err := enc.LastErr(); err != nil {
				l.Errorf("encode: %s, cat: %s, total points: %d, current part: %d, body cap: %d",
					err.Error(), b.cat().Alias(), len(w.points), parts, cap(b.sendBuf))
				return err
			}
			break
		}

		// setup body info.
		b.from = walFromMem
		b.CacheData.Payload = encodeBytes
		b.CacheData.Category = int32(w.category)
		b.CacheData.Pts = int32(nptsArr[parts])
		b.CacheData.RawLen = int32(len(encodeBytes))
		b.CacheData.PayloadType = int32(w.httpEncoding)
		b.CacheData.DynURL = w.dynamicURL
		for k, v := range w.httpHeaders {
			b.CacheData.Headers = append(b.CacheData.Headers, &HTTPHeader{Key: k, Value: v})
		}

		buildBodyCostVec.WithLabelValues(
			b.cat().String(),
			w.httpEncoding.String(),
			"enc",
		).Observe(float64(time.Since(compactStart)) / float64(time.Second))

		buildBodyBatchBytesVec.WithLabelValues(
			b.cat().String(),
			w.httpEncoding.String(),
			"raw",
		).Observe(float64(b.rawLen()))

		buildBodyBatchPointsVec.WithLabelValues(
			b.cat().String(),
			w.httpEncoding.String(),
		).Observe(float64(b.npts()))

		if w.bcb != nil {
			if err := w.bcb(w, b); err != nil {
				l.Warnf("%d points to %q bytes failed: %q, ignored",
					nptsArr[parts], w.category, err.Error())
			}
		}

		parts++
	}

	buildBodyBatchCountVec.WithLabelValues(
		w.category.String(),
		w.httpEncoding.String(),
	).Observe(float64(parts))

	return nil
}
