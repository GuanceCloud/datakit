// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ebpftrace

import (
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
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
