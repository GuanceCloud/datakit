// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	_ "embed"
)

type PLDoc struct {
	Language        string              `json:"language"`
	Doc             string              `json:"doc"`
	Prototype       string              `json:"prototype"`
	Description     string              `json:"description"`
	Deprecated      bool                `json:"deprecated"`
	RequiredVersion string              `json:"required_version"`
	FnCategory      map[string][]string `json:"fn_category"`
}

var PipelineFunctionDocs = map[string]*PLDoc{
	"add_key()":                &addKeyMarkdown,
	"add_pattern()":            &addPatternMarkdown,
	"adjust_timezone()":        &adjustTimezoneMarkdown,
	"agg_create()":             &aggCreateMarkdown,
	"agg_metric()":             &aggMetricMarkdown,
	"append()":                 &appendMarkdown,
	"b64dec()":                 &b64decMarkdown,
	"b64enc()":                 &b64encMarkdown,
	"cast()":                   &castMarkdown,
	"delete()":                 &deleteMarkdown,
	"cidr()":                   &cidrMarkdown,
	"cover()":                  &coverMarkdown,
	"datetime()":               &datetimeMarkdown,
	"decode()":                 &decodeMarkdown,
	"default_time()":           &defaultTimeMarkdown,
	"drop()":                   &dropMarkdown,
	"drop_key()":               &dropKeyMarkdown,
	"drop_origin_data()":       &dropOriginDataMarkdown,
	"duration_precision()":     &durationPrecisionMarkdown,
	"exit()":                   &exitMarkdown,
	"geoip()":                  &geoIPMarkdown,
	"get_key()":                &getKeyMarkdown,
	"grok()":                   &grokMarkdown,
	"group_between()":          &groupBetweenMarkdown,
	"group_in()":               &groupInMarkdown,
	"kv_split()":               &kvSplitMarkdown,
	"json()":                   &jsonMarkdown,
	"len()":                    &lenMarkdown,
	"load_json()":              &loadJSONMarkdown,
	"lowercase()":              &lowercaseMarkdown,
	"nullif()":                 &nullIfMarkdown,
	"parse_date()":             &parseDateMarkdown,
	"parse_duration()":         &parseDurationMarkdown,
	"query_refer_table()":      &queryReferTableMarkdown,
	"match()":                  &matchMarkdown,
	"mquery_refer_table()":     &mQueryReferTableMarkdown,
	"rename()":                 &renameMarkdown,
	"replace()":                &replaceMarkdown,
	"set_measurement()":        &setMeasurementMarkdown,
	"set_tag()":                &setTagMarkdown,
	"sql_cover()":              &sqlCoverMarkdown,
	"strfmt()":                 &strfmtMarkdown,
	"trim()":                   &trimMarkdown,
	"uppercase()":              &uppercaseMarkdown,
	"url_decode()":             &URLDecodeMarkdown,
	"use()":                    &useMarkdown,
	"user_agent()":             &userAgentMarkdown,
	"xml()":                    &xmlMarkdown,
	"sample()":                 &sampleMarkdown,
	"url_parse()":              &urlParseMarkdown,
	"timestamp()":              &timestampMarkdown,
	"value_type()":             &valueTypeMarkdown,
	"valid_json()":             &validJSONMarkdown,
	"conv_traceid_w3c_to_dd()": &convTraceID128MD,
	"create_point()":           &createPointMarkdown,
	"parse_int()":              &parseIntMarkdown,
	"format_int()":             &formatIntMarkdown,
	"pt_name()":                &ptNameMarkdown,
	"http_request()":           &HTTPRequestMarkdown,
	"cache_get()":              &CacheGetMarkdown,
	"cache_set()":              &CacheSetMarkdown,
	"gjson()":                  &gjsonMarkdown,
	"point_window()":           &pointWinodoeMarkdown,
	"window_hit()":             &winHitMarkdown,
	"pt_kvs_set()":             FnPtKvsSet.Doc[0],
	"pt_kvs_get()":             FnPtKvsGet.Doc[0],
	"pt_kvs_del()":             FnPtKvsDel.Doc[0],
	"pt_kvs_keys()":            FnPtKvsKeys.Doc[0],
	"hash()":                   FnHash.Doc[0],
	"slice_string()":           FnSliceString.Doc[0],
	"setopt()":                 &setoptMD,
	"strlen()":                 &strLen,
}

