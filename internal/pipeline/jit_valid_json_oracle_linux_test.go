// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

// Original four cases adapted from pipeline-go v1.4.3
// ptinput/funcs/fn_valid_json_test.go (MIT, Copyright Guance, Inc.).
import (
	"fmt"
	"math"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	guancegrok "github.com/GuanceCloud/grok"
	"github.com/GuanceCloud/pipeline-go/ptinput"
	"github.com/GuanceCloud/pipeline-go/ptinput/funcs"
	"github.com/GuanceCloud/pipeline-go/ptinput/plcache"
	"github.com/GuanceCloud/platypus/pkg/engine"
	"github.com/gogo/protobuf/types"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
)

type jsonOracleCase struct {
	name, source, input, key string
	want                     any
	fails                    bool
}

func TestGoRuneErrorRegexpDiagnostic(t *testing.T) {
	for _, pattern := range []string{`(?P<result>�)`, `(?P<result>�+)`, `(?P<result>[�])`, `(?P<result>.)`} {
		re := regexp.MustCompile(pattern)
		prefix, complete := re.LiteralPrefix()
		want := "\xff"
		if pattern == `(?P<result>.)` {
			want = "A"
		}
		got := re.FindStringSubmatch("A\xffB")
		if !reflect.DeepEqual(got, []string{want, want}) || prefix != "" || complete {
			t.Fatalf("pattern=%q prefix=%q complete=%t match=%q want=%q", pattern, prefix, complete, got, want)
		}
	}
}

// This pins the actual dependency's byte prefilter contract, not just Go's
// regexp semantics. A literal U+FFFD can match an invalid byte in regexp, while
// grok requires the literal's UTF-8 bytes somewhere in the original input.
func TestGoGrokRuneErrorPrefilterOracle(t *testing.T) {
	for _, tc := range []struct {
		name, pattern, input string
		want                 []string
	}{
		{"literal_raw", `(?P<result>�)`, "A\xffB", nil},
		{"literal_present_later", `(?P<result>�)`, "A\xffB�", []string{"\xff"}},
		{"repeat_raw", `(?P<result>�+)`, "A\xffB", nil},
		{"repeat_mixed", `(?P<result>�+)`, "�\xff�", []string{"�\xff�"}},
		{"class_raw", `(?P<result>[�])`, "A\xffB", nil},
		{"escaped_raw", `(?P<result>\x{FFFD})`, "A\xffB", nil},
		{"optional_raw", `(?P<result>�?)`, "A\xffB", []string{""}},
		{"alternative_raw", `(?P<result>�|B)`, "A\xffB", []string{"\xff"}},
		{"folded_raw", `(?i)(?P<result>�)`, "A\xffB", []string{"\xff"}},
		{"fold_scope", `(?i:a)(?-i:(?P<result>�))`, "A\xffB", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			compiled, err := guancegrok.CompilePattern(tc.pattern, nil)
			if err != nil {
				t.Fatal(err)
			}
			got, err := compiled.Run(tc.input, false)
			if !reflect.DeepEqual(got, tc.want) || (tc.want == nil) != (err != nil) {
				t.Fatalf("grok got=%q err=%v want=%q", got, err, tc.want)
			}
			// Read-only reflection diagnoses the pinned upstream implementation;
			// no private fields are changed and no unsafe access is used.
			pf := reflect.ValueOf(compiled).Elem().FieldByName("prefilter")
			if pf.IsValid() && !pf.IsNil() {
				t.Logf("regexp=%q literalPrefix=%q", regexp.MustCompile(tc.pattern).FindStringSubmatch(tc.input), pf.Elem().FieldByName("literalPrefix").String())
			}
		})
	}
}

func TestJITGrokFloatCaptureOracle(t *testing.T) {
	const source = `grok(_, "%{GREEDYDATA:result:float}")`
	runJSONOracle(t, []jsonOracleCase{
		{"decimal_separator", source, "1_234.5", "result", float64(1234.5), false},
		{"hex_separator", source, "0x_1.8p2", "result", float64(6), false},
		{"invalid_separator", source, "1__2", "result", float64(0), false},
		{"hex_subnormal", source, "0x1p-1074", "result", math.SmallestNonzeroFloat64, false},
		{"hex_sticky", source, "0x1.0000000000000800000000001p0", "result", math.Nextafter(1, 2), false},
	})
}

func TestJITGrokIntCaptureOracle(t *testing.T) {
	const source = `grok(_, "%{GREEDYDATA:result:int}")`
	runJSONOracle(t, []jsonOracleCase{
		{"overflow_before_suffix", source, "18446744073709551616x", "result", int64(math.MaxInt64), false},
		{"negative_overflow_before_suffix", source, "-18446744073709551616x", "result", int64(math.MinInt64), false},
		{"suffix_before_unsigned_overflow", source, "9223372036854775808x", "result", int64(0), false},
		{"overflow_bad_separator", source, "18446744073709551616__", "result", int64(math.MaxInt64), false},
		{"invalid_prefix", source, "0_b1", "result", int64(0), false},
		{"decimal", source, "1_234", "result", int64(1234), false},
		{"hex", source, "0x_ff", "result", int64(255), false},
		{"binary", source, "0b_10_10", "result", int64(10), false},
		{"octal", source, "0_7", "result", int64(7), false},
		{"invalid", source, "1__2", "result", int64(0), false},
		{"overflow", source, "9_223_372_036_854_775_808", "result", int64(math.MaxInt64), false},
	})
}

func TestJITJSONPlainRunsOracle(t *testing.T) {
	var cases []jsonOracleCase
	for _, function := range []string{"json", "gjson"} {
		source := `json(message,value,result,false)`
		if function == "gjson" {
			source = `gjson(message,"value","result")`
		}
		for _, size := range []int{0, 1, 31, 32, 63, 64, 65, 1024, 16384} {
			plain := strings.Repeat("a", size)
			input := `{"value":"` + plain + `\n\uD83D\uDE00\uD800\t\\\"` + "中\xff尾" + `","after":true}`
			cases = append(cases, jsonOracleCase{
				name:   fmt.Sprintf("%s_%d", function, size),
				source: source,
				input:  input, key: "result", want: plain + "\n😀�\t\\\"中\xff尾",
			})
		}
	}
	cases = append(cases, jsonOracleCase{
		name: "invalid_tail", source: `json(message,value,result)`,
		input: `{"value":"` + strings.Repeat("a", 16384) + `","bad":}`, key: "result", want: nil,
	})
	runJSONOracle(t, cases)
}

func TestJITGrokTailCaptureOracle(t *testing.T) {
	cases := []jsonOracleCase{
		{"greedy_prefix", `grok(_, "(?s)^(?P<head>.*) (?P<result>.*)$", false)`, "a b c\n尾", "result", "c\n尾", false},
		{"lazy_prefix", `grok(_, "(?s)^(?P<head>.*?) (?P<result>.*)$", false)`, "a b c\n尾", "result", "b c\n尾", false},
		{"duplicate_capture", `grok(_, "(?s)^%{DATA:result} %{GREEDYDATA:result}$", false)`, "a b c", "result", "b c", false},
		{"error_prefix", `grok(_, "(?s)^prefix:%{GREEDYDATA:result}$", false); value=1/divisor`, "prefix:中\xff\n尾", "result", "中\xff\n尾", true},
		{"unmatched_prefix", `grok(_, "(?s)^prefix:%{GREEDYDATA:result}$", false)`, "other:中\xff\n尾", "result", nil, false},
	}
	for _, size := range []int{0, 1024, 4096, 8192, 16384, 65536} {
		tail := "request failed\n" + strings.Repeat("\tat example.Order.process(Order.java:42)\n", size/40) + "原因:中😀\xff\xfe\n"
		cases = append(cases, jsonOracleCase{
			name:   fmt.Sprintf("stack_%d", size),
			source: `grok(_, "(?s)^%{TIMESTAMP_ISO8601:ts} %{LOGLEVEL:level} %{GREEDYDATA:result}$", false)`,
			input:  "2026-09-06T10:15:30.123Z ERROR " + tail,
			key:    "result", want: tail,
		})
	}
	runJSONOracle(t, cases)
}

func TestJITGrokRawCaptureOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{"valid_unicode", `grok(_, "(?s)^prefix:%{GREEDYDATA:result}:suffix$", false)`, "prefix:中😀\n尾:suffix", "result", "中😀\n尾", false},
		{"valid_unicode_trim", `grok(_, "^prefix:%{GREEDYDATA:result}:suffix$")`, "prefix:\u2003中😀\u2003:suffix", "result", "中😀", false},
		{"valid_empty", `grok(_, "^%{GREEDYDATA:result}$", false)`, "", "result", "", false},
		{"valid_nul", `grok(_, "^%{GREEDYDATA:result}$", false)`, "A\x00B", "result", "A\x00B", false},
		{"local", `raw=message; grok(raw, "%{GREEDYDATA:result}")`, "A\xffB", "result", "A\xffB", false},
		{"empty_capture", `grok(_, "%{DATA:result}:%{GREEDYDATA:rest}")`, ":\xff", "result", "", false},
		{"rune_error", `grok(_, "(?P<result>�)")`, "A\xffB", "result", nil, false},
		{"rune_error_present_later", `grok(_, "(?P<result>�)")`, "A\xffB�", "result", "\xff", false},
		{"rune_error_repeat", `grok(_, "(?P<result>�+)")`, "A\xffB", "result", nil, false},
		{"rune_error_class", `grok(_, "(?P<result>[�])")`, "A\xffB", "result", nil, false},
		{"rune_error_optional", `grok(_, "(?P<result>�?)")`, "A\xffB", "result", "", false},
		{"rune_error_alternative", `grok(_, "(?P<result>�|B)")`, "A\xffB", "result", "\xff", false},
		{"rune_error_folded", `grok(_, "(?i)(?P<result>�)")`, "A\xffB", "result", "\xff", false},
		{"rune_error_fold_scope", `grok(_, "(?i:a)(?-i:(?P<result>�))")`, "A\xffB", "result", nil, false},
		{"rune_error_custom", `add_pattern("RAW_RUNE", "�+"); grok(_, "%{RAW_RUNE:result}")`, "A\xffB", "result", nil, false},
		{"real_and_invalid_rune", `grok(_, "(?P<result>�+)")`, "�\xff�", "result", "�\xff�", false},
		{"typed_raw", `grok(_, "%{GREEDYDATA:result:int}")`, "1\xff", "result", int64(0), false},
		{"whole", `grok(_, "%{GREEDYDATA:result}")`, "A\xffB", "result", "A\xffB", false},
		{"truncated_utf8", `grok(_, "%{GREEDYDATA:result}")`, "\xe2\x82", "result", "\xe2\x82", false},
		{"two_captures", `grok(_, "%{DATA:first}:%{GREEDYDATA:result}")`, "\xff:A\xfeB", "result", "A\xfeB", false},
		{"unicode_and_raw", `grok(_, "%{GREEDYDATA:result}")`, "中\xff😀", "result", "中\xff😀", false},
		{"trim", `grok(_, "%{GREEDYDATA:result}")`, " \xff ", "result", "\xff", false},
		{"no_match", `grok(_, "^literal$")`, "\xff", "result", nil, false},
	})
}

func TestJITGrokRawTagOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{"tag", `grok(raw, "%{GREEDYDATA:result}")`, "test", "result", "A\xffB", false},
		{"local_shadows_tag", `raw="local"; grok(raw, "%{GREEDYDATA:result}")`, "test", "result", "local", false},
	}, func(pt *point.Point) { pt.SetTag("raw", "A\xffB") })
}

// Generated semantic combinations supplement (not count as) upstream cases.
// The same original bytes run independently through Go and the loaded JIT.
func TestJITGrokBytePrefilterCombinations(t *testing.T) {
	patterns := []string{
		`�`, `[�]`, `\x{FFFD}`, `�+`, `�?`, `�*`, `�{2}`, `�{1,3}`,
		`�|B`, `�|BC`, `AB|�`, `A�|AB`, `�A|�B`, `[B�]`, `[\x{FFFD}-\x{FFFF}]`,
		`^�`, `�$`, `^�$`, `(?m:^�$)`, `�.*tail`, `head.*�`,
		`(?i:�)`, `(?i:a)(?-i:�)`, `(?i:a(?-i:�))`, `(?i:a)|�`,
		`�(?i:a)�`, `(?:�)?B`, `(?:�|)B`, `(?:�){0,2}B`,
	}
	inputs := []string{"A\xffB", "A\xffB�", "�\xff�", "\xff\xff", "head\xfftail", "head�tail\xff", "\xff\n�", "�AB\xff"}
	cases := make([]jsonOracleCase, 0, len(patterns)*len(inputs))
	for p, pattern := range patterns {
		for i, input := range inputs {
			cases = append(cases, jsonOracleCase{
				name:   fmt.Sprintf("pattern_%02d/input_%02d", p, i),
				source: fmt.Sprintf("grok(_, %q)", "(?P<result>"+pattern+")"),
				input:  input,
			})
		}
	}
	runJSONOracle(t, cases)
}

// Adapted from pipeline-go v1.4.3 ptinput/funcs/fn_gjson_test.go, TestGJSON,
// cases 0..6. MIT License, Copyright 2021-present Guance, Inc.
// Shared fixtures preserve the source values, scripts and expected results;
// outer indentation is normalized. Selected object/array whitespace is retained.
func TestJITGJSONUpstreamOracle(t *testing.T) {
	const friends = `[
  {"first": "Dale", "last": "Murphy", "age": 44, "nets": ["ig", "fb", "tw"]},
  {"first": "Roger", "last": "Craig", "age": 68, "nets": ["fb", "tw"]},
  {"first": "Jane", "last": "Murphy", "age": 47, "nets": ["ig", "tw"]}
]`
	const input = `{
  "name": {"first": "Tom", "last": "Anderson"},
  "age": 37,
  "children": ["Sara","Alex","Jack"],
  "fav.movie": "Deer Hunter",
  "friends": ` + friends + `
}`
	runJSONOracle(t, []jsonOracleCase{
		{"upstream/0", `gjson(_, "age")`, input, "age", float64(37), false},
		{"upstream/1", `gjson(_, "children")`, input, "children", `["Sara","Alex","Jack"]`, false},
		{"upstream/2", `gjson(_, "name")`, input, "name", `{"first": "Tom", "last": "Anderson"}`, false},
		{"upstream/3", `gjson(_, "name"); gjson(name, "first")`, input, "first", "Tom", false},
		{"upstream/4", `gjson(_, "friends"); gjson(friends, "1.first", "f_first")`, input, "f_first", "Roger", false},
		{"upstream/5", `gjson(_, "friends", "friends"); gjson(friends, "1.nets.1", "f_nets")`, input, "f_nets", "tw", false},
		{"upstream/6", `gjson(_, "0.nets.2", "f_nets")`, friends, "f_nets", "tw", false},
	})
}

func TestJITNestedRawBytesOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{"raw-read", `a=pt_kvs_get("raws",true); add_key(first,a[0])`, "test", "first", "A\xffB", false},
		{"raw-length", `a=pt_kvs_get("raws",true); add_key(size,len(a[0]))`, "test", "size", int64(3), false},
		{"json-read", `add_key(encoded,pt_kvs_get("raws"))`, "test", "encoded", `["A\ufffdB","\ufffd\ufffd"]`, false},
		{"lower-element", `a=pt_kvs_get("raws",true); v=a[0]; lowercase(v); add_key(original,a[0])`, "test", "v", "a\ufffdb", false},
		{"container-output", `a=pt_kvs_get("raws",true); add_key(encoded,a)`, "test", "encoded", `["A\ufffdB","\ufffd\ufffd"]`, false},
		{"set-json", `a=pt_kvs_get("raws",true); pt_kvs_set("encoded",a)`, "test", "encoded", `["A\ufffdB","\ufffd\ufffd"]`, false},
		{"set-raw", `a=pt_kvs_get("raws",true); pt_kvs_set("copied",a,false,true)`, "test", "copied", []string{"A\xffB", "\xe2\x82"}, false},
		{"set-map-json", `a=pt_kvs_get("raws",true); count=pt_kvs_set_map({"encoded":a},include_keys=["encoded"]); add_key(count,count)`, "test", "encoded", `["A\ufffdB","\ufffd\ufffd"]`, false},
		{"set-map-raw", `a=pt_kvs_get("raws",true); count=pt_kvs_set_map({"copied":a},include_keys=["copied"],raw=true); add_key(count,count)`, "test", "copied", []string{"A\xffB", "\xe2\x82"}, false},
	}, func(pt *point.Point) { pt.AddKVs(point.NewKV("raws", []string{"A\xffB", "\xe2\x82"})) })
}

func TestJITTagOriginOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{"origin", `set_tag(_)`, "test", "message", "test", false},
		{"quoted-origin", `set_tag("_","new")`, "test", "message", "new", false},
		{"local-origin", `_=123; set_tag(_)`, "test", "message", "123", false},
		{"missing", `set_tag(absent)`, "test", "absent", "", false},
		{"bytes", `url_decode(_); set_tag(result,_)`, "%FF", "result", "\xff", false},
		{"bytes-ptset", `url_decode(_); pt_kvs_set("result",_,true)`, "%FF", "result", "\xff", false},
	})
}

func TestJITDropRawOriginOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{{
		name:   "invalid_utf8_and_nul",
		source: `add_key(before,strlen(_)); drop_origin_data(); add_key(after,true)`,
		input:  "A\xff\x00B",
		key:    "message", want: nil,
	}})
}

func TestJITRawTagTransformOracle(t *testing.T) {
	const prefix = `url_decode(_); set_tag(raw,_); `
	runJSONOracle(t, []jsonOracleCase{
		{"copy", prefix + `add_key(copied,raw)`, "%FF", "copied", "\xff", false},
		{"lower", prefix + `lowercase(raw)`, "A%FFB", "raw", "a\ufffdb", false},
		{"upper", prefix + `uppercase(raw)`, "a%FFb", "raw", "A\ufffdB", false},
		{"trim", prefix + `trim(raw)`, "%20%FF%20", "raw", "\xff", false},
		{"rename", prefix + `rename(moved,raw)`, "%FF", "moved", "\xff", false},
		{"drop", prefix + `drop_key(raw)`, "%FF", "raw", nil, false},
		{"overwrite", prefix + `add_key(raw,"text")`, "%FF", "raw", "text", false},
		{"length", prefix + `add_key(size,len(raw))`, "%E2%82", "size", int64(2), false},
	})
}

func TestJITRawTagInputOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{"copy", `add_key(copied,raw)`, "test", "copied", "A\xffB", false},
		{"lower", `lowercase(raw)`, "test", "raw", "a\ufffdb", false},
		{"rename", `rename(moved,raw)`, "test", "moved", "A\xffB", false},
		{"drop", `drop_key(raw)`, "test", "raw", nil, false},
	}, func(pt *point.Point) { pt.SetTag("raw", "A\xffB") })
}

func TestJITRawTagJSONOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{"valid", `add_key(valid,valid_json(raw))`, "test", "valid", true, false},
		{"load", `obj=load_json(raw); add_key(result,obj["value"])`, "test", "result", "A\xffB", false},
		{"json", `json(raw,value,result)`, "test", "result", "A\xffB", false},
	}, func(pt *point.Point) { pt.SetTag("raw", "{\"value\":\"A\xffB\"}") })
}

func TestJITRawJSONValidationOracle(t *testing.T) {
	var cases []jsonOracleCase
	for _, tc := range []struct {
		name, input string
		valid       bool
	}{
		{"string", "\"A\xffB\"", true},
		{"key", "{\"\xff\":1}", true},
		{"truncated_utf8", "\"\xe2\x82\"", true},
		{"outside_string", "\xff", false},
		{"nul", "\"\x00\"", false},
		{"escape", "\"\\\xff\"", false},
		{"trailing_comma", "[\"\xff\",]", false},
		{"unclosed", "\"\xff", false},
	} {
		cases = append(cases, jsonOracleCase{tc.name, `add_key(valid,valid_json(_))`, tc.input, "valid", tc.valid, false})
	}
	runJSONOracle(t, cases)
}

func TestJITRawJSONExtractionOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{"load_scalar", `v=load_json(_); add_key(result,v)`, "\"A\xffB\"", "result", "A\xffB", false},
		{"load_nested", `v=load_json(_); add_key(result,v["a"][1]["b"])`, "{\"a\":[0,{\"b\":\"\xe2\x82\"}]}", "result", "\xe2\x82", false},
		{"load_escape", `v=load_json(_); add_key(result,v[0])`, "[\"\xff\\u0041\\n\"]", "result", "\xffA\n", false},
		{"load_trim", `v=load_json(_); add_key(result,v[0])`, "\u2003[\"\xff\"]\u00a0", "result", "\xff", false},
		{"load_duplicate", `v=load_json(_); add_key(result,v["a"])`, "{\"a\":\"first\",\"a\":\"\xff\"}", "result", "\xff", false},
		{"load_invalid", `v=load_json(_); add_key(result,v==nil)`, "[\"\xff\",]", "result", true, false},
		{"json_nested", `json(_,a[1].b,result)`, "{\"a\":[0,{\"b\":\"\xff\"}]}", "result", "\xff", false},
		{"json_trim", `json(_,a,result)`, "{\"a\":\" \u2003\xff\u00a0 \"}", "result", "\xff", false},
		{"json_no_trim", `json(_,a,result,false)`, "{\"a\":\" \xff \"}", "result", " \xff ", false},
		{"json_other_raw_key", `json(_,a,result)`, "{\"\xff\":1,\"a\":\"\xff\"}", "result", "\xff", false},
		{"json_duplicate", `json(_,a,result)`, "{\"a\":\"\xff\",\"a\":\"second\"}", "result", "second", false},
		{"json_invalid", `json(_,a,result)`, "{\"a\":\"\xff\",}", "result", nil, false},
	})
}

func TestJITRawJSONAllOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{"utf8_overflow", `json_all(_,include_keys=["a"])`, `{"a":1e400}`, "a", math.Inf(1), false},
		{"skip_overflow_container", `json_all(_,include_keys=["a"])`, `{"skip":[1e400],"a":"keep"}`, "a", "keep", false},
		{"utf8_duplicate_null", `json_all(_,include_keys=["a"])`, `{"a":"keep","a":null}`, "a", "keep", false},
		{"utf8_duplicate_object", `json_all(_,include_keys=["a"])`, `{"a":"keep","a":{"b":1}}`, "a", "keep", false},
		{"utf8_empty_key", `json_all(_,include_keys=[""])`, `{"":"ignored"}`, "", nil, false},
		{"include", `json_all(_,include_keys=["a"])`, "{\"a\":\"\xff\",\"other\":2}", "a", "\xff", false},
		{"wildcard", `json_all(_,key_patterns=["a*"])`, "{\"abc\":\"\xe2\x82\",\"other\":2}", "abc", "\xe2\x82", false},
		{"array", `json_all(_,include_keys=["[1]"])`, "[false,\"\xff\"]", "[1]", "\xff", false},
		{"no_recursive_flatten", `json_all(_,include_keys=["a","b"])`, "{\"a\":{\"x\":\"\xff\"},\"b\":\"\xfe\"}", "b", "\xfe", false},
		{"duplicate_null", `json_all(_,include_keys=["a"])`, "{\"a\":\"\xff\",\"a\":null}", "a", "\xff", false},
		{"duplicate_value", `json_all(_,include_keys=["a"])`, "{\"a\":\"\xff\",\"a\":\"\xfe\"}", "a", "\xfe", false},
		{"empty_key", `json_all(_,include_keys=[""])`, "{\"\":\"\xff\"}", "", nil, false},
		{"no_filter", `json_all(_)`, "{\"a\":\"\xff\"}", "a", nil, false},
		{"invalid", `json_all(_,include_keys=["a"])`, "{\"a\":\"\xff\",}", "a", nil, false},
		{"local", `raw=message; json_all(raw,include_keys=["a"])`, "{\"a\":\"\xff\"}", "a", "\xff", false},
	})
}

func TestJITRawJSONDeleteOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{"utf8_origin", `json(_,a,result,true,true)`, "{\"a\":\"text\"}", "message", "{}", false},
		{"extract", `json(_,a,result,true,true)`, "{\"a\":\" \xff \"}", "result", "\xff", false},
		{"remaining_raw", `json(_,a,result,true,true)`, "{\"a\":1,\"b\":\"\xff\"}", "", nil, false},
		{"nested", `json(_,a.b,result,true,true)`, "{\"a\":{\"b\":\"\xff\",\"c\":2}}", "result", "\xff", false},
		{"array_member", `json(_,a[0].b,result,true,true)`, "{\"a\":[{\"b\":\"\xff\"}]}", "result", "\xff", false},
		{"missing", `json(_,missing,result,true,true)`, "{\"a\":\"\xff\"}", "result", nil, false},
		{"null", `json(_,a,result,true,true)`, "{\"a\":null,\"b\":\"\xff\"}", "result", nil, false},
		{"duplicate", `json(_,a,result,true,true)`, "{\"a\":\"old\",\"a\":\"\xff\"}", "result", "\xff", false},
		{"escaped_remaining", `json(_,a,result,true,true)`, "{\"a\":1,\"b\":\"<\xff>&\"}", "", nil, false},
	})
}

func TestJITGJSONByteAndCompositeOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{"skip_overflow", `gjson(_,"a","result")`, `{"huge":1e400,"a":"ok"}`, "result", "ok", false},
		{"skip_container_overflow", `gjson(_,"a","result")`, `{"huge":[{"n":1e400}],"a":"ok"}`, "result", "ok", false},
		{"composite_overflow", `gjson(_,"a","result")`, `{"a": [ 1e400, {"n":-1e400} ]}`, "result", `[ 1e400, {"n":-1e400} ]`, false},
		{"skip_quoted_structure", `gjson(_,"a","result")`, `{"skip":["}\\\"[",{"x":[1,2]}],"a":true}`, "result", true, false},
		{"raw_scalar", `gjson(_,"a","result")`, "{\"a\":\"\xff\"}", "result", "\xff", false},
		{"raw_array", `gjson(_,"a","result")`, "{\"a\": [ \"\xff\", 1 ]}", "result", "[ \"\xff\", 1 ]", false},
		{"object_format", `gjson(_,"a","result")`, `{"a": { "z": 2, "b" : 1 }}`, "result", `{ "z": 2, "b" : 1 }`, false},
		{"chain", `gjson(_,"a","nested"); gjson(nested,"b.1","result")`, "{\"a\":{\"b\":[0,\"\xff\"]}}", "result", "\xff", false},
		{"duplicate", `gjson(_,"a","result")`, "{\"a\":\"\xff\",\"a\":\"last\"}", "result", "\xff", false},
		{"escaped_dot", `gjson(_,"a\\.b","result")`, "{\"a.b\":\"\xff\"}", "result", "\xff", false},
		{"missing", `gjson(_,"missing","result")`, "{\"a\":\"\xff\"}", "result", nil, false},
		{"null", `gjson(_,"a","result")`, "{\"a\":null,\"b\":\"\xff\"}", "result", nil, false},
		{"numeric_key", `gjson(_,"0","result")`, "{\"0\":\"\xff\"}", "result", "\xff", false},
		{"local_path", `p="a"; gjson(_,p,"result")`, "{\"a\":\"\xff\"}", "result", "\xff", false},
	})
}

func TestJITGJSONNumberOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{"hex_huge_zero", `gjson(_,"a","result")`, `{"a":0x0p999999999999999999999}`, "result", float64(0), false},
		{"hex_huge_exponent", `gjson(_,"a","result")`, `{"a":0x1p999999999999999999999}`, "result", math.Inf(1), false},
		{"hex_min_subnormal", `gjson(_,"a","result")`, `{"a":0x1p-1074}`, "result", math.SmallestNonzeroFloat64, false},
		{"hex_half_subnormal", `gjson(_,"a","result")`, `{"a":0x1p-1075}`, "result", float64(0), false},
		{"hex_above_half_subnormal", `gjson(_,"a","result")`, `{"a":0x1.00000000000001p-1075}`, "result", math.SmallestNonzeroFloat64, false},
		{"hex_tie_even", `gjson(_,"a","result")`, `{"a":0x1.00000000000008p0}`, "result", float64(1), false},
		{"hex_sticky_round", `gjson(_,"a","result")`, `{"a":0x1.0000000000000800000000001p0}`, "result", math.Nextafter(1, 2), false},
		{"hex_float", `gjson(_,"a","result")`, `{"a":0x1.8p2}`, "result", float64(6), false},
		{"decimal_separator", `gjson(_,"a","result")`, `{"a":1_234.5_0}`, "result", float64(1234.5), false},
		{"hex_separator", `gjson(_,"a","result")`, `{"a":0x_1.8p2}`, "result", float64(6), false},
		{"invalid_separator", `gjson(_,"a","result")`, `{"a":1__2}`, "result", float64(0), false},
		{"exponent_separator", `gjson(_,"a","result")`, `{"a":1e1_0}`, "result", float64(1e10), false},
		{"invalid_exponent_separator", `gjson(_,"a","result")`, `{"a":1_e2}`, "result", float64(0), false},
		{"positive_sign", `gjson(_,"a","result")`, `{"a":+1}`, "result", float64(1), false},
		{"leading_zero", `gjson(_,"a","result")`, `{"a":01}`, "result", float64(1), false},
		{"invalid_exponent", `gjson(_,"a","result")`, `{"a":1e}`, "result", float64(0), false},
		{"inf", `gjson(_,"a","result")`, `{"a":Inf}`, "result", math.Inf(1), false},
		{"negative_inf", `gjson(_,"a","result")`, `{"a":-Inf}`, "result", math.Inf(-1), false},
		{"infinity", `gjson(_,"a","result")`, `{"a":Infinity}`, "result", math.Inf(1), false},
		{"positive_overflow", `gjson(_,"a","result")`, `{"a":1e400}`, "result", math.Inf(1), false},
		{"negative_overflow", `gjson(_,"a","result")`, `{"a":-1e400}`, "result", math.Inf(-1), false},
		{"underflow", `gjson(_,"a","result")`, `{"a":1e-400}`, "result", float64(0), false},
		{"large_integer", `gjson(_,"a","result")`, `{"a":9007199254740993}`, "result", float64(9007199254740992), false},
	})
}

func TestJITGJSONCountOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{"escaped_hash", `gjson(_,"a.\\#","result")`, `{"a":{"#":"literal"}}`, "result", "literal", false},
		{"escaped_hash_array", `gjson(_,"a.\\#","result")`, `{"a":[1,2]}`, "result", nil, false},
		{"root", `gjson(_,"#","result")`, `[1,null,true]`, "result", float64(3), false},
		{"nested", `gjson(_,"a.#","result")`, `{"a":[[1,2],{},false]}`, "result", float64(3), false},
		{"empty", `gjson(_,"a.#","result")`, `{"a":[]}`, "result", float64(0), false},
		{"raw", `gjson(_,"a.#","result")`, "{\"a\":[\"\xff\",1e400]}", "result", float64(2), false},
		{"dynamic", `p="a.#"; gjson(_,p,"result")`, `{"a":[1,2]}`, "result", float64(2), false},
		{"object_key", `gjson(_,"a.#","result")`, `{"a":{"#":"literal"}}`, "result", "literal", false},
		{"scalar", `gjson(_,"a.#","result")`, `{"a":1}`, "result", nil, false},
	})
}

func TestJITGJSONLooseInputOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{"leading_garbage", `gjson(_,"a","result")`, `garbage {"a":1}`, "result", float64(1), false},
		{"missing_colon", `gjson(_,"a","result")`, `{"a" 1}`, "result", float64(1), false},
		{"missing_comma", `gjson(_,"b","result")`, `{"a":1 "b":2}`, "result", float64(2), false},
		{"extra_colon", `gjson(_,"a","result")`, `{"a":::1}`, "result", float64(1), false},
		{"unclosed_object", `gjson(_,"a","result")`, `{"a":1`, "result", float64(1), false},
		{"trailing_text", `gjson(_,"a","result")`, `{"a":1}garbage`, "result", float64(1), false},
		{"trailing_comma", `gjson(_,"a","result")`, `{"a":1,}`, "result", float64(1), false},
		{"broken_later_value", `gjson(_,"a","result")`, `{"a":1,"b":`, "result", float64(1), false},
		{"unclosed_array", `gjson(_,"0","result")`, `["ok",`, "result", "ok", false},
		{"missing_value", `gjson(_,"a","result")`, `{"a":`, "result", nil, false},
		{"unclosed_string", `gjson(_,"a","result")`, `{"a":"unfinished`, "result", nil, false},
	})
}

func TestJITMultiLayerRawBytesOracle(t *testing.T) {
	var cases []jsonOracleCase
	for _, expr := range []string{`[a]`, `{"nested":a}`, `{"scalar":a[0]}`, `[a[0],1]`, `[{"leaf":a[1]}]`, `[a[0],"valid"]`} {
		for _, raw := range []string{"false", "true"} {
			for _, setter := range []string{`pt_kvs_set("result",v,false,RAW)`, `n=pt_kvs_set_map({"result":v},include_keys=["result"],raw=RAW); add_key(count,n)`} {
				source := `a=pt_kvs_get("raws",true); v=` + expr + `; ` + strings.ReplaceAll(setter, "RAW", raw) + `; a[0]="changed"; add_key(after_value,pt_kvs_get("result"))`
				cases = append(cases, jsonOracleCase{fmt.Sprintf("%s/%s/%d", expr, raw, len(cases)), source, "test", "", nil, false})
			}
		}
	}
	runJSONOracle(t, cases, func(pt *point.Point) { pt.AddKVs(point.NewKV("raws", []string{"A\xffB", "\xe2\x82"})) })
}

func TestJITOriginRawBytesOracle(t *testing.T) {
	cases := []jsonOracleCase{
		{"lower", `lowercase(_)`, "A\xffB", "message", "a\ufffdb", false},
		{"upper", `uppercase(_)`, "a\xffb", "message", "A\ufffdB", false},
		{"trim", `trim(_)`, " \xffA ", "message", "\xffA", false},
		{"url", `url_decode(_)`, "%FF%00+A", "message", "\xff\x00 A", false},
		{"url-error", `url_decode(_)`, "%FF%Q0", "message", "%FF%Q0", true},
		{"decode-lower", `url_decode(_); lowercase(_)`, "%FFAbC", "message", "\ufffdabc", false},
		{"decode-trim", `url_decode(_); trim(_)`, "%20%FF%20", "message", "\xff", false},
		{"decode-trim-cutset", `url_decode(_); trim(_,"�")`, "%FFa%FF", "message", "a", false},
		{"decode-trim-all", `url_decode(_); trim(_,"�")`, "%FF%FF", "message", "", false},
		{"decode-trim-truncated", `url_decode(_); trim(_)`, "%20%E2%82%20", "message", "\xe2\x82", false},
		{"rename", `rename(saved,_)`, "\xff\x00A", "saved", "\xff\x00A", false},
	}
	runJSONOracle(t, cases)
}

// Origin alias combinations compared against the pinned Go engine.
func TestJITOriginStringMutationOracle(t *testing.T) {
	var cases []jsonOracleCase
	for _, tc := range []struct{ name, call, input, want string }{
		{"lower", `lowercase(KEY)`, "AbC", "abc"},
		{"upper", `uppercase(KEY)`, "AbC", "ABC"},
		{"trim", `trim(KEY)`, "  AbC  ", "AbC"},
		{"replace", `replace(KEY,"b","x")`, "AbC", "AxC"},
		{"url", `url_decode(KEY)`, "A%20B", "A B"},
	} {
		for _, key := range []string{"_", `"_"`} {
			call := strings.ReplaceAll(tc.call, "KEY", key)
			cases = append(cases, jsonOracleCase{tc.name + "/" + key, call, tc.input, "message", tc.want, false})
			cases = append(cases, jsonOracleCase{tc.name + "/shadow/" + key, fmt.Sprintf("_=%q; %s; add_key(local,_)", tc.input, call), "original", "message", tc.want, false})
		}
	}
	runJSONOracle(t, cases)
}

func TestJITOriginRenameNullIfOracle(t *testing.T) {
	cases := []jsonOracleCase{
		{"rename-source", `rename(saved,_)`, "test", "saved", "test", false},
		{"rename-target", `add_key(a,"new"); rename(_,a)`, "test", "message", "new", false},
		{"rename-quoted-target", `add_key(a,"new"); rename("_",a)`, "test", "message", "new", false},
		{"rename-same", `rename(_,message)`, "test", "message", "test", false},
		{"rename-shadow", `_=123; rename(saved,_); add_key(local,_)`, "test", "saved", "test", false},
		{"nullif-origin", `nullif(_,"test")`, "test", "message", nil, false},
		{"nullif-quoted", `nullif("_","test")`, "test", "message", nil, false},
		{"nullif-shadow", `_=123; nullif(_,123); add_key(local,_)`, "test", "local", int64(123), false},
		{"nullif-shadow-unequal", `_=123; nullif(_,"test")`, "test", "message", "test", false},
	}
	runJSONOracle(t, cases)
}

// Upstream fn_dropkey_test.go TestDropKey (MIT, Copyright Guance, Inc.).
func TestJITDropKeyOracle(t *testing.T) {
	cases := []jsonOracleCase{
		{"upstream/0", `grok(_, "%{IPORHOST:client_ip} %{NOTSPACE} %{NOTSPACE} \\[%{HTTPDATE:time}\\] \"%{DATA} %{GREEDYDATA} HTTP/%{NUMBER}\" %{INT:status_code} %{INT:bytes}"); drop_key(client_ip)`, `162.62.81.1 - - [29/Nov/2021:07:30:50 +0000] "POST /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413 "-" "Mozilla/4.0"`, "client_ip", nil, false},
		{"missing", `drop_key(absent); add_key(done,true)`, "test", "done", true, false},
		{"repeat", `add_key(a,1); drop_key(a); drop_key(a)`, "test", "a", nil, false},
		{"local-shadow", `add_key(a,1); a=2; drop_key(a); add_key(local,a)`, "test", "local", int64(2), false},
		{"local-only", `a=2; drop_key(a); add_key(local,a)`, "test", "local", int64(2), false},
		{"tag", `add_key(a,"tag"); set_tag(a); drop_key(a)`, "test", "a", nil, false},
		{"quoted", `add_key("a b",1); drop_key("a b")`, "test", "a b", nil, false},
		{"message", `drop_key(_)`, "test", "message", nil, false},
		{"quoted-alias", `drop_key("_")`, "test", "message", nil, false},
		{"alias-local-shadow", `_=123; drop_key(_); add_key(local,_)`, "test", "", nil, false},
		{"alias-rewrite", `drop_key(_); add_key(_,"new")`, "test", "message", "new", false},
		{"alias-direct-write", `add_key(_,"new")`, "test", "message", "new", false},
		{"alias-quoted-write", `add_key("_","new")`, "test", "message", "new", false},
		{"alias-local-value", `_=123; add_key(_)`, "test", "message", int64(123), false},
		{"rewrite", `add_key(a,1); drop_key(a); add_key(a,"new")`, "test", "a", "new", false},
	}
	runJSONOracle(t, cases)
}

func TestJITDropKeyAdmissionOracle(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 16)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	for _, source := range []string{`drop_key()`, `drop_key(a,b)`, `drop_key(1)`, `drop_key(nil)`, `drop_key(true)`, `drop_key([])`, `drop_key({})`, `drop_key(key=a)`, `drop_key(a[0])`, `drop_key(load_json(message))`} {
		t.Run(source, func(t *testing.T) {
			if _, err := NewPlScriptSimple(point.Logging, "drop-key.p", source); err == nil {
				t.Fatal("Go accepted invalid fixture")
			}
			if check := runner.Check(source); check.Route == pljit.RouteJITNative || check.Detail == "" {
				t.Fatalf("invalid admission: %+v", check)
			}
		})
	}
}

// Invalid forms verified against Go's RenameChecking.
func TestJITRenameAdmissionOracle(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 16)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	for _, source := range []string{`rename()`, `rename(a)`, `rename(a,b,c)`, `rename(1,b)`, `rename(a,"b")`, `rename(a,nil)`, `rename(a,b[0])`, `rename(new_key=a,b)`, `rename(a,old_key=b)`} {
		t.Run(source, func(t *testing.T) {
			if _, err := NewPlScriptSimple(point.Logging, "rename.p", source); err == nil {
				t.Fatal("Go accepted invalid fixture")
			}
			if check := runner.Check(source); check.Route == pljit.RouteJITNative || check.Detail == "" {
				t.Fatalf("invalid admission: %+v", check)
			}
		})
	}
}

// Four originals from pipeline-go v1.4.3 fn_rename_test.go (MIT, Guance).
func TestJITRenameUpstreamOracle(t *testing.T) {
	prefix := `add_pattern("_second", "(?:(?:[0-5]?[0-9]|60)(?:[:.,][0-9]+)?)")
add_pattern("_minute", "(?:[0-5][0-9])")
add_pattern("_hour", "(?:2[0123]|[01]?[0-9])")
add_pattern("time", "([^0-9]?)%{_hour:hour}:%{_minute:minute}(?::%{_second:second})([^0-9]?)")
grok(_, "%{time}"); `
	var cases []jsonOracleCase
	for i, tc := range []struct{ call, key, want string }{
		{`rename(newhour,hour)`, "newhour", "12"},
		{`rename(newsecond,second)`, "newsecond", "15"},
		{`rename(newminute,minute)`, "newminute", "34"},
		{`rename(minute,newminute)`, "minute", "34"},
	} {
		cases = append(cases, jsonOracleCase{fmt.Sprintf("upstream/%d", i), prefix + tc.call, "12:34:15", tc.key, tc.want, false})
	}
	cases = append(cases,
		jsonOracleCase{"same", `add_key(a,1); rename(a,a)`, "test", "a", int64(1), false},
		jsonOracleCase{"overwrite", `add_key(a,1); add_key(b,2); rename(b,a)`, "test", "b", int64(1), false},
		jsonOracleCase{"local-shadow", `add_key(a,1); a=2; rename(b,a); add_key(local,a)`, "test", "b", int64(1), false},
		jsonOracleCase{"tag", `add_key(a,"value"); set_tag(a); rename(b,a)`, "test", "b", "value", false},
		jsonOracleCase{"field-over-tag", `add_key(a,1); add_key(b,"tag"); set_tag(b); rename(b,a)`, "test", "", nil, false},
		jsonOracleCase{"tag-over-field", `add_key(a,"tag"); set_tag(a); add_key(b,1); rename(b,a)`, "test", "", nil, false},
		jsonOracleCase{"quoted-target", `add_key(a,1); rename("new name",a)`, "test", "new name", int64(1), false})
	runJSONOracle(t, cases)
}

// Additional combinations checked against the upstream execution semantics.
func TestJITNullIfCombinationsOracle(t *testing.T) {
	cases := []jsonOracleCase{
		{"negative", `add_key(n,-1); nullif(n,-1)`, "test", "", nil, false},
		{"negative-float", `add_key(n,-1.5); nullif(n,-1.5)`, "test", "n", nil, false},
		{"positive-sign", `add_key(n,1); nullif(n,+1)`, "test", "", nil, false},
		{"parenthesized-negative", `add_key(n,-1); nullif(n,(-1))`, "test", "", nil, false},
		{"parenthesized", `add_key(n,1); nullif(n,(1))`, "test", "", nil, false},
		{"missing", `nullif(missing,1); add_key(done,true)`, "test", "done", true, false},
	}
	for _, value := range []string{"1", "1.0", "true", "nil", `"text"`} {
		cases = append(cases, jsonOracleCase{"cached/" + value, `add_key(n,"original"); cache_set("nullif",` + value + `); n=cache_get("nullif"); nullif(n,` + value + `)`, "test", "n", nil, false})
	}
	runJSONOracle(t, cases)
}

