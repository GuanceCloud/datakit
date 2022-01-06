package config

import (
	"fmt"
	"os"
	"testing"

	bstoml "github.com/BurntSushi/toml"
	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"
	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
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

func TestBlackWhiteList(t *testing.T) {
	wlists := []*inputHostList{
		{
			Hosts:  []string{"host1", "host2"},
			Inputs: []string{"input1", "input2"},
		},
		{
			Hosts:  []string{"hostx", "hosty"},
			Inputs: []string{"inputx", "inputy"},
		},
	}

	blists := []*inputHostList{
		{
			Hosts:  []string{"host_3", "host_4"},
			Inputs: []string{"input_3", "input_4"},
		},
		{
			Hosts:  []string{"host_i", "host_j"},
			Inputs: []string{"input_i", "input_j"},
		},
	}

	t.Logf("host1.inputx on? %v", !isDisabled(wlists, blists, "host1", "inputx"))
	t.Logf("host2.inputy on? %v", !isDisabled(wlists, blists, "host2", "inputy"))
	t.Logf("host2.input1 on? %v", !isDisabled(wlists, blists, "host2", "input1"))
	t.Logf("host2.input_foo on? %v", !isDisabled(wlists, blists, "host2", "input_foo"))
	t.Logf("host_bar.input_foo on? %v", !isDisabled(wlists, blists, "host_bar", "input_foo"))

	t.Logf("host_3.input_foo on? %v", !isDisabled(wlists, blists, "host_3", "input_foo"))
	t.Logf("host_3.input_4 on? %v", !isDisabled(wlists, blists, "host_3", "input_4"))
	t.Logf("host_3.input_j on? %v", !isDisabled(wlists, blists, "host_3", "input_j"))
}

func TestLoadCfg(t *testing.T) {
	availableInputCfgs := map[string]*ast.Table{}
	conf := map[string]string{
		"1.conf": `[[inputs.aliyunobject]]
					 ## @param - aliyun authorization informations - string - required
					 region_id = ''
					 # access_key_id = ''
					 access_key_secret = ''
					 a = ""
					 ## @param - collection interval - string - optional - default: 5m
					 interval = '5m'

					[[inputs.aliyunobject]]
					 ## @param - aliyun authorization informations - string - required
					 region_id = ''
					 # access_key_id = ''
					 access_key_secret = ''
					 ## @param - collection interval - string - optional - default: 5m
					 interval = '5m'`,
		"2.conf": `[[inputs.host_processes]]`,
	}

	for k, v := range conf {
		as, _ := toml.Parse([]byte(v))
		availableInputCfgs[k] = as
	}

	for name, creator := range inputs.Inputs {
		doLoadInputConf(name, creator, availableInputCfgs)
	}
	fmt.Println(inputs.InputsInfo)
}

func TestRemoveDepercatedInputs(t *testing.T) {
	cases := []struct {
		tomlStr      string
		res, entries map[string]string
	}{
		{
			tomlStr: `[[intputs.abc]]`,
			entries: map[string]string{"abc": "cba"},
			res:     map[string]string{"abc": "cba"},
		},
		{
			tomlStr: `[intputs.abc]`,
			entries: map[string]string{"abc": "cba"},
			res:     map[string]string{"abc": "cba"},
		},
		{
			tomlStr: `[intputs.def]`,
			entries: map[string]string{"abc": "cba"},
			res:     nil,
		},
		{
			tomlStr: `[intputs.abc.xyz]`,
			entries: map[string]string{"abc": "cba"},
			res:     map[string]string{"abc": "cba"},
		},
	}

	for _, tc := range cases {
		tbl, err := toml.Parse([]byte(tc.tomlStr))
		if err != nil {
			t.Fatal(err)
		}

		res := checkDepercatedInputs(tbl, tc.entries)
		t.Logf("res: %+#v", res)
		tu.Assert(t,
			len(res) == len(tc.res),
			"got %+#v", res)
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
			str:    "this is env from os:  $$TEST_ENV_2",
			env:    map[string]string{},
			expect: "this is env from os:  $no-value",
		},

		{
			str:    "this is env from os:  $TEST_ENV_2",
			env:    map[string]string{},
			expect: "this is env from os:  no-value",
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