var PipelineFunctionDocsEN = map[string]*PLDoc{
	"add_key()":                &addKeyMarkdownEN,
	"add_pattern()":            &addPatternMarkdownEN,
	"adjust_timezone()":        &adjustTimezoneMarkdownEN,
	"agg_create()":             &aggCreateMarkdownEN,
	"agg_metric()":             &aggMetricMarkdownEN,
	"append()":                 &appendMarkdownEN,
	"b64dec()":                 &b64decMarkdownEN,
	"b64enc()":                 &b64encMarkdownEN,
	"cast()":                   &castMarkdownEN,
	"delete()":                 &deleteMarkdownEN,
	"cidr()":                   &cidrMarkdownEN,
	"cover()":                  &coverMarkdownEN,
	"datetime()":               &datetimeMarkdownEN,
	"decode()":                 &decodeMarkdownEN,
	"default_time()":           &defaultTimeMarkdownEN,
	"drop()":                   &dropMarkdownEN,
	"drop_key()":               &dropKeyMarkdownEN,
	"drop_origin_data()":       &dropOriginDataMarkdownEN,
	"duration_precision()":     &durationPrecisionMarkdownEN,
	"exit()":                   &exitMarkdownEN,
	"geoip()":                  &geoIPMarkdownEN,
	"get_key()":                &getKeyMarkdownEN,
	"grok()":                   &grokMarkdownEN,
	"group_between()":          &groupBetweenMarkdownEN,
	"group_in()":               &groupInMarkdownEN,
	"kv_split()":               &kvSplitMarkdownEN,
	"json()":                   &jsonMarkdownEN,
	"len()":                    &lenMarkdownEN,
	"load_json()":              &loadJSONMarkdownEN,
	"lowercase()":              &lowercaseMarkdownEN,
	"nullif()":                 &nullIfMarkdownEN,
	"parse_date()":             &parseDateMarkdownEN,
	"parse_duration()":         &parseDurationMarkdownEN,
	"query_refer_table()":      &queryReferTableMarkdownEN,
	"match()":                  &matchMarkdownEN,
	"mquery_refer_table()":     &mQueryReferTableMarkdownEN,
	"rename()":                 &renameMarkdownEN,
	"replace()":                &replaceMarkdownEN,
	"set_measurement()":        &setMeasurementMarkdownEN,
	"set_tag()":                &setTagMarkdownEN,
	"sql_cover()":              &sqlCoverMarkdownEN,
	"strfmt()":                 &strfmtMarkdownEN,
	"trim()":                   &trimMarkdownEN,
	"uppercase()":              &uppercaseMarkdownEN,
	"url_decode()":             &URLDecodeMarkdownEN,
	"use()":                    &useMarkdownEN,
	"user_agent()":             &userAgentMarkdownEN,
	"xml()":                    &xmlMarkdownEN,
	"sample()":                 &sampleMarkdownEN,
	"url_parse()":              &urlParseMarkdownEN,
	"timestamp()":              &timestampMarkdownEN,
	"value_type()":             &valueTypeMarkdownEN,
	"valid_json()":             &validJSONMarkdownEN,
	"conv_traceid_w3c_to_dd()": &convTraceID128MDEN,
	"create_point()":           &createPointMarkdownEN,
	"parse_int()":              &parseIntMarkdownEN,
	"format_int()":             &formatIntMarkdownEN,
	"pt_name()":                &ptNameMarkdownEN,
	"http_request()":           &HTTPRequestMarkdownEN,
	"cache_get()":              &CacheGetMarkdownEN,
	"cache_set()":              &CacheSetMarkdownEN,
	"gjson()":                  &gjsonMarkdownEN,
	"point_window()":           &pointWinodoeMarkdownEN,
	"window_hit()":             &winHitMarkdownEN,
	"pt_kvs_set()":             FnPtKvsSet.Doc[1],
	"pt_kvs_get()":             FnPtKvsGet.Doc[1],
	"pt_kvs_del()":             FnPtKvsDel.Doc[1],
	"pt_kvs_keys()":            FnPtKvsKeys.Doc[1],
	"hash()":                   FnHash.Doc[1],
	"slice_string()":           FnSliceString.Doc[1],
	"setopt()":                 &setoptMDEN,
	"strlen()":                 &strlenMDEN,
}

