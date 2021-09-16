package nginx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

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

	cases := []struct {
		i *Input
	}{
		{
			i: &Input{
				Url:    ts.URL + "/nginx_status",
				UseVts: false,
			},
		},

		{
			i: &Input{
				Url:    ts.URL + "/status/format/json",
				UseVts: true,
			},
		},
	}

	for _, tc := range cases {
		client, err := tc.i.createHttpClient()
		if err != nil {
			l.Errorf("[error] nginx init client err:%s", err.Error())
			return
		}
		tc.i.client = client

		tc.i.getMetric()

		tu.Ok(t, tc.i.lastErr)

		tu.Assert(t, len(tc.i.collectCache) > 0, "")

		for _, v := range tc.i.collectCache {
			p, err := v.LineProto()
			if err != nil {
				t.Error(err)
			}
			t.Logf(p.String())
		}
	}
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
	w.Write([]byte(vtsModelHandleData))
}
