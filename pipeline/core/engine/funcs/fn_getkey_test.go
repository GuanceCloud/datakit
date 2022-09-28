// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetkey(t *testing.T) {
	cases := []struct {
		name, pl, in string
		expect       interface{}
		fail         bool
	}{
		{
			name: "value type: string",
			in:   `1.2.3.4 - - [29/Nov/2021:07:30:50 +0000] "POST /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413 "-" "Mozilla/4.0"`,
			pl: `
			key = "shanghai"
			add_key(key)
			key = "tokyo" 
			add_key(add_new_key, key)
`,
			expect: "tokyo",
			fail:   false,
		},
		{
			name: "value type: string",
			in:   `1.2.3.4 - - [29/Nov/2021:07:30:50 +0000] "POST /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413 "-" "Mozilla/4.0"`,
			pl: `
			key = "shanghai"
			add_key(key)
			key = "tokyo" 
			add_key(add_new_key, get_key(key))
`,
			expect: "shanghai",
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
			_, _, f, _, _, err := runScript(runner, "test", nil, map[string]any{"message": tc.in}, time.Now())

			if err == nil {
				v := f["add_new_key"]
				assert.Equal(t, tc.expect, v)
				t.Logf("[%d] PASS", idx)
			} else {
				t.Error(err)
			}
		})
	}
}
