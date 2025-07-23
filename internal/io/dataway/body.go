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

type (
	walFrom  int8
	gzipFlag int8
	bufOnwer int8
)

const (
	walFromMem    walFrom = 0
	walFromDisk   walFrom = 1
	walFromNotSet walFrom = -1

	gzipRaw    gzipFlag = 0
	gzipSet    gzipFlag = 1
	gzipNotSet gzipFlag = -1

	bufOnwerOthers bufOnwer = 0
	bufOnwerSelf   bufOnwer = 1

	encNotSet point.Encoding = -1
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

	selfBuffer bufOnwer // buffer that belongs to itself, and we should not drop it when putback
	gzon       gzipFlag
	from       walFrom
}

func (b *body) reset() {
	b.CacheData.Payload = nil
	b.CacheData.PayloadType = int32(encNotSet)
	b.CacheData.Category = int32(point.UnknownCategory)

	b.CacheData.Headers = b.CacheData.Headers[:0]
	b.CacheData.DynURL = ""
	b.CacheData.Pts = 0
	b.CacheData.RawLen = 0
	b.CacheData.PkgTime = 0

	if b.selfBuffer != bufOnwerSelf { // buffer not managed by itself
		b.sendBuf = nil
		b.marshalBuf = nil
	}

	// NOTE: do not touch b.sendBuf and b.marshalBuf, we use the buffer for encoding
	// and WAL protobuf marshal, their len(x) is always it's capacity. If len(x) changed,
	// this will **panic** body encoding and protobuf marshal.

	b.gzon = gzipNotSet
	b.from = walFromNotSet
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

	if b.enc() == encNotSet || b.cat() == point.UnknownCategory {
		l.Warnf("invalid body: %s", b.pretty())
	}

	return nil
}

func (b *body) dump() ([]byte, error) {
	// NOTE: check required size before marshal, extra Size() call may cause a bit CPU time.
	if s := b.CacheData.Size(); s > len(b.marshalBuf) {
		return nil, fmt.Errorf("too small(%d) marshal buffer, need %d", len(b.marshalBuf), s)
	} else {
		// MarshalTo() all call Size() within itself.
		if n, err := b.CacheData.MarshalToSizedBuffer(b.marshalBuf[:s]); err != nil {
			return nil, fmt.Errorf("MarshalTo: %w", err)
		} else {
			return b.marshalBuf[:n], nil
		}
	}
}

func (b *body) String() string {
	return fmt.Sprintf("from: %s, enc: %s, cat: %s, gzon: %v, headers: %d, pts: %d, buf bytes: %d, chksum: %s, rawLen: %d, cap: %d",
		b.from, b.enc(), b.cat(), b.gzon, len(b.headers()), b.npts(), len(b.buf()), b.chksum, b.rawLen(), cap(b.sendBuf))
}

func (b *body) expired(ttl time.Duration) bool {
	return ttl > 0 &&
		b.CacheData.PkgTime > 0 &&
		time.Since(time.Unix(int64(b.CacheData.PkgTime), 0)) > ttl
}

func (b *body) pretty() string {
	var arr []string
	arr = append(arr, fmt.Sprintf("\n%p from: %s", b, b.from))
	arr = append(arr, fmt.Sprintf("enc: %d/%s", b.enc(), b.enc()))
	arr = append(arr, fmt.Sprintf("cat: %d/%s", b.cat(), b.cat()))
	arr = append(arr, fmt.Sprintf("gzon: %d", b.gzon))
	arr = append(arr, fmt.Sprintf("#buf: %d", len(b.buf())))
	arr = append(arr, fmt.Sprintf("#send-buf: %d", len(b.sendBuf)))
	arr = append(arr, fmt.Sprintf("#mars-buf: %d", len(b.sendBuf)))
	arr = append(arr, fmt.Sprintf("url: %s", b.url()))
	arr = append(arr, fmt.Sprintf("raw-len: %d", b.rawLen()))
	arr = append(arr, fmt.Sprintf("pts: %d", b.npts()))

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
		point.WithIgnoreLargePoint(true),
	}

	enc := point.GetEncoder(encOpts...)

	defer func() {
		point.PutEncoder(enc)
	}()

	enc.EncodeV2(w.points)

	buildBodyPointsVec.WithLabelValues(
		w.category.String(),
		w.httpEncoding.String(),
	).Observe(float64(len(w.points)))

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
			defer putBody(b)

			if err := enc.LastErr(); err != nil {
				l.Errorf("encode: %s, cat: %s, total points: %d, current part: %d, body cap: %d",
					err.Error(), w.category.Alias(), len(w.points), parts, cap(b.sendBuf))
				return err
			}

			l.Debugf("last body: %s", b)
			break
		}

		// setup body info.
		b.from = walFromMem
		b.CacheData.Payload = encodeBytes

		if w.gzipDuringBuildBody {
			gz := getZipper()
			defer putZipper(gz)

			if zbuf, err := gz.zip(b.buf()); err != nil {
				l.Errorf("gzip: %s", err.Error())
				return err
			} else {
				ncopy := copy(b.sendBuf, zbuf)
				l.Debugf("copy %d(origin: %d) zipped bytes to buf", ncopy, len(b.buf()))
				b.CacheData.Payload = b.sendBuf[:ncopy]
			}
		}

		b.CacheData.Category = int32(w.category)
		b.CacheData.Pts = int32(nptsArr[parts])
		b.CacheData.RawLen = int32(len(encodeBytes))
		b.CacheData.PayloadType = int32(w.httpEncoding)
		b.CacheData.DynURL = w.dynamicURL
		b.CacheData.PkgTime = uint32(compactStart.Unix())
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
				l.Warnf("compact %d points on category %q failed: %q, ignored",
					nptsArr[parts], w.category, err.Error())
			}
		}

		parts++
	}

	skippedPointVec.WithLabelValues(w.category.String()).Add(float64(enc.SkippedPoints()))

	buildBodyBatchCountVec.WithLabelValues(
		w.category.String(),
		w.httpEncoding.String(),
	).Observe(float64(parts))

	return nil
}
