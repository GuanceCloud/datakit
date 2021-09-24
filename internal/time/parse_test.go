package time

import (
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	cases := []struct {
		du     string
		expect time.Duration
		fail   bool
	}{
		{du: `1ns`, expect: time.Duration(1)},
		{du: `1us`, expect: time.Microsecond},
		{du: `1ms`, expect: time.Millisecond},
		{du: `1s`, expect: time.Second},
		{du: `1h`, expect: time.Hour},
		{du: `1d`, expect: time.Hour * 24},
		{du: `1w`, expect: time.Hour * 24 * 7},
		{du: `1y`, expect: time.Hour * 24 * 365},
		{du: `1ym`, fail: true},
	}

	for _, c := range cases {
		du, err := ParseDuration(c.du)
		if err != nil && !c.fail {
			t.Fatalf("%s should ok", c.du)
		}

		if c.fail && err == nil {
			t.Fatalf("%s should fail", c.du)
		}

		if du != c.expect {
			t.Fatalf("got %q, expect %q", du, c.du)
		}

		t.Logf("%s is %v", c.du, du)
	}
}
