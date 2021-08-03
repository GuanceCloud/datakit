package dataway

import (
	"testing"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

func TestDataWayAPIs(t *testing.T) {
	dw := DataWayCfg{URLs: []string{"https://abc.com?token=tkn_abc"}}

	dw.Apply()

	for _, c := range dw.dataWayClients {
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

		err = dw.HeartBeat()
		if tc.fail {
			tu.NotOk(t, err, "")
		} else {
			tu.Ok(t, err)
		}
	}
}

func TestListDataWay(t *testing.T) {
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

		err = dw.DatawayList()
		if tc.fail {
			tu.NotOk(t, err, "")
		} else {
			t.Logf(`dataways: %+#v`, AvailableDataways)
			tu.Ok(t, err)
		}
	}
}

func TestSend(t *testing.T) {
	cases := []struct {
		urls     []string
		category string
		gz       bool
		fail     bool
	}{
		{
			urls:     []string{"http://abc.com"},
			category: "http://user-defined-url.com?token=tkn_xyz",
			gz:       true,
		},

		{
			urls:     []string{"http://abc.com"},
			category: "invalid-category",
			gz:       true,
			fail:     true,
		},
	}

	ExtraHeaders = map[string]string{
		"dkid": "not-set",
	}

	for _, tc := range cases {
		dw := &DataWayCfg{URLs: tc.urls, ontest: true}
		err := dw.Apply()
		tu.Equals(t, nil, err)

		err = dw.Send(tc.category, nil, tc.gz)
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
		for _, c := range dw.dataWayClients {
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
		for _, c := range dw.dataWayClients {
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
			fail:   false,
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
			url:   "http://1024.com?token=tkn_xyz",
			proxy: "invalid-proxy-to-1024.com",
			fail:  true,
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
		dw := DataWayCfg{
			DeprecatedURL: tc.url,
			URLs:          tc.urls,
			HttpProxy:     tc.proxy,
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
