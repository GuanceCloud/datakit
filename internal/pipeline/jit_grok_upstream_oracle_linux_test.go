// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

// Adapted from pipeline-go v1.4.3 ptinput/funcs/fn_grok_test.go TestGrok.
// Copyright 2021-present Guance, Inc. MIT License; see
// testdata/PIPELINE_GO_TEST_LICENSE.txt.
package pipeline

import "testing"

// TestGrokFastPathCompatibility, with original inputs and expectations.
func TestJITUpstreamGrokFastPathOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "inverse_space_class_captures_whole_line_without_trim", source: `add_pattern("ANY", "[\\s\\S]*")
grok(_, "%{ANY:item}", false)`, input: " not_space ", key: "item", want: " not_space "},
		{name: "inverse_space_class_captures_multiline", source: `add_pattern("ANY", "[\\s\\S]*")
grok(_, "%{ANY:item}", false)`, input: "line1\nline2", key: "item", want: "line1\nline2"},
		{name: "inverse_digit_class_preserves_punctuation_and_space", source: `add_pattern("NONDIGITS", "[\\D]*")
grok(_, "%{NONDIGITS:item}", false)`, input: "abc- ", key: "item", want: "abc- "},
		{name: "inverse_word_class_before_literal", source: `add_pattern("NONWORDS", "[\\W]+")
grok(_, "^%{NONWORDS:item}word$", false)`, input: " -\tword", key: "item", want: " -\t"},
		{name: "dot_all_style_class_between_literals", source: `add_pattern("ANY", "[\\s\\S]*")
grok(_, "^prefix:%{ANY:item}:suffix$")`, input: "prefix: a:b :suffix", key: "item", want: "a:b"},
		{name: "greedydata_keeps_text_before_following_literal", source: `grok(_, "^prefix=%{GREEDYDATA:item} suffix=%{WORD:tail}$")`, input: "prefix=alpha beta suffix=ok", key: "item", want: "alpha beta"},
		{name: "explicit_bracket_alias_one_is_false", source: `add_key(ok, grok(_, "%{WORD:item[1]}"))`, input: "first", key: "ok", want: false},
		{name: "explicit_bracket_alias_zero_is_false", source: `add_key(ok, grok(_, "%{WORD:item[0]}"))`, input: "first", key: "ok", want: false},
		{name: "explicit_bracket_alias_text_is_false", source: `add_key(ok, grok(_, "%{WORD:item[x]}"))`, input: "first", key: "ok", want: false},
		{name: "explicit_bracket_alias_mixed_is_false", source: `add_key(ok, grok(_, "%{WORD:item[1]} %{WORD:item}"))`, input: "first second", key: "ok", want: false},
		{name: "custom_pattern_bracket_alias_is_false", source: `add_pattern("CUSTOM", "%{WORD:item[1]}"); add_key(ok, grok(_, "%{CUSTOM}"))`, input: "first", key: "ok", want: false},
	})
}

func TestJITNestedGrokControlFlowOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "short_circuit", source: `add_key(capture,"keep"); result=false && grok(_,"%{WORD:capture}")`, input: "replace", key: "capture", want: "keep"},
		{name: "error_prefix", source: `result=grok(_,"%{WORD:capture}") && 1/divisor == 0`, input: "replace", key: "capture", want: "replace", fails: true},
		{name: "missing_source", source: `add_key(result,grok(absent,"%{GREEDYDATA:capture}"))`, key: "result", want: false},
		{name: "attribute_key", source: `add_key("obj.name","replace"); add_key(result,grok(obj.name,"%{WORD:capture}"))`, key: "capture", want: "replace"},
		{name: "pattern_versions", source: `add_pattern("custom","[a-z]+")
add_key(first,grok(_,"%{custom:first_capture}"))
add_pattern("custom","[0-9]+")
add_key(second,grok(_,"%{custom:second_capture}"))`, input: "abc123", key: "second_capture", want: "123"},
	})
}

func TestJITUpstreamGrokOracle(t *testing.T) {
	const parts = `add_pattern("_second", "(?:(?:[0-5]?[0-9]|60)(?:[:.,][0-9]+)?)")
add_pattern("_minute", "(?:[0-5][0-9])")
add_pattern("_hour", "(?:2[0123]|[01]?[0-9])")
`
	const timePattern = `add_pattern("time", "([^0-9]?)%{_hour:hour}:%{_minute:minute}(?::%{_second:second})([^0-9]?)")
`
	const multiline = `add_pattern("time", "%{NUMBER:time:float}")
grok(_, '''%{time}
%{WORD:word:string}
	%{WORD:code:int}
%{WORD:w1}''')`
	const multilineStr = `add_pattern("time", "%{NUMBER:time:float}")
grok(_, '''%{time}
%{WORD:word:str}
	%{WORD:code:int}
%{WORD:w1}''')`
	const trimPattern = `add_pattern("d", "[\\s\\S]*")
`
	runJSONOracle(t, []jsonOracleCase{
		{name: "00-normal_return_t", source: parts + timePattern + `add_key(grok_match_ok, grok(_, "%{time}"))`, input: "12:13:14.123", key: "grok_match_ok", want: true},
		{name: "01-normal_return_f", source: parts + timePattern + `add_key(grok_match_ok, grok(_, "%{time}"))`, input: "12 :13:14.123", key: "grok_match_ok", want: false},
		{name: "02-normal_return_sample_t", source: `add_key(grok_match_ok, grok(_, "12 :13:14.123"))`, input: "12 :13:14.123", key: "grok_match_ok", want: true},
		{name: "03-normal_return_sample_f", source: `add_key(grok_match_ok, grok(_, "12 :13:14.123"))`, input: "12:13:14.123", key: "grok_match_ok", want: false},
		{name: "04-second_fraction", source: parts + timePattern + `grok(_, "%{time}")`, input: "12:13:14.123", key: "second", want: "14.123"},
		{name: "05-minute", source: parts + timePattern + `grok(_, "%{time}")`, input: "12:13:14", key: "minute", want: "13"},
		{name: "06-hour", source: parts + timePattern + `grok(_, "%{time}")`, input: "12:13:14", key: "hour", want: "12"},
		{name: "07-second", source: parts + timePattern + `grok(_, "%{time}")`, input: "12:13:14", key: "second", want: "14"},
		{name: "08-invalid_int", source: multiline, input: "1.1\ns\n\t123cvf\naa222", key: "code", want: int64(0)},
		{name: "09-int", source: multiline, input: "1.1\ns\n\t123\naa222", key: "code", want: int64(123)},
		{name: "10-str_alias", source: multilineStr, input: "1.1\ns\n\t123\naa222", key: "code", want: int64(123)},
		{name: "11-capture_types", source: parts + `add_pattern("time", "([^0-9]?)%{_hour:hour:string}:%{_minute:minute:int}(?::%{_second:second:float})([^0-9]?)")
grok(_, "%{WORD:date} %{time}")`, input: "2021/1/11 2:13:14.123", key: "second", want: float64(14.123)},
		{name: "12-trim_default", source: trimPattern + `grok(_, "%{d:item}")`, input: " not_space ", key: "item", want: "not_space"},
		{name: "13-trim_enable", source: trimPattern + `grok(_, "%{d:item}", true)`, input: " not_space ", key: "item", want: "not_space"},
		{name: "14-trim_disable", source: trimPattern + `grok(_, "%{d:item}", false)`, input: " not_space ", key: "item", want: " not_space "},
	})
}
