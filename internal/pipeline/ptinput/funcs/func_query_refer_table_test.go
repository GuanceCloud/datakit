// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/ptinput"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/refertable"
)

func TestQueryReferTable(t *testing.T) {
	files := map[string]string{
		"a.json": testTableData,
	}
	server := newJsonDataServer(files)
	defer server.Close()
	url := server.URL

	refertable.InitReferTableRunner(url+"?name=a.json", time.Second*5, false, false)

	ok := refertable.InitFinished(time.Second * 5)
	if !ok {
		t.Fatal("init refer table timeout")
	}

	testCase := []struct {
		name string
		in   string

		script string

		key      []string
		expected []any
		fail     bool
	}{
		{
			name: "test query",
			in:   `test query"`,
			script: `add_key(f1, 123)
			query_refer_table("table1", key = "f1", value= f1)`,
			key:      []string{"key1", "key2", "f1", "f2"},
			expected: []any{"a", float64(123), int64(123), false},
		},
		{
			name: "test query multi",
			in:   `test query multi"`,
			script: `add_key(f1, 123)
			mquery_refer_table("table1", ["f1"], values= [f1])`,
			key:      []string{"key1", "key2", "f1", "f2"},
			expected: []any{"a", float64(123), int64(123), false},
		},
		{
			name: "test int",
			in:   `test int `,
			script: `# add_key(f1, 123)
			query_refer_table("table1", "f1", 123)`,
			key:      []string{"key1", "key2", "f1", "f2"},
			expected: []any{"a", float64(123), int64(123), false},
		},
		{
			name: "test string",
			in:   `test string`,
			script: `
			query_refer_table("table1", "key1", "a")`,
			key:      []string{"key1", "key2", "f1", "f2"},
			expected: []any{"a", float64(123), int64(123), false},
		},
		{
			name: "test string 2",
			in:   `test string`,
			script: `
			t = "table1"
			k = "key1"
			v = "a"
			query_refer_table(t, k, v)`,
			key:      []string{"key1", "key2", "f1", "f2"},
			expected: []any{"a", float64(123), int64(123), false},
		},
		{
			name: "test float",
			in:   `test float`,
			script: `
			# add_key(f1, 123)
			query_refer_table("table1", "key2", 123.)`,
			key:      []string{"key1", "key2", "f1", "f2"},
			expected: []any{"a", float64(123), int64(123), false},
		},
		{
			name: "test bool",
			in:   `test bool`,
			script: `
			# add_key(f1, 123)
			query_refer_table("table1", "f2", false)`,
			key:      []string{"key1", "key2", "f1", "f2"},
			expected: []any{"a", float64(123), int64(123), false},
		},
		{
			name: "test float but int",
			in:   `test float but int`,
			script: `
			# add_key(f1, 123)
			query_refer_table("table1", "key2", 123)`,
			key:      []string{"key1", "key2", "f1", "f2"},
			expected: []any{nil, nil, nil, nil},
		},
		{
			name: "test query, key not find",
			in:   `test query, key not find"`,
			script: `#add_key(f1, 123)
			query_refer_table("table1", "f1", f1)`,
			key:      []string{"key1", "key2", "f1", "f2"},
			expected: []any{nil, nil, nil, nil},
		},
		{
			name: "test query, positional keyword",
			in:   `test query, positional keyword`,
			script: `#add_key(f1, 123)
			query_refer_table("table1", "f1", value=f1)`,
			key:      []string{"key1", "key2", "f1", "f2"},
			expected: []any{nil, nil, nil, nil},
			fail:     true,
		},
		{
			name: "test-multi",
			in:   `test-multi"`,
			script: `
			key = "f2"
			value = "ab"
			mquery_refer_table("table1", ["key1", key], [value, false])`,
			key:      []string{"key1", "key2", "f1", "f2"},
			expected: []any{"ab", float64(1234), int64(123), false},
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			runner, err := NewTestingRunner(tc.script)
			if err != nil {
				t.Error(err)
				return
			}

			pt := ptinput.NewPlPoint(
				point.Logging, "test", nil, map[string]any{"message": tc.in}, time.Now())
			errR := runScript(runner, pt)

			if errR != nil {
				t.Fatal(errR.Error())
			}

			for idxK, key := range tc.key {
				v, _, ok := pt.GetWithIsTag(key)
				if !ok {
					if len(tc.expected) != 0 {
						t.Logf("key: %s, value exp: %v  act: nil",
							key, tc.expected[idxK])
					}
				}
				assert.Equal(t, tc.expected[idxK], v)
			}
		})
	}
}

func newJsonDataServer(files map[string]string) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
			default:
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			name := r.FormValue("name")
			data := files[name]
			w.Write([]byte(data))
			w.WriteHeader(http.StatusOK)
		},
	))
	return server
}

var testTableData = `
[
	{
		"table_name": "table1",
		"column_name": [
			 "key1",
			 "key2",
			 "f1",
			 "f2"
		],
		"column_type": [
			 "string",
			 "float",
			 "int",
			 "bool"
		],
		"row_data": [
			 [
				  "a",
				  123,
				  "123",
				  "false"
			 ],
			 [
				  "ab",
				  "1234",
				  "123",
				  "true"
			 ],
			 [
				"ab",
				"1234",
				"123",
				"false"
		   ]
		]
	},
	{
		"table_name": "table2",
		"primary_key": [
			 "key1",
			 "key2"
		],
		"column_name": [
			 "key1",
			 "key2",
			 "f1",
			 "f2"
		],
		"column_type": [
			 "string",
			 "float",
			 "int",
			 "bool"
		],
		"row_data": [
			 [
				  "a",
				  123,
				  "123",
				  "true"
			 ],
			 [
				  "a",
				  "1234",
				  "123",
				  "true"
			 ]
		]
	}
]

`
