package funcs

import (
	"testing"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestIfelse(t *testing.T) {
	cases := []struct {
		pl, in string
		expect interface{}
		fail   bool
	}{
		{
			in: `1.2.3.4 "POST /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413`,
			pl: `
grok(_, "%{IPORHOST:client_ip} \"%{DATA} %{GREEDYDATA} HTTP/%{NUMBER}\" %{INT:status_code} %{INT:bytes}")

if client_ip == "1.2.3.4" {
    add_key(add_new_key, "OK")
}
`,
			expect: "OK",
		},
		{
			in: `1.2.3.4 "POST /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413`,
			pl: `
grok(_, "%{IPORHOST:client_ip} \"%{DATA} %{GREEDYDATA} HTTP/%{NUMBER}\" %{INT:status_code} %{INT:bytes}")
cast(status_code, "int")

if status_code == 200 {
    add_key(add_new_key, "OK")
}
`,
			expect: "OK",
		},
		{
			in: `1.2.3.4 "POST /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413`,
			pl: `
grok(_, "%{IPORHOST:client_ip} \"%{DATA} %{GREEDYDATA} HTTP/%{NUMBER}\" %{INT:status_code} %{INT:bytes}")
cast(status_code, "int")

if status_code == 400 {
    add_key(add_new_key, "ERROR")
} elif status_code == 200 {
    add_key(add_new_key, "OK")
}
`,
			expect: "OK",
		},
		{
			in: `1.2.3.4 "POST /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413`,
			pl: `
grok(_, "%{IPORHOST:client_ip} \"%{DATA} %{GREEDYDATA} HTTP/%{NUMBER}\" %{INT:status_code} %{INT:bytes}")
cast(status_code, "int")

if status_code == 500 {
    add_key(add_new_key, "FATAL")
} elif status_code == 400 {
    add_key(add_new_key, "ERROR")
} else {
    add_key(add_new_key, "OK")
}
`,
			expect: "OK",
		},
		{
			in: `1.2.3.4 "POST /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413`,
			pl: `
grok(_, "%{IPORHOST:client_ip} \"%{DATA} %{GREEDYDATA} HTTP/%{NUMBER}\" %{INT:status_code} %{INT:bytes}")

if status_code != nil {
  add_key(add_new_key, "OK")
}
`,
			expect: "OK",
		},
		{
			in: `1.2.3.4 "POST /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413`,
			pl: `
grok(_, "%{IPORHOST:client_ip} \"%{DATA} %{GREEDYDATA} HTTP/%{NUMBER}\" %{INT:status_code} %{INT:bytes}")

if invalid_status_code == nil {
  add_key(add_new_key, "OK")
}
`,
			expect: "OK",
		},
		{
			in: `1.2.3.4 "POST /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413`,
			pl: `
grok(_, "%{IPORHOST:client_ip} \"%{DATA} %{GREEDYDATA} HTTP/%{NUMBER}\" %{INT:status_code} %{INT:bytes}")

if invalid_status_code != nil {
  add_key(add_new_key, "OK")
}
`,
			fail: true,
		},
	}
	for idx, tc := range cases {
		runner, err := NewTestingRunner(tc.pl)
		if err != nil {
			if tc.fail {
				t.Logf("[%d]expect error: %s", idx, err)
			} else {
				t.Errorf("[%d] failed: %s", idx, err)
			}
			return
		}

		if err := runner.Run(tc.in); err == nil {
			// t.Log(runner.Result())
			v, _ := runner.GetContent("add_new_key")
			tu.Equals(t, tc.expect, v)
			t.Logf("[%d] PASS", idx)
		} else {
			t.Error(err)
		}
	}
}