// Original six cases from pipeline-go v1.4.3 fn_nullif_test.go (MIT, Guance).
func TestJITNullIfOracle(t *testing.T) {
	var cases []jsonOracleCase
	for i, tc := range []struct {
		value, operand string
		want           any
	}{
		{"1", `"1"`, float64(1)}, {`"1"`, "1", "1"}, {`""`, `""`, nil},
		{"null", "nil", nil}, {"true", "true", nil}, {"2.3", "2.3", nil},
	} {
		cases = append(cases, jsonOracleCase{fmt.Sprintf("upstream/%d", i), `json(_, a.first, a_first); nullif(a_first,` + tc.operand + `)`, `{"a":{"first":` + tc.value + `,"second":2,"third":"aBC","forth":true},"age":47}`, "a_first", tc.want, false})
	}
	cases = append(cases,
		jsonOracleCase{"int-float", `add_key(n,1); nullif(n,1.0)`, "test", "n", int64(1), false},
		jsonOracleCase{"bool-int", `add_key(n,true); nullif(n,1)`, "test", "n", true, false},
		jsonOracleCase{"local-shadow", `add_key(n,1); n=2; nullif(n,2); add_key(local,n)`, "test", "local", int64(2), false},
		jsonOracleCase{"dynamic-ignored", `add_key(n,1); other=1; nullif(n,other)`, "test", "n", int64(1), false},
		jsonOracleCase{"expression-not-evaluated", `add_key(n,1); divisor=0; nullif(n,1/divisor)`, "test", "n", int64(1), false})
	runJSONOracle(t, cases)
}

// Adapted from pipeline-go v1.4.3 fn_parse_duration_test.go (MIT, Guance).
// The final upstream fail flags do not assert execution failure; the actual
// implementation preserves malformed strings and non-string input unchanged.
func TestJITParseDurationUpstreamOracle(t *testing.T) {
	var cases []jsonOracleCase
	for i, tc := range []struct {
		input string
		want  any
	}{
		{"1s", int64(1_000_000_000)}, {"1ms", int64(1_000_000)},
		{"1us", int64(1000)}, {"1µs", int64(1000)}, {"1m", int64(60_000_000_000)},
		{"1h", int64(3_600_000_000_000)}, {"-23h", int64(-82_800_000_000_000)},
		{"-23ns", int64(-23)}, {"-2.3s", int64(-2_300_000_000)}, {"1uuus", "1uuus"},
	} {
		cases = append(cases, jsonOracleCase{fmt.Sprintf("upstream/%d", i), "json(_, `str`); parse_duration(`str`)", fmt.Sprintf(`{"str":%q}`, tc.input), "str", tc.want, false})
	}
	cases = append(cases, jsonOracleCase{"upstream/11-nonstring", "json(_, `str`); parse_duration(`str`)", `{"str":1}`, "str", float64(1), false})
	for _, input := range []string{"0", "+0", ".5s", "1.s", "1h2m3.4s", "1μs", "0.1ns", "9223372036854775807ns", "-9223372036854775808ns", "9223372036854775808ns", "1d", " 1s", "1s ", "1e3s", ""} {
		// Retain the original input on errors, exactly as the Go builtin does.
		want := any(input)
		if value, err := time.ParseDuration(input); err == nil {
			want = int64(value)
		}
		cases = append(cases, jsonOracleCase{"boundary/" + input, "parse_duration(message)", input, "message", want, false})
	}
	runJSONOracle(t, cases)
}

func TestJITParseDurationContextOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{"missing", `parse_duration(absent)`, "1s", "message", "1s", false},
		{"local-shadow", `message="2s"; parse_duration(message); add_key(local,message)`, "1s", "local", "2s", false},
		{"local-only", `duration="2s"; parse_duration(duration); add_key(local,duration)`, "1s", "local", "2s", false},
		{"quoted-key", `parse_duration("message")`, "1s", "message", int64(1_000_000_000), false},
		{"tag", `pt_kvs_set("duration","2s",as_tag=true); parse_duration(duration)`, "1s", "duration", "2000000000", false},
		{"invalid-tag", `pt_kvs_set("duration","bad",as_tag=true); parse_duration(duration)`, "1s", "duration", "bad", false},
	})
}

func TestJITParseDurationAdmissionOracle(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 16)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	for _, source := range []string{`parse_duration()`, `parse_duration(message,"ns")`, `parse_duration(123)`, `parse_duration(nil)`, `parse_duration(true)`, `parse_duration([])`, `parse_duration({})`, `parse_duration(key=message)`, `parse_duration(message[0])`, `parse_duration(load_json(message))`} {
		t.Run(source, func(t *testing.T) {
			if _, err := NewPlScriptSimple(point.Logging, "duration.p", source); err == nil {
				t.Fatal("Go accepted invalid fixture")
			}
			if check := runner.Check(source); check.Route == pljit.RouteJITNative || check.Detail == "" {
				t.Fatalf("invalid admission: %+v", check)
			}
		})
	}
}

func TestJITValidJSONUpstreamOracle(t *testing.T) {
	const valid = `{"a":{"first": [2.2, 1.1], "ff": "[2.2, 1.1]","second":2,"third":"aBC","forth":true},"age":47}`
	const invalid = `{"a"??:{"first": [2.2, 1.1], "ff": "[2.2, 1.1]","second":2,"third":"aBC","forth":true},"age":47}`
	cases := []jsonOracleCase{
		{"upstream/0-map", `if valid_json(_) { d = load_json(_); add_key("abc", d["a"]["first"][0]) }`, valid, "abc", 2.2, false},
		{"upstream/1-invalid-map", `if valid_json(_) {} else { d = load_json(_); add_key("abc", d["a"]["first"][0]) }`, invalid, "", nil, true},
		{"upstream/2-empty", "add_key(`in`, valid_json(_))", "", "in", false, false},
		{"upstream/3-legacy-alias", `if vaild_json(_) { d = load_json(_); add_key("abc", d["a"]["first"][0]) }`, valid, "abc", 2.2, false},
		{"named/unknown", `add_key(valid, valid_json(unknown=_))`, valid, "valid", true, false},
		{"named/legacy-unknown", `add_key(valid, vaild_json(unknown=_))`, valid, "valid", true, false},
	}
	for _, alias := range []string{"valid_json", "vaild_json"} {
		for _, expr := range []string{"nil", "123", "true", "[]", "{}"} {
			cases = append(cases, jsonOracleCase{"types/" + alias + "/" + expr, "add_key(valid, " + alias + "(" + expr + "))", "", "valid", false, false})
		}
		for i, input := range []string{"null", "true", "1e400", "-1e400", `"\ud800"`, "{} {}", "01", "[1,]", `{"x":1,"x":2}`, strings.Repeat("[", 10_000) + "0" + strings.Repeat("]", 10_000), strings.Repeat("[", 10_001) + "0" + strings.Repeat("]", 10_001)} {
			want := i < 5 || i == 8 || i == 9
			cases = append(cases, jsonOracleCase{fmt.Sprintf("boundary/%s/%d", alias, i), "add_key(valid, " + alias + "(_))", input, "valid", want, false})
		}
	}
	runJSONOracle(t, cases)
}

// Seven cases from fn_load_json_test.go, TestLoadJson (MIT, Guance, Inc.).
func TestJITLoadJSONUpstreamOracle(t *testing.T) {
	cases := []jsonOracleCase{
		{"upstream/0", `abc = load_json(_); add_key(abc, abc["a"]["first"])`, `{"a":{"first":2.3,"second":2,"third":"aBC","forth":true},"age":47}`, "abc", 2.3, false},
		{"upstream/1", `abc = load_json(_); add_key(abc, abc["a"]["first"][-1]); add_key(len_abc, len(load_json(abc["a"]["ff"])))`, `{"a":{"first":[2.2,1.1],"ff":"[2.2, 1.1]","second":2,"third":"aBC","forth":true},"age":47}`, "len_abc", int64(2), false},
		{"upstream/2", `abc = load_json(_); add_key(abc, abc[-1])`, `[2.2,1.1]`, "abc", 1.1, false},
		{"upstream/3", `abc = load_json(_); add_key(abc, len(abc))`, `[]`, "abc", int64(0), false},
		{"upstream/4", `abc = load_json("1"); add_key(abc)`, `{}`, "abc", float64(1), false},
		{"upstream/5", `abc = load_json("true"); add_key(abc)`, `{}`, "abc", true, false},
		{"upstream/6", `abc = load_json("null"); add_key(abc)`, `{}`, "abc", nil, false},
		{"boundary/unicode-whitespace", `abc = load_json(_); add_key(abc)`, "\u00a0true\u3000", "abc", true, false},
		{"boundary/escaped-surrogate", `abc = load_json(_); add_key(abc)`, `"\ud800"`, "abc", "�", false},
		{"boundary/paired-surrogate", `abc = load_json(_); add_key(abc)`, `"\ud83d\ude00"`, "abc", "😀", false},
		{"boundary/literal-backslash", `abc = load_json(_); add_key(abc)`, `"\\ud800"`, "abc", `\ud800`, false},
		{"boundary/nested-surrogate", `abc = load_json(_); add_key(abc, abc["�"][0])`, `{"\udfff":["\ud800"]}`, "abc", "�", false},
		{"boundary/non-string", `abc = load_json(123); add_key(abc)`, `{}`, "", nil, true},
		{"named/unknown", `abc = load_json(unknown=_); add_key(abc)`, `true`, "abc", true, false},
		{"numbers/positive-overflow", `abc = load_json(_); add_key(kind, value_type(abc))`, `1e400`, "kind", "", false},
		{"numbers/negative-overflow", `abc = load_json(_); add_key(kind, value_type(abc))`, `-1e400`, "kind", "", false},
		{"numbers/array-overflow", `abc = load_json(_); add_key(abc)`, `[1e400]`, "abc", nil, false},
		{"numbers/object-overflow", `abc = load_json(_); add_key(abc)`, `{"a":1e400}`, "abc", nil, false},
		{"numbers/underflow", `abc = load_json(_); add_key(abc)`, `1e-400`, "abc", float64(0), false},
		{"numbers/trailing-junk", `abc = load_json(_); add_key(abc)`, `1e400 x`, "abc", nil, false},
	}
	runJSONOracle(t, cases)
}

// Cases of upstream TestVauleType, plus ordinary type regressions.
func TestJITValueTypeNilOracle(t *testing.T) {
	const base = `{"a":{"first":[2.2,1.1],"ff":"[2.2, 1.1]","second":2,"third":"aBC","forth":true},"age":47}`
	const numeric = `{"a":{"first":[2.2,1],"ff":"[2.2, 1.1]","second":2,"third":"aBC","forth":true},"age":47}`
	runJSONOracle(t, []jsonOracleCase{
		{"upstream/0-map", `d = load_json(_); add_key("val_type", value_type(d))`, base, "val_type", "map", false},
		{"upstream/1-list", `d = load_json(_); if value_type(d) == "map" && "a" in d && value_type(d["a"]) == "map" && "first" in d["a"] { add_key("val_type", value_type(d["a"]["first"])) }`, base, "val_type", "list", false},
		{"upstream/2-map_2", `d = load_json(_); if value_type(d) == "map" && "a" in d { add_key("val_type", value_type(d["a"])) }`, base, "val_type", "map", false},
		{"upstream/3-list-not-in", `d = load_json(_); if "a" in d && "first" in d["a"] { add_key("val_type", value_type(d["a"]["fist"])) }`, base, "val_type", "", false},
		{"upstream/4-int-float", `d = load_json(_); add_key("val_type", value_type(d["a"]["first"][1]))`, numeric, "val_type", "float", false},
		{"upstream/5-float", `d = load_json(_); add_key("val_type", value_type(d["a"]["first"][0]))`, numeric, "val_type", "float", false},
		{"upstream/6-int", `d = {"a":1}; add_key("val_type", value_type(d["a"]))`, "", "val_type", "int", false},
		{"upstream/7-bool", `d = load_json(_); add_key("val_type", value_type(d["a"]["first"][0]))`, strings.Replace(numeric, "[2.2,1]", "[true,1]", 1), "val_type", "bool", false},
		{"upstream/8-str", `d = load_json(_); add_key("val_type", value_type(d["a"]["first"][0]))`, strings.Replace(numeric, "[2.2,1]", `["true",1]`, 1), "val_type", "str", false},
		{"upstream/9-empty_nil", `add_key("val_type", value_type(nil))`, "", "val_type", "", false},
		{"upstream/10-empty_nil", `add_key("val_type", value_type(x))`, "", "val_type", "", false},
		{"types/int", `add_key("val_type", value_type(1))`, "", "val_type", "int", false},
		{"types/bool", `add_key("val_type", value_type(true))`, "", "val_type", "bool", false},
		{"types/string", `add_key("val_type", value_type("x"))`, "", "val_type", "str", false},
		{"types/list", `add_key("val_type", value_type([]))`, "", "val_type", "list", false},
		{"types/map", `add_key("val_type", value_type({}))`, "", "val_type", "map", false},
		{"error/division", `add_key("val_type", value_type(1 / divisor))`, "", "val_type", "", false},
		{"error/top-level", `value_type(1 / divisor); add_key("val_type", "continued")`, "", "val_type", "continued", false},
		{"error/index", `items = []; add_key("val_type", value_type(items[1]))`, "", "val_type", "", false},
		{"error/nested-builtin", `add_key("val_type", value_type(load_json(123)))`, "", "val_type", "", false},
		{"error/named", `add_key("val_type", value_type(val=1 / divisor))`, "", "val_type", "", false},
	})
}

func TestJITJSONArgumentAdmissionOracle(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 32)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	t.Run("delete_array_element", func(t *testing.T) {
		source := `json(_,a[0],result,true,true)`
		_, goErr := NewPlScriptSimple(point.Logging, "args.p", source)
		if goErr == nil {
			t.Fatal("Go must reject deleting an array element at compile time")
		}
		if check := runner.Check(source); check.Route == pljit.RouteJITNative {
			t.Fatalf("JIT accepted Go-invalid script: %+v", check)
		}
	})
	for _, name := range []string{"valid_json", "vaild_json", "load_json", "value_type", "len"} {
		for _, args := range []string{"", `"{}"`, `"{}", "extra"`} {
			source := name + "(" + args + ")"
			t.Run(source, func(t *testing.T) {
				_, goErr := NewPlScriptSimple(point.Logging, "args.p", source)
				want := args == `"{}"`
				if (goErr == nil) != want {
					t.Fatalf("unexpected Go admission: %v", goErr)
				}
				check := runner.Check(source)
				if (check.Route == pljit.RouteJITNative) != want {
					t.Fatalf("admission differs: Go=%v native=%+v", goErr, check)
				}
			})
		}
		for _, args := range []string{`val="{}"`, `unknown="{}"`, `"{}", val="{}"`} {
			source := name + "(" + args + ")"
			t.Run(source, func(t *testing.T) {
				_, goErr := NewPlScriptSimple(point.Logging, "args.p", source)
				check := runner.Check(source)
				if (check.Route == pljit.RouteJITNative) != (goErr == nil) {
					t.Fatalf("named admission differs: Go=%v native=%+v", goErr, check)
				}
			})
		}
	}
}