// embed docs.
var (

	//go:embed md/agg_create.md
	docAggCreate string

	//go:embed md/agg_metric.md
	docAggMetric string

	//go:embed md/add_pattern.md
	docAddPattern string

	//go:embed md/append.md
	docAppend string

	//go:embed md/b64dec.md
	docB64dec string

	//go:embed md/b64enc.md
	docB64enc string

	//go:embed md/cidr.md
	docCIDR string

	//go:embed md/grok.md
	docGrok string

	//go:embed md/json.md
	docJSON string

	//go:embed md/query_refer_table.md
	docQueryReferTable string

	//go:embed md/mquery_refer_table.md
	docMQueryReferTable string

	//go:embed md/match.md
	docMatch string

	//go:embed md/rename.md
	docRename string

	//go:embed md/url_decode.md
	docURLDecode string

	//go:embed md/geoip.md
	docGeoIP string

	//go:embed md/datetime.md
	docDatetime string

	//go:embed md/cast.md
	docCast string

	//go:embed md/delete.md
	docDelete string

	//go:embed md/get_key.md
	docGetKey string

	//go:embed md/group_between.md
	docGroupBetreen string

	//go:embed md/group_in.md
	docGroupIn string

	//go:embed md/uppercase.md
	docUppercase string

	//go:embed md/len.md
	docLen string

	//go:embed md/load_json.md
	docLoadJSON string

	//go:embed md/lowercase.md
	docLowercase string

	//go:embed md/nullif.md
	docNullif string

	//go:embed md/strfmt.md
	docStrfmt string

	//go:embed md/drop_origin_data.md
	docDropOriginData string

	//go:embed md/add_key.md
	docAddKey string

	//go:embed md/default_time.md
	docDefaultTime string

	//go:embed md/drop_key.md
	docDropKey string

	//go:embed md/trim.md
	docTrim string

	//go:embed md/user_agent.md
	docUserAgent string

	//go:embed md/parse_duration.md
	docParseDuration string

	//go:embed md/parse_date.md
	docParseDate string

	//go:embed md/cover.md
	docCover string

	//go:embed md/replace.md
	docReplace string

	//go:embed md/set_measurement.md
	docSetMeasurement string

	//go:embed md/set_tag.md
	docSetTag string

	//go:embed md/sample.md
	docSample string

	//go:embed md/drop.md
	docDrop string

	//go:embed md/exit.md
	docExit string

	//go:embed md/duration_precision.md
	docDurationPresicion string

	//go:embed md/decode.md
	docDecode string

	//go:embed md/sql_cover.md
	docSQLCover string

	//go:embed md/adjust_timezone.md
	docAdjustTimezone string

	//go:embed md/xml.md
	docXML string

	//go:embed md/use.md
	docUse string

	//go:embed md/url_parse.md
	docURLParse string

	//go:embed md/timestamp.md
	docTimestamp string

	//go:embed md/kv_split.md
	docKVSplit string

	//go:embed md/value_type.md
	docValueType string

	//go:embed md/valid_json.md
	docValidJSON string

	//go:embed md/conv_traceid_w3c_to_dd.md
	docConvTraceID string

	//go:embed md/create_point.md
	docCreatePoint string

	//go:embed md/parse_int.md
	docParseInt string

	//go:embed md/format_int.md
	docFormatInt string

	//go:embed md/pt_name.md
	docPtName string

	//go:embed md/http_request.md
	docHTTPRequest string

	//go:embed md/cache_get.md
	docCacheGet string

	//go:embed md/cache_set.md
	docCacheSet string

	//go:embed md/gjson.md
	docGJSON string

	//go:embed md/point_window.md
	docPointWindow string

	//go:embed md/window_hit.md
	docWindowHit string

	//go:embed md/setopt.md
	docSetopt string

	//go:embed md/strlen.md
	docStrlen string
)

const (
	langTagEnUS = "en-US"
	langTagZhCN = "zh-CN"
)

