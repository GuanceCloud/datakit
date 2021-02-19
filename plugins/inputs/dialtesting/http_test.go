package dialtesting

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
)

var httpCases = []struct {
	t         *httpTask
	fail      bool
	reasonCnt int
}{

	// test redirect
	{
		reasonCnt: 0,
		t: &httpTask{
			TID:       cliutils.XID("dtst"),
			Method:    "GET",
			URL:       "http://localhost:54321/_test_redirect",
			Name:      "_test_redirect",
			Locations: []string{"hangzhou"},
			Frequency: "1s",
			AdvanceOptions: []*httpAdvanceOption{
				&httpAdvanceOption{
					RequestsOptions: &httpOptRequest{FollowRedirect: true},
				},
			},

			SuccessWhen: []*httpSuccess{
				&httpSuccess{
					StatusCode: &successOption{Is: "200"}, // allow redirect, should be 200
				},
			},
		},
	},

	{
		reasonCnt: 0,
		t: &httpTask{
			TID:       cliutils.XID("dtst"),
			Method:    "GET",
			URL:       "http://localhost:54321/_test_redirect",
			Name:      "_test_redirect_disabled",
			Locations: []string{"hangzhou"},
			Frequency: "1s",
			AdvanceOptions: []*httpAdvanceOption{
				&httpAdvanceOption{
					RequestsOptions: &httpOptRequest{FollowRedirect: false},
				},
			},

			SuccessWhen: []*httpSuccess{
				&httpSuccess{
					StatusCode: &successOption{Is: "302"}, // disabled redirect, should be 302
				},
			},
		},
	},

	// test response time
	{
		reasonCnt: 1,
		t: &httpTask{
			TID:       cliutils.XID("dialt_"),
			Method:    "GET",
			URL:       "http://localhost:54321/_test_resp_time_less_10ms",
			Name:      "_test_resp_time_less_10ms",
			Frequency: "1s",
			Locations: []string{"hangzhou"},
			SuccessWhen: []*httpSuccess{
				&httpSuccess{ResponseTime: "10ms"},
			},
		},
	},

	// test response headers
	{
		reasonCnt: 2,
		t: &httpTask{
			TID:       cliutils.XID("dtst"),
			Method:    "GET",
			URL:       "http://localhost:54321/_test_header_checking",
			Name:      "_test_header_checking",
			Locations: []string{"hangzhou"},
			Frequency: "1s",
			SuccessWhen: []*httpSuccess{
				&httpSuccess{
					Header: map[string]*successOption{
						"Date":            &successOption{Contains: "GMT"},         // Date always use GMT
						"Cache-Control":   &successOption{MatchRegex: `max-ag=\d`}, // match fail
						"Server":          &successOption{Is: `Apache`},            // ok
						"NotExistHeader1": &successOption{NotMatchRegex: `.+`},     // ok
						"NotExistHeader2": &successOption{IsNot: `abc`},            // ok
						"NotExistHeader3": &successOption{NotContains: `def`},      // ok
					},
				},
			},
		},
	},
}

func TestDialHTTP(t *testing.T) {

	stopserver := make(chan interface{})

	defer close(stopserver)

	httpServer := func() {

		gin.SetMode(gin.ReleaseMode)

		r := gin.New()
		gin.DisableConsoleColor()
		r.Use(gin.Recovery())

		addTestingRoutes(r)

		// start HTTP server
		srv := &http.Server{
			Addr:    "localhost:54321",
			Handler: r,
		}

		go func() {
			if err := srv.ListenAndServe(); err != nil {
				t.Logf("ListenAndServe(): %s", err)
			}
		}()

		<-stopserver
		_ = srv.Shutdown(context.Background())
	}

	go httpServer()

	for _, c := range httpCases {

		if err := c.t.Init(); err != nil && c.fail == false {
			t.Errorf("case %s failed: %s", c.t.Name, err)
		}

		if err := c.t.Run(); err != nil && c.fail == false {
			t.Errorf("case %s failed: %s", c.t.Name, err)
		}

		reasons := c.t.CheckResult()
		if len(reasons) != c.reasonCnt {
			t.Errorf("case %s expect %d reasons, but got %d reasons:\n\t%s",
				c.t.Name, c.reasonCnt, len(reasons), strings.Join(reasons, "\n\t"))
		} else {
			if len(reasons) > 0 {
				t.Logf("case %s reasons:\n\t%s",
					c.t.Name, strings.Join(reasons, "\n\t"))
			}
		}
	}
}

func addTestingRoutes(r *gin.Engine) {
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
		l.Debug("redirect ok")
		c.Data(http.StatusOK, ``, nil)
	})
}
