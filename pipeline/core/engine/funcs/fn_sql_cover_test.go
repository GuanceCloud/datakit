// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"testing"
	"time"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestSqlCover(t *testing.T) {
	cases := []struct {
		name     string
		outKey   string
		pl, in   string
		expected interface{}
		fail     bool
	}{
		{
			name: "normal",
			pl: `json(_, str) 
			sql_cover(str)`,
			in:       `{"str": "select abc from def where x > 3 and y < 5"}`,
			outKey:   "str",
			expected: "select abc from def where x > ? and y < ?",
			fail:     false,
		},

		{
			name: "normal",
			pl: `json(_, str) 
			sql_cover(str)`,
			in:   `{"str": "SELECT $func$INSERT INTO table VALUES ('a', 1, 2)$func$ FROM users"}`,
			fail: true,
		},
		{
			name: "normal",
			pl: `json(_, str) 
			sql_cover(str)`,
			in:       `{"str": "SELECT Codi , Nom_CA AS Nom, Descripció_CAT AS Descripció FROM ProtValAptitud WHERE Vigent=1 ORDER BY Ordre, Codi"}`,
			outKey:   "str",
			expected: "SELECT Codi, Nom_CA, Descripció_CAT FROM ProtValAptitud WHERE Vigent = ? ORDER BY Ordre, Codi",
			fail:     false,
		},
		{
			name: "normal",
			pl:   `sql_cover("SELECT Codi , Nom_CA AS Nom, Descripció_CAT AS Descripció FROM ProtValAptitud WHERE Vigent=1 ORDER BY Ordre, Codi")`,
			fail: true,
		},
		{
			name: "normal",
			pl: `json(_, str) 
			sql_cover(str)`,
			in:       `{"str": "SELECT ('/uffd')"}`,
			outKey:   "str",
			expected: "SELECT ( ? )",
			fail:     false,
		},
		{
			name: "normal",
			pl:   `sql_cover(_)`,
			in: `select abc from def where x > 3 and y < 5
						SELECT ( ? )`,
			outKey:   "message",
			expected: `select abc from def where x > ? and y < ? SELECT ( ? )`,
			fail:     false,
		},
		{
			name: "normal",
			pl:   `sql_cover(_)`,
			in: `#test
select abc from def where x > 3 and y < 5`,
			outKey:   "message",
			expected: `select abc from def where x > ? and y < ?`,
			fail:     false,
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

			_, _, f, _, _, err := runScript(runner, "test", nil, map[string]interface{}{
				"message": tc.in,
			}, time.Now())
			if err != nil {
				t.Fatal(err)
			}

			t.Log(f)
			if v, ok := f[tc.outKey]; !ok {
				if !tc.fail {
					t.Errorf("[%d]expect error: %s", idx, err)
				}
			} else {
				tu.Equals(t, tc.expected, v)
				t.Logf("[%d] PASS", idx)
			}
		})
	}
}
