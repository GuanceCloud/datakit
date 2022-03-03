package dialtesting

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	dt "gitlab.jiagouyun.com/cloudcare-tools/kodo/dialtesting"
)

var httpCases = []struct {
	name      string
	t         *dt.HTTPTask
	fail      bool
	reasonCnt int
}{
	{
		name:      "test dial with certificate",
		fail:      false,
		reasonCnt: 0,
		t: &dt.HTTPTask{
			ExternalID: cliutils.XID("dtst_"),
			Method:     "GET",
			URL:        "https://localhost:45323/_test_with_cert",
			Name:       "_test_with_cert",
			Frequency:  "1s",
			AdvanceOptions: &dt.HTTPAdvanceOption{
				Certificate: &dt.HTTPOptCertificate{
					IgnoreServerCertificateError: true,
					PrivateKey:                   string(tlsData["key"]),
					Certificate:                  string(tlsData["crt"]),
				},
			},

			SuccessWhen: []*dt.HTTPSuccess{
				{
					StatusCode: []*dt.SuccessOption{
						{Is: "200"},
					},
				},
			},
		},
	},
	{
		fail:      true,
		name:      "invalid certificate",
		reasonCnt: 0,
		t: &dt.HTTPTask{
			ExternalID: cliutils.XID("dtst_"),
			Method:     "GET",
			URL:        "https://localhost:45323/_test_with_cert",
			Name:       "_test_with_cert",
			Region:     "",
			Frequency:  "1s",
			AdvanceOptions: &dt.HTTPAdvanceOption{
				Certificate: &dt.HTTPOptCertificate{
					IgnoreServerCertificateError: false, // bad certificate, expect fail
					PrivateKey:                   string(tlsData["key"]),
					Certificate:                  string(tlsData["crt"]),
				},
			},

			SuccessWhen: []*dt.HTTPSuccess{
				{
					StatusCode: []*dt.SuccessOption{
						{Is: "200"},
					},
				},
			},
		},
	},

	{
		name:      "test dial with proxy",
		fail:      false,
		reasonCnt: 0,
		t: &dt.HTTPTask{
			ExternalID: cliutils.XID("dtst_"),
			Method:     "POST",
			URL:        "http://localhost:54321/_test_with_proxy",
			Name:       "_test_with_proxy",
			Frequency:  "1s",
			AdvanceOptions: &dt.HTTPAdvanceOption{
				Proxy: &dt.HTTPOptProxy{
					URL:     "http://localhost:54325",
					Headers: map[string]string{"X-proxy-header": "proxy-foo"},
				},
			},

			SuccessWhen: []*dt.HTTPSuccess{
				{
					StatusCode: []*dt.SuccessOption{
						{Is: "200"},
					},
				},
			},
		},
	},

	{
		name:      "test dial with invalid body type",
		fail:      true,
		reasonCnt: 0,
		t: &dt.HTTPTask{
			ExternalID: cliutils.XID("dtst_"),
			Method:     "POST",
			URL:        "http://localhost:54321/_test_with_body",
			Name:       "_test_with_body",
			Frequency:  "1s",
			AdvanceOptions: &dt.HTTPAdvanceOption{
				RequestBody: &dt.HTTPOptBody{
					BodyType: "application/unknown", // XXX: invalid body type
					Body:     `{"key": "value"}`,
				},
			},

			SuccessWhen: []*dt.HTTPSuccess{
				{
					StatusCode: []*dt.SuccessOption{
						{Is: "200"},
					},
				},
			},
		},
	},
	{
		name:      "test dial with body",
		reasonCnt: 0,
		t: &dt.HTTPTask{
			ExternalID: cliutils.XID("dtst_"),
			Method:     "POST",
			URL:        "http://localhost:54321/_test_with_body",
			Name:       "_test_with_body",
			Frequency:  "1s",
			AdvanceOptions: &dt.HTTPAdvanceOption{
				RequestBody: &dt.HTTPOptBody{
					BodyType: "application/json",
					Body:     `{"key": "value"}`,
				},
			},

			SuccessWhen: []*dt.HTTPSuccess{
				{
					StatusCode: []*dt.SuccessOption{
						{Is: "200"},
					},
				},
			},
		},
	},

	{
		name:      "test dial with headers",
		reasonCnt: 0,
		t: &dt.HTTPTask{
			ExternalID: cliutils.XID("dtst_"),
			Method:     "GET",
			URL:        "http://localhost:54321/_test_with_headers",
			Name:       "_test_with_headers",
			Frequency:  "1s",
			AdvanceOptions: &dt.HTTPAdvanceOption{
				RequestOptions: &dt.HTTPOptRequest{
					Headers: map[string]string{
						"X-Header-1": "foo",
						"X-Header-2": "bar",
					},
				},
			},

			SuccessWhen: []*dt.HTTPSuccess{
				{
					StatusCode: []*dt.SuccessOption{
						{Is: "200"},
					},
				},
			},
		},
	},

	{
		name:      "test dial with auth",
		reasonCnt: 0,
		t: &dt.HTTPTask{
			ExternalID: cliutils.XID("dtst_"),
			Method:     "GET",
			URL:        "http://localhost:54321/_test_with_basic_auth",
			Name:       "_test_with_basic_auth",
			Frequency:  "1s",
			AdvanceOptions: &dt.HTTPAdvanceOption{
				RequestOptions: &dt.HTTPOptRequest{
					Auth: &dt.HTTPOptAuth{
						Username: "foo",
						Password: "bar",
					},
				},
			},

			SuccessWhen: []*dt.HTTPSuccess{
				{
					StatusCode: []*dt.SuccessOption{
						{Is: "200"},
					},
				},
			},
		},
	},

	{
		name:      "test dial with cookie",
		reasonCnt: 0,
		t: &dt.HTTPTask{
			ExternalID: cliutils.XID("dtst_"),
			Method:     "GET",
			URL:        "http://localhost:54321/_test_with_cookie",
			Name:       "_test_with_cookie",
			Frequency:  "1s",
			AdvanceOptions: &dt.HTTPAdvanceOption{
				RequestOptions: &dt.HTTPOptRequest{
					Cookies: (&http.Cookie{
						Name:   "_test_with_cookie",
						Value:  "foo-bar",
						MaxAge: 0,
						Secure: true,
					}).String(),
				},
			},

			SuccessWhen: []*dt.HTTPSuccess{
				{
					StatusCode: []*dt.SuccessOption{
						{Is: "200"},
					},
				},
			},
		},
	},

	{
		name:      "test dial for redirect",
		reasonCnt: 0,
		t: &dt.HTTPTask{
			ExternalID: cliutils.XID("dtst_"),
			Method:     "GET",
			URL:        "http://localhost:54321/_test_redirect",
			Name:       "_test_redirect",
			Frequency:  "1s",
			AdvanceOptions: &dt.HTTPAdvanceOption{
				RequestOptions: &dt.HTTPOptRequest{FollowRedirect: true},
			},

			SuccessWhen: []*dt.HTTPSuccess{
				{
					StatusCode: []*dt.SuccessOption{
						{Is: "200"}, // allow redirect, should be 200
					},
				},
			},
		},
	},

	{
		name:      "redirect disabled",
		reasonCnt: 0,
		t: &dt.HTTPTask{
			ExternalID: cliutils.XID("dtst_"),
			Method:     "GET",
			URL:        "http://localhost:54321/_test_redirect",
			Name:       "_test_redirect_disabled",
			Frequency:  "1s",
			AdvanceOptions: &dt.HTTPAdvanceOption{
				RequestOptions: &dt.HTTPOptRequest{FollowRedirect: false},
			},

			SuccessWhen: []*dt.HTTPSuccess{
				{
					StatusCode: []*dt.SuccessOption{
						{Is: "302"}, // disabled redirect, should be 302
					},
				},
			},
		},
	},

	{
		name:      "test dial with response time checking",
		reasonCnt: 1,
		t: &dt.HTTPTask{
			ExternalID: cliutils.XID("dtst_"),
			Method:     "GET",
			URL:        "http://localhost:54321/_test_resp_time_less_10ms",
			Name:       "_test_resp_time_less_10ms",
			Frequency:  "1s",
			SuccessWhen: []*dt.HTTPSuccess{
				{ResponseTime: "10ms"},
			},
		},
	},

	{
		name:      "test dial with response headers",
		reasonCnt: 2,
		t: &dt.HTTPTask{
			ExternalID: cliutils.XID("dtst_"),
			Method:     "GET",
			URL:        "http://localhost:54321/_test_header_checking",
			Name:       "_test_header_checking",
			Region:     "hangzhou",
			Frequency:  "1s",
			SuccessWhen: []*dt.HTTPSuccess{
				{
					Header: map[string][]*dt.SuccessOption{
						"Cache-Control": {
							{MatchRegex: `max-ag=\d`}, // expect fail: max-age
						},
						"Server": {
							{Is: `Apache`}, // expect fail
						},

						"Date": {
							{Contains: "GMT"}, // ok: Date always use GMT
						},
						"NotExistHeader1": {
							{NotMatchRegex: `.+`}, // ok
						},
						"NotExistHeader2": {
							{IsNot: `abc`}, // ok
						},
						"NotExistHeader3": {
							{NotContains: `def`}, // ok
						},
					},
				},
			},
		},
	},
}

