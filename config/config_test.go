// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

import (
	"os"
	"testing"

	bstoml "github.com/BurntSushi/toml"
	tu "github.com/GuanceCloud/cliutils/testutil"
	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"
)

func TestEmptyDir(t *testing.T) {
	dirname := "test123"
	if err := os.MkdirAll(dirname, 0o600); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := os.RemoveAll(dirname); err != nil {
			t.Error(err)
		}
	}()

	tu.Assert(t, emptyDir(dirname) == true, "dir %s should be empty", dirname)
}

func TestTomlSerialization(t *testing.T) {
	type X struct {
		A int    `toml:"a,___a"`
		B string `toml:"b,__b"`
	}

	cases := []struct {
		str  string
		fail bool
	}{
		{
			str: `a = 123
		b = "xyz"`,
		},
		{
			str: `___a = 123
		__b = "xyz"`,
			fail: true,
		},
	}

	for _, tc := range cases {
		var x X

		err := toml.Unmarshal([]byte(tc.str), &x)
		if tc.fail {
			tu.NotOk(t, err, "")
		} else {
			tu.Ok(t, err)
		}

		t.Logf("bstoml: %+#v", x)

		_, err = bstoml.Decode(tc.str, &x)
		if err != nil {
			t.Error(err)
		}
		t.Logf("bstoml: %+#v", x)
	}
}

func TestTomlParse2(t *testing.T) {
	cases := []struct {
		s    string
		fail bool
	}{
		{
			s: `abc = [
				"1", "2", "3",
				# some comment
			]`,
			fail: false,
		},

		{
			s:    `abc = []`,
			fail: false,
		},

		{
			s: `abc = [
				#"1", "2", "3",
				# some comment
			]`,
			fail: true,
		},
	}

	for _, tc := range cases {
		_, err := toml.Parse([]byte(tc.s))
		if tc.fail {
			tu.NotOk(t, err, "")
			t.Log(err)
			continue
		} else {
			tu.Ok(t, err)
		}
	}
}

func TestTomlParse(t *testing.T) {
	tomlParseCases := []struct {
		in string
	}{
		{
			in: `
		[[inputs.abc]]
			key1 = "1-line-string"
			key2 = '''multili
			string
			'''`,
		},

		{
			in: `
[[inputs.abc]]
	key1 = 1
	key2 = "a"
	key3 = 3.14`,
		},

		{
			in: `
[[inputs.abc]]
	key1 = 11
	key2 = "aa"
	key3 = 6.28`,
		},

		{
			in: `
[[inputs.abc]]
	key1 = 22
	key2 = "aaa"
	key3 = 6.18`,
		},
		{
			in: `
[[inputs.def]]
	key1 = 22
	key2 = "aaa"
	key3 = 6.18`,
		},
	}

	for _, tcase := range tomlParseCases {
		tbl, err := toml.Parse([]byte(tcase.in))
		if err != nil {
			t.Fatal(err)
		}

		if tbl.Fields == nil {
			t.Fatal("empty data")
		}

		for f, v := range tbl.Fields {
			switch f {
			default:
				// ignore
				t.Logf("ignore %+#v", f)

			case "inputs":
				switch tpe := v.(type) {
				case *ast.Table:
					stbl, ok := v.(*ast.Table)
					if !ok {
						t.Error("expet *ast.Table")
					}

					for _, vv := range stbl.Fields {
						switch tt := vv.(type) {
						case []*ast.Table:
							for idx, elem := range vv.([]*ast.Table) {
								t.Logf("[%d] %+#v, source: %s", idx, elem, elem.Source())
							}
						case *ast.Table:
							t.Logf("%+#v, source: %s", vv.(*ast.Table), vv.(*ast.Table).Source())
						default:
							t.Logf("bad data: %v", tt)
						}
					}

				default:
					t.Logf("unknown type: %v", tpe)
				}
			}
		}
	}
}

func TestTomlUnmarshal(t *testing.T) {
	x := []byte(`
global = "global config"
[[inputs.abc]]
	key1 = 1
	key2 = "a"
	key3 = 3.14

[[inputs.abc]]
	key1 = 11
	key2 = "aa"
	key3 = 6.28

[[inputs.def]]
	key1 = 22
	key2 = "aaa"
	key3 = 6.18

[inputs.xyz]
	key1 = 22
	key2 = "aaa"
	key3 = 6.18

	[[inputs.xyz.tags]]
		key1 = 22
		key2 = "aaa"
		key3 = 6.18
		#key4 = 7.18
	`)

	tbl, err := toml.Parse(x)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("tbl: %+#v", tbl)
	t.Log(tbl.Source())

	for f, v := range tbl.Fields {
		switch f {
		default:
			// ignore
			t.Logf("ignore %+#v", f)

		case "inputs":
			switch tpe := v.(type) {
			case *ast.Table:
				stbl, ok := v.(*ast.Table)
				if !ok {
					t.Error("expet *ast.Table")
				}

				for _, vv := range stbl.Fields {
					switch tt := vv.(type) {
					case []*ast.Table:
						for idx, elem := range vv.([]*ast.Table) {
							t.Logf("[%d] %+#v, source: %s", idx, elem, elem.Source())
						}
					case *ast.Table:
						t.Logf("%+#v, source: %s", vv.(*ast.Table), vv.(*ast.Table).Source())
					default:
						t.Logf("bad data: %v", tt)
					}
				}

			default:
				t.Logf("unknown type: %v", tpe)
			}
		}
	}
}

func TestFeedEnvs(t *testing.T) {
	cases := []struct {
		str    string
		expect string
		env    map[string]string
	}{
		{
			str: "this is env from os:  $TEST_ENV_1",

			env: map[string]string{
				"TEST_ENV_1":   "test-data",
				"TEST_ENV___1": "test-data2",
			},
			expect: "this is env from os:  test-data",
		},

		{
			str: "this is env from os:  $$TEST_ENV_1$$",
			env: map[string]string{
				"TEST_ENV_1":   "test-data",
				"TEST_ENV___1": "test-data2",
			},
			expect: "this is env from os:  $test-data$$",
		},

		{
			str: "this is env from os:  $TEST_ENV_2",
			env: map[string]string{
				"TEST_ENV_2": "test-data2",
			},
			expect: "this is env from os:  test-data2",
		},
	}

	for idx, tc := range cases {
		for k, v := range tc.env {
			if err := os.Setenv(k, v); err != nil {
				t.Fatal(err)
			}
		}

		data := feedEnvs([]byte(tc.str))
		tu.Assert(t, tc.expect == string(data), "[%d] epxect `%s', got `%s'", idx, tc.expect, string(data))
	}
}
