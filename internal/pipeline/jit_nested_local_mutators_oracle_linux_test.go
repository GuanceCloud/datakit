// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

import "testing"

// Pipeline functions return nil while applying their Point mutation. These
// compositions compare the complete Point with pipeline-go, so argument
// evaluation, key-path preservation, mutation, return value and continuation
// are all part of the oracle.
func TestJITNestedLocalPointMutatorsOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{
			name:   "url_decode",
			source: `add_key(v,"a%20b"); add_key(returned,url_decode(v)==nil)`,
			key:    "v", want: "a b",
		},
		{
			name:   "replace",
			source: `add_key(v,"abc"); add_key(returned,replace(v,"a","x")==nil)`,
			key:    "v", want: "xbc",
		},
		{
			name:   "nullif",
			source: `add_key(v,""); add_key(returned,nullif(v,"")==nil)`,
			key:    "v", want: nil,
		},
		{
			name:   "parse_duration",
			source: `add_key(v,"1s"); add_key(returned,parse_duration(v)==nil)`,
			key:    "v", want: int64(1_000_000_000),
		},
		{
			name:   "cover",
			source: `add_key(v,"abcdef"); add_key(returned,cover(v,[1,3])==nil)`,
			key:    "v", want: "***def",
		},
		{
			name:   "duration_precision",
			source: `add_key(v,7); add_key(returned,duration_precision(v,"ms","ns")==nil)`,
			key:    "v", want: int64(7_000_000),
		},
		{
			name:   "gjson",
			source: `add_key(v,"{\"name\":\"rust\"}"); add_key(returned,gjson(v,"name","parsed")==nil)`,
			key:    "parsed", want: "rust",
		},
		{
			name:   "gjson_nested_named_reordered",
			source: `add_key(v,"{\"name\":\"rust\"}"); add_key(returned,gjson(key_name="parsed",input=v,json_path="name")==nil)`,
			key:    "parsed", want: "rust",
		},
		{
			name:   "sql_cover",
			source: `add_key(v,"select abc from def where x > 3"); add_key(returned,sql_cover(v)==nil)`,
			key:    "v", want: "select abc from def where x > ?",
		},
		{
			name:   "strfmt",
			source: `add_key(returned,strfmt(v,"%d:%s",7,"ok")==nil)`,
			key:    "v", want: "7:ok",
		},
		{
			name:   "datetime",
			source: `add_key(v,0); add_key(returned,datetime(v,"s","%Y-%m-%dT%H:%M:%S","UTC")==nil)`,
			key:    "v", want: "1970-01-01T00:00:00",
		},
		{
			name:   "default_time",
			source: `add_key(v,"1970-01-02 00:00:00"); add_key(returned,default_time(v,"UTC")==nil)`,
			key:    "v", want: int64(86_400_000_000_000),
		},
		{
			name:   "default_time_with_fmt_date_only",
			source: `add_key(v,"19700103"); add_key(returned,default_time_with_fmt(v,"20060102","UTC")==nil)`,
			key:    "v", want: int64(172_800_000_000_000),
		},
		{
			name:   "json",
			source: `add_key(v,"{\"a\":1}"); add_key(returned,json(v,a,out)==nil)`,
			key:    "out", want: float64(1),
		},
		{
			name:   "json_all",
			source: `add_key(v,"{\"name\":\"Tom\",\"age\":37}"); add_key(returned,json_all(v,["age"])==nil)`,
			key:    "age", want: float64(37),
		},
		{
			name:   "kv_split",
			source: `add_key(v,"a=one b=two c=three"); add_key(returned,kv_split(v,include_keys=["a","b","c"])==nil)`,
			key:    "b", want: "two",
		},
		{
			name:   "xml",
			source: `add_key(v,"<entry><fieldx>value</fieldx></entry>"); add_key(returned,xml(v,"/entry/fieldx/text()",out)==nil)`,
			key:    "out", want: "value",
		},
		{
			name:   "user_agent",
			source: `add_key(v,"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/36.0.1985.125 Safari/537.36"); add_key(returned,user_agent(v,"ua_")==nil)`,
			key:    "ua_browser", want: "Chrome",
		},
		{
			name:   "drop_origin_data",
			source: `add_key(returned,drop_origin_data()==nil)`,
			key:    "returned", want: true,
		},
		{
			name:   "setopt",
			source: `add_key(returned,setopt(status_mapping=false)==nil)`,
			key:    "returned", want: true,
		},
	})
}
