// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build espan_linker_local_test
// +build espan_linker_local_test

package ebpftrace

import (
	"bytes"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/gin-gonic/gin"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/hash"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpcli"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/ebpftrace/espan"
)

// var b = `ebpf app_parent_id="0",app_parent_id_l=0i,app_span_sampled=1i,app_trace_id="0",app_trace_id_h=0i,app_trace_id_l=0i,direction="outgoing",duration=571i,ebpf_span_type="entry",http_method="GET",http_route="/info",http_status_code="404",message="{"http_headers":{"Accept-Encoding":"identity","Host":"0.0.0.0:9529","content-type":"application/json"},"http_param":""}",operation="HTTP",pid=175180i,recv_bytes=521i,req_seq=1228003458i,resource="GET /info",resp_seq=1493849386i,send_bytes=101i,service="flask",source_type="web",span_type="local",start=1694429605806046i,status="warning",thread_trace_id=4305341242693961874i 1694429605806046843`

var b = `ebpf app_parent_id="8294577262761556516",app_parent_id_l=8294577262761556516i,app_span_sampled=1i,app_trace_id="87422671431645263",app_trace_id_h=0i,app_trace_id_l=87422671431645263i,direction="outgoing",duration=2858i,ebpf_span_type="",http_method="GET",http_route="/svc-c",http_status_code="200",message="{"http_headers":{"Accept":"*/*","Accept-Encoding":"gzip, deflate","Connection":"keep-alive","Host":"10.200.7.127:23307","User-Agent":"python-requests/2.31.0","traceparent":"00-14a0ecef574d09a301369674dbffb84f-731c4212ecef3a24-01","tracestate":"dd=s:1;t.tid:14a0ecef574d09a3","x-datadog-parent-id":"8294577262761556516","x-datadog-sampling-priority":"1","x-datadog-tags":"_dd.p.tid=14a0ecef574d09a3","x-datadog-trace-id":"87422671431645263"},"http_param":""}",operation="HTTP",pid=175180i,recv_bytes=119i,req_seq=112870230i,resource="GET /svc-c",resp_seq=1157319362i,send_bytes=424i,service="flask",source_type="web",span_type="local",start=1694430559660711i,status="ok",thread_trace_id=-4958942488328659895i 1694430559660711643`

func TestS(t *testing.T) {
	opts := []point.Option{
		point.WithPrecision(point.PrecNS),
		point.WithTime(time.Now()),
	}
	pts, err := httpapi.HandleWriteBody([]byte(b), point.LineProtocol, opts...)
	if err != nil {
		t.Log(err)
	} else {
		t.Error(pts[0].Get("message"))
	}
}

func TestMR(t *testing.T) {
	ulid, _ := espan.NewRandID()

	ipt := &Input{
		Window:       3 * time.Second,
		DBPath:       "./spans/span_storage_test/",
		SamplingRate: 1,
	}
	if err := NewMRRunner(ipt); err != nil {
		t.Fatal(err)
	}

	lock := sync.Mutex{}
	var ptsResult [][]*point.Point

	fn := func(pts []*point.Point) error {
		lock.Lock()
		defer lock.Unlock()
		ptsResult = append(ptsResult, pts)
		return nil
	}

	mrr := ipt.mrrunner
	mrr.Run(fn)

	pts := []*point.Point{}

	kSize := 16
	jSize := 2
	for k := range make([]any, kSize) {
		for j := range make([]any, jSize) {
			var t, d string
			if j == 1 {
				t = espan.SpanTypEntry
				d = espan.INCO
			} else {
				d = espan.OUTG
				t = "exit"
			}
			kvs := point.KVs{}

			k += 1
			j += 1
			kvs = kvs.Add(espan.SpanID, k<<8+j, false, true)
			kvs = kvs.Add(espan.SpanID+"__", espan.ID64(int64(k<<8+j)).StringHex(), false, true)
			kvs = kvs.Add(espan.DirectionStr, d, false, true)
			kvs = kvs.Add(espan.ReqSeq, k, false, true)
			kvs = kvs.Add(espan.RespSeq, k, false, true)
			kvs = kvs.Add(espan.ThrTraceID, j, false, true)
			kvs = kvs.Add(espan.EBPFSpanType, t, false, true)
			pts = append(pts, point.NewPointV2("dketrace", kvs))
		}
	}

	for k := range make([]any, kSize*jSize) {
		if k%2 == 0 {
			pts[k], pts[kSize*jSize-1-k] = pts[kSize*jSize-1-k], pts[k]
		}
	}

	for _, pt := range pts {
		id, _ := ulid.ID()
		pt.Add(espan.SpanID, int64(hash.Fnv1aU8Hash(id.Byte())))
	}

	for i := 0; i < 4; i++ {
		if (i+1)*4 >= len(pts) {
			mrr.InsertSpans(pts[i*4:])
			break
		} else {
			mrr.InsertSpans(pts[i*4 : (i+1)*4])
		}
	}
}

func TestServer(t *testing.T) {
	ipt := &Input{
		Window:       10 * time.Second,
		DBPath:       "./spans/span_storage_test/",
		SamplingRate: 1,
	}
	if err := NewMRRunner(ipt); err != nil {
		t.Fatal(err)
	}
	cli := httpcli.Cli(nil)
	feed := func(pts []*point.Point) error {
		enc := point.GetEncoder(point.WithEncEncoding(point.Protobuf),
			point.WithEncBatchSize(0))
		defer point.PutEncoder(enc)

		buf, _ := enc.Encode(pts)

		req, _ := http.NewRequest("POST", "http://127.0.0.1:9529/v1/write/tracing", bytes.NewReader(buf[0]))
		req.Header.Set("Content-Type", point.Protobuf.HTTPContentType())
		resp, _ := cli.Do(req)
		if resp != nil {
			resp.Body.Close()
		}
		return nil
	}

	ipt.mrrunner.Run(feed)
	id, _ := espan.NewRandID()

	handle := apiBPFTracing(id, ipt.mrrunner)

	router := gin.Default()
	router.POST("/v1/bpftracing", wrapAPI(handle))

	router.Run("127.0.0.1:9530")
}

func wrapAPI(api httpapi.APIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		api(c.Writer, c.Request)
	}
}
