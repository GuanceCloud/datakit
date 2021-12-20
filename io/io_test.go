package io

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	ihttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dataway"
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
		if _, err := defaultIO.buildBody(pts); err != nil {
			b.Fatal(err)
		}
	}
}

func TestIODatawaySending(t *testing.T) {
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "{}") // datakit expect json response
	}))

	var cw ihttp.ConnWatcher
	ts.Config.ConnState = cw.OnStateChange

	ts.Start()

	testdw := &dataway.DataWayCfg{
		URLs: []string{ts.URL + "?tkn=tkn_TestIODatawaySending"},
	}
	if err := testdw.Apply(); err != nil {
		t.Fatal(err)
	}

	cacheCnt := 10

	ConfigDefaultIO(SetDataway(testdw), SetMaxCacheCount(int64(cacheCnt)))

	go func() {
		if err := Start(); err != nil { // start IO
			t.Error(err)
		}
	}()

	time.Sleep(time.Second) // required

	opt := &lp.Option{
		Strict:    true,
		Precision: "n",
	}

	npts := 0
	fmt.Println("feed points...")
	cache := []*Point{}
	for {
		for i := 0; i < cacheCnt; i++ {
			pt, err := doMakePoint("TestIODatawaySending",
				nil, map[string]interface{}{
					"f1": 3.14,
				}, opt)
			cache = append(cache, pt)
			if err != nil {
				t.Fatal(err)
			}

			npts++
		}

		if err := Feed("TestIODatawaySending", datakit.Metric, cache, nil); err != nil {
			t.Fatal(err)
		}
		if npts > 10000 {
			break
		}
		cache = cache[:0]
	}

	t.Logf("cw: %s", cw.String())
	tu.Assert(t, cw.Max == 1, "single dataway should only 1 HTTP client, but got %d", cw.Max)
}