func prepareSSL(t *testing.T) {
	t.Helper()
	for k, v := range tlsData {
		if err := ioutil.WriteFile("."+k+".pem", v, 0o644); err != nil {
			t.Error(err)
		}
	}
}

func cleanTLSData() {
	for k := range tlsData {
		os.Remove("." + k + ".pem") //nolint:errcheck
	}
}

func TestDialHTTP(t *testing.T) {
	stopserver := make(chan interface{})

	defer close(stopserver)
	defer cleanTLSData()

	go proxyServer(t)
	go httpServer(t, "localhost:54321", false, stopserver) // http server
	go httpServer(t, "localhost:45323", true, stopserver)  // https server

	time.Sleep(time.Second) // wait servers ok

	for _, c := range httpCases {
		t.Run(c.name, func(t *testing.T) {
			if err := c.t.Init(); err != nil {
				if c.fail == false {
					t.Errorf("case %s failed: %s", c.t.Name, err)
				} else {
					t.Logf("expected: %s", err.Error())
				}
				return
			}

			if err := c.t.Run(); err != nil {
				if c.fail == false {
					t.Errorf("case %s failed: %s", c.t.Name, err)
				} else {
					t.Logf("expected: %s", err.Error())
				}
				return
			}

			tags, fields := c.t.GetResults()

			t.Logf("tags: %+#v", tags)
			t.Logf("fields: %+#v", fields)

			reasons := c.t.CheckResult()
			tu.Equals(t, len(reasons), c.reasonCnt)

			if len(reasons) > 0 {
				t.Logf("case %s reasons:\n\t%s",
					c.t.Name, strings.Join(reasons, "\n\t"))
			}
		})
	}
}

