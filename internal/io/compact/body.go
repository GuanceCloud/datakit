// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package compact

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
)

type (
	walFrom  int8
	bufOnwer int8

	GzipFlag int8
)

const (
	WalFromMem    walFrom = 0
	WalFromDisk   walFrom = 1
	WalFromNotSet walFrom = -1

	GzipRaw    GzipFlag = 0
	GzipSet    GzipFlag = 1
	GzipNotSet GzipFlag = -1

	bufOnwerOthers bufOnwer = 0
	bufOnwerSelf   bufOnwer = 1

	encNotSet point.Encoding = -1
)

func (f walFrom) String() string {
	// nolint: exhaustive
	switch f {
	case WalFromMem:
		return "M"
	default:
		return "D"
	}
}

type Body struct {
	CacheData

	// NOTE: these 2 buffer may comes from:
	//       - reusable buffer that not allocated by body instance, or
	//       - new allocated by apply withCap() when getBody().
	// So during putBody(), do not touch these 2 buffer.
	MarshalBuf []byte // buffer used for dump pb binary
	SendBuf    []byte // buffer used for encoding points to pb/line-proto

	caller,
	chksum string

	selfBuffer bufOnwer // buffer that belongs to itself, and we should not drop it when putback
	Gzon       GzipFlag
	From       walFrom
}

func (b *Body) reset() {
	b.caller = "-"
	b.CacheData.Payload = nil
	b.CacheData.PayloadType = int32(encNotSet)
	b.CacheData.Category = int32(point.UnknownCategory)

	b.CacheData.Headers = b.CacheData.Headers[:0]
	b.CacheData.DynURL = ""
	b.CacheData.Pts = 0
	b.CacheData.RawLen = 0
	b.CacheData.PkgTime = 0

	if b.selfBuffer != bufOnwerSelf { // buffer not managed by itself
		b.SendBuf = nil
		b.MarshalBuf = nil
	}

	// NOTE: do not touch b.sendBuf and b.marshalBuf, we use the buffer for encoding
	// and WAL protobuf marshal, their len(x) is always it's capacity. If len(x) changed,
	// this will **panic** body encoding and protobuf marshal.

	b.Gzon = GzipNotSet
	b.From = WalFromNotSet
}

func (b *Body) Buf() []byte {
	return b.CacheData.Payload
}

func (b *Body) GetHeaders() []*HTTPHeader {
	return b.CacheData.Headers
}

func (b *Body) url() string {
	return b.CacheData.DynURL
}

func (b *Body) Cat() point.Category {
	return point.Category(b.CacheData.Category)
}

func (b *Body) Enc() point.Encoding {
	return point.Encoding(b.CacheData.PayloadType)
}

func (b *Body) Npts() int32 {
	return b.CacheData.Pts
}

func (b *Body) RawLen() int32 {
	return b.CacheData.RawLen
}

func (b *Body) LoadCache(data []byte) error {
	if err := b.CacheData.Unmarshal(data); err != nil {
		return fmt.Errorf("Unmarshal: %w", err)
	}

	if b.Enc() == encNotSet || b.Cat() == point.UnknownCategory {
		l.Warnf("invalid body: %s", b.Pretty())
	}

	return nil
}

func (b *Body) Dump() ([]byte, error) {
	// NOTE: check required size before marshal, extra Size() call may cause a bit CPU time.
	if s := b.CacheData.Size(); s > len(b.MarshalBuf) {
		return nil, fmt.Errorf("too small(%d) marshal buffer, need %d", len(b.MarshalBuf), s)
	} else {
		// MarshalTo() all call Size() within itself.
		if n, err := b.CacheData.MarshalToSizedBuffer(b.MarshalBuf[:s]); err != nil {
			return nil, fmt.Errorf("MarshalTo: %w", err)
		} else {
			return b.MarshalBuf[:n], nil
		}
	}
}

func (b *Body) String() string {
	return fmt.Sprintf("from: %s, enc: %s, cat: %s, gzon: %v, headers: %d, pts: %d, buf bytes: %d, chksum: %s, rawLen: %d, cap: %d",
		b.From, b.Enc(), b.Cat(), b.Gzon, len(b.GetHeaders()), b.Npts(), len(b.Buf()), b.chksum, b.RawLen(), cap(b.SendBuf))
}

func (b *Body) Expired(ttl time.Duration) bool {
	return ttl > 0 &&
		b.CacheData.PkgTime > 0 &&
		time.Since(time.Unix(int64(b.CacheData.PkgTime), 0)) > ttl
}

