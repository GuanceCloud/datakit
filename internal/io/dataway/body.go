// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
)

type body struct {
	buf        []byte
	rawLen     int
	gzon       bool
	npts       int
	payloadEnc point.Encoding
}

func (b *body) reset() {
	b.buf = nil
	b.rawLen = 0
	b.gzon = false
	b.npts = 0
	b.payloadEnc = point.LineProtocol
}

func (b *body) String() string {
	return fmt.Sprintf("gzon: %v, pts: %d, buf bytes: %d", b.gzon, b.npts, len(b.buf))
}

func (w *writer) zip(data []byte) ([]byte, error) {
	// reset zipper on multiple parts.
	// zipper may called multiple times during build HTTP bodies,
	// so zipper need to reset before next round.
	if w.parts > 0 {
		w.zipper.buf.Reset()
		w.zipper.w.Reset(w.zipper.buf)
	}

	if _, err := w.zipper.w.Write(data); err != nil {
		return nil, err
	}

	if err := w.zipper.w.Flush(); err != nil {
		return nil, err
	}

	if err := w.zipper.w.Close(); err != nil {
		return nil, err
	}

	return w.zipper.buf.Bytes(), nil
}

type bodyCallback func(w *writer, b *body) error

func (w *writer) buildPointsBody(cb bodyCallback) error {
	var (
		start   = time.Now()
		nptsArr []int
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
		buildBodyCostVec.WithLabelValues(
			w.category.String(),
			w.httpEncoding.String(),
		).Observe(float64(time.Since(start)) / float64(time.Second))

		point.PutEncoder(enc)
	}()

	enc.EncodeV2(w.points)

	if cap(w.sendBuffer) == 0 {
		// The buffer not set before, here we make a default
		// buffer to prevent too-small-buffer error from enc.Next().
		w.sendBuffer = make([]byte, w.batchBytesSize)
	}

	for {
		encodeBytes, ok := enc.Next(w.sendBuffer)
		if !ok {
			if err := enc.LastErr(); err != nil {
				log.Errorf("encode: %s", err.Error())
				return err
			}
			break
		}

		var err error
		w.body.reset()

		w.body.buf = encodeBytes
		w.body.rawLen = len(encodeBytes)
		w.body.gzon = w.gzip
		w.body.payloadEnc = w.httpEncoding
		w.body.npts = -1

		if w.gzip {
			w.body.buf, err = w.zip(encodeBytes)
			if err != nil {
				log.Errorf("datakit.GZip: %s", err.Error())
				continue
			}
		}

		w.body.npts = nptsArr[w.parts]

		buildBodyBatchBytesVec.WithLabelValues(
			w.category.String(),
			w.httpEncoding.String(),
			fmt.Sprintf("%v", w.gzip),
		).Observe(float64(len(w.body.buf)))

		buildBodyBatchPointsVec.WithLabelValues(
			w.category.String(),
			w.httpEncoding.String(),
			fmt.Sprintf("%v", w.gzip),
		).Observe(float64(w.body.npts))
		w.parts++

		if cb != nil {
			if err := cb(w, w.body); err != nil {
				log.Warnf("send %d points to %q(gzip: %v) bytes failed: %q, ignored",
					w.body.npts, w.category, w.gzip, err.Error())
			} else {
				log.Debugf("send part %d with %d points to %q ok, bytes: %d/%d(zipped)",
					w.parts, w.body.npts, w.category, len(encodeBytes), len(w.body.buf))
			}
		}
	}

	buildBodyBatchCountVec.WithLabelValues(
		w.category.String(),
		w.httpEncoding.String(),
	).Observe(float64(w.parts))

	return nil
}