func httpServer(t *testing.T, bind string, https bool, exit chan interface{}) {
	t.Helper()
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	gin.DisableConsoleColor()
	r.Use(gin.Recovery())

	addTestingRoutes(t, r, https)

	// start HTTP server
	srv := &http.Server{
		Addr:    bind,
		Handler: r,
	}

	if https {
		prepareSSL(t)
		go func() {
			if err := srv.ListenAndServeTLS(".crt.pem", ".key.pem"); err != nil {
				fmt.Printf("ListenAndServeTLS(): %v \n", err)
			}
		}()
	} else {
		go func() {
			if err := srv.ListenAndServe(); err != nil {
				fmt.Printf("ListenAndServe(): %v \n", err)
			}
		}()
	}

	<-exit
	_ = srv.Shutdown(context.Background())
}

func proxyServer(t *testing.T) {
	t.Helper()
	http.HandleFunc("/_test_with_proxy", func(w http.ResponseWriter, req *http.Request) {
		t.Logf("proxied request coming")
		for k := range req.Header {
			t.Logf("proxied header: %s: %s", k, req.Header.Get(k))
		}

		fmt.Fprintf(w, "ok")
	})
	if err := http.ListenAndServe("localhost:54325", nil); err != nil {
		t.Error(err)
	}
}

