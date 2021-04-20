package nginx

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestGetMetric(t *testing.T) {
	srv := &http.Server{Addr: ":5000"}

	http.HandleFunc("/nginx_status", httpModelHandle)
	http.HandleFunc("/status/format/json", vtsModelHandle)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			l.Fatalf("ListenAndServe(): %v", err)
		}
	}()
	time.Sleep(time.Second)

	var n = Input{
		Url:    "http://0.0.0.0:5000/nginx_status",
		UseVts: false,
	}
	client, err := n.createHttpClient()
	if err != nil {
		l.Errorf("[error] nginx init client err:%s", err.Error())
		return
	}
	n.client = client

	n.getMetric()
	for _, v := range n.collectCache {
		fmt.Println(v.LineProto())
	}
	n.collectCache = n.collectCache[:0]

	n.Url = "http://0.0.0.0:5000/status/format/json"
	n.UseVts = true
	n.getMetric()

	for _, v := range n.collectCache {
		fmt.Println(v.LineProto())
	}

	srv.Shutdown(context.Background())
}

func httpModelHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	resp := `
Active connections: 2 
server accepts handled requests
 12 12 444 
Reading: 0 Writing: 1 Waiting: 1
`
	w.Write([]byte(resp))
}

func vtsModelHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := `{"hostName":"tan-thinkpad-e450","moduleVersion":"0.1.19.dev.91bdb14","nginxVersion":"1.9.2","loadMsec":1618888188619,"nowMsec":1618888193244,"connections":{"active":1,"reading":0,"writing":1,"waiting":0,"accepted":1,"handled":1,"requests":1},"sharedZones":{"name":"ngx_http_vhost_traffic_status","maxSize":1048575,"usedSize":0,"usedNode":0},"serverZones":{"*":{"requestCounter":0,"inBytes":0,"outBytes":0,"responses":{"1xx":0,"2xx":0,"3xx":0,"4xx":0,"5xx":0,"miss":0,"bypass":0,"expired":0,"stale":0,"updating":0,"revalidated":0,"hit":0,"scarce":0},"requestMsecCounter":0,"requestMsec":0,"requestMsecs":{"times":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],"msecs":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]},"requestBuckets":{"msecs":[],"counters":[]},"overCounts":{"maxIntegerSize":18446744073709551615,"requestCounter":0,"inBytes":0,"outBytes":0,"1xx":0,"2xx":0,"3xx":0,"4xx":0,"5xx":0,"miss":0,"bypass":0,"expired":0,"stale":0,"updating":0,"revalidated":0,"hit":0,"scarce":0,"requestMsecCounter":0}}},"upstreamZones":{"test":[{"server":"10.100.64.215:8888","requestCounter":0,"inBytes":0,"outBytes":0,"responses":{"1xx":0,"2xx":0,"3xx":0,"4xx":0,"5xx":0},"requestMsecCounter":0,"requestMsec":0,"requestMsecs":{"times":[],"msecs":[]},"requestBuckets":{"msecs":[],"counters":[]},"responseMsecCounter":0,"responseMsec":0,"responseMsecs":{"times":[],"msecs":[]},"responseBuckets":{"msecs":[],"counters":[]},"weight":1,"maxFails":1,"failTimeout":10,"backup":false,"down":false,"overCounts":{"maxIntegerSize":18446744073709551615,"requestCounter":0,"inBytes":0,"outBytes":0,"1xx":0,"2xx":0,"3xx":0,"4xx":0,"5xx":0,"requestMsecCounter":0,"responseMsecCounter":0}}]}}`
	w.Write([]byte(resp))
}
