package funcs

import (
	"testing"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestIfelse(t *testing.T) {
	cases := []struct {
		name   string
		pl, in string
		expect interface{}
		fail   bool
	}{
		{
			name: "no-grok-ok",
			in:   ``,
			pl: `
add_key(score, 95)

if score == 100  {
    add_key(add_new_key, "S")
} elif 90 <= score && score < 100 {
    add_key(add_new_key, "A")
} elif 75 <= score && score < 90 {
    add_key(add_new_key, "B")
} elif 60 <= score && score < 70 {
    add_key(add_new_key, "C")
} else {
    add_key(add_new_key, "D")
}
`,
			expect: "A",
		},
		{
			name: "client-ip-2-and-stauts-code-fail",
			in:   `1.2.3.4 "POST /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413`,
			pl: `
grok(_, "%{IPORHOST:client_ip} \"%{DATA} %{GREEDYDATA} HTTP/%{NUMBER}\" %{INT:status_code} %{INT:bytes}")
cast(status_code, "int")

if invalid_status_code == 300 {
    add_key(add_new_key, "NO")
} elif (client_ip == "1.2.3.4.5" || client_ip == "1.2.3.4") && invalid_status_code == 300 {
    add_key(add_new_key, "OK")
}
`,
			fail: true,
		},
		{
			name: "elif-client-ip-2-and-stauts-code-fail",
			in:   `1.2.3.4 "POST /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413`,
			pl: `
grok(_, "%{IPORHOST:client_ip} \"%{DATA} %{GREEDYDATA} HTTP/%{NUMBER}\" %{INT:status_code} %{INT:bytes}")
cast(status_code, "int")

if status_code == 300 {
    add_key(add_new_key, "NO")
} elif (client_ip == "1.2.3.4.5" || client_ip == "1.2.3.4") && status_code == 300 {
    add_key(add_new_key, "OK")
}
`,
			fail: true,
		},
		{
			name: "client-ip-2-and-stauts-code-fail",
			in:   `1.2.3.4 "POST /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413`,
			pl: `
grok(_, "%{IPORHOST:client_ip} \"%{DATA} %{GREEDYDATA} HTTP/%{NUMBER}\" %{INT:status_code} %{INT:bytes}")
cast(status_code, "int")

if (client_ip == "1.2.3.4.5" || client_ip == "1.2.3.4") && status_code == 300 {
    add_key(add_new_key, "OK")
}
`,
			fail: true,
		},
		{
			name: "elif-client-ip-2-and-stauts-code2-ok",
			in:   `1.2.3.4 "POST /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413`,
			pl: `
grok(_, "%{IPORHOST:client_ip} \"%{DATA} %{GREEDYDATA} HTTP/%{NUMBER}\" %{INT:status_code} %{INT:bytes}")
cast(status_code, "int")

if status_code == 300 {
    add_key(add_new_key, "NO")
} elif (client_ip == "1.2.3.4.5" && client_ip == "1.2.3.4.6") || (status_code == 200 || status_code == 100) {
    add_key(add_new_key, "OK")
}
`,
			expect: "OK",
		},
		{
			name: "elif-client-ip-2-and-stauts-code-ok",
			in:   `1.2.3.4 "POST /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413`,
			pl: `
grok(_, "%{IPORHOST:client_ip} \"%{DATA} %{GREEDYDATA} HTTP/%{NUMBER}\" %{INT:status_code} %{INT:bytes}")
cast(status_code, "int")

if status_code == 300 {
    add_key(add_new_key, "NO")
} elif (client_ip == "1.2.3.4.5" || client_ip == "1.2.3.4") && status_code == 200 {
    add_key(add_new_key, "OK")
}
`,
			expect: "OK",
		},
		{
			name: "client-ip-2-and-stauts-code-ok",
			in:   `1.2.3.4 "POST /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413`,
			pl: `
grok(_, "%{IPORHOST:client_ip} \"%{DATA} %{GREEDYDATA} HTTP/%{NUMBER}\" %{INT:status_code} %{INT:bytes}")
cast(status_code, "int")

if (client_ip == "1.2.3.4.5" || client_ip == "1.2.3.4") && status_code == 200 {
    add_key(add_new_key, "OK")
}
`,
			expect: "OK",
		},
		{
			name: "client-ip-or-stauts-code-ok",
			in:   `1.2.3.4 "POST /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413`,
			pl: `
grok(_, "%{IPORHOST:client_ip} \"%{DATA} %{GREEDYDATA} HTTP/%{NUMBER}\" %{INT:status_code} %{INT:bytes}")
cast(status_code, "int")

if client_ip == "1.2.3.4.5" || status_code == 200 {
    add_key(add_new_key, "OK")
}
`,
			expect: "OK",
		},
		{
			name: "client-ip-and-stauts-code-ok",
			in:   `1.2.3.4 "POST /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413`,
			pl: `
grok(_, "%{IPORHOST:client_ip} \"%{DATA} %{GREEDYDATA} HTTP/%{NUMBER}\" %{INT:status_code} %{INT:bytes}")
cast(status_code, "int")

if client_ip == "1.2.3.4" && status_code == 200 {
    add_key(add_new_key, "OK")
}
`,
			expect: "OK",
		},
		{
			name: "client-ip-ok",
			in:   `1.2.3.4 "POST /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413`,
			pl: `
grok(_, "%{IPORHOST:client_ip} \"%{DATA} %{GREEDYDATA} HTTP/%{NUMBER}\" %{INT:status_code} %{INT:bytes}")

if client_ip == "1.2.3.4" {
    add_key(add_new_key, "OK")
}
`,
			expect: "OK",
		},
		{
			name: "status-code-ok",
			in:   `1.2.3.4 "POST /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413`,
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
			name: "status-code-list",
			in:   `1.2.3.4 "POST /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413`,
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
			name: "status-code-list2",
			in:   `1.2.3.4 "POST /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413`,
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
			name: "status-code-is-not-nil",
			in:   `1.2.3.4 "POST /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413`,
			pl: `
grok(_, "%{IPORHOST:client_ip} \"%{DATA} %{GREEDYDATA} HTTP/%{NUMBER}\" %{INT:status_code} %{INT:bytes}")

if status_code != nil {
  add_key(add_new_key, "OK")
}
`,
			expect: "OK",
		},
		{
			name: "valind-key-is-nil",
			in:   `1.2.3.4 "POST /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413`,
			pl: `
grok(_, "%{IPORHOST:client_ip} \"%{DATA} %{GREEDYDATA} HTTP/%{NUMBER}\" %{INT:status_code} %{INT:bytes}")

if invalid_status_code == nil {
  add_key(add_new_key, "OK")
}
`,
			expect: "OK",
		},
		{
			name: "invalid-key-is-not-nil",
			in:   `1.2.3.4 "POST /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413`,
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

			if err := runner.Run(tc.in); err == nil {
				// t.Log(runner.Result())
				v, _ := runner.GetContent("add_new_key")
				tu.Equals(t, tc.expect, v)
				t.Logf("[%d] PASS", idx)
			} else {
				t.Error(err)
			}
		})
	}
}