func proxyHandler(t *testing.T, target string) gin.HandlerFunc {
	t.Helper()
	remote, err := url.Parse(target)
	if err != nil {
		t.Error(err)
		return nil
	}

	return func(c *gin.Context) {
		director := func(req *http.Request) {
			c.Request.URL.Scheme = remote.Scheme
			c.Request.URL.Host = remote.Host
			c.Request.URL.RawQuery = remote.RawQuery

			c.Request.Header["X-proxy-header"] = []string{c.Request.Header.Get("X-proxy-header")}
			delete(c.Request.Header, "X-proxy-header")
		}
		proxy := &httputil.ReverseProxy{Director: director}
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

var tlsData = map[string][]byte{
	"csr": []byte(`-----BEGIN CERTIFICATE REQUEST-----
MIICtzCCAZ8CAQAwcjELMAkGA1UEBhMCQ04xDDAKBgNVBAgMA2ZvbzEMMAoGA1UE
BwwDYmFyMQ0wCwYDVQQKDARmb28xMQ0wCwYDVQQLDARiYXIxMQ0wCwYDVQQDDARm
b28yMRowGAYJKoZIhvcNAQkBFgtmb29AYmFyLmNvbTCCASIwDQYJKoZIhvcNAQEB
BQADggEPADCCAQoCggEBANEQvuwHLDTsu+QuIEXc4R8aTSFTgFl0CPz3GzAhZnYt
/MgZ66iu6W7FplTiqIPoSgTqccCcWPlOgaad0BfkmbuYaoo9SiF5/ewip6QXfpBQ
Va34Q92E3EfBv5vyuCgMyNbjXb+hHbRvYmgOUeL0J9jRMJ1l981pgYgjGdzNFJTz
S2FDVoaoexOkoOXblAI9tqbBi+9+Sbu7Q/DgeudeF07VmI+cZzZX73Oo2EzwHHXn
FI5OAbRBjKsQyU9O6TgarA/5n0hAH2bcHkoCxq4iVgHuZjK2xV8hJU8b4jGevDVE
TCwE07LJJfP2RnlYe7nBqNlNXApMSllUgX4c8RhUuQ0CAwEAAaAAMA0GCSqGSIb3
DQEBCwUAA4IBAQC4Z2ELAmnrPSoghuEyKuM2GsvRqOIUHKKwM/lCWxOE/o/pQDTY
OcC+2BwSimChoBd1TY3vM03TYxzY5jlHqfwLAFqJv51DFlTasHUhlo8+7IVR+6TE
WH9latBruNVSDZ5/qL1dfbLoBw6yyQi4kYdSg1T5CBtGVCe3iBC42NmXHqp5/XXB
dQAILNu1lzVi4dM6FbHcr6FTSZBIyYrHTYLPIj4aUQ/p5iO1jYvfM8DiXR0OWfzw
VFjOt6N0pYsfLgeOHA8v6NZMQ+N59Ne0Dl7Pg7bK56qP+l0R2hY0smXH/IPrGaFF
Qf01BfPoTUfoyV195ZF8BpeVtT1HBs3of/+6
-----END CERTIFICATE REQUEST-----`),
	"crt": []byte(`-----BEGIN CERTIFICATE-----
MIIDYDCCAkgCCQCgwJeKz0Yl5jANBgkqhkiG9w0BAQUFADByMQswCQYDVQQGEwJD
TjEMMAoGA1UECAwDZm9vMQwwCgYDVQQHDANiYXIxDTALBgNVBAoMBGZvbzExDTAL
BgNVBAsMBGJhcjExDTALBgNVBAMMBGZvbzIxGjAYBgkqhkiG9w0BCQEWC2Zvb0Bi
YXIuY29tMB4XDTIxMDIyMjA0MDAwNloXDTIyMDIyMjA0MDAwNlowcjELMAkGA1UE
BhMCQ04xDDAKBgNVBAgMA2ZvbzEMMAoGA1UEBwwDYmFyMQ0wCwYDVQQKDARmb28x
MQ0wCwYDVQQLDARiYXIxMQ0wCwYDVQQDDARmb28yMRowGAYJKoZIhvcNAQkBFgtm
b29AYmFyLmNvbTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBANEQvuwH
LDTsu+QuIEXc4R8aTSFTgFl0CPz3GzAhZnYt/MgZ66iu6W7FplTiqIPoSgTqccCc
WPlOgaad0BfkmbuYaoo9SiF5/ewip6QXfpBQVa34Q92E3EfBv5vyuCgMyNbjXb+h
HbRvYmgOUeL0J9jRMJ1l981pgYgjGdzNFJTzS2FDVoaoexOkoOXblAI9tqbBi+9+
Sbu7Q/DgeudeF07VmI+cZzZX73Oo2EzwHHXnFI5OAbRBjKsQyU9O6TgarA/5n0hA
H2bcHkoCxq4iVgHuZjK2xV8hJU8b4jGevDVETCwE07LJJfP2RnlYe7nBqNlNXApM
SllUgX4c8RhUuQ0CAwEAATANBgkqhkiG9w0BAQUFAAOCAQEAn8TzH9LvNyhH+cqa
gRc8Gqj0ccPf1LkW9dIlTlk31HBHzfKI7xhul23PimMuz6hg3YCAttXhKXrVoiIJ
1rQUngGr0e2CkesxfeaMxDPPRCRiLPRLzsryjvJI/eS2rmxtmUyC0X5aR+/2F8Ha
p2JXig4KUhYwMmttnd/Qbjmc0C397zKudBxkIoxprIN/gVhRBJJRqxN8bgeL8JsH
2HfsA/SzFDUOYPQhw0EnyLukRuQi0C3soKL4fIUGqonJPQ0TIceJRMGrtIj0h7j+
oNbJXTP7ABRYVmFRYViczu86MWsbHkif4bWqhPJeC0K+cp1UuwykJ+4XzM5WDR/+
InEHyg==
-----END CERTIFICATE-----`),

	"key": []byte(`-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDREL7sByw07Lvk
LiBF3OEfGk0hU4BZdAj89xswIWZ2LfzIGeuoruluxaZU4qiD6EoE6nHAnFj5ToGm
ndAX5Jm7mGqKPUohef3sIqekF36QUFWt+EPdhNxHwb+b8rgoDMjW412/oR20b2Jo
DlHi9CfY0TCdZffNaYGIIxnczRSU80thQ1aGqHsTpKDl25QCPbamwYvvfkm7u0Pw
4HrnXhdO1ZiPnGc2V+9zqNhM8Bx15xSOTgG0QYyrEMlPTuk4GqwP+Z9IQB9m3B5K
AsauIlYB7mYytsVfISVPG+Ixnrw1REwsBNOyySXz9kZ5WHu5wajZTVwKTEpZVIF+
HPEYVLkNAgMBAAECggEAJp/9ZgX9ONnz7LhI5h9kyCZH0bxnnh89+d59e2rwTy03
4pBHZabLIdgKXuxxTc2Av1/BHGDGX2kNswa9B20IqgwCwv+Hzp+HNjVA26QrkeYF
rlqLz0VYnTlCeUFinKOgB3OCQoE1x7w8ZhUfM9r/8aLUZIAORDkV4Vz6zjxlbQ8g
JxHrZ5eZexTzSVylVFZda3AgtqMr1N6ZzMejtYqttGGDDmh372QgykvxhmEIeHAf
g1bW86oOedxxfZ0003/F9He6qvdWmAKfbQczCNKBPHgGpdcuTTBsj/ieB/31AZG9
R1CUopzAklrUXzv1SBxw/5mJdOcmTUH4Hpdl4vXh0QKBgQD99FiKIKRxWZiHcbV4
X2wl0AZsMUbUT+BVKRbdfYk0pTstSKaQMpEB2ojvVqW8HVN83+jJxWUxxGWnT0Mn
wfw9lavhNS14klj+rJw6zI4m2lcI8t+P9HxTMDfBU+LiMnlUFK44u7Mx6Vr/dm9p
53o0aGapLOQfwps+UdJ86ZCAKwKBgQDSv9az1zHE1AtJx7UlreduzXrYjzJqrgYX
ufjLu+aTsSWNXIlIxG5gkKbkF6R4VVmpXkF7B8nJ3IrsrRuwMZpMjyhLl2LLCnGL
XgAgz/SNjxS4Clo1PVcP2ZoANVnPs/+DRlI1aTqXHZA5sJ1d2a9e385ndZ+/Qg+q
giRNOsfXpwKBgQC2dwnmtO1yQ93D839frbAWuxDiS8WIZpvYlF1JZxleKhoKv1ht
4uctXcdlr+wE7U0/O+IWly3ORD6Fp/2oY0jJNvD4Ly0spHotAfh+htrcL6S5WUgo
NpHdc5eb4JnzzDBAqVtEiBiIlBI92urSPO8hGKIqi4adC0Zf0IpcFbUtYQKBgF24
Iepn0CIPidWNkejnpPuJNRAI3grCyMLUWOeA79DN/j0W4ZYShGM88HqOaP16Nx0y
ZTwpAntaMA2ADcgUxuE06F51O+G/Cy9G5hexYrdw4W3WbLcwR/8sbWeaUg4jpYTj
SLunx/5bjz+YYuLRY0N1k3w+uoN7BSx2I16UvToRAoGAEFhhsGTxXLeNOMDU1jhJ
cbbypRkGjSoxUbn7apEMwdpeDPQwWwkwi634rjVcTIQuO/8HMbjZi2AZcM5TWNY0
HHrpiTXtbrUfbKX2TEk3DSevJ9EZEuewxALtsaRQgX4WyHlxpYDXNSjag04Nn+/x
9WKHZvRf3lbLY7GAR/emacU=
-----END PRIVATE KEY-----`),
}

func addTestingRoutes(t *testing.T, r *gin.Engine, https bool) {
	t.Helper()
	r.GET("/_test_resp_time_less_10ms", func(c *gin.Context) {
		time.Sleep(time.Millisecond * 11)
		c.Data(http.StatusOK, ``, nil)
	})

	r.GET("/_test_header_checking", func(c *gin.Context) {
		c.DataFromReader(http.StatusOK, 0, "", bytes.NewBuffer([]byte("")),
			map[string]string{
				"Cache-Control": "max-age=1024",
				"Server":        "dialtesting-server",
			})
	})

	r.GET("/_test_redirect", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/_redirect_to_me")
	})
	r.GET("/_redirect_to_me", func(c *gin.Context) {
		t.Log("redirect ok")
		c.Data(http.StatusOK, ``, nil)
	})

	r.GET("/_test_with_cookie", func(c *gin.Context) {
		cookies := c.Request.Cookies()
		for _, c := range cookies {
			t.Logf("%s", c.String())
		}

		c.Data(http.StatusOK, ``, nil)
	})

	r.GET("/_test_with_basic_auth", func(c *gin.Context) {
		user, pwd, ok := c.Request.BasicAuth()
		if !ok {
			t.Errorf("basic auth failed")
		} else {
			t.Logf("user: %s, password: %s", user, pwd)
		}

		c.Data(http.StatusOK, ``, nil)
	})

	r.GET("/_test_with_headers", func(c *gin.Context) {
		for k := range c.Request.Header {
			t.Logf("%s: %s", k, c.Request.Header.Get(k))
		}

		c.Data(http.StatusOK, ``, nil)
	})

	r.POST("/_test_with_body", func(c *gin.Context) {
		defer c.Request.Body.Close() //nolint:errcheck
		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			t.Error(err)
		}

		t.Logf("body: %s", string(body))

		c.Data(http.StatusOK, ``, nil)
	})

	r.GET("/_test_with_proxy",
		proxyHandler(t, "http://localhost:54325/_test_with_proxy" /*url must be the same*/))

	if https {
		r.GET("/_test_with_cert", func(c *gin.Context) {
			t.Logf("request tls: %+#v", c.Request.TLS)
			c.Data(http.StatusOK, ``, nil)
		})
	}
}
