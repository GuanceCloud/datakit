package funcs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSetMeasurement(t *testing.T) {
	cases := []struct {
		name, pl, in string
		del          bool
		out          string
		expect       string
		fail         bool
	}{
		{
			name: "set_measurement 0",
			in:   `162.62.81.1 - - [29/Nov/2021:07:30:50 +0000] "123 /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413 "-" "Mozilla/4.0"`,
			pl: `
		set_tag(client_ip)
		grok(_, "%{IPORHOST:client_ip} %{NOTSPACE} %{NOTSPACE} \\[%{HTTPDATE:time}\\] \"%{DATA:data} %{GREEDYDATA} HTTP/%{NUMBER}\" %{INT:status_code} %{INT:bytes}")
		cast(data, "int")
		set_measurement(client_ip)

		`,
			del:    true,
			out:    "client_ip",
			expect: "162.62.81.1",
			fail:   false,
		},
		{
			name: "set_measurement 1",
			in:   `162.62.81.1 - - [29/Nov/2021:07:30:50 +0000] "123 /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413 "-" "Mozilla/4.0"`,
			pl: `
		grok(_, "%{IPORHOST:client_ip} %{NOTSPACE} %{NOTSPACE} \\[%{HTTPDATE:time}\\] \"%{DATA:data} %{GREEDYDATA} HTTP/%{NUMBER}\" %{INT:status_code} %{INT:bytes}")
		set_tag(client_ip)
		cast(data, "int")
		set_measurement(client_ip, true)

		`,
			del:    false,
			out:    "client_ip",
			expect: "162.62.81.1",
			fail:   false,
		},
	}

	for idx, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runner, err := NewTestingRunner(tc.pl)
			if err != nil {
				if tc.fail {
					t.Logf("[%d]expect error: %s", idx, err)
				} else {
					t.Errorf("[%d] failed: %s", idx, err)
				}
				return
			}
			ret, err := runner.Run("test", map[string]string{},
				map[string]interface{}{
					"message": tc.in,
				}, "message", time.Now())
			if err != nil || ret.Error != nil {
				t.Error(err, " ", ret.Error)
				return
			}
			t.Log(ret)
			_, ok := ret.Tags[tc.out]
			assert.Equal(t, tc.del, ok)
			assert.Equal(t, tc.expect, ret.Measurement)
			t.Logf("[%d] PASS", idx)
		})
	}
}
