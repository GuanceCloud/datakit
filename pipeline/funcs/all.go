// Package funcs implement functions for datakit's pipeline.
package funcs

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

var l = logger.DefaultSLogger("funcs")

func InitLog() {
	l = logger.SLogger("funcs")
}

var FuncsMap = map[string]parser.FuncCallback{
	"grok":                  Grok,
	"add_key":               Addkey,
	"adjust_timezone":       AdjustTimezone,
	"json":                  JSON,
	"add_pattern":           AddPattern,
	"cast":                  Cast,
	"datetime":              DateTime,
	"default_time":          DefaultTime,
	"default_time_with_fmt": DefaultTimeWithFmt,
	"drop_key":              Dropkey,
	"drop_origin_data":      DropOriginData,
	"geoip":                 GeoIP,
	"group_between":         Group,
	"group_in":              GroupIn,
	"lowercase":             Lowercase,
	"nullif":                NullIf,
	"rename":                Rename,
	"strfmt":                Strfmt,
	"uppercase":             Uppercase,
	"url_decode":            URLDecode,
	"user_agent":            UserAgent,
	"parse_duration":        ParseDuration,
	"parse_date":            ParseDate,
	"cover":                 Dz,
	"replace":               Replace,
	// disable
	"json_all": JSONAll,
	"expr":     Expr,
}

var FuncsCheckMap = map[string]parser.FuncCallbackCheck{
	"grok":                  GrokChecking,
	"add_key":               AddkeyChecking,
	"adjust_timezone":       AdjustTimezoneChecking,
	"json":                  JSONChecking,
	"add_pattern":           AddPatternChecking,
	"cast":                  CastChecking,
	"datetime":              DateTimeChecking,
	"default_time":          DefaultTimeChecking,
	"default_time_with_fmt": DefaultTimeWithFmtChecking,
	"drop_key":              DropkeyChecking,
	"drop_origin_data":      DropOriginDataChecking,
	"geoip":                 GeoIPChecking,
	"group_between":         GroupChecking,
	"group_in":              GroupInChecking,
	"lowercase":             LowercaseChecking,
	"nullif":                NullIfChecking,
	"rename":                RenameChecking,
	"strfmt":                StrfmtChecking,
	"uppercase":             UppercaseChecking,
	"url_decode":            URLDecodeChecking,
	"user_agent":            UserAgentChecking,
	"parse_duration":        ParseDurationChecking,
	"parse_date":            ParseDateChecking,
	"cover":                 DzChecking,
	"replace":               ReplaceChecking,
	// disable
	"json_all": JSONAllChecking,
	"expr":     ExprChecking,
}
