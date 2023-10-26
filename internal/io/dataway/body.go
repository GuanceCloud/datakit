// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

type body struct {
	buf     []byte
	rawLen  int
	gzon    bool
	npts    int
	payload point.Encoding
}

func (b *body) String() string {
	return fmt.Sprintf("gzon: %v, pts: %d, buf bytes: %d", b.gzon, b.npts, len(b.buf))
}

func (w *writer) buildPointsBody() ([]*body, error) {
	var (
		start   = time.Now()
		nptsArr []int
	)

	// encode callback: to trace payload info.
	encFn := func(n int, _ []byte) error {
		nptsArr = append(nptsArr, n)
		return nil
	}

	enc := point.GetEncoder(
		point.WithEncEncoding(w.httpEncoding),

		// set batch size on point count
		point.WithEncBatchSize(w.batchSize),
		// or point's raw bytes(approximately)
		point.WithEncBatchBytes(w.batchBytesSize),

		point.WithEncFn(encFn),
	)

	defer func() {
		buildBodyCostVec.WithLabelValues(
			w.category.String(),
			w.httpEncoding.String(),
		).Observe(float64(time.Since(start)) / float64(time.Second))

		point.PutEncoder(enc)
	}()

	batches, err := enc.Encode(w.points)
	if err != nil {
		return nil, err
	}

	var arr []*body

	for idx, batch := range batches {
		buildBodyBatchBytesVec.WithLabelValues(
			w.category.String(),
			w.httpEncoding.String(),
		).Observe(float64(len(batch)))

		gzbuf, err := datakit.GZip(batch)
		if err != nil {
			log.Errorf("datakit.GZip: %s", err.Error())
			continue
		}

		body := &body{
			buf:     gzbuf,
			rawLen:  len(batch),
			gzon:    true,
			payload: w.httpEncoding,
			npts:    -1,
		}

		if len(nptsArr) >= idx {
			body.npts = nptsArr[idx]
		} else {
			log.Warnf("batch size not set, set %d points as single batch", len(w.points))
			body.npts = len(w.points)
		}

		arr = append(arr, body)
	}

	buildBodyBatchCountVec.WithLabelValues(
		w.category.String(),
		w.httpEncoding.String(),
	).Observe(float64(len(arr)))

	return arr, nil
}
