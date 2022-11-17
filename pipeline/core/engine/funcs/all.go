// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package funcs for pipeline
package funcs

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/runtime"
)

var l = logger.DefaultSLogger("funcs")

func InitLog() {
	l = logger.SLogger("funcs")
}

var FuncsMap = map[string]runtime.FuncCall{
	"cidr":                  CIDR,
	"grok":                  Grok,
	"add_key":               AddKey,
	"adjust_timezone":       AdjustTimezone,
	"json":                  JSON,
	"add_pattern":           AddPattern,
	"cast":                  Cast,
	"datetime":              DateTime,
	"default_time":          DefaultTime,
	"default_time_with_fmt": DefaultTimeWithFmt,
	"drop":                  Drop,
	"drop_key":              Dropkey,
	"drop_origin_data":      DropOriginData,
	"exit":                  Exit,
	"geoip":                 GeoIP,
	"get_key":               Getkey,
	"group_between":         Group,
	"group_in":              GroupIn,
	"lowercase":             Lowercase,
	"len":                   Len,
	"load_json":             LoadJSON,
	"nullif":                NullIf,
	"rename":                Rename,
	"set_tag":               SetTag,
	"set_measurement":       SetMeasurement,
	"strfmt":                Strfmt,
	"trim":                  Trim,
	"uppercase":             Uppercase,
	"use":                   Use,
	"url_decode":            URLDecode,
	"user_agent":            UserAgent,
	"parse_duration":        ParseDuration,
	"parse_date":            ParseDate,
	"cover":                 Cover,
	"query_refer_table":     QueryReferTable,
	"mquery_refer_table":    MQueryReferTableMulti,
	"replace":               Replace,
	"duration_precision":    DurationPrecision,
	"xml":                   XML,
	"match":                 Match,
	"sql_cover":             SQLCover,
	"decode":                Decode,
	"sample":                Sample,
	"url_parse":             URLParse,
	// disable
	"json_all": JSONAll,
}

var FuncsCheckMap = map[string]runtime.FuncCheck{
	"cidr":                  CIDRChecking,
	"grok":                  GrokChecking,
	"add_key":               AddkeyChecking,
	"adjust_timezone":       AdjustTimezoneChecking,
	"json":                  JSONChecking,
	"add_pattern":           AddPatternChecking,
	"cast":                  CastChecking,
	"datetime":              DateTimeChecking,
	"default_time":          DefaultTimeChecking,
	"default_time_with_fmt": DefaultTimeWithFmtChecking,
	"drop":                  DropChecking,
	"drop_key":              DropkeyChecking,
	"drop_origin_data":      DropOriginDataChecking,
	"exit":                  ExitChecking,
	"geoip":                 GeoIPChecking,
	"get_key":               GetkeyChecking,
	"group_between":         GroupChecking,
	"group_in":              GroupInChecking,
	"len":                   LenChecking,
	"load_json":             LoadJSONChecking,
	"lowercase":             LowercaseChecking,
	"nullif":                NullIfChecking,
	"rename":                RenameChecking,
	"set_measurement":       SetMeasurementChecking,
	"set_tag":               SetTagChecking,
	"strfmt":                StrfmtChecking,
	"trim":                  TrimChecking,
	"uppercase":             UppercaseChecking,
	"use":                   UseChecking,
	"url_decode":            URLDecodeChecking,
	"user_agent":            UserAgentChecking,
	"parse_duration":        ParseDurationChecking,
	"parse_date":            ParseDateChecking,
	"cover":                 CoverChecking,
	"query_refer_table":     QueryReferTableChecking,
	"mquery_refer_table":    MQueryReferTableChecking,
	"replace":               ReplaceChecking,
	"duration_precision":    DurationPrecisionChecking,
	"sql_cover":             SQLCoverChecking,
	"xml":                   XMLChecking,
	"match":                 MatchChecking,
	"decode":                DecodeChecking,
	"url_parse":             URLParseChecking,
	"sample":                SampleChecking,
	// disable
	"json_all": JSONAllChecking,
}
