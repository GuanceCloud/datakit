package dataway

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	ihttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/http"
)

func TestDataWayAPIs(t *testing.T) {
	dw := DataWayCfg{URLs: []string{"https://abc.com?token=tkn_abc"}}

	if err := dw.Apply(); err != nil {
		t.Fatal(err)
	}

	for _, c := range dw.endPoints {
		tu.Equals(t, len(apis), len(c.categoryURL))
		for k, u := range c.categoryURL {
			t.Logf("%s: %s", k, u)
		}
	}
}

func TestHeartBeat(t *testing.T) {
	cases := []struct {
		urls []string
		fail bool
	}{
		{
			urls: []string{"http://abc.com"},
		},
	}

	ExtraHeaders = map[string]string{
		"dkid": "not-set",
	}

	for _, tc := range cases {
		dw := &DataWayCfg{URLs: tc.urls, ontest: true}
		err := dw.Apply()
		tu.Equals(t, nil, err)

		_, err = dw.HeartBeat()
		if tc.fail {
			tu.NotOk(t, err, "")
		} else {
			tu.Ok(t, err)
		}
	}
}

func TestListDataWay(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"content":{"dataway_list":[],"interval":10}}`)
	}))
	defer ts.Close()

	cases := []struct {
		urls []string
		fail bool
	}{
		{
			urls: []string{ts.URL},
		},
	}

	ExtraHeaders = map[string]string{
		"dkid": "not-set",
	}

	for _, tc := range cases {
		dw := &DataWayCfg{URLs: tc.urls, ontest: true}
		err := dw.Apply()
		tu.Equals(t, nil, err)

		dws, _, err := dw.DatawayList()
		if tc.fail {
			tu.NotOk(t, err, "")
		} else {
			t.Logf(`dataways: %+#v`, dws)
			tu.Ok(t, err)
		}
	}
}

func TestSend(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "[httptest] ok")
	}))
	defer ts.Close()

	cases := []struct {
		urls     []string
		category string
		gz       bool
		fail     bool
	}{
		{
			urls:     []string{ts.URL},
			category: "invalid-category",
			gz:       true,
			fail:     true,
		},

		{
			urls:     []string{ts.URL},
			category: ts.URL + "?token=abc123",
			gz:       true,
		},
	}

	ExtraHeaders = map[string]string{
		"dkid": "not-set",
	}

	for idx, tc := range cases {
		t.Logf("===== case %d ======", idx)

		dw := &DataWayCfg{URLs: tc.urls}
		if err := dw.Apply(); err != nil {
			t.Errorf("Apply: %s", err.Error())
			continue
		}

		err := dw.Send(tc.category, []byte("abc123"), tc.gz)
		if tc.fail {
			tu.NotOk(t, err, "")
		} else {
			tu.Ok(t, err)
		}
	}
}

func TestElectionHeartBeatURL(t *testing.T) {
	cases := []struct {
		urls   []string
		expect []string
		fail   bool
	}{
		{
			urls:   []string{"https://abc.com?token=tkn_123"},
			expect: []string{"https://abc.com/v1/election/heartbeat?token=tkn_123"},
		},

		{
			urls:   []string{"abc.com?token=tkn_123"},
			expect: []string{},
			fail:   true,
		},
	}

	for _, tc := range cases {
		dw := DataWayCfg{URLs: tc.urls}
		err := dw.Apply()
		if tc.fail {
			tu.NotOk(t, err, "")
		} else {
			tu.Ok(t, err)
		}

		urls := []string{}
		for _, c := range dw.endPoints {
			urls = append(urls, c.categoryURL[datakit.ElectionHeartbeat])
		}

		for idx, u := range urls {
			tu.Equals(t, tc.expect[idx], u)
		}
	}
}

func TestElectionURL(t *testing.T) {
	cases := []struct {
		urls   []string
		expect []string
		fail   bool
	}{
		{
			urls:   []string{"https://abc.com?token=tkn_123"},
			expect: []string{"https://abc.com/v1/election?token=tkn_123"},
		},

		{
			urls:   []string{"abc.com?token=tkn_123"},
			expect: []string{},
			fail:   true,
		},
	}

	for _, tc := range cases {
		dw := DataWayCfg{URLs: tc.urls}
		err := dw.Apply()
		if tc.fail {
			tu.NotOk(t, err, "")
		} else {
			tu.Ok(t, err)
		}

		urls := []string{}
		for _, c := range dw.endPoints {
			urls = append(urls, c.categoryURL[datakit.Election])
		}

		for idx, u := range urls {
			tu.Equals(t, tc.expect[idx], u)
		}
	}
}