// First three TestLen cases, MIT, pipeline-go v1.4.3; array Point case below.
func TestJITLenUpstreamOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{"upstream/0", `abc = ["1", "2"]; add_key(abc, len(abc))`, "test", "abc", int64(2), false},
		{"upstream/1", `abc = []; add_key(abc, len(abc))`, "test", "abc", int64(0), false},
		{"upstream/2", `abc = {"a":{"first":2.3,"second":2,"third":"aBC","forth":true},"age":47}; add_key(abc, len(abc["a"]))`, "test", "abc", int64(4), false},
		{"upstream/3", `abc = pt_kvs_get("nums", true); add_key(abc, len(abc))`, "test", "abc", int64(3), false},
		{"bytes/unicode", `add_key(abc, len(_))`, "中😀", "abc", int64(7), false},
		{"bytes/nul", `add_key(abc, len(_))`, "a\x00b", "abc", int64(3), false},
		{"type/nil", `add_key(abc, len(nil))`, "", "abc", int64(0), false},
		{"type/number", `add_key(abc, len(123))`, "", "abc", int64(0), false},
		{"type/bool", `add_key(abc, len(true))`, "", "abc", int64(0), false},
		{"named/unknown", `add_key(abc, len(unknown=_))`, "xyz", "abc", int64(3), false},
		{"error/argument", `add_key(abc, len(1 / divisor))`, "", "", nil, true},
	})
}

// Upstream TestPtKvsGetAppendLenChain (MIT) and aliasing/write-back extensions.
func TestJITPointArrayMutationOracle(t *testing.T) {
	if point.EnableMixedArrayField {
		t.Fatal("this corpus pins DataKit's default mixed-array-disabled policy")
	}
	runJSONOracle(t, []jsonOracleCase{
		{"upstream/append-len", `arr = pt_kvs_get("nums", true); arr = append(arr, 4); pt_kvs_set("size", len(arr)); pt_kvs_set("arr", arr, false, true)`, "test", "size", int64(4), false},
		{"mutation/local-copy", `arr = pt_kvs_get("nums", true); arr[0] = 99; original = pt_kvs_get("nums", true); add_key(first, original[0])`, "test", "first", int64(1), false},
		{"mutation/write-back", `arr = pt_kvs_get("nums", true); arr[0] = 99; pt_kvs_set("nums", arr, false, true); original = pt_kvs_get("nums", true); add_key(size, len(original))`, "test", "size", int64(3), false},
		{"mutation/error-prefix", `arr = pt_kvs_get("nums", true); arr[0] = 99; pt_kvs_set("arr", arr, false, true); result = 1 / divisor`, "test", "", nil, true},
		{"mutation/after-write", `arr = [1,2]; pt_kvs_set("arr", arr, false, true); arr[0] = 99; saved = pt_kvs_get("arr", true); add_key(first, saved[0])`, "test", "first", int64(1), false},
		{"map/after-write", `obj = {"x":1}; pt_kvs_set("obj", obj, false, true); obj["x"] = 99; saved = pt_kvs_get("obj", true); add_key(first, saved["x"])`, "test", "first", int64(1), false},
		{"nested/after-write", `obj = {"x":[1,2]}; pt_kvs_set("obj", obj, false, true); obj["x"][0] = 99; saved = pt_kvs_get("obj", true); add_key(first, saved)`, "test", "first", `{"x":[1,2]}`, false},
		{"batch-set/after-write", `obj = {"x":[1,2]}; pt_kvs_set_map(obj, include_keys=["x"], raw=true); obj["x"][0] = 99; saved = pt_kvs_get("x", true); add_key(first, saved[0])`, "test", "first", int64(1), false},
		{"boundary/empty-array", `pt_kvs_set("arr", [], false, true); saved = pt_kvs_get("arr", true); add_key(size, len(saved))`, "test", "size", int64(0), false},
		{"boundary/empty-map", `pt_kvs_set("obj", {}, false, true); saved = pt_kvs_get("obj", true); add_key(size, len(saved))`, "test", "size", int64(0), false},
		{"boundary/nil-array", `pt_kvs_set("arr", [nil], false, true); saved = pt_kvs_get("arr", true); add_key(kind, value_type(saved))`, "test", "kind", "str", false},
		{"boundary/mixed-array", `pt_kvs_set("arr", [1,"x"], false, true); saved = pt_kvs_get("arr", true); add_key(kind, value_type(saved))`, "test", "kind", "str", false},
		{"boundary/mixed-number-array", `pt_kvs_set("arr", [1,1.5], false, true); saved = pt_kvs_get("arr", true); add_key(kind, value_type(saved))`, "test", "kind", "str", false},
	}, func(pt *point.Point) { pt.AddKVs(point.NewKV("nums", []int{1, 2, 3})) })
}

func TestJITPointMixedArrayConfigurationOracle(t *testing.T) {
	previous := point.EnableMixedArrayField
	point.EnableMixedArrayField = true
	defer func() { point.EnableMixedArrayField = previous }()
	runJSONOracleWithServices(t, []jsonOracleCase{
		{
			"raw/mixed-scalars",
			`pt_kvs_set("arr", [1,"x",true,1.5], false, true)`,
			"test",
			"arr",
			[]any{int64(1), "x", true, float64(1.5)},
			false,
		},
	}, pljit.ServiceConfig{Version: 1, EnableMixedArrayField: true})
}

func TestJITPointProtobufScalarAndArrayBoundariesOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{
			"unsigned-and-bytes",
			`add_key(wrapped, unsigned_scalar); unsigned_values = pt_kvs_get("unsigned_values", true); add_key(unsigned_first, unsigned_values[0]); raw = pt_kvs_get("raw_bytes", true); add_key(raw_copy, raw); byte_values = pt_kvs_get("byte_values", true); pt_kvs_set("byte_values_copy", byte_values, false, true)`,
			"test",
			"wrapped",
			int64(-1),
			false,
		},
	}, func(pt *point.Point) {
		pt.AddKVs(
			point.NewKV("unsigned_scalar", uint64(^uint64(0))),
			point.NewKV("unsigned_values", []uint64{^uint64(0), uint64(1)}),
			point.NewKV("raw_bytes", []byte{0xff, 0, 'A'}),
			point.NewKV("byte_values", [][]byte{{0xff}, {'A', 0}}),
		)
	})
}

func TestJITPointDictionaryAndMalformedAnyOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{
			"dictionary-and-malformed",
			`obj = pt_kvs_get("dictionary", true); add_key(unsigned_value, obj["unsigned"]); add_key(raw_kind, value_type(obj["raw"])); add_key(raw_value, obj["raw"]); add_key(map_json, obj); if pt_kvs_get("malformed", true) == nil { add_key(malformed_nil, true) }`,
			"test",
			"unsigned_value",
			int64(-1),
			false,
		},
	}, func(pt *point.Point) {
		dictionary, err := point.NewAny(&point.Map{Map: map[string]*point.BasicTypes{
			"unsigned": {X: &point.BasicTypes_U{U: ^uint64(0)}},
			"raw":      {X: &point.BasicTypes_D{D: []byte{0xff, 0, 'A'}}},
		}})
		if err != nil {
			t.Fatal(err)
		}
		pt.AddKVs(
			point.NewKV("dictionary", dictionary),
			point.NewKV("malformed", &types.Any{TypeUrl: point.DictFieldType, Value: []byte{0xff}}),
		)
	})
}

func TestJITPointNonFiniteFloatOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{
			"scalar-and-containers",
			`strfmt(nan_text, "%v", nan_value); strfmt(pos_text, "%v", pos_inf); strfmt(neg_text, "%v", neg_inf); add_key(nan_equal, nan_value == nan_value); arr = pt_kvs_get("float_array", true); add_key(array_kind, value_type(arr)); add_key(array_json, arr); obj = pt_kvs_get("float_map", true); add_key(map_kind, value_type(obj)); add_key(map_json, obj); drop_key(nan_value); drop_key(pos_inf); drop_key(neg_inf); drop_key(float_array); drop_key(float_map)`,
			"test",
			"nan_text",
			"NaN",
			false,
		},
	}, func(pt *point.Point) {
		floatMap, err := point.NewAny(&point.Map{Map: map[string]*point.BasicTypes{
			"nan": {X: &point.BasicTypes_F{F: math.NaN()}},
			"pos": {X: &point.BasicTypes_F{F: math.Inf(1)}},
		}})
		if err != nil {
			t.Fatal(err)
		}
		pt.AddKVs(
			point.NewKV("nan_value", math.NaN()),
			point.NewKV("pos_inf", math.Inf(1)),
			point.NewKV("neg_inf", math.Inf(-1)),
			point.NewKV("float_array", []float64{math.NaN(), math.Inf(1)}),
			point.NewKV("float_map", floatMap),
		)
	})
}

// This is an upstream defect reproducer, NOT a Go/Rust compatibility pass.
func TestGoRawNilMapKnownPanic(t *testing.T) {
	script, err := NewPlScriptSimple(point.Logging, "nil-map.p", `pt_kvs_set("obj", {"x":nil}, false, true); saved = pt_kvs_get("obj", true)`)
	if err != nil {
		t.Fatal(err)
	}
	panicked := false
	func() {
		defer func() { panicked = recover() != nil }()
		wrapped := ptinput.PtWrap(point.Logging, newRealScriptPoint("nil-map", map[string]any{"message": "test"}))
		if err := script.Run(wrapped, nil, nil); err != nil {
			t.Errorf("unexpected script error instead of known panic: %v", err)
		}
	}()
	if !panicked {
		t.Fatal("upstream nil-map panic changed; re-evaluate compatibility disposition")
	}
}

// DataKit's production Point normalization accepts this physical map even
// though an older pipeline-go PtWrap path panics while rereading it. Keep the
// native program reusable across alternating nil/non-nil records.
func TestJITRawNilMapSucceedsAndReusesProgram(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 4)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	for _, tc := range []struct {
		source string
		key    string
	}{
		{`obj = {"x":nil}; if message == "ok" { obj["x"] = 1 }; pt_kvs_set("obj", obj, false, true); add_key(after, true)`, "obj"},
		{`obj = {"field":{"x":nil}}; if message == "ok" { obj["field"]["x"] = 1 }; pt_kvs_set_map(obj, include_keys=["field"], raw=true); add_key(after, true)`, "field"},
	} {
		source := tc.source
		if check := runner.Check(source); check.Route != pljit.RouteJITNative {
			t.Fatalf("route: %+v", check)
		}
		projection, err := runner.Projection(source)
		if err != nil {
			t.Fatal(err)
		}
		for _, message := range []string{"bad", "ok", "bad", "ok"} {
			pt := newRealScriptPoint("safe", map[string]any{"message": message})
			input, err := encodeProjectedJITPoints(point.Logging, []*point.Point{pt}, projection)
			if err != nil {
				t.Fatal(err)
			}
			batch, err := runner.Process(source, input)
			if err != nil {
				t.Fatal(err)
			}
			if len(batch.Records) != 1 {
				t.Fatal("record count")
			}
			record := batch.Records[0]
			if record.Status != pljit.TerminalOK {
				t.Fatalf("program poisoned: %+v", record)
			}
			if batch.Static != nil {
				_, _, err = applyJITStatic(point.Logging, pt, batch.Static, 0, nil)
			} else {
				_, _, err = applyJITRecord(point.Logging, pt, record, 0, nil)
			}
			if err != nil {
				t.Fatal(err)
			}
			want := map[string]*point.BasicTypes{"x": nil}
			if message == "ok" {
				want["x"] = &point.BasicTypes{X: &point.BasicTypes_I{I: int64(1)}}
			}
			assertPhysicalPointMap(t, pt, tc.key, want)
			if pt.Get("after") != true {
				t.Fatalf("message=%q point=%#v lost continuation", message, pt.KVMap())
			}
		}
	}
}

func runJSONOracle(t *testing.T, cases []jsonOracleCase, setup ...func(*point.Point)) {
	t.Helper()
	runJSONOracleWithServices(t, cases, pljit.ServiceConfig{Version: 1}, setup...)
}

func runJSONOracleWithServices(t *testing.T, cases []jsonOracleCase, config pljit.ServiceConfig, setup ...func(*point.Point)) {
	t.Helper()
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	// Each case retains its route until the corpus finishes. Production
	// capacity/eviction tests remain separate from this semantic oracle.
	runner, err := pljit.NewRunnerWithServices(path, "pipeline-go-1.4.3-datakit", max(32, len(cases)), nil, config)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			source := "add_key(before, true)\n" + tc.source + "\nadd_key(after, true)"
			script, err := NewPlScriptSimple(point.Logging, "valid-json.p", source)
			if err != nil {
				t.Fatal(err)
			}
			if check := runner.Check(source); check.Route != pljit.RouteJITNative {
				t.Fatalf("not native: %+v", check)
			}
			projection, err := runner.Projection(source)
			if err != nil {
				t.Fatal(err)
			}
			for _, count := range []int{1, 2, 4, 8, 10} {
				actual, expected := make([]*point.Point, count), make([]*point.Point, count)
				for i := range actual {
					fields := map[string]any{"message": tc.input, "sentinel": "preserved", "divisor": int64(0)}
					actual[i], expected[i] = newRealScriptPoint("valid", fields), newRealScriptPoint("valid", fields)
					for _, init := range setup {
						init(actual[i])
						init(expected[i])
					}
					if t.Name() == "TestJITLenUpstreamOracle/upstream/3" {
						actual[i].AddKVs(point.NewKV("nums", []int{1, 2, 3}))
						expected[i].AddKVs(point.NewKV("nums", []int{1, 2, 3}))
					}
					wrapped := ptinput.PtWrap(point.Logging, expected[i])
					goErr := script.Run(wrapped, nil, nil)
					if (goErr != nil) != tc.fails {
						t.Fatalf("Go error=%v expected failure=%v", goErr, tc.fails)
					}
					if tc.key != "" && !reflect.DeepEqual(expected[i].Get(tc.key), tc.want) {
						t.Fatalf("upstream expected %v got %v", tc.want, expected[i].Get(tc.key))
					}
				}
				input, err := encodeProjectedJITPoints(point.Logging, actual, projection)
				if err != nil {
					t.Fatal(err)
				}
				batch, err := runner.Process(source, input)
				if err != nil {
					t.Fatal(err)
				}
				if len(batch.Records) != count {
					t.Fatal("record count mismatch")
				}
				for i, record := range batch.Records {
					if (record.Status == pljit.TerminalError) != tc.fails {
						t.Fatalf("terminal=%v error=%s", record.Status, record.Error)
					}
					if tc.fails && !record.CommitPrefixError {
						t.Fatal("lost error prefix")
					}
					if batch.Static != nil {
						_, _, err = applyJITStatic(point.Logging, actual[i], batch.Static, i, nil)
					} else {
						_, _, err = applyJITRecord(point.Logging, actual[i], record, uint64(i), nil)
					}
					if err != nil {
						t.Fatal(err)
					}
					// Match pl.go's successful native completion; errors retain
					// the committed prefix without consuming its time field.
					if record.Status == pljit.TerminalOK {
						ptinput.PtWrap(point.Logging, actual[i]).KeyTime2Time()
					}
					if equal, reason := equalJSONOraclePoints(actual[i], expected[i]); !equal {
						t.Fatalf("batch=%d record=%d: %s\nactual=%s\nexpected=%s", count, i, reason, actual[i].Pretty(), expected[i].Pretty())
					}
				}
			}
		})
	}
}

func TestJITNestedPointReadsOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{"len", `add_key(size, len(pt_kvs_get("nums", true)))`, "test", "size", int64(3), false},
		{"raw-type", `add_key(kind, value_type(pt_kvs_get("nums", true)))`, "test", "kind", "list", false},
		{"string-type", `add_key(kind, value_type(pt_kvs_get("nums")))`, "test", "kind", "str", false},
		{"dynamic-key", `key = "nums"; add_key(size, len(pt_kvs_get(key, true)))`, "test", "size", int64(3), false},
		{"keys", `add_key(found, "nums" in pt_kvs_keys(false, true))`, "test", "found", true, false},
		{"multiple-reads", `add_key(size, len(pt_kvs_get("nums", true)) + len(pt_kvs_get("nums", true)))`, "test", "size", int64(6), false},
		{"missing", `add_key(kind, value_type(pt_kvs_get("missing", true)))`, "test", "kind", "", false},
		{"failure", `add_key(size, len(pt_kvs_get(1 / divisor, true)))`, "test", "", nil, true},
	}, func(pt *point.Point) { pt.AddKVs(point.NewKV("nums", []int{1, 2, 3})) })
}

// TestAppend from pipeline-go v1.4.3 (MIT, Guance, Inc.), plus lazy-argument cases.
func TestJITAppendUpstreamOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{"upstream/0", `abc = ["1","2"]; abc = append(abc,5.1); add_key(arr,abc)`, "test", "arr", `["1","2",5.1]`, false},
		{"upstream/1", `abc = ["hello"]; abc = append(abc,"world"); add_key(arr,abc)`, "test", "arr", `["hello","world"]`, false},
		{"upstream/2", `abc = [1,2]; abc = append(abc,"3"); add_key(arr,abc)`, "test", "arr", `[1,2,"3"]`, false},
		{"upstream/3", `a = [1,2]; b = append(a,3); add_key(arr,b)`, "test", "arr", `[1,2,3]`, false},
		{"upstream/4", `a = [1,2]; b = [3,4]; c = append(a,b); add_key(arr,c)`, "test", "arr", `[1,2,[3,4]]`, false},
		{"upstream/5", `a = [1,2]; b = 3; append(a,b); add_key(arr,a)`, "test", "arr", `[1,2]`, false},
		{"upstream/6", `a = pt_kvs_get("nums",true); b = append(a,3); pt_kvs_set("arr",b,false,true)`, "test", "arr", []int64{1, 2, 3}, false},
		{"lazy/missing", `append(missing, 1 / divisor); add_key(done,true)`, "test", "done", true, false},
		{"lazy/non-list", `a = 1; append(a, 1 / divisor); add_key(done,true)`, "test", "done", true, false},
		{"lazy/nested", `a = 1; add_key(kind, value_type(append(a, 1 / divisor)))`, "test", "kind", "", false},
		{"lazy/nested-no-catch", `a = 1; result = append(a, 1 / divisor); add_key(done,true)`, "test", "done", true, false},
		{"error/list", `a = []; append(a, 1 / divisor); add_key(done,true)`, "test", "", nil, true},
		{"named/second", `a = [1]; b = append(a, elem=2); add_key(arr,b); add_key(assigned,elem)`, "test", "", nil, false},
		{"named/custom-assignment", `a = [1]; b = append(a, custom=2); add_key(arr,b); add_key(assigned,custom)`, "test", "assigned", int64(2), false},
		{"named/lazy", `a = 1; append(a, elem=1 / divisor); add_key(done,true)`, "test", "done", true, false},
		{"named/nested-lazy", `a = 1; b = append(a, elem=1 / divisor); add_key(done,true)`, "test", "done", true, false},
	}, func(pt *point.Point) { pt.AddKVs(point.NewKV("nums", []int{1, 2})) })
}

func TestJITAppendFirstArgumentAdmission(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 8)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	for _, expression := range []string{`append(arr=a,elem=2)`, `append(first=a,second=2)`, `append([],2)`, `append(load_json("[]"),2)`, `delete({},"a")`} {
		t.Run(expression, func(t *testing.T) {
			source := "a = []; " + expression
			if _, err := NewPlScriptSimple(point.Logging, "append.p", source); err == nil {
				t.Fatal("Go unexpectedly accepted non-identifier target")
			}
			check := runner.Check(source)
			if check.Route == pljit.RouteJITNative || !strings.Contains(check.Detail, "E_FUNCTION_ARGUMENT") {
				t.Fatalf("wrong rejection: %+v", check)
			}
		})
	}
	if check := runner.Check(`a={"a":1}; key="a"; delete(a,key)`); check.Route != pljit.RouteJITNative {
		t.Fatalf("dynamic delete key extension rejected: %+v", check)
	}
}

func TestJITValueFunctionAssignmentArguments(t *testing.T) {
	var cases []jsonOracleCase
	for _, name := range []string{"load_json", "valid_json", "vaild_json", "len"} {
		for _, prefix := range []string{"", "result = "} {
			cases = append(cases, jsonOracleCase{name + "/" + prefix, `assigned = "before"; ` + prefix + name + `(assigned="true"); add_key(observed,assigned)`, "test", "observed", "true", false})
		}
	}
	cases = append(cases, jsonOracleCase{"value_type-binding-not-assignment", `val = "before"; result = value_type(val=1); add_key(observed,val)`, "test", "observed", "before", false})
	cases = append(cases,
		jsonOracleCase{"sequence/same-expression", `assigned = "before"; add_key(size, len(assigned="abcd") + len(assigned))`, "test", "size", int64(8), false},
		jsonOracleCase{"error/rhs-not-assigned", `assigned = "before"; value_type(load_json(assigned=1 / divisor)); add_key(observed,assigned)`, "test", "observed", "before", false},
		jsonOracleCase{"error/assignment-before-builtin-failure", `assigned = "before"; value_type(load_json(assigned=123)); add_key(observed,assigned)`, "test", "observed", int64(123), false},
	)
	runJSONOracle(t, cases)
}

// TestStrlen, pipeline-go v1.4.3, plus runtime boundary cases.
func TestJITStrlenUpstreamOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{"upstream/0", `add_key("k1",strlen("你好"))`, "", "k1", int64(2), false},
		{"upstream/1", `add_key("k1",strlen("hello"))`, "", "k1", int64(5), false},
		{"upstream/2", `add_key("k1",strlen("你好hello"))`, "", "k1", int64(7), false},
		{"upstream/3", `v = []; v = append(v,strlen("hello你好")); v = append(v,len("hello你好")); add_key("v",v)`, "", "v", "[7,11]", false},
		{"boundary/codepoints", `add_key(size,strlen(_))`, "e\u0301😀", "size", int64(3), false},
		{"boundary/nul", `add_key(size,strlen(_))`, "a\x00b", "size", int64(3), false},
		{"boundary/empty", `add_key(size,strlen(_))`, "", "size", int64(0), false},
		{"error/nil", `strlen(nil)`, "", "", nil, true},
		{"error/list", `strlen([])`, "", "", nil, true},
		{"error/missing", `strlen(missing)`, "", "", nil, true},
		{"assignment/argument", `assigned = "before"; add_key(size,strlen(assigned="你好")); add_key(observed,assigned)`, "", "observed", "你好", false},
	})
}

// Runtime cases 0-5 of TestDelete, pipeline-go v1.4.3 (MIT, Guance, Inc.).
func TestJITDeleteUpstreamOracle(t *testing.T) {
	const input = `{"a":[1,{"b":2}]}`
	runJSONOracle(t, []jsonOracleCase{
		{"upstream/0", `a=load_json(_); delete(a["a"][-1],"b"); add_key(a)`, input, "a", `{"a":[1,{}]}`, false},
		{"upstream/1", `j_map=load_json(_); delete(j_map["b"][-1],"c"); delete(j_map,"a"); add_key("j_map",j_map)`, `{"a":"b","b":[0,{"c":"d"}],"e":1}`, "j_map", `{"b":[0,{}],"e":1}`, false},
		{"upstream/2", `a=load_json(_); delete(a,"a"); add_key(a)`, input, "a", `{}`, false},
		{"upstream/3", `a=load_json(_); delete(a["a"],"b"); add_key(a)`, input, "a", input, false},
		{"upstream/4", `a=load_json(_); delete(b,"b"); add_key(a)`, input, "a", input, false},
		{"upstream/5", `a=load_json(_); delete(a["a"][-7],"b"); add_key(a)`, input, "a", input, false},
		{"boundary/index-error", `a=load_json(_); delete(a[1 / divisor],"b"); add_key(a)`, input, "a", input, false},
		{"boundary/named", `a=load_json(_); delete(src=a,key="a"); add_key(a)`, input, "a", `{}`, false},
		{"order/non-container-root", `a=1; marker="before"; delete(a[len(marker="changed")],"b"); add_key(observed,marker)`, input, "observed", "before", false},
		{"order/missing-root", `marker="before"; delete(missing[len(marker="changed")],"b"); add_key(observed,marker)`, input, "observed", "before", false},
		{"order/failed-first-index", `a=[]; marker="before"; delete(a[99][len(marker="changed")],"b"); add_key(observed,marker)`, input, "observed", "before", false},
		{"order/scalar-intermediate", `a=[1]; marker="before"; delete(a[0][len(marker="changed")],"b"); add_key(observed,marker)`, input, "observed", "changed", false},
		{"order/missing-map-key", `a={}; marker="before"; delete(a["missing"][len(marker="changed")],"b"); add_key(observed,marker)`, input, "observed", "before", false},
		{"order/nil-map-value", `a={"present":nil}; marker="before"; delete(a["present"][len(marker="changed")],"b"); add_key(observed,marker)`, input, "observed", "changed", false},
	})
}

// TestB64enc, pipeline-go v1.4.3 (MIT, Guance, Inc.), plus write/read chains.
func TestJITBase64ArgumentAdmission(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 8)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	for _, name := range []string{"b64enc", "b64dec"} {
		for _, args := range []string{"", `"message"`, `123`, `key=message`, `message,message`, `load_json("1")`, `message[0]`} {
			source := name + "(" + args + ")"
			t.Run(source, func(t *testing.T) {
				if _, err := NewPlScriptSimple(point.Logging, "base64.p", source); err == nil {
					t.Fatal("Go unexpectedly accepted invalid arguments")
				}
				check := runner.Check(source)
				if check.Route == pljit.RouteJITNative || check.Detail == "" {
					t.Fatalf("JIT accepted Go-invalid arguments or omitted reason: %+v", check)
				}
			})
		}
	}
}

// TestTrim, pipeline-go v1.4.3 (MIT, Guance, Inc.), plus Unicode and write-back cases.
func TestJITTrimUpstreamOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{"upstream/0", `item=" not_space "; trim(item)`, "trim space", "item", "not_space", false},
		{"upstream/1", `item="BC_-AAACAnot_spaceABACC"; trim(item,"ABC_-")`, "trim ABC_-", "item", "not_space", false},
		{"upstream/2", `add_key(test_data,"ACCAA_test_DataA_ACBA"); trim(test_data,"ABC_")`, "trim ABC_", "test_data", "test_Data", false},
		{"boundary/unicode-space", `trim(message)`, "\u0085\u00a0\u2003你好\u3000", "message", "你好", false},
		{"boundary/non-space", `trim(message)`, "\u200btest\ufeff", "message", "\u200btest\ufeff", false},
		{"boundary/empty-cutset", `trim(message,"")`, " \t test \r\n", "message", "test", false},
		{"boundary/unicode-cutset", `trim(message,"你好")`, "好你text你好", "message", "text", false},
		{"boundary/nul", `trim(message)`, " \x00 ", "message", "\x00", false},
		{"boundary/empty", `trim(message)`, "", "message", "", false},
		{"boundary/missing", `trim(missing)`, "test", "message", "test", false},
		{"boundary/number", `item=123; trim(item,"13")`, "test", "item", "2", false},
		{"boundary/local-shadow", `item=" x "; trim(item); add_key(observed,item)`, "test", "observed", " x ", false},
		{"boundary/error-prefix", `trim(message); result=1/divisor`, " test ", "message", "test", true},
	})
}

