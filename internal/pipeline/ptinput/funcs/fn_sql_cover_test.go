// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"testing"
	"time"

	tu "github.com/GuanceCloud/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/ptinput"
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
			pl: `json(_, str_msg) 
			sql_cover(str_msg)`,
			in:       `{"str_msg": "select abc from def where x > 3 and y < 5"}`,
			outKey:   "str_msg",
			expected: "select abc from def where x > ? and y < ?",
			fail:     false,
		},

		{
			name: "normal",
			pl: `json(_, str_msg) 
			sql_cover(str_msg)`,
			in:   `{"str_msg": "SELECT $func$INSERT INTO table VALUES ('a', 1, 2)$func$ FROM users"}`,
			fail: true,
		},
		{
			name: "normal",
			pl: `json(_, str_msg) 
			sql_cover(str_msg)`,
			in:       `{"str_msg": "SELECT Codi , Nom_CA AS Nom, Descripció_CAT AS Descripció FROM ProtValAptitud WHERE Vigent=1 ORDER BY Ordre, Codi"}`,
			outKey:   "str_msg",
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
			pl: `json(_, str_msg) 
			sql_cover(str_msg)`,
			in:       `{"str_msg": "SELECT ('/uffd')"}`,
			outKey:   "str_msg",
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

			pt := ptinput.GetPoint()
			ptinput.InitPt(pt, "test", nil, map[string]any{"message": tc.in}, time.Now())
			errR := runScript(runner, pt)

			if errR != nil {
				ptinput.PutPoint(pt)
				t.Fatal(errR)
			}

			if v, ok := pt.Fields[tc.outKey]; !ok {
				if !tc.fail {
					t.Errorf("[%d]expect error: %s", idx, errR.Error())
				}
			} else {
				tu.Equals(t, tc.expected, v)
				t.Logf("[%d] PASS", idx)
			}
			ptinput.PutPoint(pt)
		})
	}
}
