// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package filter

import (
	"encoding/json"
	"testing"
	"time"

	lp "github.com/GuanceCloud/cliutils/lineproto"
	tu "github.com/GuanceCloud/cliutils/testutil"
	"github.com/stretchr/testify/assert"

	"github.com/GuanceCloud/cliutils/point"

	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
)

type pullMock struct{ pullCount int }

func (dw *pullMock) Pull(_ string) ([]byte, error) {
	filters := map[string]FilterConditions{
		"logging": {
			`{ source = "test1" and ( f1 in ["1", "2", "3"] )}`,
			`{ source = "test2" and ( f1 in ["1", "2", "3"] )}`,
			`{ source = "nginx-ingress-controller" and ( urihost notin ["mall-dev.xxxxxxxx.com", "mall-staging.xxxxxxxx.com", "mall-app.xxxxxxxx.com"] )}`,
		},

		"tracing": {
			`{ service = "test1" and ( f1 in ["1", "2", "3"] or t1 match [ 'abc.*'])}`,
			`{ service = re("test2") and ( f1 in ["1", "2", "3"] or t1 match [ 'def.*'])}`,
		},
	}

	dw.pullCount++

	return json.Marshal(&Filters{
		Filters:      filters,
		PullInterval: time.Duration(dw.pullCount) * time.Millisecond,
	})
}

func TestFilter(t *testing.T) {
	f := newFilter(&pullMock{})

	cases := []struct {
		name      string
		pts       string
		category  point.Category
		expectPts int
	}{
		{
			pts: `test1 f1="1",f2=2i,f3=3 124
test1 f1="2",f2=2i,f3=3 124
test1 f1="3",f2=2i,f3=3 125`,
			category:  point.Logging,
			expectPts: 0,
		},

		{
			pts: `test1,service=test1 f1="1",f2=2i,f3=3 123
test1,service=test1 f1="1",f2=2i,f3=3 124
test1,service=test1 f1="1",f2=2i,f3=3 125`,
			category:  point.Tracing,
			expectPts: 0,
		},

		{
			pts:       `nginx-ingress-controller,service=nginx-ingress-controller agent="Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1",body_bytes_sent="217",browser="Safari",browserVer="13.0.3",city="",client_ip="124.126.18.162",country="",engine="AppleWebKit",engineVer="605.1.15",http_method="GET",http_referer="http://127.0.0.1:9098/wtrain",http_url="/mall-web/csc/basics/company/canPay?id=xxxxxxxxxxxxxxxxxx",http_version="1.1",isBot=false,isMobile=true,isp="unknown",os="CPU iPhone OS 13_2_3 like Mac OS X",port="",province="",proxy_upstream_name="dev-mall-apisix-80",req_id="xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",request_body="",request_length="2296",request_time=0.031,status="info",status_code="200",ua="iPhone",upstream_addr="172.19.251.218:9080",upstream_response_length="258",upstream_response_time=0.031,upstream_status="200",urihost="mall-dev.xxxxxxxx.com"`,
			category:  point.Logging,
			expectPts: 1,
		},
	}

	f.pull("")
	for k, v := range f.conditions {
		t.Logf("%s: %s", k, v)
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pts, err := lp.ParsePoints([]byte(tc.pts), nil)
			if err != nil {
				t.Error(err)
				return
			}

			after, _ := f.doFilter(tc.category, dkpt.WrapPoint(pts))
			tu.Assert(t, len(after) == tc.expectPts, "expect %d pts, got %d", tc.expectPts, len(after))
		})
	}
}

func TestPull(t *testing.T) {
	f := newFilter(&pullMock{})

	round := 3
	for i := 0; i < round; i++ {
		f.pull("")
		// test if reset tick ok
		<-f.tick.C
	}

	tu.Assert(t, f.pullInterval == time.Millisecond*time.Duration(round), "expect %ds, got %s", round, f.pullInterval)
}

// go test -v -timeout 30s -run ^TestGetConds$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/filter
func TestGetConds(t *testing.T) {
	cases := []struct {
		name      string
		inFilters []string
		out       string
	}{
		{
			name:      "cpu",
			inFilters: []string{`{cpu='cpu-total'}`},
			out:       `{cpu = 'cpu-total'}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := GetConds(tc.inFilters)
			assert.NoError(t, err)

			var str string
			for _, v := range out {
				str += v.String()
			}
			assert.Equal(t, tc.out, str)
		})
	}
}
