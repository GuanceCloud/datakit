// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package funcs for pipeline
package funcs

import (
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
)

var l = logger.DefaultSLogger("pl-funcs")

func InitLog() {
	l = logger.SLogger("pl-funcs")
}

type Function struct {
	Name string
	Args []*Param

	// todo: check return type
	// Return []ast.DType

	Call  runtime.FuncCall
	Check runtime.FuncCheck

	Doc [2]*PLDoc // zh, en

	Deprecated bool
}

type Param struct {
	Name string
	Type []ast.DType

	DefaultVal func() (any, ast.DType)
	Optional   bool
	VariableP  bool
}

var FuncsMap = map[string]runtime.FuncCall{
	"agg_create":             AggCreate,
	"agg_metric":             AggAddMetric,
	"append":                 Append,
	"cidr":                   CIDR,
	"grok":                   Grok,
	"add_key":                AddKey,
	"delete":                 DeleteMapItem,
	"adjust_timezone":        AdjustTimezone,
	"json":                   JSON,
	"add_pattern":            AddPattern,
	"b64dec":                 B64dec,
	"b64enc":                 B64enc,
	"cast":                   Cast,
	"datetime":               DateTime,
	"default_time":           DefaultTime,
	"default_time_with_fmt":  DefaultTimeWithFmt,
	"drop":                   Drop,
	"drop_key":               Dropkey,
	"drop_origin_data":       DropOriginData,
	"exit":                   Exit,
	"geoip":                  GeoIP,
	"get_key":                Getkey,
	"group_between":          Group,
	"group_in":               GroupIn,
	"kv_split":               KVSplit,
	"lowercase":              Lowercase,
	"len":                    Len,
	"load_json":              LoadJSON,
	"nullif":                 NullIf,
	"rename":                 Rename,
	"set_tag":                SetTag,
	"set_measurement":        SetMeasurement,
	"strfmt":                 Strfmt,
	"trim":                   Trim,
	"timestamp":              Timestamp,
	"uppercase":              Uppercase,
	"use":                    Use,
	"url_decode":             URLDecode,
	"user_agent":             UserAgent,
	"parse_duration":         ParseDuration,
	"parse_date":             ParseDate,
	"cover":                  Cover,
	"query_refer_table":      QueryReferTable,
	"mquery_refer_table":     MQueryReferTableMulti,
	"replace":                Replace,
	"duration_precision":     DurationPrecision,
	"xml":                    XML,
	"match":                  Match,
	"sql_cover":              SQLCover,
	"decode":                 Decode,
	"sample":                 Sample,
	"url_parse":              URLParse,
	"value_type":             ValueType,
	"vaild_json":             ValidJSON,
	"valid_json":             ValidJSON,
	"conv_traceid_w3c_to_dd": ConvTraceIDW3C2DD,
	"create_point":           CreatePoint,
	"parse_int":              ParseInt,
	"format_int":             FormatInt,
	"pt_name":                PtName,
	"http_request":           HTTPRequest,
	"cache_get":              CacheGet,
	"cache_set":              CacheSet,
	"gjson":                  GJSON,
	"point_window":           PtWindow,
	"window_hit":             PtWindowHit,
	"pt_kvs_set":             FnPtKvsSet.Call,
	"pt_kvs_get":             FnPtKvsGet.Call,
	"pt_kvs_del":             FnPtKvsDel.Call,
	"pt_kvs_keys":            FnPtKvsKeys.Call,
	"hash":                   FnHash.Call,
	"slice_string":           FnSliceString.Call,

	"strlen": StrLen,

	"setopt": Setopt,

	// disable
	"json_all": JSONAll,
}

var FuncsCheckMap = map[string]runtime.FuncCheck{
	"agg_create":             AggCreateChecking,
	"agg_metric":             AggAddMetricChecking,
	"append":                 AppendChecking,
	"cidr":                   CIDRChecking,
	"grok":                   GrokChecking,
	"add_key":                AddkeyChecking,
	"delete":                 DeleteMapItemChecking,
	"adjust_timezone":        AdjustTimezoneChecking,
	"json":                   JSONChecking,
	"add_pattern":            AddPatternChecking,
	"b64dec":                 B64decChecking,
	"b64enc":                 B64encChecking,
	"cast":                   CastChecking,
	"datetime":               DateTimeChecking,
	"default_time":           DefaultTimeChecking,
	"default_time_with_fmt":  DefaultTimeWithFmtChecking,
	"drop":                   DropChecking,
	"drop_key":               DropkeyChecking,
	"drop_origin_data":       DropOriginDataChecking,
	"exit":                   ExitChecking,
	"geoip":                  GeoIPChecking,
	"get_key":                GetkeyChecking,
	"group_between":          GroupChecking,
	"group_in":               GroupInChecking,
	"kv_split":               KVSplitChecking,
	"len":                    LenChecking,
	"load_json":              LoadJSONChecking,
	"lowercase":              LowercaseChecking,
	"nullif":                 NullIfChecking,
	"rename":                 RenameChecking,
	"set_measurement":        SetMeasurementChecking,
	"set_tag":                SetTagChecking,
	"strfmt":                 StrfmtChecking,
	"trim":                   TrimChecking,
	"timestamp":              TimestampChecking,
	"uppercase":              UppercaseChecking,
	"use":                    UseChecking,
	"url_decode":             URLDecodeChecking,
	"user_agent":             UserAgentChecking,
	"parse_duration":         ParseDurationChecking,
	"parse_date":             ParseDateChecking,
	"cover":                  CoverChecking,
	"query_refer_table":      QueryReferTableChecking,
	"mquery_refer_table":     MQueryReferTableChecking,
	"replace":                ReplaceChecking,
	"duration_precision":     DurationPrecisionChecking,
	"sql_cover":              SQLCoverChecking,
	"xml":                    XMLChecking,
	"match":                  MatchChecking,
	"decode":                 DecodeChecking,
	"url_parse":              URLParseChecking,
	"sample":                 SampleChecking,
	"value_type":             ValueTypeChecking,
	"vaild_json":             ValidJSONChecking,
	"valid_json":             ValidJSONChecking,
	"conv_traceid_w3c_to_dd": ConvTraceIDW3C2DDChecking,
	"create_point":           CreatePointChecking,
	"parse_int":              ParseIntChecking,
	"format_int":             FormatIntChecking,
	"pt_name":                PtNameChecking,
	"http_request":           HTTPRequestChecking,
	"cache_get":              CacheGetChecking,
	"cache_set":              CacheSetChecking,
	"gjson":                  GJSONChecking,
	"point_window":           PtWindowChecking,
	"window_hit":             PtWindowHitChecking,

	"pt_kvs_set":   FnPtKvsSet.Check,
	"pt_kvs_get":   FnPtKvsGet.Check,
	"pt_kvs_del":   FnPtKvsDel.Check,
	"pt_kvs_keys":  FnPtKvsKeys.Check,
	"hash":         FnHash.Check,
	"slice_string": FnSliceString.Check,

	// disable
	"json_all": JSONAllChecking,

	"strlen": StrLenChecking,
	"setopt": SetoptChecking,
}
