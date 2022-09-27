// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package hostobject

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestMetaGet(t *testing.T) {
	cases := []struct {
		body, expect string
	}{
		{
			body: `multi-
lin-
data`,
			expect: `multi- lin- data`,
		},

		{
			body:   `中文balabala`,
			expect: `中文balabala`,
		},

		{
			body:   `¡™£¢∞§¶•ªº–≠‘«“æ…÷≥≤`,
			expect: `¡™£¢∞§¶•ªº–≠‘«“æ…÷≥≤`,
		},

		{
			body:   `~!@#$%^&*()_+-=|}{\][":';?><,./`,
			expect: `~!@#$%^&*()_+-=|}{\][":';?><,./`,
		},

		{
			body: `		abc`,
			expect: `		abc`,
		},
	}

	tags := map[string]string{"a": `~!@#$%^&*()_+=-|}{\][":';?><,./`}

	for _, tc := range cases {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, tc.body)
		}))

		x := metaGet(ts.URL)

		tu.Equals(t, tc.expect, x)

		ts.Close()

		pt1, err := lineproto.NewPoint("test", tags, map[string]interface{}{"extra_cloud_meta": x}, time.Now())
		if err != nil {
			t.Fatal(err)
		}

		encoder := lineproto.NewLineEncoder()
		if err := encoder.AppendPoint(pt1); err != nil {
			t.Fatal(err)
		}
		lines, err := encoder.Bytes()
		if err != nil {
			t.Fatal(err)
		}
		l1 := string(lines)

		pts, err := lineproto.ParseWithOptionSetter(lines, lineproto.WithTime(time.Now()), lineproto.WithPrecisionV2(lineproto.Nanosecond))
		if err != nil {
			t.Fatal(err)
		}
		tu.Equals(t, 1, len(pts))

		encoder.Reset()
		if err := encoder.AppendPoint(pts[0]); err != nil {
			t.Fatal(err)
		}

		b2, err := encoder.Bytes()
		if err != nil {
			t.Fatal(err)
		}
		l2 := string(b2)

		tu.Equals(t, l1, l2)

		t.Logf("pt: %s", l1)
	}
}
