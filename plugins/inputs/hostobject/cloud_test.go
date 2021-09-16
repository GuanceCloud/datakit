package hostobject

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/influxdata/influxdb1-client/models"

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

	tags := models.Tags{models.NewTag([]byte("a"), []byte(`~!@#$%^&*()_+=-|}{\][":';?><,./`))}

	for _, tc := range cases {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, tc.body)
		}))

		x := metaGet(ts.URL)

		tu.Equals(t, tc.expect, x)

		ts.Close()

		pt1, err := models.NewPoint("test", tags,
			map[string]interface{}{"extra_cloud_meta": x}, time.Now())
		if err != nil {
			t.Error(err)
		}

		pts, err := models.ParsePointsWithPrecision([]byte(pt1.String()), time.Now(), "n")
		if err != nil {
			t.Error(err)
		}

		tu.Equals(t, 1, len(pts))

		tu.Equals(t, pt1.String(), pts[0].String())

		t.Logf("pt: %s", pt1.String())
	}
}