// Adapted from TestLowercase/TestUppercase, pipeline-go v1.4.3 (MIT, Guance, Inc.).
func TestJITCaseConversionUpstreamOracle(t *testing.T) {
	cases := []jsonOracleCase{
		{"lower/0", `json(_,a.third); lowercase(a.third)`, `{"a":{"first":2.3,"second":2,"third":"aBC","forth":true},"age":47}`, "a.third", "abc", false},
		{"lower/1", `json(_,a.third); lowercase(a.third)`, `{"a":{"first":2.3,"second":2,"third":"aBC","forth":true,"age":"WWW"},"age":"wWW"}`, "a.age", nil, false},
		{"lower/2", `json(_,a.third); lowercase(a.third)`, `{"a":{"first":2.3,"second":2,"third":"aBC","forth":true,"age":"WWW"},"age":"wWW"}`, "a.forth", nil, false},
		{"lower/3", `json(_,a.third); lowercase(a.third)`, `{"a":{"first":"222SSd","second":2,"third":"aBC","forth":true,"age":"WWW"},"age":"wWW"}`, "a.first", nil, false},
		{"lower/4", `json(_,a.first); lowercase(a.first)`, `{"a":{"first":"SSd","second":2,"third":"aBC","forth":true,"age":"WWW"},"age":"wWW"}`, "a.first", "ssd", false},
		{"lower/5", `json(_,age); lowercase(age)`, `{"a":{"first":"SSd","second":2,"third":"aBC","forth":true,"age":"WWW"},"age":"wWW"}`, "age", "www", false},
		{"upper/0", `json(_,a.third); uppercase(a.third)`, `{"a":{"first":2.3,"second":2,"third":"abc","forth":true},"age":47}`, "a.third", "ABC", false},
		{"upper/1", `json(_,age); uppercase(age)`, `{"a":{"first":2.3,"second":2,"third":"abc","forth":true},"age":47}`, "age", "47", false},
		{"upper/2", `json(_,a.forth); uppercase(a.forth)`, `{"a":{"first":2.3,"second":2,"third":"abc","forth":"1a2B3c/d"},"age":47}`, "a.forth", "1A2B3C/D", false},
	}
	for _, name := range []string{"lowercase", "uppercase"} {
		for i, input := range []string{"İıßẞﬃΣςσ", "ΟΣ", "你好Éé\x00", ""} {
			want := strings.ToLower(input)
			if name == "uppercase" {
				want = strings.ToUpper(input)
			}
			cases = append(cases, jsonOracleCase{fmt.Sprintf("%s/unicode/%d", name, i), name + `(message)`, input, "message", want, false})
		}
		cases = append(cases, jsonOracleCase{name + "/missing", name + `(missing)`, "test", "message", "test", false})
		cases = append(cases, jsonOracleCase{name + "/quoted-missing", name + `("hello")`, "test", "message", "test", false})
		quotedWant := "abc"
		if name == "uppercase" {
			quotedWant = "ABC"
		}
		cases = append(cases, jsonOracleCase{name + "/quoted-present", name + `("message")`, "aBc", "message", quotedWant, false})
		cases = append(cases, jsonOracleCase{name + "/shadow", `item="aB"; ` + name + `(item); add_key(observed,item)`, "test", "observed", "aB", false})
	}
	runJSONOracle(t, cases)
}

func TestJITCaseConversionArgumentAdmission(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 8)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	for _, name := range []string{"lowercase", "uppercase"} {
		for _, args := range []string{"", `a.forth,"someArg"`, `123`, `key=message`, `load_json("1")`, `message[0]`} {
			source := name + "(" + args + ")"
			t.Run(source, func(t *testing.T) {
				if _, err := NewPlScriptSimple(point.Logging, "case.p", source); err == nil {
					t.Fatal("Go unexpectedly accepted invalid arguments")
				}
				check := runner.Check(source)
				if check.Route == pljit.RouteJITNative || check.Detail == "" {
					t.Fatalf("JIT accepted Go-invalid arguments or omitted reason: %+v", check)
				}
			})
		}
	}
}

// TestURLDecode, pipeline-go v1.4.3 (MIT, Guance, Inc.), plus escaped-byte corpus.
func TestJITURLDecodeUpstreamOracle(t *testing.T) {
	const encoded = "https:%2F%2Fkubernetes.io%2Fdocs%2Freference%2Faccess-authn-authz%2Fbootstrap-tokens%2F"
	cases := []jsonOracleCase{
		{"upstream/0", `json(_,url); url_decode(url)`, `{"url":"http%3a%2f%2fwww.baidu.com%2fs%3fwd%3d%e6%b5%8b%e8%af%95"}`, "url", "http://www.baidu.com/s?wd=测试", false},
		{"upstream/1", `json(_,url); url_decode(url)`, `{"url":"` + encoded + `"}`, "url", "https://kubernetes.io/docs/reference/access-authn-authz/bootstrap-tokens/", false},
		{"upstream/2", `json(_,url); url_decode(link)`, `{"url":"` + encoded + `"}`, "link", nil, false},
		{"upstream/3", `url_decode("` + encoded + `")`, `{"url":"` + encoded + `"}`, "message", `{"url":"` + encoded + `"}`, false},
		{"boundary/plus", `url_decode(message)`, "a+b%2Bc", "message", "a b+c", false},
		{"boundary/binary", `url_decode(message)`, "%FF%00", "message", string([]byte{255, 0}), false},
		{"boundary/once", `url_decode(message)`, "%252F", "message", "%2F", false},
		{"boundary/empty", `url_decode(message)`, "", "message", "", false},
		{"boundary/number", `item=123; url_decode(item)`, "test", "item", "123", false},
		{"boundary/shadow", `item="a%20b"; url_decode(item); add_key(observed,item)`, "test", "observed", "a%20b", false},
	}
	for i, input := range []string{"%", "%0", "%GG", "ok%20then%XZ"} {
		cases = append(cases, jsonOracleCase{fmt.Sprintf("invalid/%d", i), `url_decode(message)`, input, "message", input, true})
	}
	runJSONOracle(t, cases)
}

func TestJITURLDecodeArgumentAdmission(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 8)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	for _, args := range []string{"", `message,message`, `123`, `key=message`, `load_json("1")`, `message[0]`} {
		source := "url_decode(" + args + ")"
		t.Run(source, func(t *testing.T) {
			if _, err := NewPlScriptSimple(point.Logging, "url-decode.p", source); err == nil {
				t.Fatal("Go unexpectedly accepted invalid arguments")
			}
			check := runner.Check(source)
			if check.Route == pljit.RouteJITNative || check.Detail == "" {
				t.Fatalf("JIT accepted Go-invalid arguments or omitted reason: %+v", check)
			}
		})
	}
}

// TestSliceString, pipeline-go v1.4.3 (MIT, Guance, Inc.), successful cases and boundaries.
func TestJITSliceStringUpstreamOracle(t *testing.T) {
	cases := []jsonOracleCase{}
	for i, tc := range []struct{ expr, want string }{
		{`slice_string("█汉字15384073392",0,5)`, "█汉字15"},
		{`slice_string("15384073392",5,10)`, "07339"},
		{`slice_string("abcdefghijklmnop",0,10)`, "abcdefghij"},
		{`slice_string("abcdefghijklmnop",-1,10)`, ""},
		{`slice_string("abcdefghijklmnop",0,100)`, "abcdefghijklmnop"},
	} {
		cases = append(cases, jsonOracleCase{fmt.Sprintf("upstream/%d", i), `substring=` + tc.expr + `; pt_kvs_set("result",substring)`, "", "result", tc.want, false})
	}
	cases = append(cases, jsonOracleCase{"upstream/10", `val="123你好123123123123123123123123123"; add_key("result",slice_string(val,0,len(val)))`, "", "result", "123你好123123123123123123123123123", false})
	for i, tc := range []struct{ expr, want string }{
		{`slice_string("é😊",1,3)`, "́😊"},
		{`slice_string("abc",3,3)`, ""},
		{`slice_string("abc",4,100)`, ""},
		{`slice_string("abc",2,1)`, ""},
		{`slice_string("abc",0,-1)`, ""},
		{`slice_string("",0,9223372036854775807)`, ""},
		{`slice_string(end=2,name="abc",start=1)`, "b"},
	} {
		cases = append(cases, jsonOracleCase{fmt.Sprintf("boundary/%d", i), `add_key(result,` + tc.expr + `)`, "", "result", tc.want, false})
	}
	runJSONOracle(t, cases)
}

func TestJITSliceStringArgumentAdmission(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 8)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	for i, expr := range []string{`slice_string("abcdefghijklmnop","a","b")`, `slice_string("abcdefghijklmnop","abc","def")`, `slice_string(12345,0,3)`, `slice_string("abcdefghijklmnop",0)`, `slice_string("abcdefghijklmnop",0,1,2)`} {
		t.Run(fmt.Sprintf("upstream/%d", i+5), func(t *testing.T) {
			source := `substring=` + expr + `; pt_kvs_set("result",substring)`
			if _, err := NewPlScriptSimple(point.Logging, "slice.p", source); err == nil {
				t.Fatal("Go unexpectedly accepted invalid arguments")
			}
			check := runner.Check(source)
			if check.Route == pljit.RouteJITNative || check.Detail == "" {
				t.Fatalf("JIT accepted Go-invalid arguments or omitted reason: %+v", check)
			}
		})
	}
}

func TestJITSliceStringDynamicTypeOracle(t *testing.T) {
	cases := []jsonOracleCase{
		{"start", `idx=message; result=slice_string("abc",idx,2)`, "invalid", "", nil, true},
		{"end", `idx=message; result=slice_string("abc",0,idx)`, "invalid", "", nil, true},
		{"source", `src=divisor; result=slice_string(src,0,2)`, "test", "", nil, true},
	}
	for _, value := range []string{"1.5", "true", "nil", "[]", "{}"} {
		for _, index := range []string{"start", "end"} {
			args := `"abc",idx,2`
			if index == "end" {
				args = `"abc",0,idx`
			}
			cases = append(cases, jsonOracleCase{index + "/" + value, `idx=` + value + `; result=slice_string(` + args + `)`, "test", "", nil, true})
		}
	}
	runJSONOracle(t, cases)
}

func TestJITGrokOptionalCaptureOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{"missing-group", `add_key(optional,"old"); grok(message,"^%{WORD:first}(?: %{WORD:optional})?$")`, "hello", "optional", "", false},
		{"present-group", `add_key(optional,"old"); grok(message,"^%{WORD:first}(?: %{WORD:optional})?$")`, "hello world", "optional", "world", false},
		{"no-match", `add_key(optional,"old"); grok(message,"^%{WORD:first}(?: %{WORD:optional})?$")`, "123!", "optional", "old", false},
		{"sequential-clear", `grok(message,"^%{WORD:optional}.*$"); grok(message,"^%{WORD:first}(?: %{WORD:optional})?$")`, "hello", "optional", "", false},
	})
}

func TestJITCacheExpirationTypeOracle(t *testing.T) {
	var cases []jsonOracleCase
	for _, value := range []string{`"1"`, `"bad"`, "1.5", "true", "nil", "[]", "{}"} {
		cases = append(cases, jsonOracleCase{value, `expiry=` + value + `; cache_set("expiry-type","value",expiry)`, "test", "", nil, true})
	}
	runJSONOracle(t, cases)
}

func TestJITCacheExpirationOverflowOracle(t *testing.T) {
	var cases []jsonOracleCase
	for _, expiry := range []string{"0", "-1", "9223372036854775807"} {
		cases = append(cases, jsonOracleCase{expiry, `cache_set("ttl-boundary","old"); cache_set("ttl-boundary","new",` + expiry + `); value=cache_get("ttl-boundary"); add_key(observed,value)`, "test", "observed", "old", false})
	}
	runJSONOracle(t, cases)
}

func TestJITCacheGetAssignmentOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{"known-name", `cache_set("target","value"); key="before"; result=cache_get(key="target"); add_key(assigned,key); add_key(result,result)`, "test", "assigned", "target", false},
		{"custom-name", `cache_set("target","value"); custom="before"; result=cache_get(custom="target"); add_key(assigned,custom); add_key(result,result)`, "test", "assigned", "target", false},
		{"top-level", `custom="before"; cache_get(custom="missing"); add_key(assigned,custom)`, "test", "assigned", "missing", false},
		{"type-error", `custom="before"; cache_get(custom=123)`, "test", "", nil, true},
	})
}

func TestJITNestedCacheGetOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{"value", `cache_set("nested","value"); add_key(result,cache_get("nested"))`, "test", "result", "value", false},
		{"missing", `add_key(kind,value_type(cache_get("missing-nested")))`, "test", "kind", "", false},
		{"assignment", `cache_set("nested","value"); custom="before"; add_key(result,cache_get(custom="nested")); add_key(assigned,custom)`, "test", "assigned", "nested", false},
		{"comparison", `cache_set("nested","value"); if cache_get("nested") == "value" { add_key(result,true) }`, "test", "result", true, false},
		{"type-error", `add_key(result,cache_get(123))`, "test", "", nil, true},
		{"multi-assignment-error", `left,right=cache_get("key")`, "test", "", nil, true},
	})
}

// TestCache, pipeline-go v1.4.3 ptinput/funcs/fn_cache_test.go,
// MIT License; Copyright 2021-present Guance, Inc., plus cached value shapes.
func TestJITCacheValueOracle(t *testing.T) {
	cases := []jsonOracleCase{
		{"upstream/0", `cache_set("a","123",5); a=cache_get("a"); add_key(abc,a)`, "[]", "abc", "123", false},
		{"upstream/1", `a=cache_set("a","123"); a=cache_get("a"); add_key(abc,a)`, "[]", "abc", "123", false},
	}
	// Preserve the original isolated PlPt/cache setup and Get assertion in
	// addition to the DataKit full-Point differential below. These are the same
	// two upstream cases, not two additional migrated cases.
	for _, tc := range cases {
		t.Run(tc.name+"/original_fixture", func(t *testing.T) {
			scripts, errs := engine.ParseScript(map[string]string{"cache.p": tc.source}, funcs.FuncsMap, funcs.FuncsCheckMap)
			if len(errs) != 0 {
				t.Fatal(errs)
			}
			cache, err := plcache.NewCache(time.Second, 100)
			if err != nil {
				t.Fatal(err)
			}
			defer cache.Stop()
			pt := ptinput.NewPlPt(point.Logging, "test", nil, map[string]any{"message": tc.input}, time.Unix(1700000000, 123))
			pt.SetCache(cache)
			if err := scripts["cache.p"].Run(pt, nil); err != nil {
				t.Fatal(err)
			}
			cache.Stop()
			got, _, err := pt.Get(tc.key)
			if err != nil || !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("original cache Get: got=%#v want=%#v err=%v", got, tc.want, err)
			}
		})
	}
	for _, value := range []string{"123", "1.5", "true", "nil", "[]", "{}"} {
		cases = append(cases, jsonOracleCase{value, `cache_set("shape",` + value + `); result=cache_get("shape"); add_key(result,result)`, "test", "", nil, false})
		cases = append(cases, jsonOracleCase{"type/" + value, `cache_set("shape",` + value + `); result=cache_get("shape"); add_key(kind,value_type(result))`, "test", "kind", "str", false})
	}
	runJSONOracle(t, cases)
}

