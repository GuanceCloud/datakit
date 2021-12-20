package nginx

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

// go test -v -timeout 30s -run ^TestGetMetric$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/nginx
func TestGetMetric(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/nginx_status":
			httpModelHandle(w, r)
		case "/status/format/json":
			vtsModelHandle(w, r)
		default:
			t.Errorf("unexpected url: %s", r.URL.Path)
		}
	}))

	defer ts.Close()

	//noline:lll
	metrics := []string{
		"nginx,nginx_port=50219,nginx_server=127.0.0.1 connection_accepts=12i,connection_active=2i,connection_handled=12i,connection_reading=0i,connection_requests=444i,connection_waiting=1i,connection_writing=1i",
		"nginx,host=tan-thinkpad-e450,nginx_port=50219,nginx_server=127.0.0.1,nginx_version=1.9.2 connection_accepts=1i,connection_active=1i,connection_handled=1i,connection_reading=0i,connection_requests=1i,connection_waiting=0i,connection_writing=1i",
		"nginx_server_zone,host=tan-thinkpad-e450,nginx_port=50219,nginx_server=127.0.0.1,nginx_version=1.9.2,server_zone=* received=0i,requests=0i,response_1xx=0i,response_2xx=0i,response_3xx=0i,response_4xx=0i,response_5xx=0i,send=0i",
		"nginx_upstream_zone,host=tan-thinkpad-e450,nginx_port=50219,nginx_server=127.0.0.1,nginx_version=1.9.2,upstream_server=10.100.64.215:8888,upstream_zone=test received=0i,request_count=0i,response_1xx=0i,response_2xx=0i,response_3xx=0i,response_4xx=0i,response_5xx=0i,send=0i",
	}

	cases := []struct {
		name string
		i    *Input
	}{
		{
			name: "nginx_status",
			i: &Input{
				URL:    ts.URL + "/nginx_status",
				UseVts: false,
			},
		},

		{
			name: "nginx_status_json",
			i: &Input{
				URL:    ts.URL + "/status/format/json",
				UseVts: true,
			},
		},
	}

	count := 0
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client, err := tc.i.createHTTPClient()
			if err != nil {
				l.Errorf("[error] nginx init client err:%s", err.Error())
				return
			}
			tc.i.client = client

			mpts, err := tc.i.Collect()
			if err != nil {
				t.Errorf("Collect failed: %v", err)
			} else {
				for category, points := range mpts {
					// t.Logf("category = %s, points = %v", category, points)
					if category != "/v1/write/metric" {
						t.Error("invalid_category")
					}
					for _, v := range points {
						// t.Logf("count = %d, v = %s", count, v)

						// 为什么使用 HasPrefix？因为后面会跟时间戳，会不断变化。
						if strings.HasPrefix(v.String(), metrics[count]) {
							t.Errorf("not equal, left = %s, right = %s", v.String(), metrics[count])
						}

						count++
					}
				} // for
			}

			tu.Assert(t, len(tc.i.collectCache) > 0, "")
		})
	}
}

func httpModelHandle(w http.ResponseWriter, r *http.Request) {
	_ = r
	w.Header().Set("Content-Type", "text/plain")
	resp := `
Active connections: 2
server accepts handled requests
 12 12 444
Reading: 0 Writing: 1 Waiting: 1
`
	w.Write([]byte(resp)) //nolint:errcheck
}

func vtsModelHandle(w http.ResponseWriter, r *http.Request) {
	_ = r
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(vtsModelHandleData)) //nolint:errcheck
}
