package io

import (
	"encoding/base64"
	"math/rand"
	"testing"
	"time"
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
		l.Fatal(err.Error())
	}
	for i := 0; i < count; i++ {
		if pt, err := makePoint("mock_random_point", map[string]string{base64.StdEncoding.EncodeToString(buf): base64.StdEncoding.EncodeToString(buf[1:])}, nil, map[string]interface{}{base64.StdEncoding.EncodeToString(buf[2:]): base64.StdEncoding.EncodeToString(buf[3:])}, time.Now()); err != nil {
			l.Fatal(err.Error())
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
		defaultIO.buildBody(pts)
	}
}