// CacheGet in pipeline-go returns an ast.String type marker even for a
// container payload. Keep type, indexing, aliasing and Point snapshots separate
// from ordinary list/map semantics. The upstream len(container) panic is
// characterized separately and is deliberately not a compatibility target.
func TestJITCacheContainerOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{"list-type", `cache_set("c",[1,2]); x=cache_get("c"); add_key(kind,value_type(x))`, "test", "kind", "str", false},
		{"map-type", `cache_set("c",{"a":1}); x=cache_get("c"); add_key(kind,value_type(x))`, "test", "kind", "str", false},
		{"list-index-error", `cache_set("c",[1,2]); x=cache_get("c"); add_key(prefix,"kept"); add_key(element,x[0]); add_key(unreachable,true)`, "test", "prefix", "kept", true},
		{"list-source-alias", `a=[1,2]; cache_set("c",a); a[0]=3; x=cache_get("c"); add_key(result,x)`, "test", "result", []int64{3, 2}, false},
		{"map-source-alias", `a={"x":1}; cache_set("c",a); a["x"]=2; x=cache_get("c"); add_key(result,x)`, "test", "", nil, false},
		{"list-output-snapshot", `a=[1,2]; cache_set("c",a); x=cache_get("c"); add_key(result,x); a[0]=9`, "test", "result", []int64{1, 2}, false},
		{"map-output-snapshot", `a={"x":1}; cache_set("c",a); x=cache_get("c"); add_key(result,x); a["x"]=9`, "test", "", nil, false},
	})
}

func TestJITCacheConditionOracle(t *testing.T) {
	var cases []jsonOracleCase
	for _, value := range []string{"0", "123", "0.0", "1.5", "true", "false", "nil", "[]", "{}"} {
		want := value != "nil" && value != "[]" && value != "{}"
		cases = append(cases, jsonOracleCase{value, `cache_set("c",` + value + `); x=cache_get("c"); if x { add_key(branch,true) } else { add_key(branch,false) }`, "test", "branch", want, false})
	}
	runJSONOracle(t, cases)
}

func TestJITCacheEqualityOracle(t *testing.T) {
	var cases []jsonOracleCase
	for _, tc := range []struct{ value, text string }{
		{"123", `"123"`}, {"1.5", `"1.5"`}, {"false", `"false"`},
		{"nil", `""`}, {"[]", `""`}, {"{}", `""`},
	} {
		prefix := `cache_set("c",` + tc.value + `); x=cache_get("c"); `
		for _, expr := range []string{"x == " + tc.text, tc.text + " == x", "x == cache_get(\"c\")"} {
			cases = append(cases, jsonOracleCase{tc.value + "/" + expr, prefix + `add_key(result,` + expr + `)`, "test", "result", true, false})
		}
		cases = append(cases, jsonOracleCase{tc.value + "/not-equal", prefix + `add_key(result,x != ` + tc.text + `)`, "test", "result", false, false})
		cases = append(cases, jsonOracleCase{tc.value + "/different-string", prefix + `add_key(result,x == "different")`, "test", "result", false, false})
	}
	for _, value := range []string{"123", "1.5", "false", "nil"} {
		for _, expr := range []string{"x == " + value, value + " == x"} {
			cases = append(cases, jsonOracleCase{"different-marker/" + expr, `cache_set("c",` + value + `); x=cache_get("c"); add_key(result,` + expr + `)`, "test", "result", false, false})
		}
	}
	cases = append(cases, jsonOracleCase{"distinct-containers-same-string-cast", `cache_set("a",[1]); cache_set("b",{"x":2}); a=cache_get("a"); b=cache_get("b"); add_key(result,a==b)`, "test", "result", true, false})
	runJSONOracle(t, cases)
}

func TestJITCachePointWriteOracle(t *testing.T) {
	var cases []jsonOracleCase
	for _, value := range []string{"123", "false", "nil", "[1,2]", `{"x":1}`} {
		for _, raw := range []string{"false", "true"} {
			cases = append(cases, jsonOracleCase{value + "/raw=" + raw,
				`a=` + value + `; cache_set("c",a); x=cache_get("c"); pt_kvs_set("result",x,false,` + raw + `)`, "test", "", nil, false})
			cases = append(cases, jsonOracleCase{value + "/set-map/raw=" + raw,
				`a=` + value + `; cache_set("c",a); x=cache_get("c"); n=pt_kvs_set_map({"result":x},include_keys=["result"],raw=` + raw + `); add_key(count,n)`, "test", "count", int64(1), false})
		}
	}
	for _, raw := range []string{"false", "true"} {
		cases = append(cases,
			jsonOracleCase{"list-snapshot/" + raw, `a=[1,2]; cache_set("c",a); x=cache_get("c"); pt_kvs_set("result",x,false,` + raw + `); a[0]=9`, "test", "", nil, false},
			jsonOracleCase{"map-snapshot/" + raw, `a={"x":1}; cache_set("c",a); x=cache_get("c"); pt_kvs_set("result",x,false,` + raw + `); a["x"]=9`, "test", "", nil, false})
	}
	runJSONOracle(t, cases)
}

func TestJITCacheTagWriteOracle(t *testing.T) {
	var cases []jsonOracleCase
	for _, value := range []string{`"text"`, "123", "false", "nil", "[1,2]", `{"x":1}`} {
		for _, raw := range []string{"false", "true"} {
			cases = append(cases, jsonOracleCase{value + "/raw=" + raw,
				`a=` + value + `; cache_set("c",a); x=cache_get("c"); ok=pt_kvs_set("result",x,true,` + raw + `); add_key(success,ok)`, "test", "", nil, false})
			cases = append(cases, jsonOracleCase{value + "/set-map/raw=" + raw,
				`a=` + value + `; cache_set("c",a); x=cache_get("c"); n=pt_kvs_set_map({"result":x},include_keys=["result"],as_tag=true,raw=` + raw + `); add_key(count,n)`, "test", "count", int64(1), false})
		}
	}
	cases = append(cases,
		jsonOracleCase{"list-snapshot", `a=[1,2]; cache_set("c",a); x=cache_get("c"); pt_kvs_set("result",x,true); a[0]=9`, "test", "result", "[1,2]", false},
		jsonOracleCase{"map-snapshot", `a={"x":1}; cache_set("c",a); x=cache_get("c"); pt_kvs_set("result",x,true); a["x"]=9`, "test", "result", `{"x":1}`, false},
		jsonOracleCase{"map-write-snapshot", `a=[1,2]; cache_set("c",a); x=cache_get("c"); pt_kvs_set_map({"result":x},include_keys=["result"],as_tag=true); a[0]=9`, "test", "result", "[1,2]", false},
		jsonOracleCase{"field-to-tag", `add_key(result,7); cache_set("c",[1,2]); x=cache_get("c"); pt_kvs_set("result",x,true)`, "test", "result", "[1,2]", false},
		jsonOracleCase{"existing-tag-retained", `pt_kvs_set("result","old",true); cache_set("c",[1,2]); x=cache_get("c"); pt_kvs_set("result",x,false,true)`, "test", "result", "[1,2]", false})
	runJSONOracle(t, cases)
}

func TestJITCacheNestedContainerOracle(t *testing.T) {
	var cases []jsonOracleCase
	for _, value := range []string{"123", "false", `"text"`, "[1,2]", `{"x":1}`} {
		for _, container := range []string{"[x]", `{"child":x}`} {
			for _, raw := range []string{"false", "true"} {
				cases = append(cases, jsonOracleCase{value + "/" + container + "/raw=" + raw,
					`cache_set("c",` + value + `); x=cache_get("c"); outer=` + container + `; pt_kvs_set("result",outer,false,` + raw + `)`, "test", "", nil, false})
			}
		}
	}
	runJSONOracle(t, cases)
}

func TestJITCacheMixedArrayOracle(t *testing.T) {
	var cases []jsonOracleCase
	for _, value := range []string{"123", "1.5", "false", `"text"`, "nil", "[1,2]"} {
		for _, array := range []string{"[x,123]", "[123,x]", `[x,"text"]`, "[x,x]"} {
			for _, raw := range []string{"false", "true"} {
				cases = append(cases, jsonOracleCase{value + "/" + array + "/" + raw,
					`cache_set("c",` + value + `); x=cache_get("c"); a=` + array + `; pt_kvs_set("result",a,false,` + raw + `)`, "test", "", nil, false})
			}
		}
	}
	runJSONOracle(t, cases)
}

func TestJITTrimArgumentAdmission(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 8)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	for _, source := range []string{
		`trim()`, `trim(message,"a","b")`, `trim(123)`,
		`trim(message,123)`,
		`trim(key=message)`, `trim(message,cutset="a")`,
		`trim(load_json("1"))`, `trim(message[0])`,
	} {
		t.Run(source, func(t *testing.T) {
			if _, err := NewPlScriptSimple(point.Logging, "trim.p", source); err == nil {
				t.Fatal("Go unexpectedly accepted invalid arguments")
			}
			check := runner.Check(source)
			if check.Route == pljit.RouteJITNative || check.Detail == "" {
				t.Fatalf("JIT accepted Go-invalid arguments or omitted reason: %+v", check)
			}
		})
	}
	// Rust intentionally permits a dynamic cutset and validates its value at
	// runtime; pipeline-go's checker rejecting this form is not a JIT fallback
	// requirement.
	extension := runner.Check(`cutset=" "; trim(message,cutset)`)
	if extension.Route != pljit.RouteJITNative || extension.Capabilities.Backend != pljit.ExecutionBackendMachineCode {
		t.Fatalf("dynamic trim extension route=%+v", extension)
	}
}

func TestJITReplaceArgumentAdmission(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 8)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	for _, source := range []string{
		`replace(message,"a")`, `replace(message,2,"b")`,
		`replace(message,"[","b")`, `replace(message,"a",2)`,
		`replace(key=message,"a","b")`, `replace(message,pattern="a","b")`,
		`replace(message,"a",replacement="b")`,
	} {
		t.Run(source, func(t *testing.T) {
			if _, err := NewPlScriptSimple(point.Logging, "replace.p", source); err == nil {
				t.Fatal("Go unexpectedly accepted invalid arguments")
			}
			check := runner.Check(source)
			if check.Route == pljit.RouteJITNative || check.Detail == "" {
				t.Fatalf("JIT accepted Go-invalid arguments or omitted reason: %+v", check)
			}
		})
	}
}

// TestReplace, pipeline-go v1.4.3 (MIT, Guance, Inc.).
func TestJITReplaceUpstreamOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{"upstream/0", "json(_, `str`); replace(`str`, \"(1[0-9]{2})[0-9]{4}([0-9]{4})\", \"$1****$2\")", `{"str":"13789123014"}`, "str", "137****3014", false},
		{"upstream/1", "json(_, `str`); replace(`str`, \"([a-z]*) \\\\w*\", \"$1 ***\")", `{"str":"zhang san"}`, "str", "zhang ***", false},
		{"upstream/2", "json(_, `str`); replace(`str`, \"([1-9]{4})[0-9]{10}([0-9]{4})\", \"$1**********$2\")", `{"str":"362201200005302565"}`, "str", "3622**********2565", false},
		{"upstream/3", "json(_, `str`); replace(`str`, '([一-龥])[一-龥]([一-龥])', '$1＊$2')", `{"str":"小阿卡"}`, "str", "小＊卡", false},
		{"upstream/4", "json(_, `str`); replace(str1, '([一-龥])[一-龥]([一-龥])', '$1＊$2')", `{"str":"小阿卡"}`, "str", "小阿卡", false},
		{"capture/braces", `replace(message,"(a)","${1}x")`, "a", "message", "ax", false},
		{"capture/greedy-name", `replace(message,"(a)","$1x")`, "a", "message", "", false},
		{"capture/dollar", `replace(message,"a","$$")`, "a", "message", "$", false},
		{"capture/missing", `replace(message,"a","${missing}")`, "a", "message", "", false},
		{"boundary/empty-match", `replace(message,"a*","x")`, "ab", "message", "xbx", false},
		{"boundary/empty-pattern", `replace(message,"","x")`, "你好", "message", "x你x好x", false},
		{"boundary/non-string", `value=123; replace(value,"2","x")`, "test", "value", "1x3", false},
		{"boundary/error-prefix", `replace(message,"a","b"); result=1/divisor`, "a", "message", "b", true},
	})
}

func TestJITBase64EncodeUpstreamOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{"upstream/0", "json(_, `str`); b64enc(`str`)", `{"str":"13838130517"}`, "str", "MTM4MzgxMzA1MTc=", false},
		{"upstream/1", "json(_, `str`); b64enc(`str`)", `{"str":"hello, world"}`, "str", "aGVsbG8sIHdvcmxk", false},
		{"upstream/2", "json(_, `str`); b64enc(`str`)", `{"str":"你好"}`, "str", "5L2g5aW9", false},
		{"boundary/empty", `b64enc(message)`, "", "message", "", false},
		{"boundary/nul", `b64enc(message)`, "\x00", "message", "AA==", false},
		{"boundary/missing", `b64enc(missing)`, "test", "message", "test", false},
		{"boundary/non-string", `encoded=123; b64enc(encoded); add_key(result,encoded)`, "test", "result", int64(123), false},
		{"chain/roundtrip", `b64enc(message); b64dec(message)`, "你好\x00", "message", "你好\x00", false},
		{"chain/binary-roundtrip", `b64dec(message); b64enc(message)`, "/wA=", "message", "/wA=", false},
		{"chain/local-shadow", `encoded="a"; b64enc(encoded); add_key(observed,encoded)`, "test", "observed", "a", false},
		{"chain/error-prefix", `b64enc(message); result=1/divisor`, "a", "message", "YQ==", true},
	})
}

// TestB64dec, pipeline-go v1.4.3 (MIT, Guance, Inc.), and malformed input corpus.
func TestJITBase64DecodeUpstreamOracle(t *testing.T) {
	cases := []jsonOracleCase{
		{"upstream/0", "json(_, `str`); b64dec(`str`)", `{"str":"MTM4MzgxMzA1MTc="}`, "str", "13838130517", false},
		{"upstream/1", "json(_, `str`); b64dec(`str`)", `{"str":"aGVsbG8sIHdvcmxk"}`, "str", "hello, world", false},
		{"upstream/2", "json(_, `str`); b64dec(`str`)", `{"str":"5L2g5aW9"}`, "str", "你好", false},
		{"boundary/newlines", `b64dec(message)`, "Y\r\nQ==\n", "message", "a", false},
		{"boundary/trailing-bits", `b64dec(message)`, "YR==", "message", "a", false},
		{"boundary/binary", `b64dec(message)`, "/wA=", "message", string([]byte{255, 0}), false},
		{"boundary/empty", `b64dec(message)`, "", "message", "", false},
		{"boundary/missing", `b64dec(missing)`, "test", "message", "test", false},
		{"boundary/non-string", `encoded=123; b64dec(encoded); add_key(result,encoded)`, "test", "result", int64(123), false},
	}
	for i, input := range []string{"YQ", "YQ=", "YQ===", "Y Q==", "YQ==x", "YWJj!!!!", "_w=="} {
		cases = append(cases, jsonOracleCase{fmt.Sprintf("invalid/%d", i), `b64dec(message)`, input, "message", input, true})
	}
	runJSONOracle(t, cases)
}