const (
	cEncodeDecode    = "编解码"
	cPointOp         = "Point 操作"
	cRegExp          = "RegExp"
	cGrok            = "Grok"
	cJSON            = "JSON"
	cXML             = "XML"
	cTimeOp          = "时间操作"
	cTypeCast        = "类型转换"
	cNetwork         = "网络"
	cStringOp        = "字符串操作"
	cDesensitization = "脱敏"
	cSample          = "采样"
	cAgg             = "聚合"
	cOther           = "其他"
)

var (
	URLDecodeMarkdown = PLDoc{
		Doc: docURLDecode, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cEncodeDecode, cNetwork},
		},
	}
	addKeyMarkdown = PLDoc{
		Doc: docAddKey, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cPointOp},
		},
	}
	aggCreateMarkdown = PLDoc{
		Doc: docAggCreate, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cAgg},
		},
	}
	aggMetricMarkdown = PLDoc{
		Doc: docAggMetric, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cAgg},
		},
	}

	addPatternMarkdown = PLDoc{
		Doc: docAddPattern, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cGrok},
		},
	}
	adjustTimezoneMarkdown = PLDoc{
		Doc: docAdjustTimezone, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cTimeOp},
		},
	}
	appendMarkdown = PLDoc{
		Doc: docAppend, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cOther},
		},
	}
	b64decMarkdown = PLDoc{
		Doc: docB64dec, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cEncodeDecode},
		},
	}
	b64encMarkdown = PLDoc{
		Doc: docB64enc, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cEncodeDecode},
		},
	}
	castMarkdown = PLDoc{
		Doc: docCast, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cTypeCast},
		},
	}
	deleteMarkdown = PLDoc{
		Doc: docDelete, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cJSON, cOther},
		},
	}
	cidrMarkdown = PLDoc{
		Doc: docCIDR, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cNetwork},
		},
	}
	coverMarkdown = PLDoc{
		Doc: docCover, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cStringOp, cDesensitization},
		},
	}
	datetimeMarkdown = PLDoc{
		Doc: docDatetime, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cTimeOp},
		},
	}
	decodeMarkdown = PLDoc{
		Doc: docDecode, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cEncodeDecode},
		},
	}
	defaultTimeMarkdown = PLDoc{
		Doc: docDefaultTime, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cTimeOp, cPointOp},
		},
	}
	getKeyMarkdown = PLDoc{
		Doc: docGetKey, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cPointOp},
		},
	}
	dropKeyMarkdown = PLDoc{
		Doc: docDropKey, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cPointOp},
		},
	}
	dropMarkdown = PLDoc{
		Doc: docDrop, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cPointOp},
		},
	}
	dropOriginDataMarkdown = PLDoc{
		Doc: docDropOriginData, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cPointOp},
		},
	}
	durationPrecisionMarkdown = PLDoc{
		Doc: docDurationPresicion, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cTimeOp},
		},
	}
	exitMarkdown = PLDoc{
		Doc: docExit, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cOther},
		},
	}
	geoIPMarkdown = PLDoc{
		Doc: docGeoIP, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cNetwork},
		},
	}
	grokMarkdown = PLDoc{
		Doc: docGrok, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cGrok, cRegExp},
		},
	}
	groupBetweenMarkdown = PLDoc{
		Doc: docGroupBetreen, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cOther},
		},
	}
	groupInMarkdown = PLDoc{
		Doc: docGroupIn, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cOther},
		},
	}
	jsonMarkdown = PLDoc{
		Doc: docJSON, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cJSON},
		},
	}
	lenMarkdown = PLDoc{
		Doc: docLen, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cOther},
		},
	}
	loadJSONMarkdown = PLDoc{
		Doc: docLoadJSON, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cJSON},
		},
	}
	lowercaseMarkdown = PLDoc{
		Doc: docLowercase, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cStringOp},
		},
	}
	nullIfMarkdown = PLDoc{
		Doc: docNullif, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cOther},
		},
	}
	parseDateMarkdown = PLDoc{
		Doc: docParseDate, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cTimeOp},
		},
	}
	parseDurationMarkdown = PLDoc{
		Doc: docParseDuration, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cTimeOp},
		},
	}
	queryReferTableMarkdown = PLDoc{
		Doc: docQueryReferTable, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cOther},
		},
	}
	matchMarkdown = PLDoc{
		Doc: docMatch, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cRegExp},
		},
	}
	mQueryReferTableMarkdown = PLDoc{
		Doc: docMQueryReferTable, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cOther},
		},
	}
	renameMarkdown = PLDoc{
		Doc: docRename, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cPointOp},
		},
	}
	replaceMarkdown = PLDoc{
		Doc: docReplace, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cRegExp},
		},
	}
	sampleMarkdown = PLDoc{
		Doc: docSample, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cSample},
		},
	}

	setMeasurementMarkdown = PLDoc{
		Doc: docSetMeasurement, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cPointOp},
		},
	}
	setTagMarkdown = PLDoc{
		Doc: docSetTag, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cPointOp},
		},
	}
	sqlCoverMarkdown = PLDoc{
		Doc: docSQLCover, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cDesensitization},
		},
	}
	strfmtMarkdown = PLDoc{
		Doc: docStrfmt, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cStringOp},
		},
	}
	trimMarkdown = PLDoc{
		Doc: docTrim, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cStringOp},
		},
	}
	uppercaseMarkdown = PLDoc{
		Doc: docUppercase, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cStringOp},
		},
	}
	userAgentMarkdown = PLDoc{
		Doc: docUserAgent, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cOther},
		},
	}
	useMarkdown = PLDoc{
		Doc: docUse, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cOther},
		},
	}
	xmlMarkdown = PLDoc{
		Doc: docXML, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cXML},
		},
	}
	urlParseMarkdown = PLDoc{
		Doc: docURLParse, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cNetwork, cEncodeDecode},
		},
	}
	timestampMarkdown = PLDoc{
		Doc: docTimestamp, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cTimeOp},
		},
	}

	kvSplitMarkdown = PLDoc{
		Doc: docKVSplit, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cRegExp},
		},
	}

	valueTypeMarkdown = PLDoc{
		Doc: docValueType, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cJSON},
		},
	}

	validJSONMarkdown = PLDoc{
		Doc: docValidJSON, Deprecated: false,
		FnCategory: map[string][]string{
			langTagZhCN: {cJSON},
		},
	}

	convTraceID128MD = PLDoc{
		Doc: docConvTraceID,
		FnCategory: map[string][]string{
			langTagZhCN: {cStringOp},
		},
	}

	createPointMarkdown = PLDoc{
		Doc: docCreatePoint,
		FnCategory: map[string][]string{
			langTagZhCN: {cOther},
		},
	}
	parseIntMarkdown = PLDoc{
		Doc: docParseInt,
		FnCategory: map[string][]string{
			langTagZhCN: {cStringOp},
		},
	}

	formatIntMarkdown = PLDoc{
		Doc: docFormatInt,
		FnCategory: map[string][]string{
			langTagZhCN: {cStringOp},
		},
	}

	ptNameMarkdown = PLDoc{
		Doc: docPtName,
		FnCategory: map[string][]string{
			langTagZhCN: {cPointOp},
		},
	}

	HTTPRequestMarkdown = PLDoc{
		Doc: docHTTPRequest,
		FnCategory: map[string][]string{
			langTagZhCN: {cOther},
		},
	}

	CacheGetMarkdown = PLDoc{
		Doc: docCacheGet,
		FnCategory: map[string][]string{
			langTagZhCN: {cOther},
		},
	}

	CacheSetMarkdown = PLDoc{
		Doc: docCacheSet,
		FnCategory: map[string][]string{
			langTagZhCN: {cOther},
		},
	}

	gjsonMarkdown = PLDoc{
		Doc: docGJSON,
		FnCategory: map[string][]string{
			langTagZhCN: {cOther},
		},
	}

	pointWinodoeMarkdown = PLDoc{
		Doc: docPointWindow,
		FnCategory: map[string][]string{
			langTagZhCN: {cOther},
		},
	}

	winHitMarkdown = PLDoc{
		Doc: docWindowHit,
		FnCategory: map[string][]string{
			langTagZhCN: {cOther},
		},
	}

	setoptMD = PLDoc{
		Doc: docSetopt,
		FnCategory: map[string][]string{
			langTagZhCN: {cOther},
		},
	}

	strLen = PLDoc{
		Doc: docStrlen,
		FnCategory: map[string][]string{
			langTagZhCN: {cStringOp},
		},
	}
)
