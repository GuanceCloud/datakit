// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
	"math/rand"
	"strings"
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func randomPoints(out chan *Point, count int) {
	if out == nil {
		return
	}

	var (
		buf = make([]byte, 30)
		err error
	)
	if _, err = rand.Read(buf); err != nil {
		log.Fatal(err.Error())
	}

	for i := 0; i < count; i++ {
		opt := &lp.Option{
			Strict:    true,
			Precision: "n",
			Time:      time.Now(),
		}

		if pt, err := doMakePoint("mock_random_point",
			map[string]string{
				base64.StdEncoding.EncodeToString(buf): base64.StdEncoding.EncodeToString(buf[1:]),
			},
			map[string]interface{}{
				base64.StdEncoding.EncodeToString(buf[2:]): base64.StdEncoding.EncodeToString(buf[3:]),
			},
			opt); err != nil {
			log.Fatal(err.Error())
		} else {
			out <- pt
		}
	}
	close(out)
}

func assemblePoints(count int) []*Point {
	var (
		source = make(chan *Point)
		pts    = make([]*Point, count)
		i      = 0
	)
	go randomPoints(source, count)
	for pt := range source {
		pts[i] = pt
		i++
	}

	return pts
}

func BenchmarkBuildBody(b *testing.B) {
	pts := assemblePoints(10000)
	for i := 0; i < b.N; i++ {
		if _, err := defaultIO.buildBody(pts); err != nil {
			b.Fatal(err)
		}
	}
}

func randPt(t *testing.T) *Point {
	t.Helper()
	pt, err := NewPoint("test-point",
		map[string]string{
			"tag1": cliutils.CreateRandomString(8),
			"tag2": cliutils.CreateRandomString(8),
			"tag3": cliutils.CreateRandomString(8),
			"tag4": cliutils.CreateRandomString(8),
			"tag5": cliutils.CreateRandomString(8),
			"tag6": cliutils.CreateRandomString(8),
		},

		map[string]interface{}{
			"f1": cliutils.CreateRandomString(32),
			"f2": cliutils.CreateRandomString(32),
			"f3": cliutils.CreateRandomString(32),
			"f4": cliutils.CreateRandomString(32),
			"f5": cliutils.CreateRandomString(32),
			"f6": cliutils.CreateRandomString(32),

			"f7":  rand.Int63(),
			"f8":  rand.Int63(),
			"f9":  rand.Int63(),
			"f10": rand.Int63(),
			"f11": rand.Int63(),
			"f12": rand.Int63(),

			"f17":  rand.Float64(),
			"f18":  rand.Float64(),
			"f19":  rand.Float64(),
			"f110": rand.Float64(),
			"f111": rand.Float64(),
			"f112": rand.Float64(),
		}, nil)
	if err != nil {
		t.Fatal(err)
	}

	return pt
}

func TestBuildBody(t *testing.T) {
	pts := []*Point{}
	expect := []string{}
	for i := 0; i < 100; i++ {
		pt := randPt(t)
		expect = append(expect, pt.String())
		pts = append(pts, pt)
	}

	minGZSize = 1000
	maxKodoPack = 10000

	bodies, err := doBuildBody(pts, "")
	if err != nil {
		t.Error(err)
	}

	bufs := [][]byte{}
	for _, b := range bodies {
		if b.gzon {
			zr, err := gzip.NewReader(bytes.NewBuffer(b.buf))
			if err != nil {
				t.Error(err)
			}

			buf, err := io.ReadAll(zr)
			if err != nil {
				t.Error(err)
			}

			zr.Close() //nolint: errcheck
			bufs = append(bufs, buf)
		} else {
			bufs = append(bufs, b.buf)
		}
	}

	tu.Equals(t, strings.Join(expect, "\n"), string(bytes.Join(bufs, []byte("\n"))))
}

func TestSplitBody(t *testing.T) {
	n := 20000
	pts := []*Point{}

	for i := 0; i < n; i++ {
		pts = append(pts, randPt(t))
	}

	t.Logf("pt len: %d", len(pts[0].String()))

	// maxKodoPack = 1024 * 1024 * 32
	bodies, err := doBuildBody(pts, "")
	if err != nil {
		t.Errorf("doBuildBody: %s", err)
		return
	}

	t.Logf("bodies: %d", len(bodies))

	size := 0
	for _, b := range bodies {
		size += len(b.buf)
		t.Logf("body ssize: %d", len(b.buf))
	}

	t.Logf("body avg ssize: %d", size/len(bodies))
}