func (b *Body) Pretty() string {
	var arr []string
	arr = append(arr, fmt.Sprintf("\n%p from: %s", b, b.From))
	arr = append(arr, fmt.Sprintf("enc: %d/%s", b.Enc(), b.Enc()))
	arr = append(arr, fmt.Sprintf("cat: %d/%s", b.Cat(), b.Cat()))
	arr = append(arr, fmt.Sprintf("gzon: %d", b.Gzon))
	arr = append(arr, fmt.Sprintf("#buf: %d", len(b.Buf())))
	arr = append(arr, fmt.Sprintf("#send-buf: %d", len(b.SendBuf)))
	arr = append(arr, fmt.Sprintf("#mars-buf: %d", len(b.SendBuf)))
	arr = append(arr, fmt.Sprintf("url: %s", b.url()))
	arr = append(arr, fmt.Sprintf("raw-len: %d", b.RawLen()))
	arr = append(arr, fmt.Sprintf("pts: %d", b.Npts()))

	arr = append(arr, fmt.Sprintf("headers(%d):\n", len(b.GetHeaders())))

	for _, h := range b.GetHeaders() {
		arr = append(arr, fmt.Sprintf("  %s: %s", h.Key, h.Value))
	}

	return strings.Join(arr, "\n")
}

type bodyCallback func(w *Writer, b *Body) error

func dumpPoints(pts []*point.Point) string {
	var arr []string

	for _, pt := range pts {
		arr = append(arr, pt.Pretty())
	}
	return strings.Join(arr, "\n")
}

// BuildPointsBody build points within w into line-protocol(v1) or protobuf(v2).
//
// If there too many points, it will automatically split them on multipart on dataway's MaxRawBodySize.
func (w *Writer) BuildPointsBody() error {
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
		point.WithEncEncoding(w.HTTPEncoding),
		point.WithEncFn(encFn),
		point.WithIgnoreLargePoint(true),
	}

	enc := point.GetEncoder(encOpts...)

	defer func() {
		point.PutEncoder(enc)
	}()

	enc.EncodeV2(w.Points)

	buildBodyPointsVec.WithLabelValues(
		w.Category.String(),
		w.HTTPEncoding.String(),
	).Observe(float64(len(w.Points)))

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
					err.Error(), len(w.Points), parts, w.batchBytesSize)

				panic(fmt.Errorf("dump points: %s", dumpPoints(w.Points)))
			}
		}
	}()

	for {
		var (
			compactStart = time.Now()
			b            = GetNewBufferBody(WithNewBuffer(w.batchBytesSize), WithCaller("buildPointsBody"))
		)

		encodeBytes, ok := enc.Next(b.SendBuf)
		if !ok {
			defer PutBody(b)

			if err := enc.LastErr(); err != nil {
				l.Errorf("encode: %s, cat: %s, total points: %d, current part: %d, body cap: %d",
					err.Error(), w.Category.Alias(), len(w.Points), parts, cap(b.SendBuf))
				return err
			}

			l.Debugf("last body: %s", b)
			break
		}

		// setup body info.
		b.From = WalFromMem
		b.CacheData.Payload = encodeBytes

		if w.gzipDuringBuildBody {
			gz := GetZipper()
			defer PutZipper(gz)

			if zbuf, err := gz.Zip(b.Buf()); err != nil {
				l.Errorf("gzip: %s", err.Error())
				return err
			} else {
				ncopy := copy(b.SendBuf, zbuf)
				l.Debugf("copy %d(origin: %d) zipped bytes to buf", ncopy, len(b.Buf()))
				b.CacheData.Payload = b.SendBuf[:ncopy]
			}
		}

		b.CacheData.Category = int32(w.Category)
		b.CacheData.Pts = int32(nptsArr[parts])
		b.CacheData.RawLen = int32(len(encodeBytes))
		b.CacheData.PayloadType = int32(w.HTTPEncoding)
		b.CacheData.DynURL = w.DynamicURL
		b.CacheData.PkgTime = uint32(compactStart.Unix())
		for k, v := range w.HTTPHeaders {
			b.CacheData.Headers = append(b.CacheData.Headers, &HTTPHeader{Key: k, Value: v})
		}

		BuildBodyCostVec.WithLabelValues(
			b.Cat().String(),
			w.HTTPEncoding.String(),
			"enc",
		).Observe(float64(time.Since(compactStart)) / float64(time.Second))

		buildBodyBatchBytesVec.WithLabelValues(
			b.Cat().String(),
			w.HTTPEncoding.String(),
			"raw",
		).Observe(float64(b.RawLen()))

		buildBodyBatchPointsVec.WithLabelValues(
			b.Cat().String(),
			w.HTTPEncoding.String(),
		).Observe(float64(b.Npts()))

		if w.Callback != nil {
			if err := w.Callback(w, b); err != nil {
				l.Warnf("compact %d points on category %q failed: %q, ignored",
					nptsArr[parts], w.Category, err.Error())
			}
		}

		parts++
	}

	skippedPointVec.WithLabelValues(w.Category.String()).Add(float64(enc.SkippedPoints()))

	buildBodyBatchCountVec.WithLabelValues(
		w.Category.String(),
		w.HTTPEncoding.String(),
	).Observe(float64(parts))

	return nil
}