func TestGetToken(t *testing.T) {
	cases := []struct {
		urls   []string
		expect []string
		fail   bool
	}{
		{
			urls:   []string{"http://abc.com?token=tkn_xyz", "http://def.com?token=tkn_123"},
			expect: []string{"tkn_xyz", "tkn_123"},
		},

		{
			urls:   []string{"http://abc.com", "http://def.com?token=tkn_123"},
			expect: []string{"tkn_123"},
		},

		{ // no token
			urls: []string{"http://abc.com", "http://def.com"},
		},

		{
			urls: []string{"abc.com", "def.com"}, // invalid dataway url
			fail: true,
		},
	}

	for _, tc := range cases {
		dw := DataWayCfg{URLs: tc.urls}
		err := dw.Apply()
		if tc.fail {
			tu.NotOk(t, err, "")
			continue
		} else {
			tu.Ok(t, err)
		}

		tkns := dw.GetToken()
		for idx, x := range tkns {
			tu.Equals(t, tc.expect[idx], x)
		}
	}
}

func TestSetupDataway(t *testing.T) {
	cases := []struct {
		urls   []string
		url    string
		proxy  string
		expect []string
		fail   bool
	}{
		{
			urls:   []string{"http://abc.com", "http://def.com?token=tkn_xyz"},
			url:    "http://xyz.com",
			expect: []string{"http://abc.com", "http://def.com?token=tkn_xyz"},
		},

		{
			url:    "http://xyz.com?token=tkn_xyz",
			expect: []string{"http://xyz.com?token=tkn_xyz"},
			fail:   false,
		},

		{
			url:    "http://1024.com?token=tkn_xyz",
			proxy:  "http://proxy-to-1024.com",
			expect: []string{"http://1024.com?token=tkn_xyz"},
			fail:   false,
		},

		{
			url:    "http://1024.com?token=tkn_xyz",
			expect: []string{"http://1024.com?token=tkn_xyz"},
			proxy:  "invalid-proxy-to-1024.com", // ignored
		},

		{
			url:  "token=tkn_xyz", // invalid url
			fail: true,
		},

		{
			url:  "token=tkn_xyz", // invalid url
			fail: true,
		},

		{
			expect: []string{},
			fail:   true,
		},
	}

	for i, tc := range cases {
		t.Logf("case %d...", i)

		dw := DataWayCfg{
			DeprecatedURL: tc.url,
			URLs:          tc.urls,
			HTTPProxy:     tc.proxy,
			Proxy:         tc.proxy != "",
		}

		err := dw.Apply()
		if tc.fail {
			tu.NotOk(t, err, "")
			continue
		} else {
			tu.Ok(t, err)
		}

		tu.Assert(t, len(dw.URLs) == len(tc.expect),
			"[%d] expect len %d(%+#v), got %d(%+#v)", i, len(tc.expect), tc.expect, len(dw.URLs), dw.URLs)

		for i, x := range dw.URLs {
			tu.Assert(t, x == tc.expect[i], "[%d] epxect %s, got %s", i, tc.expect[i], x)
		}

		t.Logf(dw.String())
	}
}

func TestDatawayConnections(t *testing.T) {
	cases := []struct {
		dwCount int
		reqCnt  int
	}{
		{
			2,
			10000,
		},
	}

	for _, tc := range cases {
		for i := 0; i < tc.dwCount; i++ {
			runTestDatawayConnections(t, tc.reqCnt)
		}
	}
}

func runTestDatawayConnections(t *testing.T, nreq int) {
	t.Helper()
	i := 0

	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "{}") // datakit expect json response
	}))

	var cw ihttp.ConnWatcher
	ts.Config.ConnState = cw.OnStateChange

	ts.Start()
	defer ts.Close()

	dw := &DataWayCfg{URLs: []string{ts.URL}}
	if err := dw.Apply(); err != nil {
		t.Fatal(err)
	}

	t.Logf("dw: %+#v", dw)

	for {
		if err := dw.Send("/v1/write/metric", []byte("abc123"), false); err != nil {
			t.Fatal(err)
		}

		i++
		if i > nreq {
			break
		}
	}

	t.Logf("cw: %s", cw.String())
	tu.Assert(t, cw.Max == 1, "single dataway should only 1 http client")
}

func TestUploadLog(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "OK")
	}))
	defer ts.Close()

	dw := &DataWayCfg{URLs: []string{ts.URL}}
	if err := dw.Apply(); err != nil {
		t.Errorf("Apply: %s", err.Error())
	}
	rBody := strings.NewReader("aaaaaaaaaaaaa")
	resp, err := dw.UploadLog(rBody, "host")
	tu.Ok(t, err)
	defer resp.Body.Close() //nolint:errcheck
	respBody, err := ioutil.ReadAll(resp.Body)
	tu.Ok(t, err)
	tu.Assert(t, string(respBody) == "OK", "assert failed")
}

func TestCheckToken(t *testing.T) {
	cases := []struct {
		valid bool
		token string
	}{
		{valid: true, token: "tkn_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"},
		{valid: true, token: "token_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"},
		{valid: true, token: "tokn_xxxxxxxxxxxxxxxxxxxxxxxx"},
		{valid: false, token: "tokn_xxxxxxxxx"},
		{valid: false, token: "token_xxxxxxxxx"},
		{valid: false, token: "tkn_xxxxxxxxx"},
	}
	dw := DataWayCfg{}

	for _, info := range cases {
		err := dw.CheckToken(info.token)
		if info.valid {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}
